# AST Parser Contract

**Feature**: 002-ast-powered-checks | **Date**: 2026-04-19

This document defines the contract for the `internal/ast` package's public API. Other packages (scanner, checks) depend on this interface.

## Package: `internal/ast`

### Type: `Parser`

```go
type Parser struct { ... }

func NewParser() *Parser
func (p *Parser) ParseFile(language string, content []byte) (*ParseResult, error)
func (p *Parser) IsLanguageSupported(language string) bool
func (p *Parser) Close()
```

**Contract**:
- `NewParser()` returns a ready-to-use parser with empty grammar pools. Grammars are loaded lazily on first use for a given language.
- `ParseFile(language, content)`:
  - If `language` is not in the supported set, returns `(nil, nil)` — no result, no error. Callers treat this as "skip AST checks for this file."
  - If `content` is empty or nil, returns `(nil, error)` with a descriptive error.
  - If the source has syntax errors, returns `(*ParseResult, nil)` where `ParseResult.HasError == true`. The tree is still available (with error nodes) and checks may optionally inspect it.
  - On success, returns `(*ParseResult, nil)`.
  - The caller MUST call `ParseResult.Close()` when done to free tree memory.
  - Thread safety: `Parser` is NOT thread-safe. Each goroutine needs its own parser, or callers must synchronize access.
- `IsLanguageSupported(language)` returns true if a grammar exists for the language name.
- `Close()` frees all cached parsers and loaded grammars. Must be called when the parser is no longer needed.

### Type: `ParseResult`

```go
type ParseResult struct { ... }

func (r *ParseResult) Tree() *tree_sitter.Tree
func (r *ParseResult) Source() []byte
func (r *ParseResult) Language() string
func (r *ParseResult) HasError() bool
func (r *ParseResult) Close()
```

**Contract**:
- `Tree()` returns the parsed tree-sitter tree. Nil only if parsing was skipped (unsupported language).
- `Source()` returns the original source bytes. Needed for extracting node text and executing queries.
- `Language()` returns the language name that was used for parsing (e.g., "python", "go").
- `HasError()` returns true if the parse tree contains error nodes (syntax errors in source).
- `Close()` frees the tree memory. MUST be called exactly once. After Close(), Tree() returns nil.

### Type: `QuerySet`

```go
type QuerySet struct { ... }

func NewQuerySet(queries map[string]string) *QuerySet
func (qs *QuerySet) GetQuery(language string) (string, bool)
func (qs *QuerySet) Compile(language string, lang *tree_sitter.Language) (*tree_sitter.Query, error)
func (qs *QuerySet) Close()
```

**Contract**:
- `NewQuerySet(queries)` creates a query set from language→S-expression mappings.
- `GetQuery(language)` returns the raw query string for a language, or `("", false)` if no query exists.
- `Compile(language, lang)` compiles and caches the query for a given language. Returns the compiled `*tree_sitter.Query`. Cached queries are reused on subsequent calls for the same language.
- `Close()` frees all cached compiled queries. Must be called when the QuerySet is no longer needed.

---

## Interface: `ASTCheck`

**Package**: `internal/checks/generic`

```go
type ASTCheck interface {
    Check
    CheckFileAST(path string, content []byte, language string, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation
}
```

**Contract**:
- If a check implements `ASTCheck`, the scanner calls `CheckFileAST()` instead of `CheckFile()` when a `ParseResult` is available for the file.
- If `parseResult` is nil (unsupported language), the check should return an empty violation slice (no violations possible without AST).
- If `parseResult.HasError()` is true, the check may still attempt to analyze the tree (partial results possible) or skip entirely.
- `CheckFile()` for AST checks should return `nil` (empty violations) — they require the AST and operate only via `CheckFileAST()`.
- The check must not modify or close the `parseResult`. Ownership remains with the scanner.

---

## Scanner Integration Contract

**Package**: `internal/scanner`

The scanner's `Scan()` function is updated to:

1. Create an `ast.Parser` at the start of scanning and defer `parser.Close()`.
2. After reading each file's content and determining its language, check if `parser.IsLanguageSupported(language)`.
3. If supported, call `parser.ParseFile(language, content)` to get a `*ParseResult`.
4. For each check in the pipeline:
   - Type-assert to `ASTCheck` using: `astCheck, ok := check.(generic.ASTCheck)`
   - If `ok` and `parseResult != nil`: call `astCheck.CheckFileAST(path, content, language, cfg, parseResult)`
   - Otherwise: call `check.CheckFile(path, content, language, cfg)` as before
