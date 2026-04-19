# vibe_harness

A language-agnostic build tool that enforces a non-configurable quality floor for AI-generated code.

## Why

AI coding agents produce code that works but often lacks the qualities humans learn to include through experience — logging, error handling, input validation, reasonable function sizes. Traditional linters are configurable because humans argue about rules. This tool is intentionally **not configurable**. The rules are baked in. That's the point.

## Rules (non-configurable)

| Rule | Description |
|------|-------------|
| file-length | Files must not exceed a fixed line count |
| function-length | Functions/methods must not exceed a fixed line count |
| missing-logging | Functions performing I/O or state changes must include logging calls |
| missing-error-handling | I/O operations must have error handling |
| magic-values | Hardcoded numeric/string literals (beyond small integers) must be named constants |
| missing-input-validation | Public-facing functions must validate their parameters |
| dead-code | Unreachable code branches are flagged |
| comment-ratio | Excessive comments on obvious code are flagged |
| duplication | Copy-pasted blocks across files are flagged |

## Configuration (recognition hints only)

The tool accepts a `.vibe_harness.toml` file that tells it **how to recognize** patterns in your codebase — never **where** to put them or **whether** to enforce them.

```toml
[observability]
# These are the names your logging library uses
logging_calls = ["log", "logger", "logging", "tracing"]
metrics_calls = ["metrics", "counter", "histogram", "gauge", "timer"]

[languages]
# Map file extensions to language names (for tree-sitter)
".py" = "python"
".ts" = "typescript"
".go" = "go"
".java" = "java"
".rb" = "ruby"
```

That's it. You can't turn rules off. You can't change thresholds. You can't skip files. The floor is the floor.

## Architecture

- **Tree-sitter** for language-agnostic AST parsing
- Single static binary (no runtime dependencies)
- Exit code 0 = pass, 1 = fail
- Violations printed to stderr

## Status

Early stage. Building the spec first, then implementation.

## License

MIT