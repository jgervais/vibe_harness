# vibe-harness

vibe-harness is a static analysis tool that runs 6 non-AST quality checks on any source tree. It is language-agnostic — no AST parsing, no tree-sitter, no language-specific logic. Just text-level checks that work everywhere.

## Checks

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| VH-G001 | File Length | warning | Files must not exceed 300 non-blank, non-comment lines |
| VH-G005 | Hardcoded Secrets | error | Detects hardcoded secrets and credentials |
| VH-G006 | Magic Values | warning | Detects magic numbers and inline strings |
| VH-G007 | Copy-Paste Duplication | warning | Detects duplicated code blocks (6+ lines, 80%+ similarity) |
| VH-G008 | Comment-to-Code Ratio | note | Flags files where comments exceed 1:3 ratio |
| VH-G011 | Disabled Security Features | error | Detects disabled security verification |

## Usage

```bash
# Build
go build -o vibe-harness ./cmd/vibe-harness

# Scan a directory
./vibe-harness /path/to/source

# JSON output
./vibe-harness --format json /path/to/source

# SARIF output for CI
./vibe-harness --format sarif /path/to/source > results.sarif

# Use config
./vibe-harness --config .vibe_harness.toml /path/to/source
```

## Configuration

The tool accepts a `.vibe_harness.toml` file for recognition hints — pattern definitions that tell it how to detect issues in your codebase.

## Build & Test

```bash
go build -o vibe-harness ./cmd/vibe-harness
go test ./...
go vet ./...
```

## Install

```bash
curl -sL https://github.com/jgervais/vibe_harness/releases/latest/download/install.sh | bash
```

## Exit Codes

- `0` — all checks passed
- `1` — one or more violations found

## License

MIT