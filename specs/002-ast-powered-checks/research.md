# Research: AST-Powered Checks

**Feature**: 002-ast-powered-checks | **Date**: 2026-04-19

## R1: go-tree-sitter Integration Approach

**Decision**: Use `github.com/tree-sitter/go-tree-sitter` with CGO-compiled grammar packages for each language.

**Rationale**: go-tree-sitter is the official Go binding maintained by the tree-sitter project. It uses CGO to compile C grammar sources directly into the binary, satisfying the single-binary, zero-runtime-dependencies constraint (constitution principle IV). No `.so` files, no runtime grammar downloads, no external dependencies.

**Alternatives considered**:
- **Runtime `.so` loading via purego**: Allows loading grammars as shared libraries at runtime. Rejected because it violates principle IV (requires external grammar files) and adds complexity for binary distribution.
- **Custom parser generator**: Build a language-agnostic parser from scratch. Rejected because it would require implementing parsers for 6 languages, each with complex grammar rules, and would be far less accurate than tree-sitter's battle-tested grammars.

**Key API details**:
- `tree_sitter.NewParser()` / `parser.SetLanguage(lang)` / `parser.Parse(source, nil)`
- `tree.RootNode()` / `node.Child(i)` / `node.Kind()` / `node.Utf8Text(source)`
- `tree_sitter.NewQuery(lang, querySrc)` for S-expression queries
- `tree_sitter.NewQueryCursor()` / `cursor.Matches(query, root, source)` for executing queries
- All heap-allocated objects (`Parser`, `Tree`, `Query`, `QueryCursor`) must have `Close()` called
- `Parser` is NOT thread-safe — one parser per goroutine or mutex-protect

**CGO requirement**: Build requires `CGO_ENABLED=1` and a C compiler. This is acceptable because the existing project already uses `CGO_ENABLED=1` for cross-compilation via goreleaser.

## R2: Grammar Bundling Strategy

**Decision**: Add each language grammar as a Go module dependency. Each grammar package exposes a `Language()` function returning `unsafe.Pointer` to the compiled C grammar function. Grammars are compiled into the binary via CGO `#cgo` directives during `go build`.

**Rationale**: This aligns with how tree-sitter grammar bindings are distributed — each grammar repository includes Go bindings under `bindings/go/` with CGO directives to compile the C sources. No runtime loading needed.

**Grammar packages**:
| Language | Go Module | Import Path |
|----------|-----------|-------------|
| Python | `github.com/tree-sitter/tree-sitter-python` | `github.com/tree-sitter/tree-sitter-python/bindings/go` |
| Go | `github.com/tree-sitter/tree-sitter-go` | `github.com/tree-sitter/tree-sitter-go/bindings/go` |
| TypeScript | `github.com/tree-sitter/tree-sitter-typescript` | `github.com/tree-sitter/tree-sitter-typescript/bindings/go` |
| Java | `github.com/tree-sitter/tree-sitter-java` | `github.com/tree-sitter/tree-sitter-java/bindings/go` |
| Ruby | `github.com/tree-sitter/tree-sitter-ruby` | `github.com/tree-sitter/tree-sitter-ruby/bindings/go` |
| Rust | `github.com/tree-sitter/tree-sitter-rust` | `github.com/tree-sitter/tree-sitter-rust/bindings/go` |

**Note**: TypeScript grammar exports two functions: `LanguageTypescript()` and `LanguageTSX()`. Both `.ts` and `.tsx` extensions should map to TypeScript check behavior; `.tsx` specifically uses the TSX grammar.

**Alternatives considered**:
- **Embed `.so` files + purego runtime loading**: Allows swapping grammars without recompilation. Rejected — violates principle IV and adds distribution complexity.
- **Embed grammar source + compile at runtime**: Massive startup cost and requires a C compiler at runtime. Rejected.

## R3: Language Detection from File Extensions

**Decision**: Use the existing `config.Config.Languages` map (from `.vibe_harness.toml`) to map file extensions to language names. The language name then maps to the tree-sitter grammar. Default extensions cover all 6 supported languages.

**Rationale**: The config system already has `[languages]` mapping. Reusing it avoids duplication and lets users customize extension→language mapping for unusual setups (e.g., `.mjs` → `"typescript"`).

**Default mapping**:
```toml
[languages]
".py" = "python"
".ts" = "typescript"
".tsx" = "typescript"
".js" = "javascript"
".go" = "go"
".java" = "java"
".rb" = "ruby"
".rs" = "rust"
```

