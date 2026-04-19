<!--
Sync Impact Report
==================
Version change: N/A → 1.0.0
Modified principles: N/A (initial ratification)
Added sections:
  - I. Non-Configurable Quality Floor
  - II. Language-Agnostic by Default
  - III. Test-First (NON-NEGOTIABLE)
  - IV. Single Binary, Zero Runtime Dependencies
  - V. Recognition Hints Only
  - Technical Constraints
  - Development Workflow
  - Governance
Removed sections: None
Templates requiring updates:
  - .specify/templates/plan-template.md ✅ no update needed
    (Constitution Check section already generic; principles
    will be populated dynamically at plan time)
  - .specify/templates/spec-template.md ✅ no update needed
    (No constitution-specific mandatory sections required)
  - .specify/templates/tasks-template.md ✅ no update needed
    (Task categorization already principle-agnostic)
  - .opencode/command/speckit.plan.md ✅ no update needed
    (Loads constitution at runtime for Constitution Check)
  - .opencode/command/speckit.constitution.md ✅ no update needed
  - .opencode/command/speckit.specify.md ✅ no update needed
  - .opencode/command/speckit.tasks.md ✅ no update needed
  - .opencode/command/speckit.implement.md ✅ no update needed
  - .opencode/command/speckit.analyze.md ✅ no update needed
  - .opencode/command/speckit.checklist.md ✅ no update needed
  - .opencode/command/speckit.clarify.md ✅ no update needed
  - .opencode/command/speckit.taskstoissues.md ✅ no update needed
Follow-up TODOs: None
-->

# vibe_harness Constitution

## Core Principles

### I. Non-Configurable Quality Floor

The quality floor is the quality floor. Rules MUST NOT be
disabled, lowered, or exempted per file, directory, or project.
Thresholds are fixed and identical for every codebase scanned.
The ONLY accepted input is recognition hints — telling the
tool HOW to identify patterns, never WHETHER to enforce them.

**Rationale**: Configurable linters let teams opt out of rules
they find inconvenient. AI-generated code needs the opposite:
an immovable standard that cannot be bargained down. If rules
are configurable, the floor ceases to exist.

### II. Language-Agnostic by Default

Every generic check MUST operate across all supported languages.
Language-specific checks MUST be gated behind explicit language
detection and MUST NOT degrade the generic check baseline.
Tree-sitter grammars MUST be bundled via go-embed so the binary
has zero external grammar dependencies.

**Rationale**: AI agents write code in many languages. A
quality tool that only works for one language leaves the most
dangerous gaps uncovered. Language-agnostic checks catch the
cross-cutting problems (missing error handling, swallowed
errors, magic values) that are language-independent.

### III. Test-First (NON-NEGOTIABLE)

Tests MUST be written immediately after implementing each
functional requirement, NOT at the end of a phase or project.
The cycle is: implement FR → write test → verify test passes
→ proceed to next FR. Red-Green-Refactor MUST be followed.
No functional requirement is complete until its test exists
and passes.

**Rationale**: Deferred testing accumulates unverified
assumptions. For a tool that enforces code quality, shipping
unverified checks undermines the project's credibility and
allows regressions to hide until they become expensive to fix.

### IV. Single Binary, Zero Runtime Dependencies

The tool MUST ship as a single static binary. It MUST NOT
require a runtime, interpreter, or external grammar files.
Exit code 0 = pass, 1 = violations found, 2 = tool error.
Violations MUST be printed to stderr. JSON and SARIF output
formats MUST be supported for CI integration.

**Rationale**: Adoption friction kills developer tools.
A single binary with no dependencies can be dropped into any
CI pipeline, any Docker image, any developer laptop without
prerequisites. This is a distribution principle, not just a
build convenience.

### V. Recognition Hints Only

The `.vibe_harness.toml` config file MUST accept ONLY
recognition hints — names the codebase uses for logging,
metrics, and file-to-language mapping. The config MUST NOT
accept rule overrides, threshold modifications, file
exemptions, or severity adjustments. If the config file
attempts to modify rule behavior, the tool MUST fail with an
explicit error.

**Rationale**: Configuration that changes enforcement is a
slippery slope back to configurable linting. Recognition
hints solve a real problem (different codebases use different
logging libraries) without opening the door to disabling
checks. The distinction between "how to recognize" and "whether
to enforce" MUST remain absolute.

## Technical Constraints

- **Language**: Go (statically compiled, cross-platform)
- **Parsing**: Tree-sitter via go-tree-sitter with bundled
  grammars (Python, TypeScript, Go, Java, Ruby, Rust)
- **License**: MIT
- **Output formats**: Human-readable (stderr), JSON, SARIF
- **CI integration**: Exit codes 0/1/2, GitHub Action planned
- **Cross-compilation**: darwin/linux/windows × amd64/arm64

## Development Workflow

- Follow the Test-First principle (III) for every functional
  requirement — no exceptions
- Use subagent developers for implementation tasks; each task
  gets a fresh subagent session
- Constitution gates MUST be checked before Phase 0 research
  and re-checked after Phase 1 design in every feature plan
- Complexity violations against this constitution MUST be
  justified with a simpler-alternative-rejected rationale in
  the Complexity Tracking section of plan.md
- Code review MUST verify compliance with all five core
  principles before merge

## Governance

This constitution supersedes all other development practices
and conventions. In case of conflict between this document and
any other guidance, the constitution takes precedence.

**Amendment procedure**: Any amendment MUST be documented with
a version bump, a summary of changes, and explicit justification.
Minor amendments (new principle, materially expanded guidance)
increment the MINOR version. Major amendments (principle removal,
backward-incompatible redefinition) increment the MAJOR version.
Patch changes (clarifications, wording fixes) increment the
PATCH version.

**Compliance review**: Every feature plan MUST include a
Constitution Check section. Every pull request MUST verify
compliance with core principles. Violations of NON-NEGOTIABLE
principles block merge regardless of code review approval.

**Versioning policy**: Semantic versioning (MAJOR.MINOR.PATCH)
as described in the amendment procedure above.

Use `.specify/memory/constitution.md` for the authoritative
text. Runtime development guidance lives in the agent-specific
context files managed by the speckit toolchain.

**Version**: 1.0.0 | **Ratified**: 2026-04-19 | **Last Amended**: 2026-04-19