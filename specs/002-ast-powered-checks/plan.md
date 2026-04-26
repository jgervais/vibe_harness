# Implementation Plan: AST-Powered Checks

**Branch**: `002-ast-powered-checks` | **Date**: 2026-04-19 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/002-ast-powered-checks/spec.md`

## Summary

Add tree-sitter parsing infrastructure and 6 AST-dependent generic checks (VH-G002, G003, G004, G009, G010, G012) for Python, TypeScript, Go, Java, Ruby, and Rust. Grammars are compiled via CGO into the binary, loaded per-file based on extension detection, and queried with language-specific S-expression patterns. Each check conforms to the existing `Check` interface and integrates into the scanner pipeline alongside non-AST checks.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: go-tree-sitter v0.25+, tree-sitter grammar bindings (Python, TypeScript/TSX, Go, Java, Ruby, Rust), BurntSushi/toml (existing)
**Storage**: N/A (stateless CLI tool; config from `.vibe_harness.toml`)
**Testing**: `go test ./...` (standard Go testing; table-driven tests per check per language)
**Target Platform**: Cross-platform static binary (darwin/linux/windows × amd64/arm64)
**Project Type**: CLI tool (single binary linter)
**Performance Goals**: Parse + check a 10k LOC file in under 500ms; full scan throughput comparable to Phase 1 non-AST checks
**Constraints**: Single binary, zero runtime dependencies (constitution principle IV); CGO required for tree-sitter; grammar compilation at build time
**Scale/Scope**: 6 checks × 6 languages = 36 language-check combinations; config-driven observability hint sets

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Non-Configurable Quality Floor | PASS | AST checks enforce fixed thresholds (function length, export counts, etc.). Config only provides recognition hints (logging/metrics call names) which is explicitly permitted. No rule disable/modify/exempt controls. |
| II. Language-Agnostic by Default | PASS | All 6 AST checks are generic rules applied across 6 languages. Language-specific queries map to the same generic rule concepts. No language-specific checks are added that would degrade the generic baseline. Tree-sitter grammars are bundled via CGO (go-embed equivalent — compiled into binary). |
| III. Test-First | PASS | Plan mandates test-per-FR cycle. Each functional requirement will have tests written immediately after implementation, verified, then proceed to next FR. |
| IV. Single Binary, Zero Runtime Dependencies | PASS | Grammars compiled into binary via CGO (each grammar package compiles C sources directly). No external grammar files, shared libraries, or runtime downloads required. Exit codes and output formats remain unchanged. |
| V. Recognition Hints Only | PASS | Config additions are limited to observability hint expansion (default hint sets). The existing config validator already rejects rule override keys. No threshold/severity/skip additions. |

**Result**: All gates PASS. No violations to justify.

### Post-Design Re-Check (after Phase 1)

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Non-Configurable Quality Floor | PASS | Thresholds are hardcoded constants (50 statements for G002, 20 exports for G012). No config keys added that modify rule behavior. I/O patterns are built-in per language, not user-configurable. |
| II. Language-Agnostic by Default | PASS | All 6 AST checks operate across all supported languages with language-specific query mappings. No check is language-exclusive; G010 is naturally inapplicable to Go/Rust (which lack exception hierarchies) and correctly produces no violations for those languages. |
| III. Test-First | PASS | Each FR has a corresponding test plan. Table-driven tests per check per language are specified in the data model. |
| IV. Single Binary, Zero Runtime Dependencies | PASS | Grammars are CGO-compiled Go packages — linked directly into the binary. No `.so` files, no runtime grammar downloads. `go build` produces a single static binary. |
| V. Recognition Hints Only | PASS | The `observability` config section (logging_calls, metrics_calls) is additive with defaults per language. No new config keys for thresholds, enables, skips, or severity. I/O patterns are built-in recognition data, not user-configurable. Config validator unchanged. |

**Post-Design Result**: All gates PASS. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/002-ast-powered-checks/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── ast-parser-contract.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
cmd/vibe-harness/
└── main.go                          # Updated: no changes needed (scanner handles routing)

internal/
├── ast/
│   ├── parser.go                    # AST parser: language detection, grammar loading, file parsing
│   ├── parser_test.go               # Parser tests
│   ├── languages.go                 # Language registry: extension→language mapping, grammar init
│   └── languages_test.go
├── checks/
│   └── generic/
│       ├── file_length.go           # VH-G001 (existing)
│       ├── hardcoded_secrets.go     # VH-G005 (existing)
│       ├── magic_values.go          # VH-G006 (existing)
│       ├── duplication.go           # VH-G007 (existing)
│       ├── comment_ratio.go         # VH-G008 (existing)
│       ├── security_features.go     # VH-G011 (existing)
│       ├── function_length.go       # VH-G002 (NEW)
│       ├── function_length_test.go
│       ├── missing_logging.go       # VH-G003 (NEW)
│       ├── missing_logging_test.go
│       ├── swallowed_errors.go      # VH-G004 (NEW)
│       ├── swallowed_errors_test.go
│       ├── missing_error_handling.go# VH-G009 (NEW)
│       ├── missing_error_handling_test.go
│       ├── broad_catch.go           # VH-G010 (NEW)
│       ├── broad_catch_test.go
│       └── god_module.go            # VH-G012 (NEW)
│       └── god_module_test.go
├── config/
│   ├── config.go                    # Existing: add default observability hint sets
│   ├── config_test.go               # Existing
│   └── validate.go                   # Existing: no changes needed (already blocks rule overrides)
├── output/                           # Existing: no changes needed
├── rules/
│   ├── registry.go                  # Updated: add 6 AST check metadata entries
│   └── registry_test.go
└── scanner/
    ├── scanner.go                   # Updated: instantiate AST checks, pass parser, route AST vs non-AST
    └── scanner_test.go

testdata/
├── function_length/                 # VH-G002 fixtures
│   ├── clean.go
│   ├── violating.go
│   ├── clean.py
│   ├── violating.py
│   ├── clean.ts
│   ├── violating.ts
│   ├── clean.java
│   ├── violating.java
│   ├── clean.rb
│   ├── violating.rb
│   ├── clean.rs
│   └── violating.rs
├── missing_logging/                  # VH-G003 fixtures
├── swallowed_errors/                 # VH-G004 fixtures
├── missing_error_handling/           # VH-G009 fixtures
├── broad_catch/                      # VH-G010 fixtures
└── god_module/                       # VH-G012 fixtures

go.mod                               # Updated: add go-tree-sitter + 6 grammar dependencies
go.sum                               # Updated: checksums
```

**Structure Decision**: Single binary CLI project following existing structure. New `internal/ast/` package for tree-sitter integration. New check files in existing `internal/checks/generic/` following naming convention. New `testdata/` directories per check per language. Grammar dependencies added to `go.mod`.

## Complexity Tracking

> No violations — all constitution principles pass.