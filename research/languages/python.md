# Python — AI Code Smells & Tree-Sitter Detection

## 1. Common AI-Generated Python Anti-Patterns

### Error Handling
- **Bare `except:`** — catches everything including KeyboardInterrupt. AI's default when told to "handle errors"
- **`except Exception:` with pass** — swallowed errors, the most common AI smell in Python
- **`except:` then `continue`/`pass`** — silent failure loop, especially in batch processing
- **Missing exception specificity** — `except Exception` instead of `except ValueError`, `except IOError`

### Logging & Observability
- **`print()` instead of `logging`** — AI defaults to print for "debugging" output
- **No logging at all** — functions with I/O but zero log calls
- **Inconsistent log levels** — everything at `logger.info()` or mixing print and logging
- **Missing `__name__` in logger** — `logging.getLogger("myapp")` instead of `logging.getLogger(__name__)`

### Type Hints
- **Missing type annotations** — `def process(data):` instead of `def process(data: dict[str, Any]) -> Result`
- **`Any` overuse** — when AI can't determine a type, it defaults to `Any`
- **Return type missing** — `def calculate(x: int, y: int):` with no `-> int`

### Structural Smells
- **God functions** — single functions doing parsing, validation, transformation, and I/O
- **Mutable default arguments** — `def foo(items=[])` — classic Python bug that AI generates constantly
- **Global state** — module-level dicts/lists used as caches without thread safety
- **Star imports** — `from module import *` — AI copies example code verbatim

### Security
- **SQL string interpolation** — `f"SELECT * FROM users WHERE id = {user_id}"` instead of parameterized queries
- **`eval()` / `exec()`** — AI sometimes generates these for "dynamic" behavior
- **Hardcoded secrets** — API keys, DB passwords in source
- **`pickle.load` on untrusted data** — AI doesn't understand deserialization risks
- **`subprocess` with `shell=True`** — command injection risk

## 2. pip/PyPI Ecosystem — AI Pitfalls

- **Missing `pyproject.toml`** — AI creates scripts without project configuration
- **No virtual environment** — assumes system Python, no isolation
- **Pinned vs unpinned deps** — `requests` without version vs `requests==2.31.0`. AI mixes both
- **`requirements.txt` not sorted** — makes diff reviews harder
- **Dev dependencies in production** — `pytest`, `black` in `requirements.txt` instead of `requirements-dev.txt`
- **Missing `__init__.py`** — AI creates package directories without init files

## 3. Tree-Sitter Python AST — Key Node Types

| Node Type | What It Captures | Use For |
|---|---|---|
| `function_definition` | `def foo():` | God function detection, missing logging |
| `class_definition` | `class Foo:` | God class, missing methods |
| `try_statement` | `try:` block | Error handling presence |
| `except_clause` | `except ValueError:` | Bare except, broad except |
| `with_statement` | `with open(...)` | Resource management |
| `import_statement` | `import os` | Star imports detection |
| `import_from_statement` | `from x import *` | Star imports detection |
| `call` | `foo()` | Logging call detection, print detection |
| `decorator` | `@app.route` | Framework-specific patterns |
| `typed_parameter` | `x: int` | Type annotation presence |
| `type` | Return type annotation | Missing return types |
| `assignment` | `x = ...` | Mutable default detection |
| `fstring` | `f"..."` | SQL injection risk in string patterns |
| `global_statement` | `global x` | Global state detection |
| `assert_statement` | `assert condition` | Using assert for runtime validation |

## 4. Framework-Specific AI Issues

### Django
- **N+1 queries** — AI generates views that query per-object in loops instead of `select_related`/`prefetch_related`
- **Missing `null=True`/`blank=True`** — model fields without proper null handling
- **`objects.all()` instead of filtering** — fetching entire tables
- **Raw SQL instead of ORM** — when AI doesn't know Django ORM well enough
- **Missing migrations** — AI creates models but skips `makemigrations`
- **No `__str__` method** — model classes without string representation

### FastAPI
- **Missing Pydantic models** — route parameters without type validation
- **`Any` in response models** — defeats FastAPI's purpose
- **No dependency injection** — AI inlines DB connections instead of using `Depends()`
- **Synchronous endpoints** — `def` instead of `async def` for I/O-bound routes

### Flask
- **`app.route` with inline logic** — no separation of routes and business logic
- **Missing error handlers** — no `@app.errorhandler` for common HTTP errors
- **No application factory** — `app = Flask(__name__)` at module level instead of `create_app()`

## 5. Detection Rules — Tree-Sitter Queries

### Bare Except
```
(except_clause (wildcard_import))  → except: (no specific exception)
```

### Swallowed Exception (pass in except)
```
(except_clause body: (block (pass_statement)))
```

### Print Instead of Logging
```
(call function: (identifier) @name (#eq? @name "print"))
```

### Mutable Default Argument
```
(function_definition
  parameters: (parameters
    (default_parameter
      value: [(list) (dict) (set)])))
```

### Missing Type Annotations
```
(function_definition
  parameters: (parameters (identifier)))  → no type on parameter
(function_definition return_type: (_))     → HAS return type (absence = smell)
```

### Star Import
```
(import_from_statement (wildcard_import))
```

## 6. Quick Reference

| Smell | AST Signal | Detection |
|---|---|---|
| Bare except | `except_clause` with `wildcard_import` | Exact match |
| Swallowed error | `except_clause` → `pass_statement` | Exact match |
| Print not logging | `call` to `print` identifier | Name match |
| Mutable default | `default_parameter` with list/dict/set | Type match |
| Missing types | `function_definition` without `typed_parameter` | Absence |
| God function | Statement count in `function_definition` body | Threshold |
| Star import | `wildcard_import` | Exact match |
| f-string SQL | `fstring` containing SQL keywords | Pattern match |