# Code Smells Common in AI-Generated Code

Research compiled from developer experience, blog posts, and industry observations about patterns that AI coding agents frequently produce.

---

## 1. Structural Smells

### God Functions / Mega-Functions
- **What:** Single functions spanning 50-200+ lines that do everything — parse input, validate, transform, write output, handle errors
- **Why AI produces it:** AI generates code top-to-bottom in a single pass. It doesn't refactor into smaller units unless explicitly asked. Each "and then also..." gets appended to the current function
- **Detectability:** High — count statement nodes within function_declaration / method_declaration in tree-sitter AST

### Missing Abstractions
- **What:** Repeated inline logic instead of extracted functions/classes. Same 5-line block appears 3 times with minor variations
- **Why AI produces it:** AI doesn't see the big picture while generating. It solves each local problem independently, producing near-identical code in different locations
- **Detectability:** Medium — AST-level duplicate detection (subtree matching with parameterization)

### Flat Structure / Everything in One File
- **What:** 500-1000+ line files with no module boundaries. Related and unrelated code co-located
- **Why AI produces it:** AI tends to put new code in the current file rather than creating new modules. Creating files is a "meta" action most agents don't default to
- **Detectability:** High — line count per file, or AST node count per file

### Over-Engineering / Premature Abstraction
- **What:** Factory factories, strategy patterns for 2 variants, interface for a single implementation
- **Why AI produces it:** Training data over-represents enterprise Java/C# patterns. AI defaults to "proper" OOP even when a simple function suffices
- **Detectability:** Medium — ratio of abstract/interface nodes to concrete implementations, depth of inheritance hierarchies

### Shallow Delegation Chains
- **What:** A→B→C→D where each function just calls the next with minimal transformation
- **Why AI produces it:** AI decomposes by creating new functions for each "step" but doesn't merge trivial pass-throughs
- **Detectability:** Medium — function bodies that consist of a single call expression

---

## 2. Observability Smells

### Missing Logging
- **What:** Functions that perform I/O, state changes, or external calls with zero log statements
- **Why AI produces it:** Logging is "invisible" — it doesn't affect correctness, so AI omits it unless the prompt specifically requests it. Most code examples in training data omit logging for brevity
- **Detectability:** High — AST check: function with call expressions to I/O-adjacent APIs but no call to known logging functions

### Missing Metrics / Instrumentation
- **What:** No timing, counting, or success/failure tracking on operations that matter
- **Why AI produces it:** Same as logging — metrics are invisible to correctness. Training data almost never includes metrics instrumentation
- **Detectability:** High — similar to logging: check for I/O operations without corresponding metrics calls

### Swallowed Errors / Silent Failures
- **What:** `catch (e) { }` or `except: pass` — error caught and discarded with no logging, metrics, or re-raise
- **Why AI produces it:** AI wants to "handle all cases" but doesn't know what to do with unexpected errors. Empty catch is the path of least resistance
- **Detectability:** High — catch/except blocks with empty bodies or only a comment

### Missing Error Context
- **What:** `throw new Error("failed")` or `raise Exception("error")` with no information about what failed or why
- **Why AI produces it:** AI generates error handling as a pattern match ("this can fail → throw") without including diagnostic information
- **Detectability:** Medium — string literals in throw/raise that are very short or generic

---

## 3. Input Validation Smells

### No Input Validation
- **What:** Public functions that trust their parameters completely — no null checks, type checks, range checks, or format validation
- **Why AI produces it:** AI writes the "happy path" first and doesn't think about what happens with bad input unless prompted
- **Detectability:** Medium — public/exported function declarations with no conditional checks on parameters in the first N statements

### Hardcoded Values / Magic Numbers
- **What:** `if (retries > 3)`, `sleep(5000)`, `timeout = 30`, `"localhost:5432"` scattered inline
- **Why AI produces it:** AI picks reasonable-looking values and inlines them rather than extracting to named constants or config
- **Detectability:** High — numeric/string literals beyond small integers (0, 1, -1) or known constants (true/false/null)

### Missing Configuration Externalization
- **What:** Connection strings, URLs, timeouts, API keys embedded in code rather than environment variables or config files
- **Why AI produces it:** AI writes working code, not deployable code. "It works on my machine" is the AI's default context
- **Detectability:** High — string literals matching URL patterns, connection strings, or known config key patterns

---

## 4. Maintainability Smells

