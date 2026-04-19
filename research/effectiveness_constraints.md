# Codebase Constraints That Reduce AI Effectiveness

Research on patterns and structural issues in codebases that make AI coding agents produce worse output or operate less effectively.

---

## 1. Context Window Exhaustion

### Large Files / Monoliths
- **What:** Files exceeding the effective context window of the AI agent (typically 2K-8K lines depending on model)
- **Why it hurts AI:** The agent can't see the full file, so it generates code that contradicts unseen parts, duplicates logic that exists below the fold, or misses imports/conventions established earlier in the file
- **Detection:** Line count per file, or AST node count

### Deeply Nested Directories
- **What:** Projects with 8+ levels of directory nesting
- **Why it hurts AI:** The agent burns context navigating the file tree. By the time it finds relevant code, it has less room to reason about it
- **Detection:** Directory depth in the project tree

### Missing Entry Points
- **What:** No clear `main`, `index`, or `app` entry point that establishes the dependency graph
- **Why it hurts AI:** The agent has to guess where to start reading. It may miss critical initialization, middleware, or configuration that shapes runtime behavior
- **Detection:** Absence of well-known entry point files (`main.*`, `index.*`, `app.*`, `server.*`)

---

## 2. Implicit Conventions

### Unstated Architectural Patterns
- **What:** The codebase follows patterns (repository pattern, service layer, CQRS) but never documents them. The pattern is visible in the code but never named
- **Why it hurts AI:** AI replicates the surface syntax but not the pattern. It creates a new class that looks like a service but doesn't follow the dependency injection convention, or adds a route that skips the middleware chain
- **Detection:** Consistency checks — do similar constructs follow the same structure? (e.g., all controllers have the same method signatures, all services follow the same injection pattern)

### Magic Strings / Numbers as Conventions
- **What:** `"admin"`, `"system"`, `403`, `"INTERNAL"` used as implicit contracts between modules
- **Why it hurts AI:** The agent doesn't know these are load-bearing. It changes a string literal and breaks a cross-cutting contract
- **Detection:** Repeated string/number literals across files that aren't defined as constants

### Environment-Dependent Behavior
- **What:** Code that behaves differently based on `NODE_ENV`, feature flags, or environment variables with no type hints or documentation
- **Why it hurts AI:** The agent tests against one environment and is surprised when the other path executes in production
- **Detection:** References to `process.env`, `os.environ`, `System.getenv` without corresponding type annotations or default values

---

## 3. Poor Modularity Boundaries

### Circular Dependencies
- **What:** Module A imports B, B imports A (directly or transitively)
- **Why it hurts AI:** AI can't determine a "correct" order for reading or modifying these files. Changes cascade unpredictably
- **Detection:** Dependency graph cycle detection

### God Modules / God Classes
- **What:** A single module/class handling multiple unrelated responsibilities (e.g., a `UserService` that also sends email and generates reports)
- **Why it hurts AI:** The agent can't determine the scope of impact when modifying this module. It treats it as one big coupled unit
- **Detection:** Number of distinct responsibility clusters in a module (method naming patterns, import diversity)

### Missing Interface Boundaries
- **What:** Direct implementation coupling instead of programming to interfaces. Consumers reach into internals of other modules
- **Why it hurts AI:** The agent can't determine what's public API vs. internal detail. It may modify internals that other modules depend on
- **Detection:** Access patterns — are private/internal members being accessed from other modules?

---

## 4. Insufficient Type Information

### Dynamic Typing Without Annotations
- **What:** Python, Ruby, JS codebases with no type hints. Function signatures are `def process(data)` instead of `def process(data: Order) -> Result`
- **Why it hurts AI:** The agent has to infer types from usage, which is error-prone. It may generate code that passes wrong types, calls non-existent methods, or misses required fields
- **Detection:** Functions without type annotations in languages that support them (Python, TypeScript, PHP)

