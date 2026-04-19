# Quickstart: Phase 1 Foundation — Non-AST Checks MVP

## Prerequisites

- Go 1.22+ installed
- Git repository with source code to scan

## Local Build and Run

```bash
# Clone and build
git clone https://github.com/jgervais/vibe_harness.git
cd vibe_harness
go build -o vibe-harness ./cmd/vibe-harness

# Run against a source tree
./vibe-harness /path/to/source

# Run with JSON output
./vibe-harness --format json /path/to/source

# Run with SARIF output
./vibe-harness --format sarif /path/to/source > results.sarif

# Check exit code
./vibe-harness /path/to/source && echo "All checks passed" || echo "Violations found"
```

## Configuration (Optional)

Create `.vibe_harness.toml` in the root of your project:

```toml
[observability]
logging_calls = ["log", "logger", "slog"]

[languages]
".py" = "python"
".ts" = "typescript"
".go" = "go"
```

The tool auto-discovers this file by walking up from the target directory.
No configuration is required — the tool works with built-in defaults.

## CI Integration

### GitHub Actions

```yaml
- name: Quality Gate
  run: |
    curl -sL https://github.com/jgervais/vibe_harness/releases/latest/download/vibe-harness-$(uname -s)-$(uname -m) -o vibe-harness
    chmod +x vibe-harness
    ./vibe-harness --format sarif . > vibe-results.sarif
    ./vibe-harness .  # Also set pipeline exit code
  continue-on-error: true

- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: vibe-results.sarif
```

### Makefile

```makefile
lint: vibe-harness .
.PHONY: lint
```

## Test Scenarios

### Scenario 1: Clean Repository

A repository with well-structured code produces zero violations and
exit code 0.

### Scenario 2: Violations Detected

A repository with a 500-line file, an inline API key, and a duplicated
8-line block produces 3 violations and exit code 1.

### Scenario 3: Invalid Configuration

A `.vibe_harness.toml` containing `enabled = false` for any rule
produces an error message and exit code 2.

### Scenario 4: No Configuration File

Running without any `.vibe_harness.toml` uses built-in defaults and
completes normally.

## Verification

After building, verify the binary works:

```bash
# Should print version and exit 0
./vibe-harness --version

# Should scan the tool's own source and report results
./vibe-harness .

# Should produce valid JSON
./vibe-harness --format json . | python3 -m json.tool > /dev/null

# Should produce valid SARIF
./vibe-harness --format sarif . | python3 -m json.tool > /dev/null
```