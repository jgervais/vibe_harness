# Quickstart: AST-Powered Checks

**Feature**: 002-ast-powered-checks | **Date**: 2026-04-19

## Prerequisites

- Go 1.22+ with CGO enabled (`CGO_ENABLED=1`)
- C compiler (gcc/clang) — required by tree-sitter grammar compilation
- All Phase 1 checks (VH-G001, G005, G006, G007, G008, G011) implemented and passing

## Build & Run

```bash
# Install dependencies (adds go-tree-sitter + 6 grammar packages)
go mod tidy

# Build (CGO required for tree-sitter grammars)
CGO_ENABLED=1 go build -o vibe-harness ./cmd/vibe-harness

# Run all checks (non-AST + AST) against a codebase
./vibe-harness /path/to/codebase

# Run with JSON output
./vibe-harness --format json /path/to/codebase

# Run with SARIF output
./vibe-harness --format sarif /path/to/codebase

# Run with custom config
./vibe-harness --config .vibe_harness.toml /path/to/codebase
```

## Test

```bash
# Run all tests
go test ./...

# Run only AST package tests
go test ./internal/ast/...

# Run a specific check's tests
go test ./internal/checks/generic/ -run TestFunctionLength

# Run with verbose output
go test -v ./internal/checks/generic/ -run TestSwallowedErrors
```

## What Gets Scanned

| Language | Extensions | AST Checks |
|----------|-----------|------------|
| Python | `.py` | G002, G003, G004, G009, G010, G012 |
| TypeScript | `.ts`, `.tsx` | G002, G003, G004, G009, G010, G012 |
| JavaScript | `.js` | G002, G003, G004, G009, G010, G012 |
| Go | `.go` | G002, G003, G004, G009, G012 |
| Java | `.java` | G002, G003, G004, G009, G010, G012 |
| Ruby | `.rb` | G002, G003, G004, G009, G010, G012 |
| Rust | `.rs` | G002, G003, G004, G009, G012 |
| Other | (all others) | Non-AST checks only (G001, G005, G006, G007, G008, G011) |

**Notes**:
- Go doesn't have exception hierarchies; VH-G010 is skipped for Go.
- Rust doesn't have try/catch; VH-G010 is skipped for Rust.
- `.js` files use the TypeScript grammar (superset of JavaScript).

## New Checks

| ID | Name | What It Detects | Threshold |
|----|------|----------------|-----------|
| VH-G002 | Function Length | Functions/methods with too many statements | 50 statements |
| VH-G003 | Missing Logging in I/O | I/O calls without logging in the same scope | Pattern-based |
| VH-G004 | Swallowed Errors | Empty catch/except blocks (error not re-raised, returned, or logged) | Pattern-based |
| VH-G009 | Missing Error Handling on I/O | I/O calls not in error-handling constructs | Pattern-based |
| VH-G010 | Broad Exception Catching | Catching root exception types (Exception, Throwable, bare except) | Pattern-based |
| VH-G012 | God Module | Files with too many public exports | 20 exports |

## Observability Hints

The `.vibe_harness.toml` `observability` section controls how VH-G003 recognizes logging and metrics calls:

```toml
[observability]
logging_calls = ["log", "logger", "logging"]
metrics_calls = ["metrics", "counter", "histogram"]
```

These are **additive** with built-in defaults per language. Specifying custom calls does NOT remove the defaults — it extends them. This is the ONLY config that affects check behavior, per constitution principle V.

## Exit Codes

Unchanged from Phase 1:
- `0` — no violations found
- `1` — violations found
- `2` — tool error (config error, scan failure)

AST check violations are included in the violation count that determines exit code 1.

## Expected Violation Output (Human)

```
src/main.py:42:VH-G002 — Function 'process_data' has 73 statements (threshold: 50)
src/main.py:15:VH-G004 — Error is swallowed in except block at line 15
src/main.go:89:VH-G009 — I/O call 'os.ReadFile' at line 89 is not wrapped in error handling
src/app.py:5:VH-G010 — Broad exception type 'Exception' caught at line 5
src/utils.py:1:VH-G012 — File has 34 public exports (threshold: 20)
```