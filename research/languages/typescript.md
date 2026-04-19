# TypeScript/JavaScript — AI Code Smells & Tree-Sitter Detection

## 1. Common AI-Generated TS/JS Anti-Patterns

### Type Safety
- **`any` type** — AI's escape hatch when it can't determine a type. Most common TS smell
- **`as any` casts** — forcing types to bypass type checking
- **`@ts-ignore` / `@ts-expect-error`** — suppressing errors instead of fixing them
- **Missing return types** — functions with no `: Type` on return
- **`unknown` overuse** — opposite problem, too conservative
- **Non-null assertions** — `user!.name` instead of proper null checking

### Error Handling
- **Empty catch blocks** — `catch (e) {}` — the JS equivalent of Python's bare except
- **Catch without type** — `catch (e)` with `e: any` (or no annotation)
- **Missing try/catch on async** — `await` calls with no error boundary
- **Unhandled promise rejections** — `.then()` without `.catch()`
- **`console.log` in catch** — not a real error handling strategy

### Logging & Observability
- **`console.log` instead of proper logger** — the #1 AI JS smell. Debug output in production
- **No structured logging** — `console.log("error:", err)` instead of `logger.error({ err })`
- **Missing request tracing** — no correlation IDs, no request logging
- **`console.log` left in production** — AI writes it for debugging, never removes it

### Async Patterns
- **`async` without `await`** — functions marked async that don't await anything
- **`await` in loops** — sequential processing instead of `Promise.all()`
- **Missing error boundaries in React** — no ErrorBoundary wrapping
- **Race conditions** — multiple async operations modifying shared state

### Structural Smells
- **God components** — React components doing data fetching, transformation, and rendering
- **Prop drilling** — passing props through 5+ levels instead of context
- **Huge useEffect** — dependency arrays with 10+ items
- **Missing cleanup** — useEffect without return cleanup function
- **Inline styles** — `style={{...}}` everywhere instead of CSS modules or styled-components

## 2. npm Ecosystem — AI Pitfalls

- **Dependency bloat** — AI adds packages it "knows" without checking if they're needed
- **Mixed package managers** — both `package-lock.json` and `yarn.lock` in repo
- **No `.nvmrc`** — Node version not pinned
- **Dev deps in production** — `typescript`, `eslint` in `dependencies` instead of `devDependencies`
- **Outdated deps** — AI generates code for API versions that don't match installed packages
- **No `engines` field** — Node version requirements not specified
- **Missing scripts** — no `lint`, `test`, `build` scripts in package.json

## 3. Tree-Sitter TypeScript AST — Key Node Types

| Node Type | What It Captures | Use For |
|---|---|---|
| `function_declaration` | `function foo()` | God function detection |
| `arrow_function` | `() => {}` | Arrow function size, async detection |
| `class_declaration` | `class Foo` | God class detection |
| `interface_declaration` | `interface Foo` | Type coverage |
| `type_alias_declaration` | `type Foo =` | Type coverage |
| `try_statement` | `try { }` | Error handling presence |
| `catch_clause` | `catch (e)` | Empty catch, broad catch |
| `call_expression` | `foo()` | console.log, logging detection |
| `member_expression` | `obj.prop` | Method call patterns |
| `async_function` / `async_...` | `async` modifier | Async patterns |
| `await_expression` | `await x` | Missing await detection |
| `jsx_element` | `<Component>` | React component detection |
| `jsx_self_closing_element` | `<Comp />` | Component usage |
| `decorator` | `@decorator` | NestJS decorators |
| `import_statement` | `import x from` | Import analysis |
| `export_statement` | `export` | Public API surface |

## 4. Framework-Specific AI Issues

### React / Next.js
- **No Error Boundaries** — unhandled errors crash the entire UI
- **Client components that should be server** — `'use client'` everywhere in Next.js 13+
- **Missing loading states** — no Suspense boundaries or loading indicators
- **useEffect for data fetching** — instead of React Query / SWR / server components
- **Missing dependency arrays** — `useEffect(() => { ... })` without deps
- **Stale closures** — useEffect capturing stale state values
- **No memo on expensive renders** — missing `React.memo`, `useMemo`, `useCallback`
- **Inline object/array props** — `style={{ color: 'red' }}` causing re-renders
- **God components** — one component doing fetch + transform + render

### Express / NestJS
- **No error middleware** — missing error handler at end of middleware chain
- **No request validation** — accepting `req.body` without schema validation
- **Missing `next(error)`** — errors not propagated to error handler
- **No rate limiting** — endpoints exposed without throttling
- **Synchronous operations** — blocking the event loop

### Node.js General
- **No graceful shutdown** — `process.exit()` instead of SIGTERM handlers
- **Uncaught exception handling** — missing `process.on('uncaughtException')`
- **No health check endpoint** — `/health` or `/readiness` absent

## 5. Detection Rules — Tree-Sitter Queries

### Empty Catch Block
```
(catch_clause body: (statement_block . (_)))  → HAS content
(catch_clause body: (statement_block))         → empty = smell
```

### console.log Usage
```
(call_expression
  function: (member_expression
    object: (identifier) @obj (#eq? @obj "console")
    property: (property_identifier) @method
    (#match? @method "^(log|debug|info|warn|error)$")))
```

### `any` Type Usage
```
(predefined_type) @type (#eq? @type "any")
(type_assertion expression: (_) type: (predefined_type) @t (#eq? @t "any"))
```

### Async Without Await
```
(function_declaration (modifier "async") body: (statement_block !contains await_expression))
```

### Missing Return Type
```
(function_declaration name: (_) parameters: (_) body: (_)) 
  → no type_annotation after parameters = missing return type
```

## 6. Quick Reference

| Smell | AST Signal | Detection |
|---|---|---|
| `any` type | `predefined_type "any"` | Exact match |
| Empty catch | `catch_clause` → empty `statement_block` | Body empty |
| console.log | `call_expression` to `console.log|debug|...` | Name match |
| Async no await | `async` function without `await_expression` | Absence |
| Missing return type | `function_declaration` without type after params | Absence |
| God component | JSX element count in single `function_declaration` | Threshold |
| Missing deps | `call_expression` with `useEffect` + analysis | Pattern |
| Unhandled promise | `.then()` without `.catch()` | Absence |