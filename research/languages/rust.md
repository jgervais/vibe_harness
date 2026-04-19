# Rust — AI Code Smells & Tree-Sitter Detection

## 1. Common AI-Generated Rust Anti-Patterns

### Error Handling
- **`.unwrap()` everywhere** — the #1 AI Rust smell. Crashes on None/Err instead of proper handling
- **`.expect("msg")` with vague messages** — `expect("failed")` instead of `expect("database connection should be established from pool")`
- **`unwrap_or_default()` masking real errors** — silently replacing errors with defaults
- **`panic!()` for control flow** — panicking instead of returning `Result`
- **Boxed errors without context** — `Box<dyn Error>` instead of `thiserror`/`anyhow` with context
- **Missing `?` operator** — manual match on Result instead of propagating with `?`

### Option/Result Handling
- **Nested match on Option** — `match x { Some(y) => match y { ... } }` instead of `and_then`/`map`
- **`as_ref().unwrap()` pattern** — still crashes, just with extra steps
- **Ignoring `Result` return values** — calling a function that returns Result and discarding it

### Concurrency
- **`.unwrap()` on `lock()`** — `mutex.lock().unwrap()` instead of handling PoisonError
- **Missing `.join()` on threads** — spawned threads never joined
- **Deadlock-prone lock ordering** — multiple mutexes acquired in inconsistent order

### Logging & Observability
- **`println!` instead of `tracing`** — using print macros instead of structured logging
- **`eprintln!` for errors** — not using proper error logging with context
- **No span instrumentation** — missing `#[tracing::instrument]` on functions
- **Missing metrics** — no prometheus or metrics instrumentation on handlers

### Structural Smells
- **`.clone()` overuse** — fighting the borrow checker by cloning everything instead of proper lifetime design
- **God functions** — single `fn` with 100+ lines
- **Everything in `main.rs`** — no module decomposition
- **`unsafe` without justification** — AI uses `unsafe` to bypass borrow checker issues
- **Large `enum` variants** — unboxing large variants causing memory waste

## 2. Cargo Ecosystem — AI Pitfalls

- **Missing `Cargo.toml`** — AI writes standalone Rust without project setup
- **Dependency version not pinned** — `version = "*"` instead of semver range
- **Feature flags ignored** — using crate features without specifying them
- **Dev dependencies in production** — test crates in `[dependencies]` instead of `[dev-dependencies]`
- **Missing `edition` field** — `edition = "2021"` absent
- **No workspace for multi-crate projects** — multiple Cargo.toml files instead of workspace

## 3. Tree-Sitter Rust AST — Key Node Types

| Node Type | What It Captures | Use For |
|---|---|---|
| `function_item` | `fn foo()` | God function detection |
| `impl_item` | `impl Foo` | Method organization |
| `struct_item` | `struct Foo` | Struct design |
| `enum_item` | `enum Foo` | Enum variant analysis |
| `match_block` | `match x { }` | Exhaustiveness, nesting |
| `call_expression` | `foo()` | unwrap detection, logging |
| `method_call_expression` | `x.method()` | .unwrap(), .clone() detection |
| `try_expression` | `expr?` | Proper error propagation |
| `closure_expression` | `\|x\| x + 1` | Functional patterns |
| `use_declaration` | `use crate::` | Import analysis |
| `mod_item` | `mod foo;` | Module organization |
| `attribute_item` | `#[derive(...)]` | Derive patterns |
| `macro_invocation` | `println!()` | Macro usage detection |
| `unsafe_block` | `unsafe { }` | Unsafe usage |
| `type_arguments` | `<T>` | Generic types |
| `lifetime` | `'a` | Lifetime annotations |
| `let_declaration` | `let x =` | Variable bindings |
| `if_expression` | `if let` | Control flow |

## 4. Framework-Specific AI Issues

### Tokio
- **Missing `#[tokio::main]`** — async main without runtime
- **Blocking in async context** — `std::fs` calls inside async functions instead of `tokio::fs`
- **No graceful shutdown** — missing signal handlers
- **Unbounded channels** — `mpsc::channel()` instead of bounded `mpsc::channel(N)`

### Axum / Actix-web
- **Missing error handling in handlers** — returning `impl IntoResponse` without error type
- **No middleware for logging** — routes without tracing middleware
- **State as global** — `lazy_static!` instead of Axum State / Actix Data
- **Missing extract validation** — accepting raw `Json<T>` without validating T

### Serde
- **Missing `#[serde(deny_unknown_fields)]`** — silently ignoring unknown JSON fields
- **`String` where `Cow<str>` would do** — unnecessary allocations in deserialization

## 5. Detection Rules — Tree-Sitter Queries

### .unwrap() Usage
```
(method_call_expression
  method: (field_identifier) @method (#eq? @method "unwrap"))
```
Flag all instances. In production code, `.unwrap()` should be extremely rare.

### .expect() With Vague Message
```
(method_call_expression
  method: (field_identifier) @method (#eq? @method "expect")
  arguments: (arguments (string_literal) @msg
    (#match? @msg "^\")(failed|error|oops|something)")))
```

### println! Instead of Logging
```
(macro_invocation
  macro: (identifier) @name
  (#match? @name "^(println|eprintln|dbg)$"))
```

### .clone() Overuse
```
(method_call_expression
  method: (field_identifier) @method (#eq? @method "clone"))
```
Track count per function — high clone count suggests borrow checker fights.

### Missing ? on Result Returns
In functions returning `Result`, check for `match` on `Result` values that could use `?`:
```
(match_expression
  value: (call_expression)  ; that returns Result
  arms: (match_arm pattern: (tuple_pattern))  ; Ok/Err destructuring
)
```

### Unsafe Block
```
(unsafe_block) → flag with "must have safety comment" rule
```

### Mutex .lock().unwrap()
```
(method_call_expression
  method: (field_identifier) @method (#eq? @method "lock")
  (method_call_expression
    method: (field_identifier) @m2 (#eq? @m2 "unwrap")))
```

## 6. Quick Reference

| Smell | AST Signal | Detection |
|---|---|---|
| .unwrap() | `method_call_expression` `.unwrap()` | Name match |
| Vague .expect() | `.expect("short/vague")` | String pattern |
| println! | `macro_invocation` `println!` | Name match |
| .clone() overuse | `.clone()` count per function | Threshold |
| Missing ? | `match` on Result where `?` suffices | Pattern |
| unsafe block | `unsafe_block` node | Presence |
| lock().unwrap() | `lock` + `unwrap` chain | Sequence match |
| God function | Statement count in `function_item` | Threshold |
| Missing module | everything in single file | File structure |