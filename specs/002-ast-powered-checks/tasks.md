# Tasks: AST-Powered Checks

**Input**: Design documents from `/specs/002-ast-powered-checks/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included per constitution principle III (Test-First: implement FR → write test → verify → proceed).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add tree-sitter dependencies and create the `internal/ast` package.

- [x] T001 Add go-tree-sitter and 6 grammar dependencies to `go.mod` (go get github.com/tree-sitter/go-tree-sitter, github.com/tree-sitter/tree-sitter-python/bindings/go, github.com/tree-sitter/tree-sitter-go/bindings/go, github.com/tree-sitter/tree-sitter-typescript/bindings/go, github.com/tree-sitter/tree-sitter-java/bindings/go, github.com/tree-sitter/tree-sitter-ruby/bindings/go, github.com/tree-sitter/tree-sitter-rust/bindings/go); run `go mod tidy`
- [x] T002 [P] Create `internal/ast/languages.go` — language registry mapping file extensions to language names and language names to grammar initializer functions; include default mapping for Python (.py), TypeScript (.ts/.tsx), JavaScript (.js→typescript grammar), Go (.go), Java (.java), Ruby (.rb), Rust (.rs); expose `SupportedLanguages()` and `LanguageForExtension(ext string) (string, bool)`
- [x] T003 [P] Create `internal/ast/languages_test.go` — table-driven tests for language detection: known extensions return correct language, unknown extensions return false, `.js` maps to "javascript" which uses TypeScript grammar, `.tsx` maps to "typescript"
- [x] T004 Create `internal/ast/parser.go` — `Parser` type with `NewParser()`, `ParseFile(language, content)`, `IsLanguageSupported(language)`, `Close()`; lazy grammar initialization; `ParseResult` type with `Tree()`, `Source()`, `Language()`, `HasError()`, `Close()`; per-contract behavior for unsupported languages (returns nil, nil) and syntax errors (returns result with HasError=true); `QuerySet` type with `NewQuerySet()`, `GetQuery()`, `Compile()`, `Close()`
- [x] T005 Create `internal/ast/parser_test.go` — tests: `NewParser()` creates valid parser; `ParseFile` with unsupported language returns (nil, nil); `ParseFile` with valid Python source returns tree with correct language; `ParseFile` with syntax errors returns result with `HasError=true`; `Close()` frees resources without panic; `QuerySet.Compile` caches compiled queries; `IsLanguageSupported` returns correct bool for known/unknown languages
- [x] T006 Create `internal/ast/io_patterns.go` — per-language I/O call pattern sets for G003 and G009; function `IOPatternsForLanguage(language string) []string` returning built-in I/O function/method name patterns for each supported language; function `IOQueryPatterns(language string) map[string]string` returning language-specific tree-sitter queries for I/O call detection

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core integration that MUST be complete before ANY user story can be implemented. Adds AST support to the scanner pipeline, rule registry, and config.

- [x] T007 [P] Add 6 AST check metadata entries to `internal/rules/registry.go` — VH-G002 (Function Length, warning, RequiresAST: true, Threshold: "50 statements"), VH-G003 (Missing Logging in I/O, warning, RequiresAST: true, Threshold: ""), VH-G004 (Swallowed Errors, error, RequiresAST: true, Threshold: ""), VH-G009 (Missing Error Handling on I/O, error, RequiresAST: true, Threshold: ""), VH-G010 (Broad Exception Catching, warning, RequiresAST: true, Threshold: ""), VH-G012 (God Module, warning, RequiresAST: true, Threshold: "20 exports")
- [x] T008 [P] Add `ASTCheck` interface to `internal/checks/generic/` — define `ASTCheck` interface embedding the existing `Check` interface and adding `CheckFileAST(path string, content []byte, language string, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation`; ensure it's in a shared file accessible by all check implementations
- [x] T009 Update `internal/scanner/scanner.go` — instantiate `ast.Parser` at scan start with `defer parser.Close()`; for each file: determine language, call `parser.ParseFile()` if supported; for each check: type-assert to `ASTCheck`, call `CheckFileAST()` with parse result if AST check, else call `CheckFile()`; defer `parseResult.Close()` per file; handle `ParseFile` errors by logging warning and skipping AST checks
- [x] T010 [P] Add per-language default observability hints to `internal/config/config.go` — create `defaults.go` with `defaultLanguageHints map[string]perLanguageDefaults`; add `MergedLoggingCalls(language string) []string` and `MergedMetricsCalls(language string) []string` methods on `Config` that merge user-provided hints with language defaults (additive); default sets for Python, TypeScript, Go, Java, Ruby, Rust as specified in research.md R6
- [x] T011 [P] Add observability hint merging tests to `internal/config/config_test.go` — test that `MergedLoggingCalls("python")` returns defaults when no config provided; test that user-provided hints are additive (merged, not replacing); test that unknown language returns user-provided hints only (no crash)
- [x] T012 Verify existing non-AST checks still pass — run `go test ./internal/checks/generic/ -run "TestFileLength|TestHardcodedSecrets|TestMagicValues|TestDuplication|TestCommentRatio|TestSecurityFeatures"` and confirm all pass; run `go vet ./...` and confirm no errors

**Checkpoint**: Foundation ready — scanner can parse files with tree-sitter, route AST vs non-AST checks, merge observability hints. User story implementation can begin.

---

## Phase 3: User Story 1 - Run AST checks on a codebase (Priority: P1) 🎯 MVP

**Goal**: All 6 AST checks parse source files, detect violations, and report them alongside existing non-AST checks.

**Independent Test**: Scan a directory with known-violating fixture files in each supported language; verify all 6 AST checks produce correct violations; verify existing non-AST checks still pass.

### VH-G002: Function Length

- [ ] T013 [P] [US1] Create `testdata/function_length/` fixture files — `clean.py`, `violating.py`, `clean.go`, `violating.go`, `clean.ts`, `violating.ts`, `clean.java`, `violating.java`, `clean.rb`, `violating.rb`, `clean.rs`, `violating.rs`; clean files have short functions (under 50 statements); violating files have functions exceeding 50 statements
- [ ] T014 [US1] Implement VH-G002 `FunctionLengthCheck` in `internal/checks/generic/function_length.go` — `struct FunctionLengthCheck` implementing `Check` and `ASTCheck`; `ID()` returns "VH-G002", `Name()` returns "Function Length"; `CheckFile()` returns nil (delegates to `CheckFileAST`); `CheckFileAST()` uses tree-sitter queries per language to find function/method definitions, count body statements, flag those exceeding 50; language-specific query patterns as defined in research; include `QuerySet` for compiled queries with `Close()`; handle language not in query set (return empty)
- [ ] T015 [US1] Write test for VH-G002 in `internal/checks/generic/function_length_test.go` — table-driven tests per language: Python (`function_definition`, `async_function_definition`), Go (`function_declaration`, `method_declaration`), TypeScript (`function_declaration`, `method_definition`, `arrow_function`), Java (`method_declaration`, `constructor_declaration`), Ruby (`method`, `singleton_method`), Rust (`function_item`); test clean files produce zero violations; test violating files produce violations with correct RuleID, Line, Message; test unsupported languages produce zero violations; test function with only comments in body counts as 0 statements

### VH-G004: Swallowed Errors

- [ ] T016 [P] [US1] Create `testdata/swallowed_errors/` fixture files — per-language files with: empty catch/except blocks, catch blocks with only comments or `pass`, catch blocks with error handling (not swallowed); Go: `_ = errValue` assignments, `if err != nil { }` empty bodies; Rust: `.unwrap()` calls
- [ ] T017 [US1] Implement VH-G004 `SwallowedErrorsCheck` in `internal/checks/generic/swallowed_errors.go` — detect empty or comment-only catch/except blocks; language-specific queries: Python `except_clause` with `pass`/comments, TypeScript `catch_clause` with empty `statement_block`, Java `catch_clause` with empty `block`, Ruby `rescue_clause` with empty/`nil` body, Go `_ =` assignment and empty `if err != nil` blocks, Rust `.unwrap()`/`.expect()` calls; check that caught error is not re-raised, returned, or logged before flagging; use observability hints from config to recognize logging calls
- [ ] T018 [US1] Write test for VH-G004 in `internal/checks/generic/swallowed_errors_test.go` — per-language: flagged empty catch blocks, not-flagged catch blocks with `raise`/`return`/logging, Go `_ = err` flagged, Go `if err != nil { return err }` not flagged, Rust `.unwrap()` flagged, Python bare `except:` flagged as both G004 and G010; test unsupported languages produce zero violations

### VH-G009: Missing Error Handling on I/O

- [ ] T019 [P] [US1] Create `testdata/missing_error_handling/` fixture files — per-language files with: I/O calls inside try/catch (handled), I/O calls outside try/catch (not handled), Go: `val, err := ioCall(); if err != nil { ... }` (handled) vs `ioCall()` with no error check (not handled), Rust: `io_call()?` (handled) vs `io_call();` (not handled)
- [ ] T020 [US1] Implement VH-G009 `MissingErrorHandlingCheck` in `internal/checks/generic/missing_error_handling.go` — find I/O function calls not wrapped in error-handling constructs; per-language queries detecting I/O calls and checking for enclosing try/catch, if err != nil, Result handling; use built-in I/O patterns from `internal/ast/io_patterns.go`; identify I/O calls by matching function/method names against known I/O patterns; walk up AST to check for error-handling ancestor nodes; handle Go-specific pattern where `val, err := ioCall()` followed by `if err != nil` IS handled
- [ ] T021 [US1] Write test for VH-G009 in `internal/checks/generic/missing_error_handling_test.go` — per-language: I/O call inside try/catch → not flagged; I/O call outside try/catch → flagged; Go: `os.ReadFile` with `if err != nil` → not flagged; Go: bare `os.ReadFile` call → flagged; Rust: `fs::read_to_string` with `?` → not flagged; Rust: bare call → flagged; test unsupported languages produce zero violations

### VH-G012: God Module

- [ ] T022 [P] [US1] Create `testdata/god_module/` fixture files — per-language files with: files under 20 exports (clean), files over 20 exports (violating); Python: public defs/classes (not starting with _); Go: exported functions/types (uppercase); Java: public methods/classes; TypeScript: export declarations; Ruby: public methods/classes; Rust: `pub fn`/`pub struct`/`pub enum`
- [ ] T023 [US1] Implement VH-G012 `GodModuleCheck` in `internal/checks/generic/god_module.go` — count public exports per file; per-language definition of "public": Python (top-level def/class not starting with `_`), TypeScript (`export` declarations), Go (uppercase-first names), Java (`public` modifier), Ruby (method/class/module definitions not starting with `_`), Rust (`pub` without restriction); threshold: 20 exports; flag files exceeding threshold with message `File has {count} public exports (threshold: 20)`
- [ ] T024 [US1] Write test for VH-G012 in `internal/checks/generic/god_module_test.go` — per-language: clean file < 20 exports → zero violations; violating file > 20 exports → violation with correct count; file with zero exports → not flagged; test Python private functions (starting with _) not counted; test Go unexported names not counted; test Rust `pub(crate)` not counted as public; test unsupported languages produce zero violations

### VH-G003: Missing Logging in I/O

- [ ] T025 [P] [US1] Create `testdata/missing_logging/` fixture files — per-language files with: I/O call with nearby logging (clean), I/O call without logging (violating); include examples using default observability hints (`log`, `logger`, `print` etc.) and examples that would need custom hints
- [ ] T026 [US1] Implement VH-G003 `MissingLoggingCheck` in `internal/checks/generic/missing_logging.go` — identify I/O calls and check for logging calls in the same scope; use built-in I/O patterns from `internal/ast/io_patterns.go` for I/O call detection; use `cfg.MergedLoggingCalls(language)` and `cfg.MergedMetricsCalls(language)` for logging/metrics call detection; walk AST scope (enclosing block/function) to find logging calls; flag I/O calls where no logging call exists in the same scope
- [ ] T027 [US1] Write test for VH-G003 in `internal/checks/generic/missing_logging_test.go` — per-language: I/O call with logging in same block → not flagged; I/O call without any logging in same block → flagged; I/O call with logging in parent scope → depends on scope model; test that config-provided `logging_calls` are recognized; test unsupported languages produce zero violations

### VH-G010: Broad Exception Catching

- [ ] T028 [P] [US1] Create `testdata/broad_catch/` fixture files — per-language files with: Python `except Exception` and bare `except:`, Java `catch (Exception e)` and `catch (Throwable t)`, Ruby `rescue Exception` and bare `rescue`, TypeScript `catch (e)`; also include specific catches that should NOT be flagged (e.g., `except ValueError`, `catch (IOException e)`)
- [ ] T029 [US1] Implement VH-G010 `BroadCatchCheck` in `internal/checks/generic/broad_catch.go` — detect catch/except blocks catching root exception types; per-language queries: Python `except_clause` with `Exception` type or bare except, Java `catch_formal_parameter` with `Exception`/`Throwable`/`RuntimeException` type, Ruby `rescue_clause` with `Exception`/`StandardError` type or bare rescue, TypeScript `catch_clause` (always catches all types); skip for Go and Rust (no exception hierarchies); message format: `Broad exception type '{type}' caught at line {line}`
- [ ] T030 [US1] Write test for VH-G010 in `internal/checks/generic/broad_catch_test.go` — per-language: broad catch flagged with correct type name; specific catch (e.g., `except ValueError`, `catch (IOException e)`) not flagged; Go and Rust produce zero violations; Python bare `except:` flagged; TypeScript `catch (e)` flagged; test unsupported languages produce zero violations

### Scanner Integration Tests for US1

- [ ] T031 [US1] Integrate all 6 AST checks into scanner in `internal/scanner/scanner.go` — instantiate `FunctionLengthCheck`, `MissingLoggingCheck`, `SwallowedErrorsCheck`, `MissingErrorHandlingCheck`, `BroadCatchCheck`, `GodModuleCheck`; add them to the check pipeline; ensure they implement `ASTCheck` and scanner routes to `CheckFileAST` when parse result available
- [ ] T032 [US1] Write integration test for US1 in `internal/scanner/scanner_test.go` — scan a testdata directory containing violating files in multiple languages; verify all 6 AST check violations appear; verify existing non-AST checks (G001, G005-G008, G011) still produce correct results; verify files in unsupported languages only produce non-AST results; verify file with syntax error doesn't crash the scan (AST checks skipped, non-AST checks proceed)
- [ ] T033 [US1] Write rule registry test in `internal/rules/registry_test.go` — verify `rules.Checks()` returns 12 entries (6 existing + 6 new); verify new entries have correct ID, Name, Severity, RequiresAST=true; verify existing entries unchanged

**Checkpoint**: User Story 1 complete — all 6 AST checks parse files, detect violations, and report them. Existing checks unaffected.

---

## Phase 4: User Story 2 - View AST check results in output formats (Priority: P2)

**Goal**: AST check violations appear correctly in human-readable, JSON, and SARIF output formats.

**Independent Test**: Run vibe-harness with `--format json` and `--format sarif` on violating files; verify violations appear in structured output with correct metadata.

- [ ] T034 [P] [US2] Write test for JSON output with AST violations in `internal/output/json_test.go` — scan a directory with known AST violations; verify JSON output contains entries with correct `rule_id` (VH-G002, G003, G004, G009, G010, G012), `file`, `line`, `column`, `end_line`, `message`, `severity`; verify `stats.violations_by_rule` includes AST check rule IDs
- [ ] T035 [P] [US2] Write test for SARIF output with AST violations in `internal/output/sarif_test.go` — verify SARIF `tool.driver.rules` array includes all 6 AST check rule definitions; verify `results` array contains entries for AST violations; verify each result has correct `ruleId`, `level`, `message.text`, and `location` with file URI and region
- [ ] T036 [P] [US2] Write test for human-readable output with AST violations in `internal/output/human_test.go` — verify violations render in `<path>:<line>:<RuleID> — <Message>` format; verify new check IDs appear correctly; verify total count line includes AST violations

**Checkpoint**: User Story 2 complete — all 3 output formats correctly render AST check violations.

---

## Phase 5: User Story 3 - Configure observability recognition hints (Priority: P3)

**Goal**: Custom `.vibe_harness.toml` observability hints are recognized by VH-G003, and default hints work without config.

**Independent Test**: Create a config with custom `logging_calls`, scan code using those call names, verify VH-G003 recognizes them.

- [ ] T037 [P] [US3] Create test config `testdata/missing_logging/custom_hints.toml` with `logging_calls = ["LoggerFactory", "LOG"]` and `metrics_calls = ["micrometer"]`
- [ ] T038 [US3] Write test for custom observability hints in `internal/checks/generic/missing_logging_test.go` — load config with custom `logging_calls`; verify VH-G003 recognizes `LoggerFactory.log(...)` as logging and does not flag I/O calls near it; verify custom `metrics_calls` are recognized as metrics calls; verify additive behavior: defaults + custom hints are both recognized
- [ ] T039 [US3] Write test for default hints (no config) in `internal/checks/generic/missing_logging_test.go` — run VH-G003 with default config (no `.vibe_harness.toml`); verify default logging call names (`log`, `logger`, `logging`, etc.) are recognized per language; verify default metrics call names are recognized

**Checkpoint**: User Story 3 complete — observability hints work correctly with both defaults and custom config.

---

## Phase 6: User Story 4 - Per-language AST pattern correctness (Priority: P4)

**Goal**: Each supported language has correct language-specific query patterns that handle edge cases properly.

**Independent Test**: Run each check against per-language fixture files with known edge cases and verify correct detection.

- [ ] T040 [P] [US4] Create per-language edge case fixtures in `testdata/swallowed_errors/` — Go: `if err != nil { return err }` (not swallowed, should NOT be flagged); Go: `if err != nil { log.Fatal(err) }` (handled, should NOT be flagged); Python: `except:` bare except (flagged by both G004 and G010); Python: `except ValueError as e: raise` (not swallowed)
- [ ] T041 [P] [US4] Create per-language edge case fixtures in `testdata/missing_error_handling/` — Rust: `io_call()?` with `?` operator (handled, not flagged); Rust: `match io_call() { Ok(v) => ..., Err(e) => ... }` (handled, not flagged); Python: I/O inside `with` statement (context manager, considered handled); TypeScript: `await ioCall()` inside try/catch (handled)
- [ ] T042 [P] [US4] Create per-language edge case fixtures in `testdata/broad_catch/` — Python: `except (ValueError, OSError) as e:` (specific, not flagged); Java: `catch (IOException | SQLException e)` (specific multi-catch, not flagged); Ruby: `rescue RuntimeError => e` (specific, not flagged)
- [ ] T043 [P] [US4] Write edge case tests in `internal/checks/generic/swallowed_errors_test.go` — Go: `if err != nil { return err }` not flagged; Go: `_ = errValue` flagged; Python bare `except:` flagged; Python `except Exception as e: logging.error(e)` not flagged
- [ ] T044 [P] [US4] Write edge case tests in `internal/checks/generic/missing_error_handling_test.go` — Rust `?` operator results not flagged; Python `with` statement I/O not flagged; Go `val, err := ioCall()` followed by `if err != nil` not flagged
- [ ] T045 [US4] Write edge case tests in `internal/checks/generic/broad_catch_test.go` — Python tuple exception types not flagged if all specific; Java multi-catch with specific types not flagged; Ruby specific rescue types not flagged; Go and Rust produce zero violations for G010

**Checkpoint**: User Story 4 complete — all per-language edge cases are correctly handled.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Integration validation, performance, and code quality.

- [ ] T046 Add `testdata/parse_error/` fixture — a Python file with deliberate syntax errors; verify scan completes without crash, AST checks are skipped for that file, non-AST checks still produce results
- [ ] T047 Performance validation — create a testdata file with >1000 LOC; verify parse + check completes under 500ms (per quickstart.md goal); test scanner with mixed-language directory (all 6 languages present); verify no memory leaks from ParseResult.Close()
- [ ] T048 Verify SARIF rule metadata complete — run full scan with `--format sarif` and verify all 12 rules (6 existing + 6 new) appear in `tool.driver.rules` with correct id, name, shortDescription, and properties
- [ ] T049 Run `go vet ./...` and `go test ./...` — ensure all tests pass, no vet warnings
- [ ] T050 Update `CLAUDE.md` / project documentation with AST check descriptions (if applicable)
- [ ] T051 Run quickstart.md validation — build binary with `CGO_ENABLED=1 go build -o vibe-harness ./cmd/vibe-harness`; run against testdata directory; verify all output formats render correctly; verify exit codes (0 for clean, 1 for violations, 2 for errors)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — can start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 completion — BLOCKS all user stories
- **Phase 3 (US1)**: Depends on Phase 2 — core AST check implementation
- **Phase 4 (US2)**: Depends on Phase 3 — output format integration requires working checks
- **Phase 5 (US3)**: Depends on Phase 3 (T026 VH-G003 implementation) — config hint integration
- **Phase 6 (US4)**: Depends on Phase 3 — edge case correctness requires working checks
- **Phase 7 (Polish)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after Phase 2 — no dependencies on other stories
- **US2 (P2)**: Depends on US1 — needs working AST checks to test output formats
- **US3 (P3)**: Depends on US1 (specifically VH-G003) — needs working G003 to test config hints
- **US4 (P4)**: Depends on US1 — needs working checks to test edge cases

### Within US1 (Phase 3)

Tasks T013-T015 (G002), T016-T018 (G004), T019-T021 (G009), T022-T024 (G012), T025-T027 (G003), T028-T030 (G010) can each proceed independently after Phase 2. The first 4 check groups (G002, G004, G009, G012) have no inter-dependencies. G003 depends on `io_patterns.go` (T006) and config hints (T010).

### Parallel Opportunities

- All fixture creation tasks (T013, T016, T019, T022, T025, T028) can run in parallel
- All check implementation tasks (T014, T017, T020, T023, T026, T029) can run in parallel after their fixtures are ready
- All test tasks (T015, T018, T021, T024, T027, T030) can run in parallel after their implementations are complete
- US2, US3, US4 tasks can partially overlap once US1 core checks are implemented

---

## Parallel Example: User Story 1 Checks

```bash
# After Phase 2 is complete, these can all launch in parallel:
Task T013: "Create function_length fixtures"      → T014: "Implement G002"      → T015: "Test G002"
Task T016: "Create swallowed_errors fixtures"     → T017: "Implement G004"      → T018: "Test G004"
Task T019: "Create error_handling fixtures"       → T020: "Implement G009"      → T021: "Test G009"
Task T022: "Create god_module fixtures"            → T023: "Implement G012"      → T024: "Test G012"
Task T025: "Create missing_logging fixtures"        → T026: "Implement G003"      → T027: "Test G003"
Task T028: "Create broad_catch fixtures"            → T029: "Implement G010"      → T030: "Test G010"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T006)
2. Complete Phase 2: Foundational (T007-T012)
3. Complete Phase 3: User Story 1 — all 6 checks (T013-T033)
4. **STOP and VALIDATE**: Run full test suite, verify all checks produce correct violations
5. Build binary, run against real codebase

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (6 AST checks) → Test independently → Build and validate (MVP!)
3. Add US2 (output formats) → Test SARIF/JSON/human → Deploy/Demo
4. Add US3 (config hints) → Test custom observability → Deploy/Demo
5. Add US4 (edge cases) → Test language-specific correctness → Deploy/Demo
6. Polish (Phase 7) → Performance, edge case validation → Release v0.2.0

### Parallel Team Strategy

With multiple developers:

1. Team completes Phase 1 + Phase 2 together
2. Once Foundational is done:
   - Developer A: G002 + G004 (Function Length + Swallowed Errors)
   - Developer B: G009 + G012 (Missing Error Handling + God Module)
   - Developer C: G003 + G010 (Missing Logging + Broad Catch)
3. After US1 checks complete:
   - Developer A: US2 (Output formats)
   - Developer B: US3 (Config hints)
   - Developer C: US4 (Edge cases)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (constitution principle III: Test-First)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- CGO_ENABLED=1 is required for all builds and tests
- All 6 grammar packages increase binary size and build time — this is expected
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence