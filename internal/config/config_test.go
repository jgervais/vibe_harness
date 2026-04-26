package config

import (
	"os"
	"path/filepath"
	"testing"
)

var validConfigContent = `
source_directories = ["src/**"]
test_file_pattern = ["_test.", "testdata"]

[observability]
logging_calls = ["custom_log"]
metrics_calls = ["custom_metric"]

[languages]
".py" = "python3"
".go" = "go"
`

func TestLoadConfig_ValidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".vibe_harness.toml")
	if err := os.WriteFile(path, []byte(validConfigContent), 0644); err != nil {
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
	if len(cfg.Languages) != 2 {
		t.Errorf("expected 2 languages, got %d", len(cfg.Languages))
	}
	if len(cfg.SourceDirs) != 1 || cfg.SourceDirs[0] != "src/**" {
		t.Errorf("expected SourceDirs = [\"src/**\"], got %v", cfg.SourceDirs)
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

func TestLoadConfig_ReplacesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".vibe_harness.toml")
	content := `
source_directories = ["cmd"]
test_file_pattern = [".spec."]

[languages]
".go" = "go"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if len(cfg.Languages) != 1 || cfg.Languages[".go"] != "go" {
		t.Errorf("expected only .go language, got %v", cfg.Languages)
	}
	if len(cfg.SourceDirs) != 1 || cfg.SourceDirs[0] != "cmd" {
		t.Errorf("expected SourceDirs = [\"cmd\"], got %v", cfg.SourceDirs)
	}
	if len(cfg.TestFilePattern) != 1 || cfg.TestFilePattern[0] != ".spec." {
		t.Errorf("expected TestFilePattern = [\".spec.\"], got %v", cfg.TestFilePattern)
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

func TestMergedLoggingCalls_DefaultsOnly(t *testing.T) {
	cfg := Config{}
	result := cfg.MergedLoggingCalls("python")

	expected := []string{"log", "logger", "logging", "print"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d results, got %d: %v", len(expected), len(result), result)
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("result[%d]: expected %q, got %q", i, v, result[i])
		}
	}
}

func TestMergedLoggingCalls_AdditiveMerge(t *testing.T) {
	cfg := Config{
		Observability: ObservabilityConfig{
			LoggingCalls: []string{"my_custom_logger"},
		},
	}
	result := cfg.MergedLoggingCalls("python")

	hasCustom := false
	for _, v := range result {
		if v == "my_custom_logger" {
			hasCustom = true
			break
		}
	}
	if !hasCustom {
		t.Error("expected user-provided hint to appear in merged result")
	}

	hasDefault := false
	for _, v := range result {
		if v == "log" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		t.Error("expected language default hint to appear in merged result")
	}
}

func TestMergedLoggingCalls_Deduplication(t *testing.T) {
	cfg := Config{
		Observability: ObservabilityConfig{
			LoggingCalls: []string{"log", "my_custom_logger"},
		},
	}
	result := cfg.MergedLoggingCalls("python")

	count := 0
	for _, v := range result {
		if v == "log" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 'log' to appear exactly once, got %d", count)
	}
}

func TestMergedLoggingCalls_UnknownLanguage(t *testing.T) {
	cfg := Config{
		Observability: ObservabilityConfig{
			LoggingCalls: []string{"my_logger"},
		},
	}
	result := cfg.MergedLoggingCalls("brainfuck")

	if len(result) != 1 || result[0] != "my_logger" {
		t.Errorf("expected [\"my_logger\"], got %v", result)
	}
}

func TestMergedLoggingCalls_UnknownLanguageEmpty(t *testing.T) {
	cfg := Config{}
	result := cfg.MergedLoggingCalls("brainfuck")

	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestMergedMetricsCalls_DefaultsOnly(t *testing.T) {
	cfg := Config{}
	result := cfg.MergedMetricsCalls("python")

	expected := []string{"metrics", "counter", "histogram", "gauge", "timer"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d results, got %d: %v", len(expected), len(result), result)
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("result[%d]: expected %q, got %q", i, v, result[i])
		}
	}
}

func TestMergedMetricsCalls_AdditiveMerge(t *testing.T) {
	cfg := Config{
		Observability: ObservabilityConfig{
			MetricsCalls: []string{"my_custom_metric"},
		},
	}
	result := cfg.MergedMetricsCalls("go")

	hasCustom := false
	for _, v := range result {
		if v == "my_custom_metric" {
			hasCustom = true
			break
		}
	}
	if !hasCustom {
		t.Error("expected user-provided hint to appear in merged result")
	}

	hasDefault := false
	for _, v := range result {
		if v == "prometheus" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		t.Error("expected language default hint to appear in merged result")
	}
}

func TestMergedMetricsCalls_UnknownLanguage(t *testing.T) {
	cfg := Config{
		Observability: ObservabilityConfig{
			MetricsCalls: []string{"my_metric"},
		},
	}
	result := cfg.MergedMetricsCalls("brainfuck")

	if len(result) != 1 || result[0] != "my_metric" {
		t.Errorf("expected [\"my_metric\"], got %v", result)
	}
}

type configBuilder struct {
	SourceDirs      []string
	Languages       map[string]string
	TestFilePattern []string
	LoggingCalls    []string
	MetricsCalls    []string
}

func testConfig(b configBuilder) Config {
	return Config{
		Observability: ObservabilityConfig{
			LoggingCalls: b.LoggingCalls,
			MetricsCalls: b.MetricsCalls,
		},
		Languages:       b.Languages,
		SourceDirs:      b.SourceDirs,
		TestFilePattern: b.TestFilePattern,
	}
}

func TestIsTestFile_DefaultPatterns(t *testing.T) {
	cfg := testConfig(configBuilder{
		TestFilePattern: []string{"_test.", "testdata"},
	})
	if !cfg.IsTestFile("foo_test.go") {
		t.Error("expected foo_test.go to match default _test. pattern")
	}
	if !cfg.IsTestFile("bar_test.ts") {
		t.Error("expected bar_test.ts to match default _test. pattern")
	}
	if !cfg.IsTestFile("internal/checks/testdata/fixture.go") {
		t.Error("expected testdata directory path to match default testdata pattern")
	}
	if !cfg.IsTestFile("pkg/testdata/sub/fixture.go") {
		t.Error("expected testdata directory path to match default testdata pattern")
	}
	if cfg.IsTestFile("main.go") {
		t.Error("expected main.go to NOT match any test file pattern")
	}
}

func TestIsTestFile_CustomPatterns(t *testing.T) {
	cfg := testConfig(configBuilder{
		TestFilePattern: []string{".tst."},
	})
	if !cfg.IsTestFile("foo.tst.ts") {
		t.Error("expected foo.tst.ts to match custom pattern")
	}
	if cfg.IsTestFile("main.go") {
		t.Error("expected main.go to NOT match custom pattern")
	}
}

func TestIsTestFile_CustomPatterns_Multiple(t *testing.T) {
	cfg := testConfig(configBuilder{
		TestFilePattern: []string{"_test.", "testdata", ".spec."},
	})
	if !cfg.IsTestFile("foo_test.go") {
		t.Error("expected foo_test.go to match _test. pattern")
	}
	if !cfg.IsTestFile("internal/testdata/fixture.go") {
		t.Error("expected testdata path to match testdata pattern")
	}
	if !cfg.IsTestFile("bar.spec.ts") {
		t.Error("expected bar.spec.ts to match .spec. pattern")
	}
	if cfg.IsTestFile("internal/checks/fixture.go") {
		t.Error("expected non-test path to NOT match any pattern")
	}
}

func TestIsInSourceDir_GlobMatch(t *testing.T) {
	cfg := testConfig(configBuilder{
		SourceDirs: []string{"src/**", "cmd/**"},
	})
	tests := []struct {
		path   string
		match  bool
		desc   string
	}{
		{"src/main.go", true, "direct file in src"},
		{"src/sub/lib.go", true, "nested file under src"},
		{"src/a/b/c/file.go", true, "deeply nested under src"},
		{"cmd/vibe-harness/main.go", true, "file under cmd"},
		{"internal/checks/generic/magic.go", false, "internal not in patterns"},
		{"readme.md", false, "root file"},
		{"vendor/pkg.go", false, "vendor not in patterns"},
	}
	for _, tt := range tests {
		got := cfg.IsInSourceDir(tt.path)
		if got != tt.match {
			t.Errorf("IsInSourceDir(%q) = %v, want %v (%s)", tt.path, got, tt.match, tt.desc)
		}
	}
}

func TestIsInSourceDir_NoSourceDirs(t *testing.T) {
	cfg := testConfig(configBuilder{})
	if cfg.IsInSourceDir("any/file.go") {
		t.Error("expected IsInSourceDir to return false when no SourceDirs configured")
	}
}

func TestIsSourceDirAncestor(t *testing.T) {
	cfg := testConfig(configBuilder{
		SourceDirs: []string{"cmd/**", "internal/**"},
	})
	tests := []struct {
		dir     string
		ancestor bool
		desc    string
	}{
		{".", true, "root is always ancestor"},
		{"", true, "empty path is root"},
		{"cmd", true, "cmd matches as exact pattern start"},
		{"cmd/sub", true, "nested under cmd"},
		{"internal", true, "internal matches"},
		{"internal/checks", true, "nested under internal"},
		{"vendor", false, "vendor has no matching pattern"},
		{".git", false, ".git has no matching pattern"},
	}
	for _, tt := range tests {
		got := cfg.IsSourceDirAncestor(tt.dir)
		if got != tt.ancestor {
			t.Errorf("IsSourceDirAncestor(%q) = %v, want %v (%s)", tt.dir, got, tt.ancestor, tt.desc)
		}
	}
}

func TestIsSourceDirAncestor_PatternWithGlobstar(t *testing.T) {
	cfg := testConfig(configBuilder{
		SourceDirs: []string{"pkg/**"},
	})
	if !cfg.IsSourceDirAncestor("pkg") {
		t.Error("expected pkg to be ancestor of pkg/**")
	}
	if !cfg.IsSourceDirAncestor("pkg/sub") {
		t.Error("expected pkg/sub to be descendant of pkg/**")
	}
	if cfg.IsSourceDirAncestor("internal") {
		t.Error("expected internal to NOT be ancestor of pkg/**")
	}
}

func TestDefaultConfig_DefensiveCopy(t *testing.T) {
	cfg1 := DefaultConfig()
	cfg1.Observability.LoggingCalls[0] = "MODIFIED"
	cfg1.Observability.MetricsCalls[0] = "MODIFIED"
	cfg1.Languages[".py"] = "MODIFIED"
	cfg1.SourceDirs[0] = "MODIFIED"

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
	if cfg2.SourceDirs[0] == "MODIFIED" {
		t.Error("modifying returned config affected future DefaultConfig() SourceDirs")
	}
}
