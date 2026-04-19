# Monorepo Structure

Everything lives in one repo: `github.com/jgervais/vibe_harness`. One `git tag` triggers all builds and publishes all packages.

## Directory Layout

```
vibe_harness/
├── cmd/
│   └── vibe-harness/           # Go CLI entry point
│       └── main.go
├── internal/                    # Go internal packages
│   ├── checks/                 # Check implementations
│   │   ├── generic/            # VH-G001 through VH-G012
│   │   └── languages/          # Language-specific checks
│   ├── parser/                 # Tree-sitter integration
│   ├── config/                 # .vibe_harness.toml loading + validation
│   ├── output/                 # JSON, SARIF, human-readable formatters
│   └── rules/                  # Rule definitions and thresholds
├── pkg/                         # Go public packages (if any)
│   └── rules/                   # Rule registry (for plugin system later)
├── grammars/                    # Tree-sitter grammar sources
│   ├── python/
│   ├── typescript/
│   ├── go/
│   ├── java/
│   ├── ruby/
│   └── rust/
├── dist/                        # Package manager wrappers
│   ├── npm/                    # npm wrapper package
│   │   ├── package.json
│   │   ├── index.js            # CLI entry (exec's the binary)
│   │   ├── install.js          # postinstall: download binary
│   │   └── platforms.json      # OS/arch → download URL mapping
│   ├── pypi/                   # PyPI wrapper package
│   │   ├── pyproject.toml
│   │   ├── vibe_harness/
│   │   │   ├── __init__.py
│   │   │   └── cli.py          # CLI entry (exec's the binary)
│   │   └── install.py          # Downloads binary on install
│   ├── homebrew/               # Homebrew formula
│   │   └── vibe-harness.rb     # Copied to homebrew-tap on release
│   ├── gradle/                 # Gradle plugin
│   │   ├── build.gradle.kts
│   │   └── src/
│   │       └── main/kotlin/
│   │           └── VibeHarnessPlugin.kt
│   └── cargo/                  # Cargo subcommand (optional)
│       └── Cargo.toml           # cargo-vibe-harness binary
├── .github/
│   ├── workflows/
│   │   ├── ci.yml              # PR checks: test, lint, build
│   │   ├── release.yml         # Tag-triggered: build all platforms + publish
│   │   └── update-formula.yml  # Sync homebrew formula to tap repo
│   └── dependabot.yml
├── docs/                        # Documentation
│   ├── generic-checks.md
│   ├── roadmap.md
│   ├── monorepo.md             # This file
│   ├── python.md
│   ├── typescript.md
│   ├── go.md
│   ├── java.md
│   ├── ruby.md
│   └── rust.md
├── research/                    # Research artifacts
│   ├── code_smells.md
│   ├── effectiveness_constraints.md
│   ├── existing_tools.md
│   ├── build_integration.md
│   └── languages/
├── .specify/                    # SpecKit
├── .opencode/                   # OpenCode integration
├── go.mod
├── go.sum
├── Makefile                     # Build, test, lint targets
├── README.md
└── LICENSE
```

## Release Workflow

One tag push triggers everything:

```
git tag v0.1.0 && git push --tags
```

```yaml
# .github/workflows/release.yml (simplified)
on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        include:
          - goos: darwin, goarch: arm64
          - goos: darwin, goarch: amd64
          - goos: linux,  goarch: arm64
          - goos: linux,  goarch: amd64
          - goos: windows, goarch: amd64
    steps:
      - run: CGO_ENABLED=0 go build -ldflags="-s -w" -o vibe-harness-${{ matrix.goos }}-${{ matrix.goarch }}
      - uses: actions/upload-artifact@v4

  release:
    needs: build
    steps:
      - uses: actions/download-artifact@v4
      - run: shasum -a 256 vibe-harness-* > checksums.txt
      - uses: softprops/action-gh-release@v2  # Creates GitHub Release with all binaries

  publish-npm:
    needs: release
    steps:
      - run: cd dist/npm && npm publish

  publish-pypi:
    needs: release
    steps:
      - run: cd dist/pypi && python -m build && twine upload dist/*

  publish-homebrew:
    needs: release
    steps:
      - run: |
          # Copy formula to homebrew-tap repo with updated version + sha256
          cp dist/homebrew/vibe-harness.rb ../homebrew-tap/Formula/vibe-harness.rb
          cd ../homebrew-tap && git add . && git commit -m "vibe-harness $VERSION" && git push

  publish-gradle:
    needs: release
    steps:
      - run: cd dist/gradle && ./gradlew publishPlugin
```

## Package Manager Details

### npm (`dist/npm/`)

The npm package is a thin wrapper. It contains no linting logic — just downloads the native binary and exec's it.

