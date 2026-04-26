# Feature Specification: AST-Powered Checks

**Feature Branch**: `002-ast-powered-checks`  
**Created**: 2026-04-19  
**Status**: Draft  
**Input**: User description: "Add tree-sitter parsing for 6 AST-dependent generic checks (VH-G002, G003, G004, G009, G010, G012) as defined in Phase 2 of docs/roadmap.md"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run AST checks on a codebase (Priority: P1)

A developer runs vibe-harness against a source tree containing Python, TypeScript, Go, Java, Ruby, or Rust files. The tool parses each supported file into an AST, identifies structural code quality issues (long functions, swallowed errors, missing logging, broad catches, unhandled I/O errors, god modules), and reports violations alongside existing non-AST checks.

**Why this priority**: This is the core value proposition — without AST parsing and check execution, no AST-dependent violations can be reported. All other stories depend on this working.

**Independent Test**: Can be fully tested by scanning a directory with known-violating files in each supported language and verifying that all 6 AST checks produce correct violations, while existing non-AST checks continue working unchanged.

**Acceptance Scenarios**:

1. **Given** a directory containing Python files with functions exceeding the statement threshold, **When** the user runs vibe-harness, **Then** VH-G002 violations are reported with the correct file, line, function name, and statement count.
2. **Given** a directory containing Java files with empty catch blocks, **When** the user runs vibe-harness, **Then** VH-G004 violations are reported identifying each swallowed error location.
3. **Given** a directory containing files in an unsupported language (e.g., C++), **When** the user runs vibe-harness, **Then** AST checks are skipped for those files and non-AST checks still run normally.
4. **Given** a file that cannot be parsed (syntax error), **When** the user runs vibe-harness, **Then** the tool gracefully skips AST checks for that file and continues processing others without crashing.

---

### User Story 2 - View AST check results in output formats (Priority: P2)

A developer or CI system receives AST check violations through the existing output formats (human-readable text, JSON, SARIF) with the same structure and consistency as non-AST check results.

**Why this priority**: Output integration is essential for the checks to be actionable, but depends on the checks actually running first.

**Independent Test**: Can be tested by running vibe-harness with `--format json` and `--format sarif` on violating files and verifying that AST check violations appear in the structured output with correct rule metadata.

**Acceptance Scenarios**:

1. **Given** AST check violations exist, **When** the user runs with `--format human`, **Then** violations are displayed in the standard `<path>:<line>:<RuleID> — <Message>` format.
2. **Given** AST check violations exist, **When** the user runs with `--format json`, **Then** violations appear in the `results` array with `rule_id`, `file`, `line`, `column`, `end_line`, `message`, and `severity` fields.
3. **Given** AST check violations exist, **When** the user runs with `--format sarif`, **Then** violations appear in the SARIF `results` array and the 6 AST rules are listed in `tool.driver.rules`.

---

### User Story 3 - Configure observability recognition hints (Priority: P3)

A team maintains a `.vibe_harness.toml` config file with custom logging and metrics call names specific to their framework (e.g., Spring Boot's `LoggerFactory`, Rails' `Rails.logger`). The AST checks use these hints to correctly identify logging and metrics calls when evaluating VH-G003 (missing logging in I/O) and related rules.

**Why this priority**: Recognition hints make checks more accurate for real-world frameworks, but the checks function with defaults for common cases.

**Independent Test**: Can be tested by creating a config file with custom `logging_calls` and `metrics_calls`, scanning code that uses those custom names, and verifying that VH-G003 correctly recognizes them as logging calls.

**Acceptance Scenarios**:

1. **Given** a config file specifying `logging_calls = ["LoggerFactory", "LOG"]` and a Java file that uses `LoggerFactory.getLogger()` near an I/O call, **When** the user runs vibe-harness, **Then** VH-G003 does not flag the I/O call as missing logging.
2. **Given** no config file, **When** the user runs vibe-harness on a file using standard `log` calls, **Then** the default hint set correctly identifies `log` as a logging call.

---

### User Story 4 - Understand per-language AST patterns (Priority: P4)

A developer wants assurance that each supported language has appropriate query mappings. For example, Go's error-return pattern, Python's bare `except`, and Rust's `Result` handling are each addressed with language-specific queries that map to the same generic rule concepts.

**Why this priority**: Language-specific correctness is important for production quality but the generic rule concepts and query system are the prerequisite.

**Independent Test**: Can be tested by running each check against fixture files per language and verifying that language-specific patterns are correctly detected (e.g., Go `if err != nil { return err }` is not flagged by G004, Python bare `except:` is flagged by G004).

**Acceptance Scenarios**:

1. **Given** a Go file with `if err != nil { return err }`, **When** VH-G004 runs, **Then** this is not flagged as a swallowed error because the error is explicitly returned.
2. **Given** a Python file with `except:` (bare except), **When** VH-G004 runs, **Then** this is flagged as a broad exception catch by both G004 and G010 appropriately.
3. **Given** a Rust file with `.unwrap()` on a `Result`, **When** VH-G009 runs, **Then** it is flagged as missing error handling on I/O.

---

### Edge Cases

