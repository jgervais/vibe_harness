# Python — Build System & Existing Linters

## Build Systems

### pip + setuptools
- **Config:** `pyproject.toml`, `setup.py`, `setup.cfg`
- **Run lint:** `pip run lint` or direct CLI invocation
- **Integration:**
  ```toml
  [project.scripts]
  lint = "vibe-harness ."
  ```
  Or add to scripts:
  ```toml
  [tool.setuptools]
  script-files = ["scripts/lint.sh"]
  ```

### pip + requirements.txt
- **Run lint:** Add `vibe-harness` to requirements, run directly
  ```bash
  pip install vibe-harness
  vibe-harness .
  ```
- **CI:** `vibe-harness .` in pipeline step

### Poetry
- **Config:** `pyproject.toml` under `[tool.poetry.scripts]`
- **Integration:**
  ```toml
  [tool.poetry.scripts]
  lint = "vibe-harness ."
  ```
  ```bash
  poetry run lint
  ```

### pipenv
- **Config:** `Pipfile` scripts section
- **Integration:**
  ```toml
  [scripts]
  lint = "vibe-harness ."
  ```
  ```bash
  pipenv run lint
  ```

### uv
- **Config:** `pyproject.toml` with uv scripts
- **Integration:**
  ```bash
  uv add --dev vibe-harness
  uv run vibe-harness .
  ```

## Frameworks

### Django
- **Build:** `manage.py` commands
- **Integration:** Custom management command
  ```python
  # management/commands/vibe_check.py
  from django.core.management.base import BaseCommand
  import subprocess

  class Command(BaseCommand):
      def handle(self, *args, **options):
          subprocess.run(["vibe-harness", "."])
  ```
  ```bash
  python manage.py vibe_check
  ```
- **CI:** Direct `vibe-harness .` call in CI before test step

### FastAPI
- **No built-in task runner**
- **Integration:** Use `pyproject.toml` scripts or Makefile
  ```bash
  # Makefile
  lint:
      vibe-harness src/
  ```

### Flask
- **No built-in task runner**
- **Integration:** Same as FastAPI — pyproject.toml or Makefile

## Existing Linters to Leverage

| Linter | What It Catches | Overlap with Vibe Harness |
|--------|----------------|--------------------------|
| **Ruff** | Style, complexity, imports, unused vars, pyflakes | File length (C0301), function length (C0302), bare except (E722), unused imports (F401) |
| **Pylint** | Complexity, design, metrics | God class (R0902), too many methods (R0904), empty docstring (C0114), broad except (W0703) |
| **mypy** | Type checking | Missing type hints, `Any` usage |
| **bandit** | Security | Hardcoded secrets (S105, S106), SQL injection (S608), eval usage (S307) |
| **flake8-bugbear** | Bug patterns | Mutable defaults (B006), assert in production (B101) |

### Leverage Strategy
- **Run Ruff first** for style/basic checks — it's fast and catches low-hanging fruit
- **Vibe Harness adds what Ruff misses:** missing logging, missing error handling context, observability gaps, magic values in context
- **bandit complements** for security — but VH-G005 (hardcoded secrets) and VH-G011 (disabled security) overlap
- **mypy catches type gaps** — VH doesn't duplicate type checking, but VH-G006 (magic values) catches untyped constants that mypy won't flag

### Ruff Configuration for Maximum Overlap
```toml
[tool.ruff]
line-length = 300  # Match VH-G001 threshold
select = ["E", "F", "W", "C90", "S", "B"]
ignore = []  # Don't disable anything — let Ruff catch what it can

[tool.ruff.mccabe]
max-complexity = 10
```