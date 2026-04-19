# Feature Specification: Phase 1 Foundation — Non-AST Checks MVP

**Feature Branch**: `001-non-ast-mvp`
**Created**: 2026-04-19
**Status**: Draft
**Input**: User description: "phase1 from ./docs/roadmap.md"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run Quality Checks on a Source Tree (Priority: P1)

A developer points the tool at a directory of source code and gets a list of violations from the six non-AST checks. The tool scans all files matching configured extensions, applies each check, and reports every violation with file path, line number, and rule identifier. No configuration is required to get started — the tool works out of the box with sensible defaults.

**Why this priority**: This is the core value proposition. Without the ability to scan and report violations, nothing else matters. All other stories depend on this working.

**Independent Test**: Can be fully tested by running the tool against a fixture directory containing known violations and verifying that each expected violation appears in the output with correct file, line, and rule ID.

**Acceptance Scenarios**:

1. **Given** a directory with source files containing known violations for each of the 6 rules, **When** the tool is run against that directory, **Then** it reports all violations with accurate file paths, line numbers, and rule IDs
2. **Given** a directory with clean source files (no violations), **When** the tool is run, **Then** it exits with code 0 and reports no violations
3. **Given** a directory with mixed clean and violating files, **When** the tool is run, **Then** only the violating files produce violations in the output

---

### User Story 2 - Get Machine-Readable Output for CI (Priority: P2)

A continuous integration pipeline runs the tool on every pull request and needs structured output to integrate with code review dashboards. The tool produces JSON and SARIF formats so that CI systems and GitHub Code Scanning can consume results programmatically.

**Why this priority**: CI integration is the primary distribution channel. Without machine-readable output, the tool cannot gate pull requests or populate security dashboards.

**Independent Test**: Can be tested by running the tool with each output format flag and verifying the output is valid JSON or SARIF with the expected structure and content.

**Acceptance Scenarios**:

1. **Given** a directory with violations, **When** the tool is run with the JSON output flag, **Then** it outputs a valid JSON document containing all violations with structured fields
2. **Given** a directory with violations, **When** the tool is run with the SARIF output flag, **Then** it outputs a valid SARIF document compatible with GitHub Code Scanning
3. **Given** any run that produces violations, **When** the tool exits, **Then** the exit code is 1; when no violations, exit code is 0; when the tool itself errors, exit code is 2

---

### User Story 3 - Configure Recognition Hints (Priority: P3)

A team wants the tool to recognize their project's logging library and source file extensions. They create a `.vibe_harness.toml` file specifying which function names count as logging calls and which extensions map to which languages. The tool uses these hints to improve detection accuracy without changing which rules are enforced or what thresholds apply.

**Why this priority**: Recognition hints improve accuracy but the tool works without them using built-in defaults. This is an enhancement, not a blocker.

**Independent Test**: Can be tested by providing a `.vibe_harness.toml` with custom logging call names and verifying the tool uses them during checks, then providing a config that attempts to modify a rule and verifying the tool rejects it with an error.

**Acceptance Scenarios**:

1. **Given** a `.vibe_harness.toml` with valid observability and languages sections, **When** the tool runs, **Then** it uses the provided recognition hints for pattern detection
2. **Given** a `.vibe_harness.toml` that attempts to disable a rule or modify a threshold, **When** the tool loads the config, **Then** it fails with an explicit error message identifying the disallowed field
3. **Given** no `.vibe_harness.toml` file, **When** the tool runs, **Then** it uses built-in default recognition hints and completes successfully

---

### User Story 4 - Install and Run Anywhere (Priority: P4)

A developer downloads a single binary for their operating system and architecture, places it on their PATH, and immediately runs it against any source tree. No runtime, package manager, or external grammar files are needed. GitHub Releases provide signed binaries with checksums and a one-command install script.

**Why this priority**: Distribution is essential for adoption but can follow the core functionality. A working local build suffices for early development.

**Independent Test**: Can be tested by downloading the binary from the release artifact, verifying the checksum, running it on a fresh system without prerequisites, and confirming it scans a directory and produces output.

**Acceptance Scenarios**:

1. **Given** a GitHub Release with cross-compiled binaries, **When** a user downloads the binary for their OS/arch, **Then** it runs without errors and produces the expected output
2. **Given** the install script, **When** a user runs it, **Then** it detects OS/arch, downloads the correct binary, verifies the checksum, and places it on the PATH
3. **Given** the cross-compilation matrix covers darwin/linux/windows on amd64/arm64, **When** any combination is downloaded, **Then** the binary executes and produces correct exit codes

---

### Edge Cases

