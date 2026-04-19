# Generic Checks

Checks that apply to all source code files regardless of language. These rules require no AST parsing — they operate on raw file content, line counts, and structural patterns detectable with text analysis.

---

## VH-G001: File Length

**Rule:** A source file must not exceed a fixed maximum number of lines.

**Threshold:** 300 lines (non-blank, non-comment)

**Rationale:** Long files indicate poor modularity and exceed the effective context window of most AI agents. When a file is too long, the AI generating code against it cannot see the full picture and will produce code that contradicts, duplicates, or conflicts with existing content.

**Detection:** Count non-blank, non-comment lines per file. Flag files exceeding the threshold.

**Non-configurable:** The threshold is fixed. You cannot raise it, lower it, or exempt files.

---

## VH-G002: Function Length

**Rule:** A single function or method must not exceed a fixed maximum number of lines.

**Threshold:** 50 lines (non-blank, non-comment)

**Rationale:** Long functions indicate mixed responsibilities. AI agents generate functions top-to-bottom without refactoring. Each "and then also" gets appended, producing god functions that are hard to test, hard to debug, and hard to understand.

**Detection:** Requires AST parsing to identify function/method boundaries. Count non-blank, non-comment lines within the function body.

**Non-configurable:** The threshold is fixed.

---

## VH-G003: Missing Logging in I/O Functions

**Rule:** Functions that perform I/O operations (network calls, file reads/writes, database queries) must contain at least one logging call.

**Rationale:** AI-generated code optimizes for correctness on the happy path. Logging is invisible to correctness, so it gets omitted. In production, unlogged I/O operations become blind spots — failures happen silently, and there is no way to diagnose them.

**Detection:** Requires AST parsing. Identify functions containing call expressions to known I/O APIs (HTTP clients, file operations, database drivers). Check whether the same function body contains any call to a recognized logging function.

**Configurable hints:** The `.vibe_harness.toml` file specifies which function names constitute logging calls in your codebase. This does not change where logging is required — only how the tool recognizes it.

```toml
[observability]
logging_calls = ["log", "logger", "logging", "tracing", "slog"]
```

**Non-configurable:** You cannot disable this rule. You cannot exempt functions. You cannot lower the I/O threshold.

---

## VH-G004: Swallowed Errors

**Rule:** Catch/except blocks must not have empty bodies.

**Rationale:** AI agents generate empty catch blocks as the path of least resistance — "handle all cases" but with no idea what to do with unexpected errors. Empty catch blocks make failures invisible. The error is caught and discarded, leaving no trace.

**Detection:** Requires AST parsing. Identify catch/except blocks whose body contains zero statements (comments do not count as statements).

**Non-configurable:** You cannot disable this rule. You cannot add exceptions.

---

## VH-G005: Hardcoded Secrets

**Rule:** Source files must not contain string literals matching known secret patterns.

**Patterns detected:**
- AWS access keys (`AKIA...`)
- AWS secret keys (40-char base64 strings in assignment context)
- Generic API key patterns (`api_key = "..."`, `apikey = "..."`)
- Database connection strings with embedded credentials (`postgres://user:pass@...`)
- Bearer tokens (`Bearer ey...`)
- Private key markers (`-----BEGIN PRIVATE KEY-----`)

**Rationale:** AI agents hardcode credentials because that makes code "work." They don't understand the difference between working code and secure code.

**Detection:** Regex pattern matching on string literals and assignment statements. Not AST-dependent — works on raw text.

**Non-configurable:** You cannot disable this rule. You cannot mark patterns as false positives globally.

---

## VH-G006: Magic Values

**Rule:** Numeric literals and string literals beyond common small values must be assigned to named constants rather than used inline.

**Allowed inline values (not flagged):**
- Integers: `0`, `1`, `-1`, `2`
- Booleans: `true`, `false`
- Null values: `null`, `nil`, `None`, `undefined`
- Empty collections: `[]`, `{}`, `""`

**Threshold:** Any other numeric or string literal used inline more than once in a file, or any literal exceeding 20 characters used inline.

**Rationale:** AI agents inline reasonable-looking values (`retries = 3`, `timeout = 30`, `"localhost:5432"`) instead of extracting them to named constants or configuration. This makes values hard to change, hard to find, and hard to understand the intent behind.

**Detection:** Scan for numeric and string literals in expression contexts (not constant declarations). Flag values not in the allowed list.

**Non-configurable:** The allowed values list is fixed. You cannot add to it.

---

## VH-G007: Copy-Paste Duplication

**Rule:** Identical or near-identical code blocks (6+ lines) must not appear in multiple locations within the same repository.

**Threshold:** 6 lines, 80% token similarity

