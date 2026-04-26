package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/BurntSushi/toml"
)

type ObservabilityConfig struct {
	LoggingCalls []string `toml:"logging_calls"`
	MetricsCalls []string `toml:"metrics_calls"`
}

type Config struct {
	Observability   ObservabilityConfig `toml:"observability"`
	Languages       map[string]string   `toml:"languages"`
	SourceDirs      []string            `toml:"source_directories"`
	TestFilePattern []string            `toml:"test_file_pattern"`
}

var defaultLoggingCalls = []string{"log", "logger", "logging", "tracing", "slog", "logr"}
var defaultMetricsCalls = []string{"metrics", "counter", "histogram", "gauge", "timer", "prometheus"}
var defaultLanguages = map[string]string{
	".py":   "python",
	".ts":   "typescript",
	".tsx":  "typescript",
	".js":   "javascript",
	".go":   "go",
	".java": "java",
	".rb":   "ruby",
	".rs":   "rust",
}

func DefaultConfig() Config {
	loggingCalls := make([]string, len(defaultLoggingCalls))
	copy(loggingCalls, defaultLoggingCalls)

	metricsCalls := make([]string, len(defaultMetricsCalls))
	copy(metricsCalls, defaultMetricsCalls)

	languages := make(map[string]string, len(defaultLanguages))
	for k, v := range defaultLanguages {
		languages[k] = v
	}

	return Config{
		Observability: ObservabilityConfig{
			LoggingCalls: loggingCalls,
			MetricsCalls: metricsCalls,
		},
		Languages:       languages,
		SourceDirs:      []string{"**"},
		TestFilePattern: []string{"_test.", "testdata"},
	}
}

func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	var fileCfg Config
	if _, err := toml.DecodeFile(path, &fileCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	cfg := DefaultConfig()
	cfg.Observability = fileCfg.Observability
	cfg.Languages = fileCfg.Languages
	cfg.SourceDirs = fileCfg.SourceDirs
	cfg.TestFilePattern = fileCfg.TestFilePattern

	return &cfg, nil
}

func AutoDiscoverConfig(target string) (string, error) {
	dir, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %s: %w", target, err)
	}

	for {
		candidate := filepath.Join(dir, ".vibe_harness.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", nil
}

func (c *Config) IsTestFile(path string) bool {
	for _, pattern := range c.TestFilePattern {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func (c *Config) IsInSourceDir(relPath string) bool {
	for _, pattern := range c.SourceDirs {
		match, err := doublestar.Match(pattern, relPath)
		if err == nil && match {
			return true
		}
	}
	return false
}

func (c *Config) IsSourceDirAncestor(relDir string) bool {
	if relDir == "" || relDir == "." {
		return true
	}
	for _, pattern := range c.SourceDirs {
		cleaned := strings.TrimPrefix(pattern, "./")
		if cleaned == "**" {
			return true
		}
		base := strings.TrimSuffix(cleaned, "/**")
		base = strings.TrimSuffix(base, "/*")
		base = strings.TrimSuffix(base, "/")
		if relDir == base || strings.HasPrefix(relDir, base+"/") {
			return true
		}
		if strings.HasPrefix(cleaned, relDir+"/") {
			return true
		}
	}
	return false
}
