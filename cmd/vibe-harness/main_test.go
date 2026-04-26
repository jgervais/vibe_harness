package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	repoRoot := filepath.Join("..", "..")
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve repo root: %v\n", err)
		os.Exit(1)
	}
	binaryPath = filepath.Join(absRoot, "test-vibe-harness")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/vibe-harness")
	cmd.Dir = absRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build test binary: %v\n%s\n", err, out)
		os.Exit(1)
	}

	code := m.Run()
	os.Remove(binaryPath)
	os.Exit(code)
}

func run(args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	} else {
		exitCode = 0
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestVersionOutput(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}
	binPath := filepath.Join(absRoot, "test-vibe-harness-versioned")

	cmd := exec.Command("go", "build",
		"-ldflags", "-X main.version=0.1.0 -X main.rulesHash=abc123",
		"-o", binPath,
		"./cmd/vibe-harness",
	)
	cmd.Dir = absRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build versioned binary: %v\n%s", err, out)
	}
	defer os.Remove(binPath)

	out, _ := exec.Command(binPath, "--version").CombinedOutput()

	output := string(out)
	if !strings.Contains(output, "vibe-harness v0.1.0") {
		t.Errorf("expected version format, got: %s", output)
	}
	if !strings.Contains(output, "rules hash: abc123") {
		t.Errorf("expected rules hash, got: %s", output)
	}
	if !strings.Contains(output, runtime.GOOS) || !strings.Contains(output, runtime.GOARCH) {
		t.Errorf("expected os/arch in version, got: %s", output)
	}
}

func TestHelpFlag(t *testing.T) {
	_, stderr, exitCode := run("--help")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	lower := strings.ToLower(stderr)
	if !strings.Contains(lower, "usage") && !strings.Contains(lower, "flags") {
		t.Fatalf("expected output to contain 'Usage' or 'flags', got: %s", stderr)
	}
}

func TestVersionFlag(t *testing.T) {
	stdout, _, exitCode := run("--version")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "vibe-harness vdev") {
		t.Fatalf("expected output to contain 'vibe-harness vdev', got: %s", stdout)
	}
	if !strings.Contains(stdout, "rules hash: unknown") {
		t.Fatalf("expected output to contain 'rules hash: unknown', got: %s", stdout)
	}
}

func TestNoPathArgument(t *testing.T) {
	_, _, exitCode := run()
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
}

func TestInvalidFormat(t *testing.T) {
	_, stderr, exitCode := run("--format", "xml", ".")
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(stderr, "invalid format") {
		t.Fatalf("expected stderr to contain 'invalid format', got: %s", stderr)
	}
}

func TestScanTestdata(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "hardcoded_secrets")
	configPath := filepath.Join("..", "..", "testdata", "config", "valid.toml")
	_, stderr, exitCode := run("--config", configPath, testdataPath)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, " — ") {
		t.Fatalf("expected stderr to contain violation format, got: %s", stderr)
	}
}

func TestScanEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join("..", "..", "testdata", "config", "valid.toml")
	_, _, exitCode := run("--config", configPath, tmpDir)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
}

func TestScanNonExistentPath(t *testing.T) {
	_, _, exitCode := run("/non/existent/path/that/does/not/exist")
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
}

func TestScanJSONFormat(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "hardcoded_secrets")
	configPath := filepath.Join("..", "..", "testdata", "config", "valid.toml")
	stdout, _, exitCode := run("--config", configPath, "--format", "json", testdataPath)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", err, stdout)
	}
	for _, key := range []string{"version", "tool", "results"} {
		if _, ok := parsed[key]; !ok {
			t.Fatalf("JSON missing key %q, keys: %v", key, parsed)
		}
	}
}

func TestScanSARIFFormat(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "hardcoded_secrets")
	configPath := filepath.Join("..", "..", "testdata", "config", "valid.toml")
	stdout, _, exitCode := run("--config", configPath, "--format", "sarif", testdataPath)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", err, stdout)
	}
	for _, key := range []string{"$schema", "runs"} {
		if _, ok := parsed[key]; !ok {
			t.Fatalf("SARIF JSON missing key %q, keys: %v", key, parsed)
		}
	}
	if v, _ := parsed["version"].(string); v != "2.1.0" {
		t.Fatalf("expected SARIF version 2.1.0, got %v", parsed["version"])
	}
}

func TestConfigValid(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}
	configPath := filepath.Join(absRoot, "testdata", "config", "valid.toml")
	tmpDir := t.TempDir()
	_, _, exitCode := run("--config", configPath, tmpDir)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0 with valid config, got %d", exitCode)
	}
}

func TestConfigInvalid(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}
	configPath := filepath.Join(absRoot, "testdata", "config", "invalid_rule_mod.toml")
	_, stderr, exitCode := run("--config", configPath, ".")
	if exitCode != 2 {
		t.Fatalf("expected exit code 2 with invalid config, got %d", exitCode)
	}
	if !strings.Contains(stderr, "invalid configuration") {
		t.Fatalf("expected stderr to contain 'invalid configuration', got: %s", stderr)
	}
}

func TestConfigMissing(t *testing.T) {
	tmpDir := t.TempDir()
	_, stderr, exitCode := run(tmpDir)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2 when no config file found, got %d", exitCode)
	}
	if !strings.Contains(stderr, ".vibe_harness.toml") {
		t.Fatalf("expected stderr to mention .vibe_harness.toml, got: %s", stderr)
	}
}