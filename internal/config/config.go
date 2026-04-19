package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type ObservabilityConfig struct {
	LoggingCalls []string `toml:"logging_calls"`
	MetricsCalls []string `toml:"metrics_calls"`
}

type Config struct {
	Observability ObservabilityConfig `toml:"observability"`
	Languages     map[string]string   `toml:"languages"`
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
		Languages: languages,
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

	if len(fileCfg.Observability.LoggingCalls) > 0 {
		cfg.Observability.LoggingCalls = fileCfg.Observability.LoggingCalls
	}
	if len(fileCfg.Observability.MetricsCalls) > 0 {
		cfg.Observability.MetricsCalls = fileCfg.Observability.MetricsCalls
	}
	for k, v := range fileCfg.Languages {
		cfg.Languages[k] = v
	}

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