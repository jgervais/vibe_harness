package config

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

var disallowedTopLevelKeys = map[string]bool{
	"enabled":   true,
	"threshold": true,
	"skip":      true,
	"ignore":    true,
	"severity":  true,
	"rules":     true,
}

func ValidateTOML(path string) error {
	var raw map[string]interface{}
	if _, err := toml.DecodeFile(path, &raw); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	if err := checkDisallowed(raw); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if err := validateLanguages(raw); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if err := validateSourceDirs(raw); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if err := validateTestFilePattern(raw); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	return nil
}

func checkDisallowed(m map[string]interface{}) error {
	for key, val := range m {
		if disallowedTopLevelKeys[key] {
			return fmt.Errorf("'%s' key is not allowed (cannot modify rule behavior)", key)
		}

		switch v := val.(type) {
		case map[string]interface{}:
			if err := checkDisallowed(v); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateLanguages(raw map[string]interface{}) error {
	val, ok := raw["languages"]
	if !ok {
		return fmt.Errorf("'[languages]' section is required")
	}
	langs, ok := val.(map[string]interface{})
	if !ok {
		return fmt.Errorf("'[languages]' must be a table")
	}
	if len(langs) == 0 {
		return fmt.Errorf("'[languages]' must define at least one language extension")
	}
	for ext := range langs {
		if !strings.HasPrefix(ext, ".") {
			return fmt.Errorf("language key %q must start with '.' (e.g. '.go')", ext)
		}
	}
	return nil
}

func validateSourceDirs(raw map[string]interface{}) error {
	val, ok := raw["source_directories"]
	if !ok {
		return fmt.Errorf("'source_directories' is required")
	}
	dirs, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("'source_directories' must be an array of strings")
	}
	if len(dirs) == 0 {
		return fmt.Errorf("'source_directories' must contain at least one pattern")
	}
	for i, item := range dirs {
		dir, ok := item.(string)
		if !ok {
			return fmt.Errorf("'source_directories[%d]' must be a string", i)
		}
		if strings.HasPrefix(dir, "/") {
			return fmt.Errorf("'source_directories[%d]' %q must be a relative path", i, dir)
		}
		if dir == "." || dir == ".." || strings.HasPrefix(dir, "../") || strings.HasPrefix(dir, "..\\") {
			return fmt.Errorf("'source_directories[%d]' %q is not allowed", i, dir)
		}
		if dir == "" {
			return fmt.Errorf("'source_directories[%d]' must not be empty", i)
		}
	}
	return nil
}

func validateTestFilePattern(raw map[string]interface{}) error {
	val, ok := raw["test_file_pattern"]
	if !ok {
		return nil
	}

	switch v := val.(type) {
	case string:
		return validateSinglePattern(v)
	case []interface{}:
		for i, item := range v {
			pattern, ok := item.(string)
			if !ok {
				return fmt.Errorf("'test_file_pattern[%d]' must be a string", i)
			}
			if err := validateSinglePattern(pattern); err != nil {
				return fmt.Errorf("'test_file_pattern[%d]': %w", i, err)
			}
		}
	default:
		return fmt.Errorf("'test_file_pattern' must be a string or array of strings")
	}
	return nil
}

func validateSinglePattern(pattern string) error {
	if pattern == "" {
		return nil
	}
	if !strings.Contains(pattern, "test") && !strings.Contains(pattern, "tst") {
		return fmt.Errorf("must contain 'test' or 'tst', got %q", pattern)
	}
	if pattern == "*" || pattern == "**" || pattern == "." || pattern == "/" {
		return fmt.Errorf("%q is too broad and would match all files", pattern)
	}
	return nil
}
