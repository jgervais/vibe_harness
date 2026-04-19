# Rust — Build System & Existing Linters

## Build Systems

### Cargo
- **Config:** `Cargo.toml`
- **Integration:**
  - As a Cargo subcommand: binary named `cargo-vibe-harness` enables `cargo vibe-harness`
  - Makefile for CI:
    ```makefile
    lint:
        vibe-harness src/
    .PHONY: lint
    ```

### Cargo + Just
- **Config:** `justfile`
- **Integration:**
  ```makefile
  lint:
      cargo clippy
      vibe-harness src/
  ```

### Cargo + Task
- **Config:** `Taskfile.yml`
- **Integration:**
  ```yaml
  tasks:
    lint:
      cmds:
        - cargo clippy
        - vibe-harness src/
  ```

## Frameworks

### Tokio
- **Build:** Cargo
- **Integration:** Standard Cargo + vibe-harness
- **Specific concerns:** async functions without context, missing graceful shutdown, unbounded channels

### Axum
- **Build:** Cargo
- **Integration:** Standard Cargo + vibe-harness
- **Specific concerns:** missing error handling in handlers, no middleware for logging

### Actix-web
- **Build:** Cargo
- **Integration:** Standard Cargo + vibe-harness

### Serde-heavy projects
- **Build:** Cargo
- **Specific concerns:** missing deny_unknown_fields, String where Cow<str> would be better

## Existing Linters to Leverage

| Linter | What It Catches | Overlap with Vibe Harness |
|--------|----------------|--------------------------|
| **Clippy** | Style, correctness, complexity, perf |unwrap_used, expect_used, print_stdout, print_stderr, panic, module_name_repetitions |
| **rustc** | Compiler warnings | Unused variables, dead code, missing docs |
| **rust-analyzer** | IDE-level diagnostics | Type errors, missing imports |
| **cargo-audit** | Dependency vulnerabilities | Complementary — different domain |
| **cargo-deny** | Dependency licensing, banning | Complementary — different domain |
| **cargo-machete** | Unused dependencies | Complementary — dead dep detection |

### Clippy Key Lints for Overlap
```toml
# Clippy.toml
cognitive-complexity-threshold = 10
too-many-arguments-threshold = 7
type-complexity-threshold = 250
single-char-binding-names-threshold = 4
```

```rust
// lib.rs — enable aggressive Clippy lints
#![warn(clippy::unwrap_used)]
#![warn(clippy::expect_used)]
#![warn(clippy::print_stdout)]
#![warn(clippy::print_stderr)]
#![warn(clippy::panic)]
#![warn(clippy::cognitive_complexity)]
#![warn(clippy::too_many_lines)]
#![warn(clippy::too_many_arguments)]
```

### Leverage Strategy
- **Clippy first** — it's the Rust standard linter and catches a lot
- **Clippy's `unwrap_used`** overlaps directly with Rust's version of VH-G009 (ignored errors)
- **Clippy's `print_stdout`** overlaps with VH's "no println" rule
- **Vibe Harness adds what Clippy misses:** missing tracing instrumentation on handlers, .clone() overuse detection (threshold-based), missing context.Context equivalent (Rust: missing tokio context), missing graceful shutdown patterns
- **Clippy is configurable** (allows #[allow]) — VH is not, providing defense in depth
- **cargo-audit + cargo-deny** handle dependency security — complementary domain

### Key Difference: Rust's Compiler is Already Strict
Rust's type system and borrow checker catch many issues that VH catches in other languages:
- **Null handling** — Rust has no nulls, so VH's "missing null check" doesn't apply
- **Error handling** — Rust's Result type forces explicit handling, but `.unwrap()` bypasses it
- **Thread safety** — Rust's ownership system prevents data races

Vibe Harness for Rust focuses on the patterns the compiler allows but that still indicate AI-generated code quality issues: excessive `.unwrap()`, missing observability, `.clone()` overuse as a borrow-checker workaround.