---

description: "Task list template for feature implementation"
---

# Tasks: Phase 1 Foundation — Non-AST Checks MVP

**Input**: Design documents from `/specs/001-non-ast-mvp/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Constitution Principle III (Test-First) is NON-NEGOTIABLE. Tests are included for every functional requirement per the TDD cycle: implement FR → write test → verify pass → proceed.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Single Go project**: `cmd/`, `internal/`, `testdata/` at repository root
- Paths shown below match plan.md structure

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Initialize Go module and project structure with `go mod init github.com/jgervais/vibe_harness` and create directories: `cmd/vibe-harness/`, `internal/checks/generic/`, `internal/config/`, `internal/output/`, `internal/rules/`, `internal/scanner/`, `testdata/`
- [X] T002 [P] Create Makefile with targets: `build`, `test`, `vet`, `lint`, `install` at Makefile
- [X] T003 [P] Create LICENSE file (MIT) at LICENSE
- [X] T004 [P] Create `.gitignore` with Go patterns (*.exe, *.test, *.out, vendor/, dist/, vibe-harness) at .gitignore

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Define Violation struct (RuleID, File, Line, Column, EndLine, Message, Severity) and Check struct (ID, Name, Description, Severity, RequiresAST, Threshold) in internal/rules/registry.go
- [X] T006 Write test for Violation and Check struct construction and validation in internal/rules/registry_test.go
- [X] T007 Define Config struct (Observability with LoggingCalls/MetricsCalls, Languages map) and default values in internal/config/config.go
- [X] T008 Write test for Config defaults (verify default logging calls, metrics calls, and language mappings) in internal/config/config_test.go
- [X] T009 Define ScanResult struct (Tool, Target, Violations, Stats, ExitCode) and ScanStats/ToolInfo in internal/scanner/scanner.go
- [X] T010 Implement file discovery using `filepath.WalkDir` with extension filtering, `.git`/`vendor`/`node_modules` skip, and binary file detection (null-byte heuristic) in internal/scanner/scanner.go
- [X] T011 Write test for file discovery (creates temp dirs with source/binary/hidden files, verifies correct filtering) in internal/scanner/scanner_test.go

**Checkpoint**: Foundation ready — user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Run Quality Checks on a Source Tree (Priority: P1) 🎯 MVP

**Goal**: Scan a directory and report all violations from the 6 non-AST checks

**Independent Test**: Run the compiled binary against `testdata/` directories containing known violations and verify output contains expected rule IDs, file paths, and line numbers

### Implementation for User Story 1

- [X] T012 [P] [US1] Implement VH-G001 File Length check: count non-blank, non-comment lines per file, flag files exceeding 300 in internal/checks/generic/file_length.go
- [X] T013 [P] [US1] Write test for VH-G001 with fixtures: clean file under 300 lines, violating file over 300 lines in testdata/file_length/ and internal/checks/generic/file_length_test.go
- [X] T014 [P] [US1] Implement VH-G005 Hardcoded Secrets check: regex pattern matching for AWS keys (AKIA...), API key assignments, credential connection strings, bearer tokens, private key markers in internal/checks/generic/hardcoded_secrets.go
- [X] T015 [P] [US1] Write test for VH-G005 with fixtures: clean Python file, violating Python file with embedded secrets in testdata/hardcoded_secrets/ and internal/checks/generic/hardcoded_secrets_test.go
- [X] T016 [P] [US1] Implement VH-G006 Magic Values check: regex literal scanning, allowed list (0, 1, -1, 2, true, false, null/nil/None/undefined, empty collections), flag repeated values and strings exceeding 20 chars, heuristic constant declaration detection in internal/checks/generic/magic_values.go
- [X] T017 [P] [US1] Write test for VH-G006 with fixtures: clean TypeScript file, violating TypeScript file with magic numbers and long inline strings in testdata/magic_values/ and internal/checks/generic/magic_values_test.go
- [X] T018 [US1] Implement VH-G007 Copy-Paste Duplication check: Winnowing algorithm for file fingerprinting, token normalization (strip identifiers), flag 6+ line blocks with 80%+ similarity across files in internal/checks/generic/duplication.go
- [X] T019 [US1] Write test for VH-G007 with fixtures: two Go files with duplicated 8-line block, clean file pair in testdata/duplication/ and internal/checks/generic/duplication_test.go
- [X] T020 [P] [US1] Implement VH-G008 Comment-to-Code Ratio check: per-language comment pattern table, block comment state machine, flag files where comment lines exceed 1:3 ratio in internal/checks/generic/comment_ratio.go
- [X] T021 [P] [US1] Write test for VH-G008 with fixtures: clean Go file, violating Ruby file with excessive comments in testdata/comment_ratio/ and internal/checks/generic/comment_ratio_test.go
- [X] T022 [P] [US1] Implement VH-G011 Disabled Security Features check: regex pattern matching for verify=False, InsecureSkipVerify: true, rejectUnauthorized: false, --no-verify-ssl, ssl_verify: false, CURLOPT_SSL_VERIFYPEER: false in internal/checks/generic/security_features.go
- [X] T023 [P] [US1] Write test for VH-G011 with fixtures: clean Python file, violating Go file with InsecureSkipVerify in testdata/security_features/ and internal/checks/generic/security_features_test.go
- [X] T024 [US1] Implement per-language comment pattern table (C-family: //, /_.._/; Python/Ruby: #; SQL: --) and line classifier function in internal/scanner/scanner.go
- [X] T025 [US1] Write test for comment line classification across languages (Go, Python, Ruby) in internal/scanner/scanner_test.go
- [X] T026 [US1] Wire scanner to run all 6 checks on discovered files, collect violations, build ScanResult with stats in internal/scanner/scanner.go
- [X] T027 [US1] Write integration test: scan testdata/ directory, verify all expected violations appear in internal/scanner/scanner_test.go
- [X] T028 [US1] Implement CLI entry point in cmd/vibe-harness/main.go: parse flags (--format, --config, --version, --help), positional path arg, wire to scanner, output via human formatter to stderr, set exit codes (0/1/2)
- [X] T029 [US1] Write test for CLI exit codes: build binary, run against clean dir (expect 0), violating dir (expect 1), bad path (expect 2) in cmd/vibe-harness/main_test.go

**Checkpoint**: At this point, User Story 1 should be fully functional — the tool scans directories and reports violations

---

## Phase 4: User Story 2 - Get Machine-Readable Output for CI (Priority: P2)

**Goal**: JSON and SARIF output formats for CI pipeline consumption

**Independent Test**: Run with `--format json` and `--format sarif`, verify output parses as valid JSON and conforms to SARIF v2.1.0 structure

### Implementation for User Story 2

- [X] T030 [P] [US2] Implement JSON formatter: marshal ScanResult to JSON with version, tool info, stats, and results array per contracts/cli.md in internal/output/json.go
- [X] T031 [P] [US2] Write test for JSON formatter: verify output is valid JSON, contains all fields from contract, empty-violation case in internal/output/json_test.go
- [X] T032 [P] [US2] Implement SARIF v2.1.0 formatter: $schema, version, runs with tool.driver (name, version, rules list), results with ruleId/level/message/locations per contracts/cli.md in internal/output/sarif.go
- [X] T033 [P] [US2] Write test for SARIF formatter: verify output is valid JSON, contains required SARIF fields, compatible with GitHub Code Scanning schema in internal/output/sarif_test.go
- [X] T034 [US2] Implement human-readable formatter: `<path>:<line>:<rule-id> — <message>` to stderr per contracts/cli.md in internal/output/human.go
- [X] T035 [US2] Write test for human-readable formatter: verify format string, verify output written to correct writer in internal/output/human_test.go
- [X] T036 [US2] Wire `--format` flag in cmd/vibe-harness/main.go to select formatter (human → stderr, json → stdout, sarif → stdout)
- [X] T037 [US2] Write CLI integration test: run binary with `--format json` and `--format sarif`, verify stdout contains valid structured output in cmd/vibe-harness/main_test.go

**Checkpoint**: User Stories 1 AND 2 both work independently — tool scans and outputs in 3 formats

---

## Phase 5: User Story 3 - Configure Recognition Hints (Priority: P3)

**Goal**: Load `.vibe_harness.toml` for recognition hints; reject rule-modifying config

**Independent Test**: Provide valid config and verify hints are used; provide invalid config with rule modifications and verify explicit error

### Implementation for User Story 3

- [X] T038 [P] [US3] Implement TOML config loading: parse `.vibe_harness.toml` using BurntSushi/toml, auto-discover by walking up from target dir, merge with defaults in internal/config/config.go
- [X] T039 [P] [US3] Write test for config loading: valid TOML, missing file (use defaults), malformed TOML (error) in internal/config/config_test.go
- [X] T040 [US3] Implement config validation: reject any keys that modify rule behavior (disable, threshold, skip, ignore, severity, rules.*, ignore.*), produce explicit error identifying the disallowed field in internal/config/validate.go
- [X] T041 [US3] Write test for config validation: valid config passes, config with `enabled = false` rejected, config with `threshold = 500` rejected, config with `[ignore]` section rejected in internal/config/validate_test.go
- [X] T042 [US3] Create test fixture: valid `.vibe_harness.toml` with observability/logging_calls and languages sections in testdata/config/valid.toml
- [X] T043 [US3] Create test fixture: invalid `.vibe_harness.toml` attempting to disable a rule in testdata/config/invalid_rule_mod.toml
- [X] T044 [US3] Wire `--config` flag in cmd/vibe-harness/main.go: explicit config path or auto-discover, pass Config to scanner and checks
- [X] T045 [US3] Write CLI integration test: run with valid config (uses custom hints), run with invalid config (exit code 2 with error message), run without config (uses defaults) in cmd/vibe-harness/main_test.go

**Checkpoint**: User Stories 1, 2, AND 3 all work independently

---

## Phase 6: User Story 4 - Install and Run Anywhere (Priority: P4)

**Goal**: Cross-compiled binaries, GitHub Release pipeline, install script

**Independent Test**: Download binary from release artifact, verify checksum, run on clean system

### Implementation for User Story 4

- [X] T046 [P] [US4] Create `.goreleaser.yml` with build matrix (darwin/linux/windows × amd64/arm64), `CGO_ENABLED=0`, ldflags for version injection, checksum generation, GitHub Release creation at .goreleaser.yml
- [X] T047 [P] [US4] Create GitHub Actions CI workflow: run `go vet`, `go test ./...`, `go build` on PRs at .github/workflows/ci.yml
- [X] T048 [P] [US4] Create GitHub Actions release workflow: trigger on tag `v*`, run goreleaser, publish binaries and checksums at .github/workflows/release.yml
- [X] T049 [US4] Create install script `install.sh`: detect OS/arch via `uname`, download correct binary from GitHub Releases, verify SHA-256 checksum, install to `/usr/local/bin`, make executable at scripts/install.sh
- [X] T050 [US4] Write test for install script: verify script detects current OS/arch correctly, verify checksum comparison logic (mock download) in scripts/install_test.sh
- [X] T051 [US4] Add version injection: set `Version` and `RulesHash` variables via `-ldflags` in build, wire into `--version` flag output in cmd/vibe-harness/main.go
- [X] T052 [US4] Write test for `--version` output: verify format `vibe-harness v0.1.0 (<os>/<arch>)` and rules hash present in cmd/vibe-harness/main_test.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T053 Update README.md with: tool description, 6 non-AST check summary table, usage examples, build instructions, install script instructions
- [X] T054 [P] Add edge case handling: unreadable files (log warning, skip), binary files (skip via null-byte heuristic), empty directories (exit 0), symlink cycles (skip with warning) in internal/scanner/scanner.go
- [X] T055 [P] Write test for edge cases: unreadable file, binary file content, empty dir, symlink cycle in internal/scanner/scanner_test.go
- [X] T056 [P] Verify all checks handle multi-violation files correctly (single file triggers multiple rules) in internal/checks/generic/*_test.go
- [X] T057 Run `go vet ./...` and fix any issues
- [X] T058 Run `go test ./...` and verify all tests pass
- [X] T059 Verify quickstart.md scenarios: build, run against own source, validate JSON output, validate SARIF output

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1, US2, US3, US4 can proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after Foundational (Phase 2) — No dependencies on other stories
- **US2 (P2)**: Can start after Foundational (Phase 2) — Formatting parallel to US1 checks, but CLI wiring (T036) depends on US1 scanner
- **US3 (P3)**: Can start after Foundational (Phase 2) — Config parallel to US1, but CLI wiring (T044) depends on US1 scanner
- **US4 (P4)**: Can start after Foundational (Phase 2) — Release pipeline independent of check implementations. Version injection (T051) depends on US1 main.go

### Within Each User Story

- Tests written alongside or immediately after implementation tasks
- Check implementations can run in parallel (T012-T023 are all [P])
- Scanner wiring depends on all checks being implemented
- CLI entry point depends on scanner being wired

### Parallel Opportunities

- All Setup tasks (T002-T004) can run in parallel
- All check implementations (T012, T014, T016, T018, T020, T022) can run in parallel
- All check tests (T013, T015, T017, T019, T021, T023) can run in parallel with each other
- JSON/SARIF formatters (T030-T033) can run in parallel
- Config loading and validation (T038-T041) can run in parallel
- CI/release workflows (T046-T048) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all check implementations together:
Task: "Implement VH-G001 File Length check in internal/checks/generic/file_length.go"
Task: "Implement VH-G005 Hardcoded Secrets check in internal/checks/generic/hardcoded_secrets.go"
Task: "Implement VH-G006 Magic Values check in internal/checks/generic/magic_values.go"
Task: "Implement VH-G007 Copy-Paste Duplication check in internal/checks/generic/duplication.go"
Task: "Implement VH-G008 Comment-to-Code Ratio check in internal/checks/generic/comment_ratio.go"
Task: "Implement VH-G011 Disabled Security Features check in internal/checks/generic/security_features.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Run binary against testdata/, verify all 6 checks report violations
5. Deploy/demo if ready — tool is already useful

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → MVP (scanning works)
3. Add User Story 2 → Test independently → CI-ready (JSON/SARIF output)
4. Add User Story 3 → Test independently → Configurable hints
5. Add User Story 4 → Test independently → Distributable binary
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:
1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (checks + scanner + CLI)
   - Developer B: User Story 2 (output formatters — parallel with A)
   - Developer C: User Story 3 (config loading — parallel with A)
3. Developer D: User Story 4 (release pipeline — can start early)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Constitution Principle III requires tests — they are integrated into each story phase
- T018 (duplication) is the most complex check — consider implementing after simpler checks
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence