# Research: Phase 1 Foundation — Non-AST Checks MVP

## Decision 1: CLI Framework

**Decision**: Use Go's standard `flag` package for CLI parsing.

**Rationale**: The constitution requires a single static binary with zero
runtime dependencies. Cobra adds a transitive dependency
(`github.com/spf13/cobra` + `pflag`). For a tool with 3-5 flags
(`--format`, `--config`, `--version`, `--help`, positional path),
the standard library suffices. Cobra becomes worthwhile if subcommands
are needed later (e.g., `vibe-harness check`, `vibe-harness version`).

**Alternatives considered**:
- **Cobra**: Feature-rich, standard for complex Go CLIs. Overkill for
  this scope. Adds binary size. Deferred to a future phase if
  subcommands emerge.
- **urfave/cli**: Similar tradeoffs to Cobra. Heavier than needed.
- **kong**: Good for struct-based config. Adds dependency for minimal
  benefit at this stage.

## Decision 2: TOML Parsing Library

**Decision**: Use `github.com/BurntSushi/toml` for `.vibe_harness.toml`
parsing.

**Rationale**: This is the most widely-adopted TOML library in the Go
ecosystem. Pure Go (no CGO), actively maintained, compatible with
`CGO_ENABLED=0` static builds. The spec requires parsing a small,
well-defined TOML structure with only 2 sections.

**Alternatives considered**:
- **github.com/pelletier/go-toml**: V2 is faster but API churn between
  v1 and v2 caused ecosystem issues. `BurntSushi/toml` is more stable.
- **github.com/naoina/toml**: Abandoned. Not viable.

## Decision 3: SARIF Output Format

**Decision**: Produce SARIF v2.1.0 using hand-constructed JSON via
`encoding/json`.

**Rationale**: The SARIF schema is stable and well-documented. No Go
SARIF library has wide adoption. Since our output structure is simple
(tool metadata + rule list + results), hand-constructing the structs
and marshaling to JSON avoids a dependency that would need to track
SARIF spec updates.

**Alternatives considered**:
- **github.com/owenrumney/go-sarif**: Exists but low adoption, risk of
  abandonment. Adds dependency for something trivially buildable.
- **github.com/sourcegraph/src-cli**: Includes SARIF output but is a
  monolithic tool, not a library.

## Decision 4: File Discovery Strategy

**Decision**: Use `filepath.WalkDir` with extension-based filtering
from the config's `[languages]` section, with built-in defaults for
common extensions.

**Rationale**: `filepath.WalkDir` is more efficient than the older
`filepath.Walk` (avoids `os.Lstat` calls). Extension filtering is
sufficient for non-AST checks since they operate on raw text.
Directories commonly ignored (`.git`, `vendor`, `node_modules`) are
skipped. Binary files are detected by checking the first 512 bytes
for null bytes (the same heuristic `file` command uses).

**Alternatives considered**:
- **gitignore-aware walking**: Respects `.gitignore`. More correct but
  adds a dependency (`go-gitignore` or similar). Can be added later
  without breaking the interface.
- **`.vibe_harness_ignore` file**: Rejected per constitution principle
  V (Recognition Hints Only) — no way to exempt files.

## Decision 5: Duplication Detection Algorithm

**Decision**: Use a sliding-window token hash approach with
Winnowing (document fingerprinting).

**Rationale**: The duplication check (VH-G007) requires detecting
6+ line blocks with 80%+ token similarity across files. Winnowing
selects a subset of hash values from every possible k-gram substring,
producing a compact fingerprint per file. Matching fingerprints between
files indicate duplicate code blocks. This is how MOSS (Measure of
Software Similarity) works. It handles large repos efficiently (O(n)
per file, O(f²) comparison where f = number of files).

**Alternatives considered**:
- **Full AST diff**: Overkill for Phase 1. Requires tree-sitter.
  Deferred to Phase 2.
- **Rabin-Karp string matching**: Simpler but O(n²) for cross-file
  comparison. Too slow for 1000-file repos within the 10-second
  target.
- **Line-based Jaccard similarity**: Fast but produces too many false
  positives (renamed variables defeat it). The spec requires token
  normalization.

## Decision 6: Magic Value Detection Approach

**Decision**: Use regex-based literal scanning with a token classifier.

**Rationale**: Without AST parsing, we cannot distinguish declaration
contexts from usage contexts with 100% accuracy. However, we can
achieve the spec's requirements by: (1) scanning for numeric literals
via regex, (2) scanning for string literals via quote detection,
(3) maintaining a per-file counter for repeated values, (4) flagging
values not in the allowed list. Constants (named assignments) are
detected heuristically by checking if the literal appears on the right
side of an assignment to an ALL_CAPS or PascalCase name.

**Alternatives considered**:
- **Full AST-based detection**: More accurate but requires tree-sitter.
  Deferred to Phase 2 where function-scope tracking is possible.
- **Pure regex without heuristic**: Too many false positives on
  constant declarations.

## Decision 7: Comment Detection Heuristic

**Decision**: Use a per-language comment pattern table for line
classification.

**Rationale**: Non-AST checks need to distinguish comment lines from
code lines for VH-G001 (file length) and VH-G008 (comment-to-code
ratio). The recognition hints config provides language detection per
file extension. A table maps each language to its comment prefix
patterns:
- `//`, `/*..*/`, `///` for C-family (Go, Java, TypeScript, Rust)
- `#` for Python, Ruby
- `--` for SQL (if needed later)
- Block comments are tracked via an in-block state machine per file.

**Alternatives considered**:
- **No comment filtering**: All lines counted. Violates the spec which
  requires "non-blank, non-comment" counting.
- **Tree-sitter comment parsing**: Accurate but requires AST — Phase 2.

## Decision 8: Release Pipeline Architecture

**Decision**: GitHub Actions with a matrix strategy for
cross-compilation, goreleaser for release automation.

**Rationale**: `goreleaser` is the de-facto standard for Go binary
releases. It handles the cross-compilation matrix, checksum generation,
GitHub Release creation, and Homebrew formula updates in a single
configured workflow. It produces the exact output structure described
in the monorepo doc. Using goreleaser avoids hand-writing all the
release steps.

**Alternatives considered**:
- **Manual GitHub Actions steps**: More control but significantly more
  YAML to maintain. goreleaser encapsulates all best practices.
- **Makefile-based release**: Not portable across developer machines.
  CI is where releases should happen.

## Decision 9: Testing Strategy

**Decision**: Table-driven tests using Go's `testing` package with
fixture directories.

**Rationale**: Each check is a pure function: input = file contents
+ config, output = list of violations. Table-driven tests map naturally
to this shape. Fixture directories under `testdata/` contain source
files with known violations and clean files. Each test case specifies
the expected violations per check.

**Alternatives considered**:
- **Testify/assert**: Popular but adds a dependency. The standard
  library `t.Errorf` suffices for table-driven tests.
- **Ginkgo/Gomega**: BDD-style. Overkill. Standard library is more
  idiomatic Go.

---

_Compiled 2026-04-19_