# Roadmap

## Phase 1: Foundation (MVP)

**Goal:** A working binary that runs the 6 non-AST generic checks on any source tree.

1. **Go project setup**
   - Initialize Go module, project structure
   - CLI entry point with Cobra or standard `flag` package
   - File discovery (walk directory tree, filter by extension)
   - Output formatting: human-readable to stderr, JSON and SARIF formats

2. **Non-AST checks (VH-G001, G005, G006, G007, G008, G011)**
   - VH-G001: File length — line counting, blank/comment filtering
   - VH-G005: Hardcoded secrets — regex pattern matching
   - VH-G006: Magic values — literal scanning with allowed-list
   - VH-G007: Copy-paste duplication — token sequence matching
   - VH-G008: Comment-to-code ratio — line classification and counting
   - VH-G011: Disabled security features — known parameter pattern matching

3. **Configuration loading**
   - Parse `.vibe_harness.toml` for recognition hints only
   - Validate that config only contains hint fields, not rule modifications
   - Fail explicitly if config attempts to disable or modify rules

4. **Exit codes and CI output**
   - Exit 0 = pass, 1 = violations, 2 = tool error
   - JSON output for machine consumers
   - SARIF output for GitHub Code Scanning

5. **Release pipeline**
   - GitHub Actions cross-compilation matrix (darwin/linux/windows × amd64/arm64)
   - Binary signing / checksums
   - GitHub Release with all platform assets
   - Install script with OS/arch detection

**Deliverable:** `vibe-harness v0.1.0` — 6 checks, zero dependencies, single binary.

---

## Phase 2: AST-Powered Checks

**Goal:** Add tree-sitter parsing for the 6 AST-dependent generic checks.

1. **Tree-sitter integration**
   - Add `go-tree-sitter` dependency
   - Bundle language grammars (Python, TypeScript, Go, Java, Ruby, Rust) via go-embed
   - Language detection from file extensions
   - Parse files → AST, handle parse errors gracefully

2. **AST checks (VH-G002, G003, G004, G009, G010, G012)**
   - VH-G002: Function length — count statements in function/method bodies
   - VH-G003: Missing logging in I/O — identify I/O calls, check for logging calls in same scope
   - VH-G004: Swallowed errors — find empty catch/except blocks
   - VH-G009: Missing error handling on I/O — find I/O calls outside error-handling constructs
   - VH-G010: Broad exception catching — identify root-type catches
   - VH-G012: God module — count public exports per file

3. **Tree-sitter query system**
   - Define queries per language per rule
   - Map generic rule concepts to language-specific AST patterns
   - Handle language-specific edge cases (Go error returns, Rust Result, Python bare except)

4. **Recognition hints in action**
   - Use `[observability].logging_calls` from config to identify logging calls
   - Use `[observability].metrics_calls` from config to identify metrics calls
   - Default hint sets for common frameworks (Spring Boot, Rails, FastAPI, etc.)

**Deliverable:** `vibe-harness v0.2.0` — all 12 generic checks, tree-sitter parsing for 6 languages.

---

## Phase 3: Language-Specific Checks

**Goal:** Add rules that only make sense for specific languages.

1. **Python-specific rules**
   - Bare `except:` detection
   - Mutable default arguments (`def foo(items=[])`)
   - `print()` instead of logging
   - Star imports (`from x import *`)
   - Missing type annotations on public functions

2. **TypeScript-specific rules**
   - `any` type usage
   - `console.log` instead of structured logger
   - Missing return type annotations
   - `async` function without `await`
   - Missing Error Boundaries in React components

3. **Go-specific rules**
   - Ignored error return values
   - Missing `context.Context` on I/O functions
   - `fmt.Println` instead of structured logging
   - `log.Fatal` outside `main` package
   - `panic` usage in library code

4. **Java-specific rules**
   - `System.out.println` instead of Logger
   - `@Autowired` field injection (constructor injection preferred)
   - Missing `@Transactional` on service methods
   - Empty catch blocks with only comments
   - `throws Exception` on method signatures

5. **Ruby-specific rules**
   - `puts` / `p` / `pp` instead of Rails logger
   - Missing `permit` on controller params
   - `rescue Exception` catching too broadly
   - N+1 query patterns (basic detection)

6. **Rust-specific rules**
   - `.unwrap()` usage (flag all instances)
   - `.expect()` with vague messages
   - `println!` / `eprintln!` instead of tracing
   - Excessive `.clone()` per function
   - `unsafe` blocks without safety comments

**Deliverable:** `vibe-harness v0.3.0` — 12 generic + ~30 language-specific checks.

---

## Phase 4: Distribution & Ecosystem

**Goal:** Make it trivial to adopt in any project.

1. **Package manager wrappers**
   - npm package (`npm install -g vibe-harness`) with post-install binary download
   - PyPI package (`pip install vibe-harness`) with post-install binary download
   - Homebrew tap (`brew install jgervais/tap/vibe-harness`)
   - Cargo install support (`cargo install vibe-harness-cli` or `cargo vibe-harness`)

2. **Build system templates**
   - Pre-commit hook definition (`.pre-commit-hooks.yaml`)
   - GitHub Action (`action.yml`) for one-step CI integration
   - GitLab CI template
   - Example Makefile targets

3. **Editor integration**
   - LSP-style output for editor diagnostics
   - VS Code extension (wrapper that runs the binary)
   - JetBrains plugin (wrapper that runs the binary)

4. **Documentation**
   - Website with rule reference, install guides, CI examples
   - Per-language guides with framework-specific advice
   - Migration guide from existing linters (ESLint, Ruff, etc.)

**Deliverable:** `vibe-harness v1.0.0` — stable API, stable rules, multiple install methods, CI/editor integrations.

---

## Phase 5: Advanced Detection

**Goal:** Push beyond static patterns into deeper analysis.

1. **Cross-file analysis**
   - Dead code detection (symbols exported but never imported)
   - Circular dependency detection
   - Inconsistent naming patterns across files

2. **AI behavior detection**
   - Patterns suggesting agent-generated code that circumvented rules
   - Suspicious suppression of detection (e.g., wrapping I/O in a "helper" just to avoid the logging check)
   - Gaming detection — flag patterns that look like workarounds

3. **Metrics reporting**
   - Aggregate quality scores per directory/module
   - Trend tracking (is the codebase getting better or worse?)
   - Diff mode — only flag violations in changed files (for PR review)

4. **Custom grammar support**
   - Load external tree-sitter grammars for niche languages
   - Community-contributed grammar packs

**Deliverable:** `vibe-harness v2.0.0` — cross-file analysis, gaming detection, metrics.

---

## Timeline Estimate

| Phase | Scope | Effort |
|-------|-------|--------|
| Phase 1 | Non-AST checks + release pipeline | 2-3 weeks |
| Phase 2 | Tree-sitter + AST checks | 3-4 weeks |
| Phase 3 | Language-specific checks | 4-6 weeks |
| Phase 4 | Distribution & ecosystem | 2-3 weeks |
| Phase 5 | Advanced detection | Ongoing |

**v0.1.0 (MVP):** ~3 weeks
**v1.0.0 (stable):** ~3-4 months
**v2.0.0 (advanced):** ~6 months