**package.json:**
```json
{
  "name": "vibe-harness",
  "version": "0.1.0",
  "description": "AI code quality floor — non-configurable linter",
  "bin": { "vibe-harness": "index.js" },
  "scripts": { "postinstall": "node install.js" },
  "os": ["darwin", "linux", "win32"],
  "cpu": ["x64", "arm64"],
  "files": ["index.js", "install.js", "platforms.json"]
}
```

**install.js** detects `process.platform` + `process.arch`, downloads from GitHub Releases, saves to `node_modules/.cache/vibe-harness/bin/`.

**index.js:**
```js
const { execFileSync } = require('child_process');
const bin = require('path').join(__dirname, 'bin', process.platform, process.arch, 'vibe-harness');
execFileSync(bin, process.argv.slice(2), { stdio: 'inherit' });
```

This is the same pattern used by esbuild, Biome, @swc/core, and Turbo.

### PyPI (`dist/pypi/`)

Two options:

**Option A: Binary wheels (recommended)**
Build platform-specific `.whl` files with the binary baked in. Each wheel targets `manylinux`, `macos`, or `win`. `pip install vibe-harness` just works — no download step.

```bash
# Build wheels for each platform
python -m build --wheel --config-setting=--platform=manylinux2014_x86_64
python -m build --wheel --config-setting=--platform=macosx_arm64
```

This is how Ruff and uv distribute their Rust binaries.

**Option B: Download wrapper (simpler)**
Python package that downloads the binary on first run. Less polished but faster to implement.

### Homebrew (`dist/homebrew/`)

The formula lives in this repo but gets **copied to a separate tap repo** on release. Homebrew requires formulas in a git repo with the right structure.

**Formula:**
```ruby
class VibeHarness < Formula
  desc "Non-configurable quality floor for AI-generated code"
  homepage "https://github.com/jgervais/vibe_harness"
  version "0.1.0"

  on_macos do
    on_arm do
      url "https://github.com/jgervais/vibe_harness/releases/download/v0.1.0/vibe-harness-darwin-arm64"
      sha256 "..."
    end
    on_intel do
      url "https://github.com/jgervais/vibe_harness/releases/download/v0.1.0/vibe-harness-darwin-amd64"
      sha256 "..."
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/jgervais/vibe_harness/releases/download/v0.1.0/vibe-harness-linux-arm64"
      sha256 "..."
    end
    on_intel do
      url "https://github.com/jgervais/vibe_harness/releases/download/v0.1.0/vibe-harness-linux-amd64"
      sha256 "..."
    end
  end

  def install
    bin.install Dir["vibe-harness-*"].first => "vibe-harness"
  end

  test do
    system "#{bin}/vibe-harness", "--version"
  end
end
```

**Tap repo:** `github.com/jgervais/homebrew-tap` (separate repo, auto-updated by CI)

**User install:** `brew install jgervais/tap/vibe-harness`

### Gradle Plugin (`dist/gradle/`)

A Gradle plugin that downloads the binary and creates a `vibeCheck` task.

```kotlin
// src/main/kotlin/VibeHarnessPlugin.kt
class VibeHarnessPlugin : Plugin<Project> {
    override fun apply(project: Project) {
        project.tasks.register("vibeCheck", Exec::class) {
            val binary = downloadBinary() // Downloads if not cached
            commandLine(binary.absolutePath, project.projectDir.absolutePath)
        }
        project.tasks.named("check").configure { dependsOn("vibeCheck") }
    }
}
```

**Publishing:** Push to [plugins.gradle.org](https://plugins.gradle.org) via `gradle publishPlugin`.

**User install:**
```kotlin
// build.gradle.kts
plugins {
    id("com.vibeharness.plugin") version "0.1.0"
}
```

### Cargo (`dist/cargo/`)

A separate binary crate named `cargo-vibe-harness`. When installed, enables `cargo vibe-harness` as a subcommand.

```toml
# dist/cargo/Cargo.toml
[package]
name = "cargo-vibe-harness"
version = "0.1.0"

[[bin]]
name = "cargo-vibe-harness"
path = "src/main.rs"
```

**Publishing:** `cargo publish` to crates.io.

**User install:** `cargo install cargo-vibe-harness` (requires Rust toolchain).

**Note:** This is a convenience for Rust teams only. Most users will use the direct binary download.

## Version Management

All packages share the same version number, derived from the git tag:

- Git tag `v0.1.0` → npm `0.1.0`, PyPI `0.1.0`, Homebrew `0.1.0`, etc.
- The release workflow sets the version in all `package.json`, `pyproject.toml`, formula, and plugin metadata from the tag

## Anti-Tampering

The release workflow includes:

1. **Binary checksums** — `checksums.txt` published alongside binaries
2. **Rules hash** — compiled into the binary at build time, printed in output so CI can detect binary tampering
3. **Signed releases** — GitHub releases with provenance (Sigstore/cosign, optional)
4. **No suppress mechanisms** — the binary has no `--disable`, `--skip`, or inline comment suppression flags