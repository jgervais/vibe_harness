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

func writeTestTOML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	return path
}

func TestValidateTOML_EnabledKey(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["src/**"]

[languages]
".go" = "go"

enabled = false`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for 'enabled' key, got nil")
	}
	if !strings.Contains(err.Error(), "enabled") {
		t.Errorf("expected error to mention 'enabled', got: %v", err)
	}
}

func TestValidateTOML_ThresholdKey(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["src/**"]

[languages]
".go" = "go"

threshold = 500`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for 'threshold' key, got nil")
	}
	if !strings.Contains(err.Error(), "threshold") {
		t.Errorf("expected error to mention 'threshold', got: %v", err)
	}
}

func TestValidateTOML_IgnoreSection(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["src/**"]

[languages]
".go" = "go"

[ignore]
paths = ["vendor/"]`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for '[ignore]' section, got nil")
	}
	if !strings.Contains(err.Error(), "ignore") {
		t.Errorf("expected error to mention 'ignore', got: %v", err)
	}
}

func TestValidateTOML_RulesSection(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["src/**"]

[languages]
".go" = "go"

[rules.VH-G001]
enabled = false`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for '[rules]' section, got nil")
	}
	if !strings.Contains(err.Error(), "rules") {
		t.Errorf("expected error to mention 'rules', got: %v", err)
	}
}

func TestValidateTOML_SeverityKey(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["src/**"]

[languages]
".go" = "go"

severity = "note"`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for 'severity' key, got nil")
	}
	if !strings.Contains(err.Error(), "severity") {
		t.Errorf("expected error to mention 'severity', got: %v", err)
	}
}

func TestValidateTOML_NestedDisallowed(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["src/**"]

[languages]
".go" = "go"

[observability]
logging_calls = ["log"]

[observability.threshold]
value = 500`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for nested 'threshold' key, got nil")
	}
	if !strings.Contains(err.Error(), "threshold") {
		t.Errorf("expected error to mention 'threshold', got: %v", err)
	}
}

func TestValidateTOML_MissingLanguages(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["src/**"]`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for missing '[languages]' section")
	}
	if !strings.Contains(err.Error(), "languages") {
		t.Errorf("expected error to mention 'languages', got: %v", err)
	}
}

func TestValidateTOML_EmptyLanguages(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["src/**"]

[languages]`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for empty '[languages]' section")
	}
	if !strings.Contains(err.Error(), "languages") {
		t.Errorf("expected error to mention 'languages', got: %v", err)
	}
}

func TestValidateTOML_LanguageExtensionFormat(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["src/**"]

[languages]
go = "go"`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for language key without '.' prefix")
	}
	if !strings.Contains(err.Error(), "'.'") {
		t.Errorf("expected error to mention '.' prefix, got: %v", err)
	}
}

func TestValidateTOML_MissingSourceDirs(t *testing.T) {
	path := writeTestTOML(t, `[languages]
".go" = "go"`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for missing 'source_directories'")
	}
	if !strings.Contains(err.Error(), "source_directories") {
		t.Errorf("expected error to mention 'source_directories', got: %v", err)
	}
}

func TestValidateTOML_EmptySourceDirs(t *testing.T) {
	path := writeTestTOML(t, `source_directories = []

[languages]
".go" = "go"`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for empty 'source_directories'")
	}
	if !strings.Contains(err.Error(), "source_directories") {
		t.Errorf("expected error to mention 'source_directories', got: %v", err)
	}
}

func TestValidateTOML_SourceDirsAbsolute(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["/usr/src"]

[languages]
".go" = "go"`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for absolute source_directories")
	}
	if !strings.Contains(err.Error(), "must be a relative path") {
		t.Errorf("expected error about relative path, got: %v", err)
	}
}

func TestValidateTOML_SourceDirsDotDot(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["../src"]

[languages]
".go" = "go"`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for '../' in source_directories")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("expected error about not allowed, got: %v", err)
	}
}

func TestValidateTOML_SourceDirsDot(t *testing.T) {
	path := writeTestTOML(t, `source_directories = ["."]

[languages]
".go" = "go"`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for '.' in source_directories")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("expected error about not allowed, got: %v", err)
	}
}

func TestValidateTOML_SourceDirsEmptyString(t *testing.T) {
	path := writeTestTOML(t, `source_directories = [""]

[languages]
".go" = "go"`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for empty string in source_directories")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Errorf("expected error about empty, got: %v", err)
	}
}

func TestValidateTOML_TestFilePatternValid(t *testing.T) {
	path := writeTestTOML(t, `test_file_pattern = ["_test."]
source_directories = ["src/**"]

[languages]
".go" = "go"`)

	if err := ValidateTOML(path); err != nil {
		t.Errorf("expected test_file_pattern with 'test' to pass, got: %v", err)
	}
}

func TestValidateTOML_TestFilePatternTst(t *testing.T) {
	path := writeTestTOML(t, `test_file_pattern = [".tst."]
source_directories = ["src/**"]

[languages]
".go" = "go"`)

	if err := ValidateTOML(path); err != nil {
		t.Errorf("expected test_file_pattern with 'tst' to pass, got: %v", err)
	}
}

func TestValidateTOML_TestFilePatternNoTest(t *testing.T) {
	path := writeTestTOML(t, `test_file_pattern = [".src."]
source_directories = ["src/**"]

[languages]
".go" = "go"`)

	err := ValidateTOML(path)
	if err == nil {
		t.Fatal("expected error for pattern without 'test' or 'tst'")
	}
	if !strings.Contains(err.Error(), "test_file_pattern") {
		t.Errorf("expected error to mention 'test_file_pattern', got: %v", err)
	}
}

func TestValidateTOML_TestFilePatternTooBroad(t *testing.T) {
	for _, pattern := range []string{"*", "**", ".", "/"} {
		content := `test_file_pattern = ["` + pattern + `"]
source_directories = ["src/**"]

[languages]
".go" = "go"`
		path := writeTestTOML(t, content)
		err := ValidateTOML(path)
		if err == nil {
			t.Errorf("expected error for too-broad pattern %q", pattern)
		}
	}
}

func TestValidateTOML_TestFilePatternArray(t *testing.T) {
	path := writeTestTOML(t, `test_file_pattern = ["_test.", "testdata"]
source_directories = ["src/**"]

[languages]
".go" = "go"`)

	if err := ValidateTOML(path); err != nil {
		t.Errorf("expected array test_file_pattern to pass, got: %v", err)
	}
}

func TestValidateTOML_TestFilePatternStringValid(t *testing.T) {
	path := writeTestTOML(t, `test_file_pattern = "_test."
source_directories = ["src/**"]

[languages]
".go" = "go"`)

	if err := ValidateTOML(path); err != nil {
		t.Errorf("expected string test_file_pattern to pass, got: %v", err)
	}
}
