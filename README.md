# vibe-harness

vibe-harness is a static analysis tool with both text-level and AST-powered quality checks. A `.vibe_harness.toml` config file is required.

## Checks

| ID | Name | Severity | AST | Description |
|----|------|----------|-----|-------------|
| VH-G001 | File Length | warning | no | Files must not exceed 300 non-blank, non-comment lines |
| VH-G002 | Function Length | warning | yes | Functions must not exceed 50 statements |
| VH-G003 | Missing Logging in I/O | warning | yes | I/O calls should have logging in the same scope |
| VH-G004 | Swallowed Errors | error | yes | Catch/except blocks must handle errors (re-raise, return, or log) |
| VH-G005 | Hardcoded Secrets | error | no | Detects hardcoded secrets and credentials |
| VH-G006 | Magic Values | error | no | Detects magic numbers and inline strings repeated across the codebase |
| VH-G007 | Copy-Paste Duplication | warning | no | Detects duplicated code blocks across files |
| VH-G008 | Comment-to-Code Ratio | note | no | Flags files where comments exceed 1:3 ratio |
| VH-G009 | Missing Error Handling on I/O | error | yes | I/O calls must be wrapped in error-handling constructs |
| VH-G010 | Broad Exception Catching | warning | yes | Catching root exception types (Exception, Throwable, bare except) |
| VH-G011 | Disabled Security Features | error | no | Detects disabled security verification |
| VH-G012 | God Module | warning | yes | Files must not exceed 20 public exports |

## Usage

```bash
# Build
go build -o vibe-harness ./cmd/vibe-harness

# Scan a directory (requires .vibe_harness.toml)
./vibe-harness .

# JSON output
./vibe-harness --format json .

# SARIF output for CI
./vibe-harness --format sarif . > results.sarif

# Specify config path
./vibe-harness --config path/to/.vibe_harness.toml .
```

## Configuration

A `.vibe_harness.toml` file is required. It defines:

```toml
source_directories = ["cmd/**", "internal/**"]
test_file_pattern = ["_test.", "testdata"]

[languages]
".go" = "go"
".py" = "python"

[observability]
logging_calls = ["log", "logger", "logging", "tracing", "slog", "logr"]
metrics_calls = ["metrics", "counter", "histogram", "gauge", "timer", "prometheus"]
```

- **`source_directories`** — glob patterns (relative to scan root) specifying which directories to scan. Required.
- **`test_file_pattern`** — substrings/path patterns identifying test files. Default `["_test.", "testdata"]`.
- **`[languages]`** — file extension to language mapping. Required, must have at least one entry.
- **`[observability]`** — hints for logging/metrics detection checks.

## Build & Test

```bash
go build -o vibe-harness ./cmd/vibe-harness
go test ./...
go vet ./...
make
```

## Install

```bash
curl -sL https://github.com/jgervais/vibe_harness/releases/latest/download/install.sh | bash
```

## Exit Codes

- `0` — all checks passed
- `1` — one or more violations found
- `2` — configuration or usage error

## License

MIT
