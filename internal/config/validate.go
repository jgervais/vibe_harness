package config

import (
	"fmt"

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