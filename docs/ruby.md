# Ruby — Build System & Existing Linters

## Build Systems

### Bundler + Rake
- **Config:** `Gemfile`, `Rakefile`
- **Integration:**
  ```ruby
  # Rakefile
  task :lint do
    sh 'vibe-harness .'
  end
  ```
  ```bash
  bundle exec rake lint
  ```

### Bundler alone
- **Config:** `Gemfile`
- **Integration:** Direct CLI or Gem script
  ```bash
  bundle exec vibe-harness .
  ```

### Make
- **Config:** `Makefile`
- **Integration:**
  ```makefile
  lint:
      vibe-harness .
  .PHONY: lint
  ```

## Frameworks

### Rails
- **Build:** Rake + Bundler
- **Integration:**
  ```ruby
  # lib/tasks/vibe.rake
  namespace :vibe do
    task :check do
      sh 'vibe-harness app/'
    end
  end
  ```
  ```bash
  bundle exec rake vibe:check
  ```
- **CI:** Add to CI pipeline before test step

### Sinatra
- **No built-in task runner**
- **Integration:** Rakefile or Makefile

### Hanami
- **Build:** Rake + Bundler
- **Integration:** Same Rake task pattern as Rails

## Existing Linters to Leverage

| Linter | What It Catches | Overlap with Vibe Harness |
|--------|----------------|--------------------------|
| **RuboCop** | Style, complexity, layout, Lint | Method length (Metrics/MethodLength), class length (Metrics/ClassLength), empty rescue (Lint/RescueException), puts (Style/GlobalStdStream) |
| **Brakeman** | Security (Rails-specific) | SQL injection, mass assignment, hardcoded secrets, CSRF |
| **Reek** | Code smells | God class (TooManyInstanceVariables), long method (TooManyStatements), feature envy, data clump |
| **Fasterer** | Performance | Inefficient Ruby idioms |
| **Rails Best Practices** | Rails patterns | Fat controller, model callbacks |
| **Bundler-audit** | Dependency vulnerabilities | Complementary — different domain |
| **Solargraph** | Type checking (LSP) | Missing type annotations |

### Leverage Strategy
- **RuboCop first** for style and conventional linting — it's the standard
- **Brakeman** for Rails security — overlaps with VH-G005 and VH-G011 but is configurable
- **Reek** for design smells — overlaps with VH-G002 and VH-G012 but can be suppressed
- **Vibe Harness adds what they miss:** missing logging in controller actions and service methods, bare rescue detection (stricter than RuboCop), mass assignment without strong parameters, N+1 query patterns
- **RuboCop EmptyRescue** can be configured to match VH-G004 — but VH is non-configurable

### RuboCop Configuration for Maximum Overlap
```yaml
# .rubocop.yml
Metrics/MethodLength:
  Max: 50  # Match VH-G002

Metrics/ClassLength:
  Max: 300  # Match VH-G001

Lint/RescueException:
  Enabled: true

Lint/SuppressedException:
  Enabled: true

Style/GlobalStdStream:
  Enabled: true

Naming/PredicateName:
  Enabled: true

Metrics/AbcSize:
  Max: 30

Metrics/CyclomaticComplexity:
  Max: 10
```