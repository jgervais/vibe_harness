package config

type perLanguageDefaults struct {
	LoggingCalls []string
	MetricsCalls []string
}

var defaultLanguageHints = map[string]perLanguageDefaults{
	"python": {
		LoggingCalls: []string{"log", "logger", "logging", "print"},
		MetricsCalls: []string{"metrics", "counter", "histogram", "gauge", "timer"},
	},
	"typescript": {
		LoggingCalls: []string{"log", "logger", "console.log", "console.error"},
		MetricsCalls: []string{"metrics", "counter", "histogram"},
	},
	"go": {
		LoggingCalls: []string{"log", "slog", "fmt.Println"},
		MetricsCalls: []string{"prometheus", "metrics"},
	},
	"java": {
		LoggingCalls: []string{"log", "logger", "LOG"},
		MetricsCalls: []string{"metrics", "counter", "meter"},
	},
	"ruby": {
		LoggingCalls: []string{"log", "logger", "puts"},
		MetricsCalls: []string{"metrics", "counter"},
	},
	"rust": {
		LoggingCalls: []string{"log", "tracing"},
		MetricsCalls: []string{"metrics", "prometheus"},
	},
}

func (c *Config) MergedLoggingCalls(language string) []string {
	seen := make(map[string]struct{})
	var result []string

	if defaults, ok := defaultLanguageHints[language]; ok {
		for _, v := range defaults.LoggingCalls {
			if _, exists := seen[v]; !exists {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}

	for _, v := range c.Observability.LoggingCalls {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}

func (c *Config) MergedMetricsCalls(language string) []string {
	seen := make(map[string]struct{})
	var result []string

	if defaults, ok := defaultLanguageHints[language]; ok {
		for _, v := range defaults.MetricsCalls {
			if _, exists := seen[v]; !exists {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}

	for _, v := range c.Observability.MetricsCalls {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}