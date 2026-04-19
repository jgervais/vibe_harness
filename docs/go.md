# Go — Build System & Existing Linters

## Build Systems

### Go Modules (go build)
- **Config:** `go.mod`
- **Integration:** Makefile or `go generate`
  ```makefile
  lint: vibe-harness .
  .PHONY: lint
  ```
  Or via `go generate`:
  ```go
  //go:generate vibe-harness .
  ```

### Task (taskfile.dev)
- **Config:** `Taskfile.yml`
- **Integration:**
  ```yaml
  tasks:
    lint:
      cmds:
        - vibe-harness .
  ```

### Mage
- **Config:** `magefile.go`
- **Integration:**
  ```go
  func Lint() error {
      return sh.RunV("vibe-harness", ".")
  }
  ```

### just
- **Config:** `justfile`
- **Integration:**
  ```makefile
  lint:
      vibe-harness .
  ```

## Frameworks

### net/http (Standard Library)
- **No task runner**
- **Integration:** Makefile or justfile

### Gin
- **No task runner**
- **Integration:** Makefile
  ```makefile
  lint:
      go vet ./...
      vibe-harness .
  ```

### Echo
- **No task runner**
- **Integration:** Same as Gin

### gRPC
- **Build:** protoc + go generate
- **Integration:** Add to Makefile after protoc step

## Existing Linters to Leverage

| Linter | What It Catches | Overlap with Vibe Harness |
|--------|----------------|--------------------------|
| **go vet** | Common mistakes | Printf format errors, unreachable code, unmarshal args |
| **golangci-lint** | Meta-linter (runs 50+ linters) | See breakdown below |
| **errcheck** | Unchecked errors | VH-G009 overlap — errcheck catches ignored error returns |
| **staticcheck** | Advanced static analysis | SA series checks, deprecated APIs |
| **gosec** | Security | Hardcoded credentials (G101), SQL injection (G201-G203), weak crypto |
| **gocognit** | Cognitive complexity | Overlap with VH-G002 |
| **gocyclo** | Cyclomatic complexity | Overlap with VH-G002 |
| **funlen** | Function length | Direct overlap with VH-G002 |
| **golint / revive** | Style | Missing doc comments, naming conventions |
| **nilerr** | Nil error handling | Checks err == nil patterns |

### golangci-lint Key Rules for Overlap
```yaml
# .golangci.yml
linters:
  enable:
    - errcheck      # VH-G009 overlap
    - funlen        # VH-G002 overlap
    - gocognit      # Complexity
    - gosec         # VH-G005, VH-G011 overlap
    - govet         # Basic correctness
    - staticcheck   # Advanced analysis
    - nilerr        # Error handling patterns

linters-settings:
  funlen:
    lines: 50       # Match VH-G002
  gocognit:
    min-complexity: 10
  gocyclo:
    min-complexity: 10
```

### Leverage Strategy
- **golangci-lint as primary** — it aggregates most Go linters
- **Vibe Harness adds what Go linters miss:** missing logging in I/O functions, missing context.Context propagation, fmt.Println instead of structured logging, log.Fatal in library code
- **errcheck** catches ignored errors (VH-G009 equivalent), but VH also catches *missing* error handling (no try at all), not just ignored returns
- **gosec** catches hardcoded secrets and disabled security features — overlaps with VH-G005 and VH-G011, but VH is non-configurable while gosec rules can be disabled
- **funlen** matches VH-G002 threshold — run both for defense in depth