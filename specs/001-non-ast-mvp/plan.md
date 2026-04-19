# Implementation Plan: Phase 1 Foundation — Non-AST Checks MVP

**Branch**: `001-non-ast-mvp` | **Date**: 2026-04-19 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-non-ast-mvp/spec.md`

## Summary

Build a working Go binary that runs 6 non-AST generic quality checks
(VH-G001, G005, G006, G007, G008, G011) on any source tree. The tool
accepts recognition hints only via `.vibe_harness.toml`, outputs
violations in human-readable, JSON, and SARIF formats, and ships as a
single static binary with a cross-platform release pipeline. This is
the MVP deliverable: `vibe-harness v0.1.0`.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: BurntSushi/toml (config parsing), goreleaser (release pipeline)
**Storage**: N/A (no persistence, ephemeral per-scan)
**Testing**: Go standard `testing` package, table-driven tests, `testdata/` fixtures
**Target Platform**: darwin/linux/windows × amd64/arm64
**Project Type**: CLI tool
**Performance Goals**: Scan 1,000 files in under 10 seconds
**Constraints**: Single static binary (CGO_ENABLED=0), zero runtime deps, no config-based rule modification
**Scale/Scope**: Repositories up to 10,000 source files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Non-Configurable Quality Floor | PASS | All 6 checks have fixed thresholds. Config accepts recognition hints only. |
| II. Language-Agnostic by Default | PASS | All Phase 1 checks operate on raw text. No language-specific paths. Default language mappings cover all 6 languages. AST integration deferred to Phase 2. |
| III. Test-First (NON-NEGOTIABLE) | PASS | Table-driven tests specified per FR. Test fixtures in testdata/. TDD cycle enforced in task plan. |
| IV. Single Binary, Zero Runtime Dependencies | PASS | CGO_ENABLED=0 static build. Only dependency (BurntSushi/toml) is pure Go. goreleaser for distribution. |
| V. Recognition Hints Only | PASS | Config limited to [observability] and [languages]. Validation rejects rule-modifying keys. No skip/ignore/exempt. |

**Post-design re-check**: All 5 principles still pass. No violations introduced.

## Project Structure

### Documentation (this feature)

```text
specs/001-non-ast-mvp/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli.md           # CLI interface contract
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── vibe-harness/
    └── main.go              # CLI entry point

internal/
├── checks/
│   └── generic/
│       ├── file_length.go       # VH-G001
│       ├── hardcoded_secrets.go # VH-G005
│       ├── magic_values.go     # VH-G006
│       ├── duplication.go      # VH-G007
│       ├── comment_ratio.go    # VH-G008
│       └── security_features.go # VH-G011
├── config/
│   ├── config.go               # .vibe_harness.toml loading
│   └── validate.go             # Config validation (reject rule mods)
├── output/
│   ├── human.go                # Human-readable formatter
│   ├── json.go                 # JSON formatter
│   └── sarif.go                # SARIF formatter
├── rules/
│   └── registry.go             # Rule definitions and thresholds
└── scanner/
    └── scanner.go              # File discovery and orchestration

testdata/
├── file_length/
│   ├── clean.go                # Under threshold
│   └── violating.go           # Over threshold
├── hardcoded_secrets/
│   ├── clean.py
│   └── violating.py
├── magic_values/
│   ├── clean.ts
│   └── violating.ts
├── duplication/
│   ├── file_a.go
│   └── file_b.go
├── comment_ratio/
│   ├── clean.go
│   └── violating.rb
├── security_features/
│   ├── clean.py
│   └── violating.go
└── config/
    ├── valid.toml
    └── invalid_rule_mod.toml

.github/
└── workflows/
    ├── ci.yml                  # PR checks: test, vet, build
    └── release.yml             # Tag-triggered: goreleaser

.goreleaser.yml                 # Release configuration
go.mod
go.sum
Makefile
README.md
LICENSE
```

**Structure Decision**: Single Go project following standard Go layout.
`cmd/` for CLI entry point, `internal/` for implementation packages,
`testdata/` for test fixtures. This matches the Go community conventions
and the monorepo layout in docs/monorepo.md.

## Complexity Tracking

> No constitution violations to justify.