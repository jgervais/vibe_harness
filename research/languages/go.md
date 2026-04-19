# Go — AI Code Smells & Tree-Sitter Detection

## 1. Common AI-Generated Go Anti-Patterns

### Error Handling
- **Ignored errors** — `doSomething()` without checking the returned error. The #1 Go smell
- **`_ = doSomething()`** — explicitly discarding the error return value
- **`log.Fatal()` instead of returning error** — crashes the program instead of propagating
- **Panic for control flow** — `panic("not found")` instead of returning an error
- **Error wrapping missing** — `return err` instead of `return fmt.Errorf("doing X: %w", err)`

### Context & Cancellation
- **Missing `context.Context`** — function signatures without ctx parameter for I/O operations
- **`context.Background()` in handlers** — should use the request context
- **No context timeout** — HTTP calls without `context.WithTimeout`
- **Ignoring context cancellation** — long-running loops not checking `ctx.Done()`

### Concurrency
- **Goroutine leaks** — goroutines started but never guaranteed to exit
- **Missing `defer close()`** — channels/resources opened but not closed
- **Race conditions** — shared state accessed without synchronization
- **Unbuffered channels in hot paths** — potential deadlock patterns

### Logging & Observability
- **`fmt.Println()` instead of structured logging** — the Go equivalent of console.log
- **`log.Printf()` instead of `slog`** — using stdlib log instead of structured logger
- **No request tracing** — missing trace IDs in HTTP handlers
- **Missing metrics** — no prometheus instrumentation on handlers

### Structural Smells
- **God functions** — single `func` doing everything, common in `main()`
- **Package-level state** — `var cache = map[string]string{}` without sync protection
- **Huge interfaces** — interfaces with 10+ methods (Go favors small interfaces)
- **Returning structs instead of interfaces** — violates Go's implicit interface idiom

## 2. Go Modules — AI Pitfalls

- **No `go.mod`** — AI writes standalone scripts without module initialization
- **Wrong module path** — `module myapp` instead of `module github.com/user/myapp`
- **Indirect dependencies** — deps appearing in `go.mod` that aren't directly imported
- **Missing `go.sum`** — or committing only `go.mod`
- **No `vendor` directory** — when reproducible builds matter
- **Ignoring `cmd/` layout** — AI puts everything in root instead of `cmd/server/main.go`

## 3. Tree-Sitter Go AST — Key Node Types

| Node Type | What It Captures | Use For |
|---|---|---|
| `function_declaration` | `func foo()` | God function detection |
| `method_declaration` | `func (r *Repo) foo()` | Method receiver analysis |
| `call_expression` | `doSomething()` | Ignored error detection |
| `assignment_statement` | `x, y := foo()` | Multi-return value tracking |
| `if_statement` | `if err != nil` | Error handling patterns |
| `return_statement` | `return err` | Error propagation |
| `go_statement` | `go func()` | Goroutine detection |
| `defer_statement` | `defer close()` | Resource cleanup |
| `interface_declaration` | `type Foo interface` | Interface size |
| `type_declaration` | `type Foo struct` | Struct definition |
| `import_declaration` | `import "fmt"` | Import analysis |
| `package_clause` | `package main` | Package identification |
| `selector_expression` | `fmt.Println` | Logging call detection |
| `func_literal` | `func() { }` | Anonymous functions |
| `communication_statement` | `ch <- x` | Channel operations |
| `for_statement` | `for { }` | Infinite loops without ctx check |

## 4. Framework-Specific AI Issues

### net/http (Standard Library)
- **Missing `WriteHeader`** — setting status after writing body
- **No timeout on `http.Client`** — `http.DefaultClient` has no timeouts
- **Missing `Close()` on response body** — `defer resp.Body.Close()` absent
- **No context in HTTP requests** — `http.NewRequest` instead of `http.NewRequestWithContext`

### Gin / Echo
- **No middleware for logging/metrics** — endpoints without observability middleware
- **Missing error response** — handlers that don't write response on error paths
- **Global state in handlers** — package-level variables instead of dependency injection

### gRPC
- **Missing context in RPC methods** — not passing ctx through
- **No error codes** — returning generic errors instead of status codes
- **Missing deadlines** — no `grpc.WithTimeout` on client calls

## 5. Detection Rules — Tree-Sitter Queries

### Ignored Error Return
The most important rule for Go. When a function call returns multiple values and the last is an `error` type, all return values must be captured or explicitly discarded with `_`.

```
(call_expression) where result is not assigned and last return type is error
(assignment_statement left: (expression_list len < right_return_count))
```

Practically: detect `call_expression` that is not the RHS of an assignment and is known to return error.

### fmt.Println Instead of Logging
```
(call_expression
  function: (selector_expression
    operand: (identifier) @pkg (#eq? @pkg "fmt")
    field: (field_identifier) @method
    (#match? @method "^(Println|Printf|Sprintf)$")))
```

### Missing Context Parameter
```
(function_declaration
  parameters: (parameter_list)
  body: (contains call_expression to known I/O))
  → no `context.Context` in first parameter position
```

### Panic Instead of Error
```
(call_expression
  function: (identifier) @name (#eq? @name "panic"))
```

### log.Fatal In Library Code
```
(call_expression
  function: (selector_expression
    operand: (identifier) @pkg (#eq? @pkg "log")
    field: (field_identifier) @method
    (#match? @method "^(Fatal|Fatalf|Panic|Panicf)$")))
```
Only acceptable in `main()`, not in library packages.

## 6. Quick Reference

| Smell | AST Signal | Detection |
|---|---|---|
| Ignored error | `call_expression` without assignment | Return value tracking |
| fmt.Println | `selector_expression` `fmt.Println` | Name match |
| Missing context | I/O function without ctx param | Absence + pattern |
| Panic usage | `call_expression` to `panic` | Exact match |
| log.Fatal in lib | `log.Fatal` outside `main` package | Package + name |
| God function | Statement count in `function_declaration` | Threshold |
| Missing defer close | Resource creation without `defer` | Absence |
| Goroutine leak | `go_statement` without exit condition | Pattern |