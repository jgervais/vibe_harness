# Data Model: AST-Powered Checks

**Feature**: 002-ast-powered-checks | **Date**: 2026-04-19

## Primary Entities

### AST Parser

**Purpose**: Centralized parser that manages language grammars and parses source files into tree-sitter ASTs.

| Field | Type | Description |
|-------|------|-------------|
| `parsers` | `map[string]*tree_sitter.Parser` | Lazy-initialized parser pool, keyed by language name |
| `languages` | `map[string]*tree_sitter.Language` | Loaded language objects, keyed by language name |

**Operations**:
- `NewParser() *Parser` — create parser with empty pools
- `ParseFile(language string, content []byte) (*ParseResult, error)` — parse source content; returns nil result if language unsupported (not an error)
- `IsLanguageSupported(language string) bool` — check if grammar exists for language

**Validation rules**:
- Unsupported language → return nil result, no error (check skips AST processing)
- Parse error (syntax error) → return ParseResult with HasError flag; checks decide whether to proceed
- Empty content → return nil tree; non-nil error

### ParseResult

**Purpose**: Wraps a parsed tree-sitter tree with its source bytes and metadata.

| Field | Type | Description |
|-------|------|-------------|
| `Tree` | `*tree_sitter.Tree` | The parsed AST tree (must call Close()) |
| `Source` | `[]byte` | Original source bytes (needed for node text extraction and queries) |
| `Language` | `string` | Language name that was used for parsing |
| `HasError` | `bool` | Whether the tree contains error nodes (syntax errors in source) |

**Validation rules**:
- `Tree` must not be nil when HasError is false
- Caller must call `Close()` to free tree memory
- `Source` is retained for query execution (cursor.Matches requires source bytes)

### LanguageRegistry

**Purpose**: Maps file extensions to language names and language names to grammar initializers. Populated from config defaults + user config.

| Field | Type | Description |
|-------|------|-------------|
| `extensionToLanguage` | `map[string]string` | e.g., `".py" → "python"`, `".go" → "go"` |
| `availableGrammars` | `map[string]func() unsafe.Pointer` | Language name → grammar Language() function |
| `loadedGrammars` | `map[string]*tree_sitter.Language` | Cache of already-loaded grammars |

**Relationships**: 
- Uses `config.Config.Languages` to override/extend `extensionToLanguage`
- Referenced by `ASTParser` when looking up which grammar to use

### QuerySet

**Purpose**: Maps language names to tree-sitter S-expression query strings for a specific check.

| Field | Type | Description |
|-------|------|-------------|
| `queries` | `map[string]string` | Language → S-expression query pattern |

**Operations**:
- `NewQuerySet(queries map[string]string) *QuerySet` — create with language→query mapping
- `GetQuery(language string) (string, bool)` — get query string for a language; returns false if language not in set
- `Compile(language string, lang *tree_sitter.Language) (*tree_sitter.Query, error)` — compile query for language; result cached

**Validation rules**:
- Query strings must be valid S-expressions for the target grammar
- Unsupported languages simply have no entry (check returns no violations for that language)

### ObservabilityHintSet

**Purpose**: Language-aware collection of logging and metrics call name patterns, resolved from config defaults and user overrides.

| Field | Type | Description |
|-------|------|-------------|
| `LoggingCalls` | `map[string][]string` | Language → list of logging call name patterns |
| `MetricsCalls` | `map[string][]string` | Language → list of metrics call name patterns |
| `IOPatterns` | `map[string][]string` | Language → list of I/O call name patterns |

**Relationships**:
- Built from `config.Config.Observability` (user-provided) merged with defaults per language
- Used by VH-G003 to recognize logging calls and VH-G009 to recognize I/O patterns
- IOPatterns is built-in (not user-configurable, per constitution principle V)

**Validation rules**:
- User-provided hints are additive (merged with defaults, not replacing)
- Empty user config → defaults only
- Patterns are matched as prefixes/suffixes of AST node text (function call names)

## Existing Entities (Modified)

