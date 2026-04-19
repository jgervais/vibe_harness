package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ValidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".vibe_harness.toml")
	content := `
[observability]
logging_calls = ["custom_log"]
metrics_calls = ["custom_metric"]

[languages]
".py" = "python3"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if len(cfg.Observability.LoggingCalls) != 1 || cfg.Observability.LoggingCalls[0] != "custom_log" {
		t.Errorf("expected LoggingCalls = [\"custom_log\"], got %v", cfg.Observability.LoggingCalls)
	}
	if len(cfg.Observability.MetricsCalls) != 1 || cfg.Observability.MetricsCalls[0] != "custom_metric" {
		t.Errorf("expected MetricsCalls = [\"custom_metric\"], got %v", cfg.Observability.MetricsCalls)
	}
	if cfg.Languages[".py"] != "python3" {
		t.Errorf("expected Languages[\".py\"] = \"python3\", got %q", cfg.Languages[".py"])
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/.vibe_harness.toml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadConfig_MalformedTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".vibe_harness.toml")
	content := `this is not [valid toml = `
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for malformed TOML, got nil")
	}
}

func TestLoadConfig_DefaultsPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".vibe_harness.toml")
	content := `
[languages]
".py" = "python3"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if len(cfg.Observability.LoggingCalls) == 0 {
		t.Error("expected default LoggingCalls to be preserved, got empty slice")
	}
	if len(cfg.Observability.MetricsCalls) == 0 {
		t.Error("expected default MetricsCalls to be preserved, got empty slice")
	}
}

func TestAutoDiscoverConfig_Found(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".vibe_harness.toml")
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	found, err := AutoDiscoverConfig(dir)
	if err != nil {
		t.Fatalf("AutoDiscoverConfig returned error: %v", err)
	}
	if found != configPath {
		t.Errorf("expected %q, got %q", configPath, found)
	}
}

func TestAutoDiscoverConfig_NotFound(t *testing.T) {
	dir := t.TempDir()

	found, err := AutoDiscoverConfig(dir)
	if err != nil {
		t.Fatalf("AutoDiscoverConfig returned error: %v", err)
	}
	if found != "" {
		t.Errorf("expected empty string, got %q", found)
	}
}

func TestAutoDiscoverConfig_WalksUp(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, ".vibe_harness.toml")
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	subdir := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	found, err := AutoDiscoverConfig(subdir)
	if err != nil {
		t.Fatalf("AutoDiscoverConfig returned error: %v", err)
	}
	if found != configPath {
		t.Errorf("expected %q, got %q", configPath, found)
	}
}

func TestDefaultConfig_LoggingCalls(t *testing.T) {
	cfg := DefaultConfig()
	expected := []string{"log", "logger", "logging", "tracing", "slog", "logr"}
	if len(cfg.Observability.LoggingCalls) != len(expected) {
		t.Fatalf("expected %d LoggingCalls, got %d", len(expected), len(cfg.Observability.LoggingCalls))
	}
	for i, v := range expected {
		if cfg.Observability.LoggingCalls[i] != v {
			t.Errorf("LoggingCalls[%d]: expected %q, got %q", i, v, cfg.Observability.LoggingCalls[i])
		}
	}
}

func TestDefaultConfig_MetricsCalls(t *testing.T) {
	cfg := DefaultConfig()
	expected := []string{"metrics", "counter", "histogram", "gauge", "timer", "prometheus"}
	if len(cfg.Observability.MetricsCalls) != len(expected) {
		t.Fatalf("expected %d MetricsCalls, got %d", len(expected), len(cfg.Observability.MetricsCalls))
	}
	for i, v := range expected {
		if cfg.Observability.MetricsCalls[i] != v {
			t.Errorf("MetricsCalls[%d]: expected %q, got %q", i, v, cfg.Observability.MetricsCalls[i])
		}
	}
}

func TestDefaultConfig_Languages(t *testing.T) {
	cfg := DefaultConfig()
	expected := map[string]string{
		".py":  "python",
		".ts":  "typescript",
		".tsx": "typescript",
		".js":  "javascript",
		".go":  "go",
		".java": "java",
		".rb":  "ruby",
		".rs":  "rust",
	}
	if len(cfg.Languages) != len(expected) {
		t.Fatalf("expected %d Languages entries, got %d", len(expected), len(cfg.Languages))
	}
	for k, v := range expected {
		if cfg.Languages[k] != v {
			t.Errorf("Languages[%q]: expected %q, got %q", k, v, cfg.Languages[k])
		}
	}
}

func TestDefaultConfig_DefensiveCopy(t *testing.T) {
	cfg1 := DefaultConfig()
	cfg1.Observability.LoggingCalls[0] = "MODIFIED"
	cfg1.Observability.MetricsCalls[0] = "MODIFIED"
	cfg1.Languages[".py"] = "MODIFIED"

	cfg2 := DefaultConfig()
	if cfg2.Observability.LoggingCalls[0] == "MODIFIED" {
		t.Error("modifying returned config affected future DefaultConfig() LoggingCalls")
	}
	if cfg2.Observability.MetricsCalls[0] == "MODIFIED" {
		t.Error("modifying returned config affected future DefaultConfig() MetricsCalls")
	}
	if cfg2.Languages[".py"] == "MODIFIED" {
		t.Error("modifying returned config affected future DefaultConfig() Languages")
	}
}