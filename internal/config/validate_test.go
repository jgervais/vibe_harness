package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateTOML_ValidConfig(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "config", "valid.toml")
	if err := ValidateTOML(path); err != nil {
		t.Fatalf("expected valid config to pass, got error: %v", err)
	}
}

func TestValidateTOML_EnabledKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	content := `enabled = false` + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for 'enabled' key, got nil")
	}
	if !strings.Contains(err.Error(), "enabled") {
		t.Errorf("expected error to mention 'enabled', got: %v", err)
	}
}

func TestValidateTOML_ThresholdKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	content := `threshold = 500` + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for 'threshold' key, got nil")
	}
	if !strings.Contains(err.Error(), "threshold") {
		t.Errorf("expected error to mention 'threshold', got: %v", err)
	}
}

func TestValidateTOML_IgnoreSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	content := `
[ignore]
paths = ["vendor/"]
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for '[ignore]' section, got nil")
	}
	if !strings.Contains(err.Error(), "ignore") {
		t.Errorf("expected error to mention 'ignore', got: %v", err)
	}
}

func TestValidateTOML_RulesSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	content := `
[rules.VH-G001]
enabled = false
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for '[rules]' section, got nil")
	}
	if !strings.Contains(err.Error(), "rules") {
		t.Errorf("expected error to mention 'rules', got: %v", err)
	}
}

func TestValidateTOML_SeverityKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	content := `severity = "note"` + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for 'severity' key, got nil")
	}
	if !strings.Contains(err.Error(), "severity") {
		t.Errorf("expected error to mention 'severity', got: %v", err)
	}
}

func TestValidateTOML_InvalidRuleModFixture(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "config", "invalid_rule_mod.toml")
	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected invalid_rule_mod.toml to be rejected, got nil")
	}
}

func TestValidateTOML_NestedDisallowed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	content := `
[observability]
logging_calls = ["log"]

[observability.threshold]
value = 500
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for nested 'threshold' key, got nil")
	}
	if !strings.Contains(err.Error(), "threshold") {
		t.Errorf("expected error to mention 'threshold', got: %v", err)
	}
}