### rules.Check (Extended)

| Field | Type | Existing? | Change |
|-------|------|-----------|--------|
| `ID` | `string` | Yes | No change |
| `Name` | `string` | Yes | No change |
| `Description` | `string` | Yes | No change |
| `Severity` | `string` | Yes | No change |
| `RequiresAST` | `bool` | Yes | No change (set to `true` for new checks) |
| `Threshold` | `string` | Yes | No change (human-readable description of fixed threshold) |

6 new entries added to `rules.Checks()` return value:

| ID | Name | Severity | RequiresAST | Threshold |
|----|------|----------|-------------|-----------|
| VH-G002 | Function Length | warning | true | 50 statements |
| VH-G003 | Missing Logging in I/O | warning | true | N/A (pattern-based) |
| VH-G004 | Swallowed Errors | error | true | N/A (pattern-based) |
| VH-G009 | Missing Error Handling on I/O | error | true | N/A (pattern-based) |
| VH-G010 | Broad Exception Catching | warning | true | N/A (pattern-based) |
| VH-G012 | God Module | warning | true | 20 public exports |

### Scanner (Modified)

The scanner's `Scan()` function will be updated to:
1. Create an `ASTParser` instance
2. Before iterating over files, determine which languages are present
3. For each file, after reading content and determining language, call `ASTParser.ParseFile()` if the language is supported
4. Store `ParseResult` in a `map[string]*ParseResult` keyed by file path
5. For each check, test if it implements `ASTCheck` interface
6. If `ASTCheck`, call `CheckFileAST()` with the parse result
7. If only `Check`, call `CheckFile()` as before (non-AST path)
8. Parse results are cleaned up after all checks run

### config.Config (Extended)

| Field | Type | Existing? | Change |
|-------|------|-----------|--------|
| `Observability.LoggingCalls` | `[]string` | Yes | No change |
| `Observability.MetricsCalls` | `[]string` | Yes | No change |
| `Languages` | `map[string]string` | Yes | No change |

The `ObservabilityConfig` struct is extended internally with per-language defaults at load time (not in the TOML struct, but computed during `LoadConfig`/`ApplyDefaults`). The check implementations receive the full `Config` with merged defaults.

## New Interface: ASTCheck

```go
type ASTCheck interface {
    Check
    CheckFileAST(path string, content []byte, language string, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation
}
```

Checks that implement `ASTCheck` receive the pre-parsed tree. Checks that don't implement it fall back to `CheckFile` (text-based analysis).

The scanner detects which interface a check implements at runtime via type assertion and calls the appropriate method.

## State Transitions

### Parse Lifecycle

```
File discovered → Language determined from extension
                  ↓
          Is language supported?
          ↓              ↓
         Yes            No
          ↓              ↓
   Parse with grammar   Skip AST checks
          ↓              ↓
   Has parse errors?    Run non-AST checks only
   ↓           ↓
  Yes         No
   ↓           ↓
 Skip AST    Run AST + non-AST
 checks or    checks
 flag errors
```

### Check Lifecycle

```
For each file:
  1. Read content, determine language
  2. Run non-AST checks (CheckFile)
  3. If language is supported:
     a. Parse file → ParseResult
     b. For each AST check:
        - If implements ASTCheck → CheckFileAST(path, content, language, cfg, parseResult)
        - If only Check → skip (AST checks require tree)
  4. Cleanup parse result
```

### Violation Flow

AST checks produce `rules.Violation` structs identical to non-AST checks:
- `RuleID`: e.g., `"VH-G002"`
- `File`: source file path
- `Line`: start line of the violating construct (from AST node)
- `Column`: start column (from AST node, or 0 if not applicable)
- `EndLine`: end line (for multi-line constructs like function bodies)
- `Message`: human-readable description (e.g., `"Function 'processData' has 73 statements (threshold: 50)"`)
- `Severity`: `"error"`, `"warning"`, or `"note"` (per rule definition)

Violations are sorted by file then line (existing behavior) and passed to output formatters unchanged.