- What happens when a file has a parse error (syntax error)? The tool should skip AST checks for that file and continue processing other files.
- What happens when a recognized language grammar fails to load? The tool should log a warning and skip AST checks for that language while continuing with other languages.
- What happens for files that match a language extension but contain content in another language (e.g., HTML with embedded JS)? Only the primary language's grammar should be applied; mixed-content handling is out of scope.
- What happens when a function body contains only comments (no statements)? It should be treated as having zero statements for VH-G002.
- What happens when an I/O call is nested inside a lambda or closure that has error handling? The check should recognize the error-handling construct at the same scope level.
- What happens when a file has zero public exports? VH-G012 should not flag it (a file with no exports is not a "god module").

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST parse source files into ASTs using tree-sitter for Python, TypeScript, Go, Java, Ruby, and Rust.
- **FR-002**: The tool MUST detect the programming language from file extensions (matching the existing `[languages]` config mapping) and apply the appropriate grammar.
- **FR-003**: The tool MUST handle parse errors gracefully — when a file cannot be parsed, AST-dependent checks are skipped for that file and non-AST checks continue to run.
- **FR-004**: VH-G002 (Function Length) MUST count statements in function/method bodies and flag functions exceeding the configured or default threshold.
- **FR-005**: VH-G003 (Missing Logging in I/O) MUST identify I/O function calls (file reads/writes, network calls, database queries) and check whether logging calls exist in the same scope, flagging I/O calls without accompanying logging.
- **FR-006**: VH-G004 (Swallowed Errors) MUST detect empty or comment-only catch/except blocks where the caught error is not re-raised, returned, logged, or handled.
- **FR-007**: VH-G009 (Missing Error Handling on I/O) MUST find I/O function calls that are not wrapped in error-handling constructs (try/catch, if err != nil, Result handling, etc.) appropriate to the language.
- **FR-008**: VH-G010 (Broad Exception Catching) MUST identify catch/except blocks that catch root-level exception types (e.g., `Exception` in Java/Python, `rescue Exception` in Ruby, `catch (...)` in TypeScript) rather than specific subtypes.
- **FR-009**: VH-G012 (God Module) MUST count public exports per file and flag files exceeding the configured or default threshold.
- **FR-010**: The tool MUST define tree-sitter queries per language per rule, mapping generic rule concepts to language-specific AST patterns (e.g., Go error returns, Rust Result, Python bare except).
- **FR-011**: The tool MUST use the `[observability].logging_calls` and `[observability].metrics_calls` config values as recognition hints when evaluating VH-G003 and related rules.
- **FR-012**: The tool MUST provide default observability hint sets for common frameworks (Spring Boot, Rails, FastAPI, etc.) that apply when no user configuration is provided.
- **FR-013**: The tool MUST bundle language grammars (Python, TypeScript, Go, Java, Ruby, Rust) with the binary so no external grammar fetching is required at runtime.
- **FR-014**: The tool MUST register all 6 AST checks in the rule metadata (`rules.Checks()`) with `RequiresAST: true`.
- **FR-015**: The tool MUST register all 6 AST checks in the scanner so they execute alongside existing non-AST checks during a scan.
- **FR-016**: AST check violations MUST follow the same `rules.Violation` structure and appear in all output formats (human, JSON, SARIF) identically to non-AST check violations.
- **FR-017**: When a language is not supported by tree-sitter grammars, AST checks MUST be silently skipped for files of that language while non-AST checks still execute.

### Key Entities

- **AST Parser**: Parses source files into tree-sitter ASTs; manages grammar loading and language detection; handles parse errors gracefully.
- **Tree-sitter Query**: A pattern definition that matches specific AST node structures for a given rule and language; maps generic rule concepts to language-specific node types.
- **Recognition Hint Set**: A collection of function/method name patterns used to identify logging calls and metrics calls in code; sourced from user config or built-in defaults per framework.
- **AST Check**: A check that requires a parsed AST to evaluate violations; integrates with the existing check pipeline but operates on AST nodes rather than text patterns.
- **Grammar Bundle**: Compiled tree-sitter grammars for supported languages, bundled with the tool; loaded on demand based on detected language.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 6 AST checks (VH-G002, G003, G004, G009, G010, G012) produce correct violations on fixture files for each of the 6 supported languages (Python, TypeScript, Go, Java, Ruby, Rust).
- **SC-002**: Scanning a directory containing files in all 6 supported languages completes without errors, and violations from AST and non-AST checks both appear in output.
- **SC-003**: A file with a syntax error does not cause the scan to fail — AST checks are skipped for that file and non-AST checks still produce results.
- **SC-004**: Custom observability hints from `.vibe_harness.toml` are correctly applied so that framework-specific logging calls (e.g., `LoggerFactory`, `Rails.logger`) are recognized by VH-G003.
- **SC-005**: The binary requires no external dependencies or runtime grammar downloads — all grammars are bundled.
- **SC-006**: Existing non-AST checks (VH-G001, G005, G006, G007, G008, G011) continue to produce identical results when AST checks are added to the scan pipeline.

## Assumptions

- The tool will integrate tree-sitter for AST parsing, matching the technology direction stated in the roadmap.
- Language grammars will be bundled with the binary so no external downloads are required at runtime.
- Default function length threshold for VH-G002 will follow existing convention (the roadmap states "count statements in function/method bodies"); a reasonable default of 50 statements is assumed unless a configurable threshold is added.
- Default god module threshold for VH-G012 will be a reasonable default (e.g., 20 public exports per file) unless made configurable.
- AST checks will conform to the existing check interface pattern so they integrate seamlessly with the current scan pipeline.
- Observability hint defaults will cover common frameworks: Spring Boot (Java), Rails (Ruby), FastAPI/Django (Python), Express (TypeScript), stdlib (Go), and tracing (Rust).
- VH-G007 (Copy-Paste Duplication) is a multi-file check that already has a separate pattern; the new AST checks are single-file checks and will follow the standard `Check` interface.
- Mixed-language files (e.g., HTML with embedded JS) are out of scope; only the primary language identified by file extension is parsed.
- Language-specific edge cases (Go explicit error returns not being "swallowed", Rust `Result` handling patterns) will be handled by language-specific tree-sitter queries rather than generic heuristics.