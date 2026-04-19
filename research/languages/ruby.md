# Ruby — AI Code Smells & Tree-Sitter Detection

## 1. Common AI-Generated Ruby Anti-Patterns

### Error Handling
- **Bare `rescue`** — `rescue` without specifying exception type (catches StandardError by default, not what you usually want)
- **Swallowed exceptions** — `rescue => e; end` with empty body or only comment
- **`rescue Exception`** — catches everything including SystemExit and Interrupt
- **Missing `ensure`** — cleanup code (file closes, connection returns) not in ensure block
- **`begin`/`rescue` around single line** — over-cautious error handling on trivial operations

### Logging & Observability
- **`puts` instead of Rails logger** — `puts "processing order"` instead of `Rails.logger.info`
- **No structured logging** — string concatenation in log messages instead of key-value pairs
- **Missing request tracing** — no Lograge or custom request ID logging in Rails
- **`p` for debugging** — `p object` left in production code

### Structural Smells
- **God methods** — single methods spanning 50+ lines, especially in controllers
- **Fat controllers** — business logic in controllers instead of models/services
- **Callback soup** — `before_save`, `after_commit` chains that trigger cascading side effects
- **Concerns as dumping grounds** — `include EverythingModule` instead of proper decomposition
- **Class methods where instances belong** — `self.method` overuse

### Security
- **SQL string interpolation** — `where("name = '#{params[:name]}'")` instead of parameterized queries
- **`eval` / `send` with user input** — dynamic dispatch on untrusted data
- **Mass assignment** — `Model.create(params)` without strong parameters
- **Hardcoded secrets** — API keys in initializers, `secrets.yml` with actual values

## 2. RubyGems/Bundler — AI Pitfalls

- **No `Gemfile.lock`** — AI generates Gemfile but doesn't run `bundle install`
- **Version constraints missing** — `gem 'rails'` without version, `gem 'sidekiq', '~> 6.0'`
- **Dev gems in production group** — `gem 'pry'` in default group instead of `group :development`
- **Platform-specific gems** — missing `platforms` declarations for OS-specific dependencies
- **Mixed gem sources** — both `rubygems.org` and private repos without proper source blocks

## 3. Tree-Sitter Ruby AST — Key Node Types

| Node Type | What It Captures | Use For |
|---|---|---|
| `method` | `def foo` | God method detection |
| `singleton_method` | `def self.foo` | Class method overuse |
| `class` | `class Foo` | God class detection |
| `module` | `module Foo` | Module organization |
| `rescue` | `rescue` clause | Bare rescue, swallowed errors |
| `ensure` | `ensure` block | Missing cleanup |
| `begin` | `begin` block | Error handling structure |
| `call` | `foo.bar` | Logging call detection |
| `string_interpolation` | `"#{x}"` | SQL injection risk |
| `if` | `if` statement | Complexity detection |
| `block` | `do \|x\|` | Block patterns |
| `lambda` / `proc` | `-> {}` / `Proc.new` | Functional patterns |
| `accessor` | `attr_accessor` | Class design |
| `constant` | `CONSTANT` | Magic constant detection |
| `hash` | `{ key: value }` | Options hash patterns |
| `yield` | `yield` | Method yield patterns |

## 4. Framework-Specific AI Issues

### Rails
- **N+1 queries** — `Post.all.each { |p| p.comments }` instead of `includes(:comments)`
- **Logic in views** — business logic in `.erb` templates instead of helpers/decorators
- **Missing strong parameters** — `permit` not called on controller params
- **`skip_before_action :verify_authenticity_token`** — disabling CSRF protection
- **No pagination** — `Model.all` instead of `Model.page(params[:page])`
- **Missing index on foreign keys** — AI creates migrations without adding DB indexes
- **God controllers** — 300+ line controllers with 20 actions
- **Callbacks instead of service objects** — `before_save :do_ten_things` instead of extracted service
- **Missing `dependent: :destroy`** — has_many without cascade deletion

### Sinatra
- **No error handlers** — missing `error` blocks
- **Inline everything** — routes with full business logic
- **No middleware** — missing Rack middleware for logging, auth, etc.

## 5. Detection Rules — Tree-Sitter Queries

### Bare Rescue
```
(rescue (rescue_clause))  → rescue without specific exception class
```
Look for `rescue` nodes without an exception type reference.

### Swallowed Exception
```
(rescue body: (body_statement)) where body is empty or contains only comment
```

### puts Instead of Logger
```
(call method: (identifier) @name (#match? @name "^(puts|p|pp|print)$"))
```

### SQL String Interpolation
```
(string_interpolation) inside call expressions to known DB methods (where, find_by_sql, execute)
```

### Mass Assignment Risk
```
(call
  method: (identifier) @method (#match? @method "^(create|update|new)$")
  arguments: (argument_list (identifier) @arg))
  → where @arg is not result of `.permit`
```

## 6. Quick Reference

| Smell | AST Signal | Detection |
|---|---|---|
| Bare rescue | `rescue` without exception type | Absence |
| Swallowed error | `rescue` with empty body | Body empty |
| puts/p debug | `call` to `puts`, `p`, `pp` | Name match |
| God method | Statement count in `method` body | Threshold |
| SQL interpolation | `string_interpolation` in DB call | Pattern context |
| Missing ensure | `begin`/`rescue` without `ensure` | Absence |
| Class method overuse | `singleton_method` count ratio | Ratio |
| Fat controller | Method count in Rails controller class | Threshold |