**Rationale:** AI agents generate code independently for each problem. They don't check whether similar logic already exists elsewhere. The same 8-line block appears three times with minor variable name changes, creating a maintenance burden and inconsistency risk.

**Detection:** Token-level sequence matching across files within the repository. Normalize variable names before comparison to catch "same logic, different names" patterns.

**Non-configurable:** The threshold is fixed. You cannot exempt files or directories.

---

## VH-G008: Comment-to-Code Ratio

**Rule:** Files with excessive comment density (more than 1 comment line per 3 code lines) are flagged.

**Rationale:** AI agents over-comment obvious code as a side effect of RLHF training to "be helpful." Comments like `// increment counter` above `counter++` add noise without information. Real documentation is valuable; obvious paraphrasing is not.

**Detection:** Count comment lines vs. non-blank code lines per file. Flag files where the ratio exceeds the threshold.

**Non-configurable:** The ratio is fixed.

---

## VH-G009: Missing Error Handling on I/O

**Rule:** I/O operations (network calls, file operations, database queries) must be wrapped in error handling constructs (try/catch, error return checking, etc.).

**Rationale:** AI agents write the happy path. I/O operations that fail without error handling cause unhandled exceptions, silent data loss, or undefined behavior in production.

**Detection:** Requires AST parsing. Identify call expressions to known I/O APIs. Check whether they are contained within try/catch blocks, followed by error checks, or use error-propagation patterns (`?` in Rust, `throws` in Java).

**Non-configurable:** You cannot disable this rule. You cannot exempt specific I/O calls.

---

## VH-G010: Broad Exception Catching

**Rule:** Catch/except blocks that catch root exception types (Exception, Error, Throwable, any) are flagged.

**Language-specific root types:**
- Python: `Exception`, `BaseException`
- Java: `Exception`, `Throwable`, `RuntimeException`
- TypeScript: no type annotation on catch
- Go: not applicable (uses error returns)
- Ruby: `rescue` without type (catches StandardError), `rescue Exception`
- Rust: not applicable (uses Result)
- C#: `Exception`

**Rationale:** AI agents catch the widest exception type because they don't know which specific exceptions a library can throw. This catches more than intended, masks bugs, and makes error handling unpredictable.

**Detection:** Requires AST parsing. Identify catch/except clauses and check the exception type against the root types list.

**Non-configurable:** The root types list is fixed per language.

---

## VH-G011: Disabled Security Features

**Rule:** Known parameters that disable security features must not be set to their "off" values.

**Patterns detected:**
- `verify=False` (Python requests, SSL verification)
- `InsecureSkipVerify: true` (Go TLS)
- `rejectUnauthorized: false` (Node.js TLS)
- `--no-verify-ssl` (CLI flags)
- `ssl_verify: false` (Ruby, various clients)
- `CURLOPT_SSL_VERIFYPEER: false` (PHP cURL)

**Rationale:** AI agents encounter SSL errors in development environments and "fix" them by disabling verification rather than properly configuring trust. This works locally and fails in production — or worse, works in production with no encryption validation.

**Detection:** Regex/AST pattern matching on known parameter names and their values.

**Non-configurable:** The pattern list is fixed.

---

## VH-G012: God Module — Too Many Exports

**Rule:** A single file/module must not export more than a fixed number of public symbols (functions, classes, constants).

**Threshold:** 20 public exports

**Rationale:** Files with dozens of exports have no single responsibility. They are dumping grounds. AI agents add new code to existing files rather than creating new modules, and files grow without bound.

**Detection:** Count exported/public symbols per file. Requires language-specific knowledge of what constitutes "public" (export statements, public modifier, etc.).

**Non-configurable:** The threshold is fixed.

---

## Summary

| ID | Rule | Requires AST | Threshold |
|----|------|-------------|-----------|
| VH-G001 | File Length | No | 300 lines |
| VH-G002 | Function Length | Yes | 50 lines |
| VH-G003 | Missing Logging in I/O | Yes | 1 log call per I/O function |
| VH-G004 | Swallowed Errors | Yes | 0 empty catch bodies |
| VH-G005 | Hardcoded Secrets | No | Pattern match |
| VH-G006 | Magic Values | No | Inline literal detection |
| VH-G007 | Copy-Paste Duplication | No | 6 lines, 80% similarity |
| VH-G008 | Comment-to-Code Ratio | No | 1:3 ratio |
| VH-G009 | Missing Error Handling on I/O | Yes | All I/O wrapped |
| VH-G010 | Broad Exception Catching | Yes | No root type catches |
| VH-G011 | Disabled Security Features | No | Pattern match |
| VH-G012 | God Module — Too Many Exports | Yes | 20 public exports |