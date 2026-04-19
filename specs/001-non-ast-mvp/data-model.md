# Data Model: Phase 1 Foundation — Non-AST Checks MVP

## Entities

### Violation

A single rule breach detected in a source file.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| RuleID | string | yes | Rule identifier (e.g., "VH-G001") |
| File | string | yes | Absolute or relative path to the violating file |
| Line | int | yes | 1-based line number where the violation occurs |
| Column | int | no | 1-based column offset (0 if not applicable) |
| EndLine | int | no | Ending line for multi-line violations (0 if not applicable) |
| Message | string | yes | Human-readable description of the violation |
| Severity | string | yes | One of: "error", "warning", "note" |

**Validation rules**:
- RuleID MUST match the pattern `VH-G\d{3}`
- File MUST be a non-empty string
- Line MUST be >= 1
- Severity MUST be one of the three allowed values

### Check

A stateless check that examines source file(s) and produces violations.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| ID | string | yes | Rule identifier (e.g., "VH-G001") |
| Name | string | yes | Short human-readable name |
| Description | string | yes | One-line description of what the check detects |
| Severity | string | yes | Default severity for violations from this check |
| RequiresAST | bool | yes | false for all Phase 1 checks |
| Threshold | string | yes | Human-readable threshold description |

**Validation rules**:
- ID is unique across all checks
- Severity is one of: "error", "warning", "note"
- Phase 1 checks all have RequiresAST = false

**Predefined instances** (Phase 1):

| ID | Name | Severity | Threshold |
|----|------|----------|-----------|
| VH-G001 | File Length | warning | 300 lines |
| VH-G005 | Hardcoded Secrets | error | Pattern match |
| VH-G006 | Magic Values | warning | Inline literal detection |
| VH-G007 | Copy-Paste Duplication | warning | 6 lines, 80% similarity |
| VH-G008 | Comment-to-Code Ratio | note | 1:3 ratio |
| VH-G011 | Disabled Security Features | error | Pattern match |

### Config (Recognition Hints)

User-provided configuration loaded from `.vibe_harness.toml`.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Observability | ObservabilityConfig | no | Logging/metrics recognition hints |
| Languages | map[string]string | no | File extension → language name mapping |

#### ObservabilityConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| LoggingCalls | []string | no | Function names recognized as logging calls |
| MetricsCalls | []string | no | Function names recognized as metrics calls |

**Validation rules**:
- Config MUST NOT contain any key that modifies rule behavior
  (e.g., disable, threshold, skip, ignore, severity)
- Unknown keys at the top level produce a warning
- Unknown keys in `[observability]` or `[languages]` produce an error
- If no config file is found, defaults are used

**Default values**:
- `LoggingCalls`: `["log", "logger", "logging", "tracing", "slog", "logr"]`
- `MetricsCalls`: `["metrics", "counter", "histogram", "gauge", "timer", "prometheus"]`
- `Languages`: `{".py": "python", ".ts": "typescript", ".tsx": "typescript", ".js": "javascript", ".go": "go", ".java": "java", ".rb": "ruby", ".rs": "rust"}`

### ScanResult

The aggregate output of scanning a directory tree.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Tool | ToolInfo | yes | Tool name and version |
| Target | string | yes | Scanned directory path |
| Violations | []Violation | yes | All violations found (empty if none) |
| Stats | ScanStats | yes | Summary statistics |
| ExitCode | int | yes | 0, 1, or 2 |

#### ToolInfo

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Name | string | yes | "vibe-harness" |
| Version | string | yes | Semantic version (e.g., "0.1.0") |
| RulesHash | string | yes | SHA-256 of compiled rule set |

#### ScanStats

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| FilesScanned | int | yes | Number of source files processed |
| FilesSkipped | int | yes | Files skipped (binary, unreadable, wrong extension) |
| ViolationsByRule | map[string]int | yes | Count of violations per rule ID |
| Duration | string | yes | Human-readable scan duration |

## Relationships

- A `ScanResult` contains zero or more `Violation` records
- Each `Violation` references exactly one `Check` via RuleID
- A `Config` is loaded once per scan and passed to all checks
- Each `Check` receives the `Config` and file contents, returns `[]Violation`

## State Transitions

There are no state transitions in Phase 1. All entities are
ephemeral — created during a scan and discarded after output.
No persistence is required.