**Important**: JavaScript (`.js`) files are currently in the default config but no JavaScript-specific grammar is included in Phase 2. The TypeScript grammar should be applied to `.js` files (since TS grammar is a superset), or `.js` files should be treated as unsupported for AST checks. Decision: apply the TypeScript grammar to `.js` files — TypeScript parser handles valid JavaScript. This is a reasonable trade-off since `.js` → `"javascript"` maps to the TypeScript grammar.

## R4: AST Parser Architecture

**Decision**: Create an `internal/ast` package that provides:
1. A `LanguageRegistry` — maps language names to `*tree_sitter.Language` instances, initialized lazily
2. A `Parser` — wraps `tree_sitter.Parser`, provides a `ParseFile(path, content, language string) (*tree_sitter.Tree, error)` method
3. `ParseResult` — contains the parsed tree and source bytes; handles cleanup via `Close()`

**Rationale**: Centralizing grammar loading and parsing in one package avoids duplication across 6 check implementations. Each check receives a `*ast.ParseResult` (or nil if parsing failed/unsupported) and can execute its own queries against the tree.

**Error handling**: If a language is not supported (no grammar), `ParseFile` returns `nil` with no error — checks simply skip. If a parse error occurs (syntax error in source), `ParseFile` returns a `ParseResult` where `tree.RootNode().HasError()` is true — checks can decide whether to proceed or skip. This aligns with FR-003 and FR-017.

**Alternatives considered**:
- **Parse per check**: Each check creates its own parser and parses independently. Rejected — wasteful; same file would be parsed 6 times per scan.
- **Parse in scanner, pass tree to checks**: Scanner parses all files once, stores results in a map. Rejected — increases memory usage and couples scanner to AST internals. Compromise: the `internal/ast` package caches parsers and provides a simple API; the scanner calls `ParseFile` once per file per scan scope and passes the result to AST checks.

**Final architecture decision**: The scanner will parse each file once and pass the `*ast.ParseResult` to AST checks via an extended interface. Checks that require AST access will implement an `ASTCheck` interface with a `CheckFileAST(path, content, language, cfg, parseResult)` method, or the existing `CheckFile` signature can be extended to accept an optional parse result via a context/map parameter. Simpler approach: use a `ParsedFiles` map in the scanner that AST checks can access.

## R5: Query System Design

**Decision**: Each check defines its queries as Go string constants, organized by language. A `QuerySet` maps language names to S-expression query strings. At check time, the query for the detected language is compiled and executed against the parsed tree.

**Rationale**: Tree-sitter queries are S-expression patterns that are language-specific. Organizing them as constants per check per language keeps them co-located with check logic and easy to maintain. Queries are compiled at runtime (cheap, but should be cached per language per check).

**Pattern**:
```go
// In swallowed_errors.go
var g004Queries = map[string]string{
    "python": `(except_clause body: (block (pass_statement))) @swallowed`,
    "typescript": `(catch_clause body: (statement_block)) @swallowed`,
    "go": `...` ,
    "java": `...`,
    "ruby": `...`,
    "rust": `...`,
}
```

**Query caching**: Since queries are compiled per-language per-check, and there are 6 checks × 6 languages = 36 combinations, queries should be compiled once and cached. The `internal/ast` package can provide a `QueryCache` that stores `*tree_sitter.Query` by language+check ID.

**Alternatives considered**:
- **External query files (`.scm`)**: Store queries as separate files loaded at runtime. Rejected — violates principle IV (external file dependency). Also harder to version and test.
- **Code-generate queries from a DSL**: Abstract away language differences with a metalanguage. Rejected — adds complexity without benefit for 36 query strings.

## R6: Observability Hint Default Sets

**Decision**: Expand the existing `ObservabilityConfig` with default hint sets per language/framework, using the existing `LoggingCalls` and `MetricsCalls` fields. When no config file is present or these fields are empty, defaults are applied.

**Default hint sets**:

| Framework/Language | `logging_calls` | `metrics_calls` |
|---|---|---|
| Python (general) | `log`, `logger`, `logging`, `print` | `metrics`, `counter`, `histogram`, `gauge`, `timer` |
| Python (FastAPI) | + `app.logger`, `logger` | + `fastapi_instrumentor` |
| Python (Django) | + `django.logger` | + `django_metrics` |
| TypeScript (general) | `log`, `logger`, `console.log`, `console.error` | `metrics`, `counter`, `histogram` |
| TypeScript (Express) | + `winston`, `morgan` | + `prom-client` |
| Go (general) | `log`, `slog`, `fmt.Println` | `prometheus`, `metrics` |
| Java (general) | `log`, `logger`, `LOG` | `metrics`, `counter`, `meter` |
| Java (Spring) | + `LoggerFactory`, `log4j` | + `micrometer` |
| Ruby (general) | `log`, `logger`, `puts` | `metrics`, `counter` |
| Ruby (Rails) | + `Rails.logger` | + `ActiveSupport::Notifications` |
| Rust (general) | `log`, `tracing` | `metrics`, `prometheus` |

