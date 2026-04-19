# vibe_harness Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-19

## Active Technologies

- Go 1.22+ (001-non-ast-mvp)
- BurntSushi/toml for config parsing (001-non-ast-mvp)
- goreleaser for release pipeline (001-non-ast-mvp)

## Project Structure

```text
cmd/vibe-harness/          # CLI entry point
internal/checks/generic/   # VH-G001..G012 check implementations
internal/config/           # .vibe_harness.toml loading + validation
internal/output/           # JSON, SARIF, human-readable formatters
internal/rules/            # Rule definitions and thresholds
testdata/                  # Fixture directories for tests
.github/workflows/         # CI and release workflows
```

## Commands

- Build: `go build -o vibe-harness ./cmd/vibe-harness`
- Test: `go test ./...`
- Lint: `go vet ./...`
- Install: `go install ./cmd/vibe-harness`

## Code Style

Go: Follow standard Go conventions (gofmt, go vet, effective Go)

## Recent Changes

- 001-non-ast-mvp: Added

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
