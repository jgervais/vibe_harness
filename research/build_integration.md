# Build System & CI Integration

## 1. How Existing Linters Integrate

### ESLint (npm)
- `npm install --save-dev eslint` + `eslint .` in npm scripts
- Config: `.eslintrc.*` or `eslint.config.js`
- Exit: 0 = pass, 1 = lint errors, 2 = fatal/crash
- Output: `--format json` for machine-readable, `--format stylish` for humans
- CI: `npx eslint .` in GitHub Actions, GitLab CI, etc.

### Ruff (Python)
- `pip install ruff` + `ruff check .`
- Config: `pyproject.toml` `[tool.ruff]`
- Exit: 0 = pass, 1 = violations, 2 = error
- Output: `--output-format json` for SARIF/machine, default for humans
- CI: `ruff check .` in any pipeline

### golangci-lint (Go)
- `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` + `golangci-lint run`
- Config: `.golangci.yml`
- Exit: 0 = pass, 1-3 = issues, 4+ = errors
- Output: `--out-format json` or `--out-format code-climate`
- CI: `golangci-lint run` in pipeline

### Common Patterns
- **npm scripts:** `"lint": "vibe-harness ."` in `package.json`
- **Make:** `lint: vibe-harness .` in Makefile
- **Gradle:** Custom task type executing the binary
- **Maven:** Exec plugin `<executable>vibe-harness</executable>`
- **Cargo:** `cargo-xyz` binary naming convention for `cargo xyz`

## 2. Build System Integration

### npm / Node.js
```json
{
  "scripts": {
    "lint": "vibe-harness .",
    "lint:ci": "vibe-harness . --format sarif"
  }
}
```
Add to `npm run lint` alongside ESLint. Runs in same CI step.

### pip / Python
```toml
# pyproject.toml
[project.scripts]
lint = "vibe-harness ."
```
Or directly: `vibe-harness .` after `pip install vibe-harness`

### Go
```makefile
lint: vibe-harness .
```
Or as a `go generate` directive, or in `golangci-lint` as a custom linter.

### Maven
```xml
<plugin>
  <groupId>org.codehaus.mojo</groupId>
  <artifactId>exec-maven-plugin</artifactId>
  <executions>
    <execution>
      <phase>validate</phase>
      <goals><goal>exec</goal></goals>
      <configuration>
        <executable>vibe-harness</executable>
        <arguments><argument>src/main</argument></arguments>
      </configuration>
    </execution>
  </executions>
</plugin>
```

### Gradle
```kotlin
tasks.register("vibeCheck", Exec::class) {
    commandLine("vibe-harness", "src/main")
}
tasks.check { dependsOn("vibeCheck") }
```

### Cargo
Binary named `cargo-vibe-harness` enables `cargo vibe-harness`.

### Make (Universal)
```makefile
VIBE_HARNES ?= vibe-harness
lint: ## Run quality gate
	$(VIBE_HARNES) .
.PHONY: lint
```

## 3. CI/CD Integration

### GitHub Actions
```yaml
- name: Vibe Harness Quality Gate
  run: |
    curl -sL https://github.com/jgervais/vibe_harness/releases/latest/download/vibe-harness-$(uname -s)-$(uname -m) -o vibe-harness
    chmod +x vibe-harness
    ./vibe-harness . --format sarif > vibe-results.sarif
  continue-on-error: true  # Don't block, upload results

- name: Upload SARIF results
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: vibe-results.sarif
```

### GitLab CI
```yaml
vibe_check:
  stage: test
  image: alpine:latest
  before_script:
    - apk add curl
    - curl -sL $VIBE_HARNES_URL -o vibe-harness && chmod +x vibe-harness
  script:
    - ./vibe-harness .
  artifacts:
    reports:
      codequality: vibe-results.json
```

### Pre-commit Framework
```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/jgervais/vibe_harness
    rev: v0.1.0
    hooks:
      - id: vibe-harness
        args: ['.']
```
Requires a `.pre-commit-hooks.yaml` in the repo:
```yaml
- id: vibe-harness
  name: vibe-harness
  description: AI code quality gate
  entry: vibe-harness
  language: golang  # or rust
  types: [text]
```

## 4. Exit Codes & Output Formats

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | All checks passed |
| 1 | Quality gate violations found |
| 2 | Tool error (misconfiguration, parse failure) |

### Output Formats

**Human (default, stderr):**
```
src/handlers/orders.py:42: missing-logging — function "process_order" performs I/O but has no logging calls
src/main.go:15: ignored-error — "doSomething()" return value (error) is discarded
src/App.tsx:88: console-debug — "console.log" used instead of structured logger
```

**JSON (machine-readable):**
```json
{
  "version": "1.0",
  "tool": "vibe-harness",
  "rules_hash": "abc123def",
  "results": [
    {
      "rule": "missing-logging",
      "file": "src/handlers/orders.py",
      "line": 42,
      "function": "process_order",
      "message": "function performs I/O but has no logging calls"
    }
  ]
}
```

**SARIF (for GitHub Code Scanning):**
```json
{
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  "runs": [{
    "tool": {
      "driver": {
        "name": "vibe-harness",
        "rules": [...]
      }
    },
    "results": [...]
  }]
}
```

## 5. Distribution Strategy

### Single Binary (Recommended)
Build as a static binary using Go or Rust:
- **Go:** `CGO_ENABLED=0 go build -ldflags="-s -w"` → single binary, no dependencies
- **Rust:** `cargo build --release` → single binary, musl target for fully static

### Release Assets
GitHub Releases with platform-specific binaries:
- `vibe-harness-darwin-arm64` (macOS Apple Silicon)
- `vibe-harness-darwin-amd64` (macOS Intel)
- `vibe-harness-linux-arm64` (Linux ARM)
- `vibe-harness-linux-amd64` (Linux x86_64)
- `vibe-harness-windows-amd64.exe` (Windows)

### Install Methods
```bash
# Direct download
curl -sL https://github.com/jgervais/vibe_harness/releases/latest/download/vibe-harness-$(uname -s)-$(uname -m) | sudo tee /usr/local/bin/vibe-harness > /dev/null && sudo chmod +x /usr/local/bin/vibe-harness

# Homebrew
brew install jgervais/tap/vibe-harness

# npm (wrapper)
npm install -g vibe-harness

# pip (wrapper)
pip install vibe-harness
```

## 6. Opacity / Anti-Tampering

To prevent AI agents from subverting the rules:

- **Rules hash in output** — include a hash of the active rule set so CI can detect binary tampering
- **No `--disable` or `--skip` flags** — literally no way to turn rules off
- **No inline comments to suppress** — no `// vibe-harness-ignore` mechanism
- **No `.vibe_harness_ignore` file** — no file-level ignoring
- **Binary is signed** — releases include checksums, CI verifies before running
- **Rules are compiled in** — not loaded from external config files that an agent could modify

The only configuration is `.vibe_harness.toml` for recognition hints (what your logging library is called), which cannot modify rule behavior.

---

_Compiled 2026-04-19_