5. After all checks for a file, if `parseResult != nil`: call `parseResult.Close()`
6. After all files processed, `parser.Close()` is called.

**Contract**:
- Parse results are scoped per-file (created before checks, destroyed after all checks for that file).
- If `ParseFile()` returns an error (not nil), the scanner logs a warning and skips AST checks for that file but still runs non-AST checks.
- Memory: Only one parse tree is held in memory at a time (per-file scope).

---

## Rule Metadata Contract

**Package**: `internal/rules`

The `Checks()` function is updated to return 12 entries (6 existing + 6 new):

| ID | Name | Severity | RequiresAST | Threshold |
|----|------|----------|-------------|-----------|
| VH-G001 | File Length | warning | false | 300 lines |
| VH-G002 | Function Length | warning | true | 50 statements |
| VH-G003 | Missing Logging in I/O | warning | true | N/A |
| VH-G004 | Swallowed Errors | error | true | N/A |
| VH-G005 | Hardcoded Secrets | error | false | N/A |
| VH-G006 | Magic Values | warning | false | N/A |
| VH-G007 | Copy-Paste Duplication | warning | false | N/A |
| VH-G008 | Comment-to-Code Ratio | note | false | 25% |
| VH-G009 | Missing Error Handling on I/O | error | true | N/A |
| VH-G010 | Broad Exception Catching | warning | true | N/A |
| VH-G011 | Disabled Security Features | error | false | N/A |
| VH-G012 | God Module | warning | true | 20 exports |

**Contract**:
- `RequiresAST` is metadata only — the scanner does not use it to decide whether to parse. All supported files are parsed; AST checks simply return empty results for unsupported languages.
- Threshold values are human-readable descriptions. They are NOT configurable (constitution principle I).

---

## Config Extension Contract

**Package**: `internal/config`

The existing `Config` struct is unchanged in its TOML representation. Internal defaults are expanded:

```go
type perLanguageDefaults struct {
    LoggingCalls []string
    MetricsCalls []string
}

var defaultLanguageHints = map[string]perLanguageDefaults{
    "python":     {LoggingCalls: []string{"log", "logger", "logging", "print"}, MetricsCalls: []string{"metrics", "counter", "histogram", "gauge", "timer"}},
    "typescript": {LoggingCalls: []string{"log", "logger", "console.log", "console.error"}, MetricsCalls: []string{"metrics", "counter", "histogram"}},
    "go":         {LoggingCalls: []string{"log", "slog", "fmt.Println"}, MetricsCalls: []string{"prometheus", "metrics"}},
    "java":       {LoggingCalls: []string{"log", "logger", "LOG"}, MetricsCalls: []string{"metrics", "counter", "meter"}},
    "ruby":       {LoggingCalls: []string{"log", "logger", "puts"}, MetricsCalls: []string{"metrics", "counter"}},
    "rust":       {LoggingCalls: []string{"log", "tracing"}, MetricsCalls: []string{"metrics", "prometheus"}},
}
```

**Contract**:
- User-provided `logging_calls` and `metrics_calls` from config are ADDITIVE (merged with language defaults, not replacing).
- The merged result is available via a new method on `Config`: `MergedLoggingCalls(language string) []string` and `MergedMetricsCalls(language string) []string`.
- I/O pattern lists are built-in per language and are NOT configurable (constitution principle V — they are "how to recognize" patterns, but they aren't user-configurable because I/O calls are structural, not conventions).

---

## Violation Message Format

AST checks produce violations following the existing `rules.Violation` structure. Messages must be specific and actionable:

| Check | Message Format |
|-------|---------------|
| VH-G002 | `Function '{name}' has {count} statements (threshold: {threshold})` |
| VH-G003 | `I/O call '{call}' at line {line} is missing logging in the same scope` |
| VH-G004 | `Error is swallowed in catch/except block at line {line}` |
| VH-G009 | `I/O call '{call}' at line {line} is not wrapped in error handling` |
| VH-G010 | `Broad exception type '{type}' caught at line {line}` |
| VH-G012 | `File has {count} public exports (threshold: {threshold})` |

**Contract**:
- `Line` and `Column` are extracted from AST node positions where available.
- `EndLine` is set for multi-line constructs (function bodies, catch blocks).
- `Severity` is taken from the rule definition, not configurable.