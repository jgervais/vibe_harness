# CLI Contract: vibe-harness

## Command Schema

```
vibe-harness [flags] <path>
```

### Positional Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| path | yes | Directory or file to scan |

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --format | string | "human" | Output format: "human", "json", "sarif" |
| --config | string | "" | Path to config file (default: auto-discover `.vibe_harness.toml`) |
| --version | bool | false | Print version and exit |
| --help | bool | false | Print usage and exit |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No violations found |
| 1 | One or more violations found |
| 2 | Tool error (invalid flag, unreadable path, config error) |

### Output: Human-Readable (stderr)

```
<path>:<line>:<rule-id> — <message>
```

Example:
```
src/main.go:42:VH-G001 — file exceeds 300 non-blank, non-comment lines (412)
src/config.py:7:VH-G005 — hardcoded secret: AWS access key pattern "AKIA..."
src/utils.ts:15:VH-G006 — magic value: "localhost:5432" used inline (20+ chars)
```

### Output: JSON (stdout)

```json
{
  "version": "1.0",
  "tool": {
    "name": "vibe-harness",
    "version": "0.1.0",
    "rules_hash": "sha256:abc123..."
  },
  "target": "/path/to/scanned/dir",
  "stats": {
    "files_scanned": 42,
    "files_skipped": 3,
    "violations_by_rule": {
      "VH-G001": 2,
      "VH-G005": 1
    },
    "duration": "1.2s"
  },
  "results": [
    {
      "rule_id": "VH-G001",
      "file": "src/main.go",
      "line": 1,
      "column": 0,
      "end_line": 0,
      "message": "file exceeds 300 non-blank, non-comment lines (412)",
      "severity": "warning"
    }
  ]
}
```

### Output: SARIF (stdout)

```json
{
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "vibe-harness",
          "version": "0.1.0",
          "rules": [
            {
              "id": "VH-G001",
              "name": "FileLength",
              "shortDescription": { "text": "Files must not exceed 300 lines" },
              "defaultConfiguration": { "level": "warning" }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "VH-G001",
          "level": "warning",
          "message": { "text": "file exceeds 300 non-blank, non-comment lines (412)" },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": { "uri": "src/main.go" },
                "region": { "startLine": 1 }
              }
            }
          ]
        }
      ]
    }
  ]
}
```

### Config File Contract: `.vibe_harness.toml`

```toml
[observability]
logging_calls = ["log", "logger", "logging", "tracing", "slog"]
metrics_calls = ["metrics", "counter", "histogram", "gauge", "timer"]

[languages]
".py" = "python"
".ts" = "typescript"
".tsx" = "typescript"
".go" = "go"
".java" = "java"
".rb" = "ruby"
".rs" = "rust"
```

**Rejected keys** (produce error on load):
```toml
[rules.VH-G001]
enabled = false           # ERROR: cannot disable rules
threshold = 500          # ERROR: cannot modify thresholds

[ignore]
paths = ["vendor/"]      # ERROR: cannot exempt paths
```

### Version Output

```
vibe-harness v0.1.0 (darwin/arm64)
rules hash: sha256:abc123...
```