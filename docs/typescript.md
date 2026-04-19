# TypeScript/JavaScript — Build System & Existing Linters

## Build Systems

### npm
- **Config:** `package.json`
- **Integration:**
  ```json
  {
    "scripts": {
      "lint": "vibe-harness .",
      "lint:ci": "vibe-harness . --format sarif"
    }
  }
  ```
  ```bash
  npm run lint
  ```

### yarn
- **Config:** `package.json` (same as npm)
- **Integration:**
  ```bash
  yarn lint
  ```

### pnpm
- **Config:** `package.json` (same as npm)
- **Integration:**
  ```bash
  pnpm lint
  ```

### bun
- **Config:** `package.json` (same as npm)
- **Integration:**
  ```bash
  bun run lint
  ```

### turborepo / nx
- **Config:** `turbo.json` or `nx.json`
- **Integration:**
  ```json
  {
    "pipeline": {
      "lint": {
        "inputs": ["src/**"],
        "outputs": []
      }
    }
  }
  ```

## Frameworks

### React (Create React App / Vite)
- **Integration:** Add to package.json scripts
  ```json
  {
    "scripts": {
      "lint": "eslint . && vibe-harness src/"
    }
  }
  ```
- Run ESLint for style, Vibe Harness for quality floor

### Next.js
- **Integration:** Add to package.json scripts, run before build
  ```json
  {
    "scripts": {
      "build": "vibe-harness src/ && next build"
    }
  }
  ```

### Express / NestJS
- **No built-in task runner**
- **Integration:** package.json scripts or Makefile

### Angular
- **Build:** `ng build`, `ng lint`
- **Integration:**
  ```json
  {
    "scripts": {
      "lint": "ng lint && vibe-harness src/"
    }
  }
  ```

## Existing Linters to Leverage

| Linter | What It Catches | Overlap with Vibe Harness |
|--------|----------------|--------------------------|
| **ESLint** | Style, complexity, best practices | max-lines, max-lines-per-function, no-empty, no-console, complexity |
| **@typescript-eslint** | TS-specific rules | no-explicit-any, explicit-function-return-type, no-unsafe-assignment |
| **Prettier** | Formatting | No overlap — Prettier is style only |
| **Biome** | Formatting + linting (ESLint replacement) | Similar to ESLint overlap, faster |
| **oxlint** | Fast ESLint alternative | Similar overlap, Rust-based |
| **ts-prune** | Unused exports | Partial overlap with VH-G012 |
| **knip** | Unused files, exports, deps | Partial overlap with VH-G012 |
| **npm audit** | Dependency vulnerabilities | Complementary — different domain |

### Leverage Strategy
- **ESLint/Biome first** for style and conventional linting
- **Vibe Harness adds:** missing logging in I/O functions, console.log detection (stricter than ESLint's no-console), missing error boundaries in React, async without await patterns
- **ESLint max-lines** can be set to match VH-G001 threshold
- **@typescript-eslint/no-explicit-any** catches the `any` type but VH-G010 catches broad catch types that ESLint misses
- **knip** finds dead code — complementary to VH-G007 (duplication) and VH-G012

### ESLint Configuration for Maximum Overlap
```json
{
  "rules": {
    "max-lines": ["error", { "max": 300, "skipBlankLines": true, "skipComments": true }],
    "max-lines-per-function": ["error", { "max": 50, "skipBlankLines": true, "skipComments": true }],
    "no-empty": ["error", { "allowEmptyCatch": false }],
    "no-console": "error",
    "complexity": ["error", { "max": 10 }]
  }
}
```