**Rationale**: The existing config has `LoggingCalls` and `MetricsCalls` string slices. These are merged with defaults at config load time (user-provided values are additive, not exclusive). Per-language default sets allow the VH-G003 check to match framework-specific call patterns.

**Implementation**: Defaults live in `internal/config/defaults.go` as a `map[string]ObservabilityConfig` keyed by language name. At config load time, the defaults for each language are merged with any user-provided values. The check receives the full merged config and can match against all known call names.

## R7: Integration with Existing Scanner Pipeline

**Decision**: Extend the scanner to parse files once and provide the `*ast.ParseResult` to AST checks. Checks that need AST access implement an `ASTCheck` interface alongside the existing `Check` interface.

**Rationale**: The existing scanner iterates over files, reads content, and runs checks. The most efficient approach is to parse each file once, then pass the parse result to all AST checks. The existing `Check` interface signature is `CheckFile(path, content, language, cfg) []Violation`. Rather than changing this signature, we add an optional `ASTCheck` interface:

```go
type ASTCheck interface {
    Check
    CheckFileAST(path string, content []byte, language string, cfg *config.Config, tree *ast.ParseResult) []rules.Violation
}
```

The scanner checks if a check implements `ASTCheck`; if so, it passes the parse result. If a check only implements `Check`, the scanner calls `CheckFile` as before.

**Alternatives considered**:
- **Modify `CheckFile` signature**: Add a parse result parameter. Rejected — would require changing all 6 existing checks' signatures.
- **Use a context/package-level variable**: Store parsed files in a map accessible to all checks. Rejected — implicit coupling, harder to test.

## R8: VH-G002 Function Length — Threshold Design

**Decision**: Default threshold of 50 statements. Not configurable (constitution principle I). The threshold counts statements (not lines) in function/method bodies, as specified by the roadmap.

**Rationale**: 50 statements is a widely accepted upper bound for function complexity. Counting statements rather than lines is more accurate (ignores blank lines, comments, closing braces). Fixed threshold satisfies principle I.

**Statement counting approach**: Walk the AST body node and count direct child statements (not recursively into nested blocks). Different languages count differently:
- Python: count top-level children of `block` node (excluding comment nodes)
- Go: count direct children of `block` node
- Java: count statements in `block` node
- TypeScript: count statements in `statement_block` node
- Ruby: count statements in `body_statement` node
- Rust: count statements in `block` node

## R9: VH-G012 God Module — Threshold Design

**Decision**: Default threshold of 20 public exports. Not configurable (constitution principle I).

**Rationale**: 20 exports is a reasonable upper bound before a module becomes difficult to understand and maintain. "Public" is defined per language:
- Python: top-level `def`/`class` not starting with `_`
- TypeScript: `export` keyword on any declaration
- Go: names starting with uppercase letter
- Java: `public` modifier on any declaration
- Ruby: `method`/`class`/`module` definitions (public by default, excluding `_` prefix)
- Rust: `pub` modifier without restriction (not `pub(crate)`, `pub(super)`)

## R10: VH-G003 and VH-G009 — I/O Pattern Detection

**Decision**: Build per-language I/O call pattern sets based on standard library and common framework names. These are recognition hints (matching how observability hints work) rather than configurable rules.

**Rationale**: I/O call identification is a recognition problem (like logging detection), not a rule-configuration problem. Standard library patterns per language:
- Python: `open`, `requests.*`, `socket.*`, `urllib.*`, `conn.*`, `cursor.*`
- TypeScript: `fetch`, `fs.*`, `http.*`, `axios.*`, DB client methods
- Go: `os.*`, `net.*`, `http.*`, `sql.*`, `io.*`
- Java: `Files.*`, `Socket.*`, `PreparedStatement.*`, `Connection.*`
- Ruby: `open`, `IO.*`, `Net::*`, `File.*`, DB client methods
- Rust: `std::fs::*`, `std::net::*`, `std::io::*`, `reqwest::*`

These lists are fixed per language and extended by observability hints from config for framework-specific call names.