### Over-Commenting / Obvious Comments
- **What:** `// increment counter` above `counter++`, `// returns true if valid` above `function isValid()`
- **Why AI produces it:** AI has been RLHF'd to "be helpful" and explain itself. It treats comments as documentation requirements rather than tools for non-obvious information
- **Detectability:** Medium — comment-to-code ratio, or comments that paraphrase the next line

### Copy-Paste Duplication
- **What:** Identical or near-identical code blocks in multiple locations with minor variable name changes
- **Why AI produces it:** Each generation is independent. AI doesn't check "did I already write this?" before producing the same pattern again
- **Detectability:** High — AST-level duplicate subtree detection, or simpler: token-level sequence matching across files

### Dead Code / Unreachable Branches
- **What:** Variables assigned but never read, if-branches that can never be true, import statements for unused modules
- **Why AI produces it:** AI generates scaffolding and then doesn't clean up. Or it writes defensive code that's logically unreachable
- **Detectability:** High — unused variable analysis, unreachable code flow analysis

### Inconsistent Style Within a File
- **What:** Mix of arrow functions and function declarations, inconsistent naming (camelCase and snake_case in same file), mixed quote styles
- **Why AI produces it:** Different parts of the file may be generated in different sessions or contexts. AI doesn't always adopt the existing style
- **Detectability:** High — pattern consistency checks within a single file scope

---

## 5. Security Smells

### Hardcoded Secrets
- **What:** API keys, passwords, tokens directly in source code
- **Why AI produces it:** AI doesn't understand the difference between "working code" and "secure code." It includes credentials inline because that's what makes the code run
- **Detectability:** High — regex for known secret patterns (base64 strings, key patterns, password variable names)

### SQL Injection / String Interpolation in Queries
- **What:** `f"SELECT * FROM users WHERE id = {user_id}"` instead of parameterized queries
- **Why AI produces it:** String interpolation is the simplest way to compose a query. AI picks the path that "works" without considering injection risk
- **Detectability:** Medium — string concatenation/interpolation within SQL-adjacent call expressions

### Broad Exception Catching
- **What:** `catch (Exception e)` or `except Exception` catching everything instead of specific error types
- **Why AI produces it:** AI doesn't know which specific exceptions a library call can throw, so it catches the widest type
- **Detectability:** High — catch blocks referencing root exception types

### Disabled Security Features
- **What:** `verify=False` on HTTP clients, `--no-verify-ssl`, disabled certificate checks
- **Why AI produces it:** SSL errors are common in development. AI "fixes" them by disabling verification rather than properly configuring trust
- **Detectability:** High — known parameter names that disable security features

---

## 6. Testing Smells

### Missing Edge Case Tests
- **What:** Tests that only cover the happy path. Empty inputs, null values, extreme sizes — all untested
- **Why AI produces it:** AI generates tests that pass, not tests that find bugs. Happy path is the easiest to verify
- **Detectability:** Low-Medium — check test files for absence of null/empty/boundary test patterns

### Assertions on Implementation Details
- **What:** Tests that check internal state or method call order rather than observable behavior
- **Why AI produces it:** AI generates tests by "reading" the implementation rather than reasoning about contracts
- **Detectability:** Low — requires understanding test intent

### Test Mirrors Implementation
- **What:** Test code that's essentially a copy of the production code — same logic, same structure
- **Why AI produces it:** AI sees the implementation and writes tests that replicate it, providing no independent verification
- **Detectability:** Medium — similarity analysis between test and source code

---

## Summary: Highest-Value Detection Targets

Ranked by (prevalence × detectability × impact):

| Priority | Smell | Prevalence | Detectability | Impact |
|----------|-------|-----------|---------------|--------|
| 1 | Missing Logging | Very High | High | High |
| 2 | God Functions | Very High | High | High |
| 3 | Magic Values / Hardcoded Config | Very High | High | Medium |
| 4 | Missing Error Handling | High | High | High |
| 5 | Swallowed Errors | High | High | High |
| 6 | File Length / Flat Structure | High | High | Medium |
| 7 | Copy-Paste Duplication | High | High | Medium |
| 8 | Missing Input Validation | High | Medium | High |
| 9 | Over-Commenting | High | Medium | Low |
| 10 | Hardcoded Secrets | Medium | High | Very High |
| 11 | Broad Exception Catching | Medium | High | Medium |
| 12 | Disabled Security Features | Medium | High | Very High |
| 13 | Missing Metrics | High | High | Medium |
| 14 | Dead Code | Medium | High | Low |
| 15 | Over-Engineering | Medium | Medium | Medium |

---

_Compiled from developer experience and community observations, 2026-04-19_