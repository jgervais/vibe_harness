# Existing Tools for AI-Generated Code Quality

Survey of existing tools, research, and approaches that address AI-generated code quality concerns.

---

## 1. Traditional Static Analysis (Applicable but Not AI-Specific)

### ESLint / TSLint
- **What:** Pluggable JavaScript/TypeScript linter with configurable rules
- **Relevance:** Rules like `max-lines`, `max-lines-per-function`, `no-unused-vars`, `no-empty` catch AI patterns
- **Gaps:** Fully configurable — agents can disable rules. Not AI-specific. JS/TS only

### Pylint / Flake8 / Ruff
- **What:** Python linters with similar structural rules
- **Relevance:** `too-many-lines`, `too-many-branches`, `bare-except`, `unused-variable`
- **Gaps:** Same configurability problem. Python only

### SonarQube / SonarLint
- **What:** Multi-language static analysis platform with code quality gates
- **Relevance:** Tracks code smells, bugs, security vulnerabilities, duplication. Quality gate concept is close to what we want
- **Gaps:** Heavyweight (requires server). Configurable rules. Not AI-specific. Enterprise pricing

### Semgrep
- **What:** Pattern-based multi-language static analysis
- **Relevance:** Write custom rules in a declarative pattern language. Can express "function with I/O but no logging" type rules
- **Gaps:** Requires rule authoring. Configurable by definition. Good building block but not the product
- **Interesting:** Could be a backend for some rules

### CodeQL (GitHub)
- **What:** Semantic code analysis for security vulnerabilities
- **Relevance:** Deep data flow analysis. Can detect hardcoded secrets, injection patterns
- **Gaps:** Security-focused, not quality-focused. Complex query language. GitHub-proprietary

---

## 2. AI-Specific Code Quality Research

### GitHub Copilot Research (2022-2023)
- **What:** GitHub's own studies on Copilot code quality
- **Findings:** Users accepted ~30% of suggestions. Code with Copilot had more bugs per KLOC in security scans. Users wrote more code but didn't necessarily write better code
- **Relevance:** Validates the problem — AI code has measurable quality gaps

### "Asleep at the Keyboard" (Perry et al., 2023)
- **What:** Academic study on security of AI-generated code
- **Findings:** Users with AI assistance wrote significantly less secure code. They trusted AI output and didn't review for security
- **Relevance:** Confirms security-specific smell patterns

### "Do Users Write More Insecure Code with AI?" (Pearce et al., 2022)
- **What:** Study comparing security of code written with and without AI assistance
- **Findings:** AI-assisted code had more CVE-level vulnerabilities, especially in crypto and input validation
- **Relevance:** Reinforces missing-validation as a top AI smell

### Counterfactual Testing / Mutation Testing
- **What:** Generate mutations of code and check if tests catch them
- **Relevance:** AI code with poor tests will have low mutation scores — revealing the "test mirrors implementation" smell
- **Gaps:** Runtime-dependent, not static analysis

---

## 3. AI Code Detection Tools

### GPTDetector / AI Text Classifier
- **What:** Detects whether text (including code) was AI-generated
- **Relevance:** Interesting but orthogonal — we don't care IF code is AI-generated, we care about the QUALITY patterns regardless of origin
- **Gaps:** Classification, not quality assessment

### DetectGPT (Mitchell et al., 2023)
- **What:** Academic approach using probability curvature to detect AI text
- **Relevance:** Same as above — detection not quality
- **Gaps:** Not applicable to our use case

---

## 4. AI Code Review Tools

### CodeRabbit
- **What:** AI-powered code review on PRs
- **Relevance:** Reviews code for quality issues including some AI-specific patterns
- **Gaps:** AI reviewing AI — fox guarding henhouse? Also not a build tool, it's a PR review tool

### Sourcery
- **What:** Automated code review and refactoring suggestions
- **Relevance:** Detects duplication, complexity, missing error handling
- **Gaps:** Suggests fixes instead of enforcing a floor. Configurable. Not AI-specific

### Codacy
- **What:** Automated code quality platform with multi-language support
- **Relevance:** Quality gates, pattern detection, duplication analysis
- **Gaps:** Configurable, SaaS-dependent, not AI-specific

---

## 5. Observability-Specific Tools

### OpenTelemetry Instrumentation
- **What:** Auto-instrumentation libraries that add tracing/metrics without code changes
- **Relevance:** Sidesteps the missing-observability problem by adding it at runtime
- **Gaps:** Not a quality gate — it patches the problem rather than requiring good code. Doesn't help with missing error handling, validation, etc.

### Datadog / New Relic Auto-Instrumentation
- **What:** Vendor-specific auto-instrumentation
- **Relevance:** Same approach as OpenTelemetry
- **Gaps:** Vendor lock-in, doesn't enforce code quality

---

## 6. Tree-Sitter Based Tools

### tree-sitter-cli
- **What:** Parser generator and CLI for tree-sitter grammars
- **Relevance:** Foundation for language-agnostic AST analysis. Parses 50+ languages into uniform tree format
- **Gaps:** Just a parser — no rules, no analysis. Building block only

### ast-grep (astgrep)
- **What:** Code search and rewrite tool based on tree-sitter AST patterns
- **Relevance:** Can express structural rules like "catch block with empty body" as AST patterns. Multi-language
- **Gaps:** Search/replace tool, not a linter. No built-in rules. Could be a backend
- **Interesting:** Very close to what we need — might be worth using as the pattern-matching engine

### Semgrep (tree-sitter based)
- **What:** Uses tree-sitter under the hood for pattern matching
- **Relevance:** Proves tree-sitter works at scale for multi-language analysis
- **Gaps:** Requires custom rules; fully configurable

---

## 7. Non-Configurable / Opinionated Tools (Closest Relatives)

### StandardJS
- **What:** Opinionated JS linter with zero configuration
- **Relevance:** Proves the "no config" model works. Users adopted it specifically because they were tired of config arguments
- **Gaps:** JS only. Style-focused, not AI-quality-focused. No structural rules

### Standard Ruby
- **What:** Same concept for Ruby
- **Relevance:** Same proof point
- **Gaps:** Same limitations

### Black (Python formatter)
- **What:** Opinionated Python formatter with no config
- **Relevance:** "Any color you like, as long as it's black." Proves that rigid, non-configurable tools get adopted when they remove bike-shedding
- **Gaps:** Formatter, not linter. Doesn't check quality, only style

### gofmt / go vet
- **What:** Go's built-in formatter and linter
- **Relevance:** go vet checks for specific constructs that are almost always bugs — closer to our model
- **Gaps:** Go only. Limited rule set

---

## Key Takeaways

1. **No existing tool does what we want.** Everything is either configurable, single-language, style-focused, or a building block without rules.

2. **Tree-sitter is the right foundation.** It's proven (Semgrep, ast-grep), multi-language, and produces uniform ASTs for rule matching.

3. **ast-grep could be a backend.** It does the tree-sitter pattern matching we need. We'd provide the non-configurable rule set on top.

4. **The "no config" model has precedent.** StandardJS, Black, and gofmt prove that opinionated tools with zero configuration get adopted. The key is having good defaults.

5. **Quality gates exist but are heavyweight.** SonarQube's quality gate concept is right, but it requires a server and configuration. We need the same concept as a CLI tool.

6. **Nobody is targeting AI-specific smells.** Traditional linters can catch some of these patterns, but nobody has organized them as "the quality floor for AI-generated code." That framing is new.

---

_Compiled 2026-04-19_