- What happens when a source file cannot be read (permissions, encoding)?
- How does the tool handle binary files or very large files in the source tree?
- What happens when `.vibe_harness.toml` exists but is malformed TOML?
- How does the tool handle a directory with zero source files?
- What happens when a file contains lines that match multiple violation patterns?
- How does the duplication check handle files larger than available memory?
- What happens when the tool is run on a symlink cycle?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST scan a directory tree recursively and discover source files based on configured file extensions
- **FR-002**: The tool MUST apply the VH-G001 (File Length) check: flag files exceeding 300 non-blank, non-comment lines
- **FR-003**: The tool MUST apply the VH-G005 (Hardcoded Secrets) check: flag files containing string literals matching known secret patterns (AWS keys, API key assignments, credential connection strings, bearer tokens, private key markers)
- **FR-004**: The tool MUST apply the VH-G006 (Magic Values) check: flag inline numeric and string literals beyond an allowed list (0, 1, -1, 2, true, false, null/nil/None/undefined, empty collections), when used more than once or exceeding 20 characters
- **FR-005**: The tool MUST apply the VH-G007 (Copy-Paste Duplication) check: flag code blocks of 6+ lines with 80%+ token similarity appearing in multiple locations
- **FR-006**: The tool MUST apply the VH-G008 (Comment-to-Code Ratio) check: flag files where comment lines exceed 1 per 3 code lines
- **FR-007**: The tool MUST apply the VH-G011 (Disabled Security Features) check: flag known parameters that disable security features (verify=False, InsecureSkipVerify: true, rejectUnauthorized: false, etc.)
- **FR-008**: The tool MUST output violations in a human-readable format to stderr with file path, line number, rule ID, and description
- **FR-009**: The tool MUST output violations in JSON format when requested, containing structured violation records
- **FR-010**: The tool MUST output violations in SARIF format when requested, compatible with GitHub Code Scanning
- **FR-011**: The tool MUST exit with code 0 when no violations, code 1 when violations found, code 2 when the tool encounters an internal error
- **FR-012**: The tool MUST parse a `.vibe_harness.toml` configuration file for recognition hints only (observability logging/metrics call names, file extension to language mappings)
- **FR-013**: The tool MUST reject any configuration that attempts to disable rules, modify thresholds, exempt files, or change severity — with an explicit error identifying the disallowed field
- **FR-014**: The tool MUST work with built-in default recognition hints when no configuration file is present
- **FR-015**: The tool MUST ship as a single static binary with no external runtime dependencies
- **FR-016**: The tool MUST provide a cross-compilation and release pipeline producing binaries for darwin/linux/windows on amd64/arm64
- **FR-017**: The tool MUST produce checksums for release binaries
- **FR-018**: The tool MUST provide an install script with automatic OS/arch detection, binary download, checksum verification, and PATH installation

### Key Entities

- **Violation**: A single rule breach detected in a source file. Attributes: rule ID, file path, line number(s), description, severity. Each violation is independently identifiable and reportable.
- **Check**: One of the six non-AST quality rules. Attributes: rule ID, name, description, fixed threshold, detection method. Checks are stateless and order-independent.
- **Recognition Hint**: A user-provided mapping in `.vibe_harness.toml` that helps the tool identify patterns (logging call names, file extensions). Hints affect detection accuracy only, never enforcement.
- **Scan Result**: The aggregate output of a full directory scan. Contains the list of violations, the scan target path, the output format, and the exit code.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can scan a source tree and receive violation reports in under 10 seconds for a repository with 1,000 source files
- **SC-002**: Each of the 6 checks correctly identifies 100% of intentional violations in a curated fixture set with zero false negatives
- **SC-003**: False positive rate for each check is below 5% on a set of real-world open-source repositories
- **SC-004**: A user can download, install, and run the tool on a new source tree within 60 seconds on a standard developer machine
- **SC-005**: CI pipelines can parse the machine-readable output without custom parsing logic — JSON output validates against its schema and SARIF output is accepted by GitHub Code Scanning
- **SC-006**: The tool runs successfully on all 6 target platform combinations (darwin/linux/windows × amd64/arm64) without OS-specific failures

## Assumptions

- Users run the tool on source code repositories (not arbitrary text files or binary blobs)
- Source files use UTF-8 encoding; files with other encodings may produce best-effort results
- The six non-AST checks are sufficient for an MVP; AST-dependent checks are deferred to Phase 2
- Comment detection for line-based checks uses heuristic patterns (lines starting with //, #, --, etc.) rather than language-aware parsing
- The duplication check can operate on files that fit in memory; extremely large repositories may require future optimization
- The tool's Go implementation with static compilation inherently satisfies the single-binary, zero-dependency requirement
- The release pipeline uses GitHub Actions, which is already available for the project
- Binary signing uses checksum files; GPG signing or code signing certificates are deferred to a future phase
- The install script targets POSIX systems (macOS, Linux); Windows users receive the binary directly or via package managers in later phases