### Undocumented Return Types
- **What:** Functions that return different types depending on input, or return `Any`/`object`/`interface{}`
- **Why it hurts AI:** The agent can't determine what to do with the result. It may chain methods that don't exist on the actual return type
- **Detection:** Return type annotations that are `Any`, `object`, `interface{}`, or absent entirely

### Unstructured Data Bags
- **What:** Dictionaries, maps, or objects used as ad-hoc data structures with no defined schema. `data["user"]["settings"]["theme"]`
- **Why it hurts AI:** The agent has to guess what keys exist, what types values are, and what's required vs. optional. It often guesses wrong
- **Detection:** Chained dictionary/map access patterns, especially with string literal keys

---

## 5. Missing or Misleading Documentation

### Stale Comments
- **What:** Comments that describe what the code *used* to do, not what it does now
- **Why it hurts AI:** The agent trusts comments as truth and generates code consistent with the comment, not the actual implementation. This is worse than no comments
- **Detection:** Mismatch between comment content and code behavior (hard to detect automatically, but comment age vs. code modification date is a proxy)

### Missing README / Architecture Docs
- **What:** No top-level documentation explaining project structure, conventions, or how to get started
- **Why it hurts AI:** The agent has no orientation. It starts making assumptions about project organization that may be wrong
- **Detection:** Absence of README.md, CONTRIBUTING.md, ARCHITECTURE.md, or similar

### Undocumented Side Effects
- **What:** Functions that modify global state, write to databases, send network requests, or have other side effects not visible from their signature
- **Why it hurts AI:** The agent calls a function thinking it's pure, triggering unexpected side effects. Or it fails to call a function that has necessary side effects
- **Detection:** Functions that call I/O operations but have "pure" sounding names (e.g., `getUser`, `validateOrder`)

---

## 6. Inconsistent Code Style

### Mixed Paradigms
- **What:** Some files use OOP, others functional. Some use async/await, others callbacks. Some use classes, others use closures
- **Why it hurts AI:** The agent doesn't know which paradigm to follow for new code. It picks whatever feels natural, creating inconsistency
- **Detection:** Paradigm-specific constructs in the same project (class vs. function exports, async/await vs. .then() chains)

### Inconsistent Naming
- **What:** Some variables use camelCase, others snake_case. Some use abbreviations, others full words. `userId`, `user_id`, `uid`, `usrId` all coexist
- **Why it hurts AI:** The agent can't determine the naming convention and produces new names that don't match any existing pattern
- **Detection:** Variable/identifier naming pattern analysis within a single scope

---

## 7. Testing Gaps

### Missing Integration Tests
- **What:** Unit tests exist but nothing tests the connections between modules
- **Why it hurts AI:** The agent can make changes that pass unit tests but break integrations. It has no signal that its changes broke cross-cutting behavior
- **Detection:** Test file structure — presence of only unit-level test patterns (mocking, single-module imports)

### No Test Infrastructure
- **What:** Project has no test framework, no test directory, no CI configuration
- **Why it hurts AI:** The agent can't verify its changes. It writes code that "looks right" but has no automated feedback loop
- **Detection:** Absence of test directories, test framework dependencies, CI configuration files

---

## Summary: Highest-Impact Constraints

| Priority | Constraint | AI Impact | Detectability | Fix Difficulty |
|----------|-----------|-----------|----------------|----------------|
| 1 | Large files / monoliths | Very High | High | Medium |
| 2 | Missing type annotations | Very High | High | Medium |
| 3 | Unstated architectural patterns | High | Medium | Hard |
| 4 | Missing entry points | High | High | Easy |
| 5 | Undocumented side effects | High | Medium | Medium |
| 6 | Circular dependencies | High | High | Hard |
| 7 | Missing README/architecture docs | Medium | High | Easy |
| 8 | Magic strings as conventions | Medium | Medium | Easy |
| 9 | Mixed paradigms | Medium | Medium | Hard |
| 10 | Missing integration tests | Medium | Medium | Medium |

---

_Compiled 2026-04-19_