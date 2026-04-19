package output

import (
	"encoding/json"
	"testing"

	"github.com/jgervais/vibe_harness/internal/rules"
	"github.com/jgervais/vibe_harness/internal/scanner"
)

func TestFormatJSON_WithViolations(t *testing.T) {
	result := &scanner.ScanResult{
		Tool: scanner.ToolInfo{
			Name:      "vibe-harness",
			Version:   "0.1.0",
			RulesHash: "sha256:abc123",
		},
		Target: "/path/to/dir",
		Violations: []rules.Violation{
			{
				RuleID:   "VH-G001",
				File:     "src/main.go",
				Line:     1,
				Column:  0,
				EndLine:  0,
				Message:  "file exceeds 300 non-blank, non-comment lines (412)",
				Severity: "warning",
			},
			{
				RuleID:   "VH-G005",
				File:     "src/auth.go",
				Line:     42,
				Column:   10,
				EndLine:  42,
				Message:  "hardcoded secret detected",
				Severity: "error",
			},
		},
		Stats: scanner.ScanStats{
			FilesScanned:    42,
			FilesSkipped:    3,
			ViolationsByRule: map[string]int{"VH-G001": 2, "VH-G005": 1},
			Duration:         "1.2s",
		},
		ExitCode: 1,
	}

	data, err := FormatJSON(result)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if v, _ := parsed["version"].(string); v != "1.0" {
		t.Errorf("version = %q, want %q", v, "1.0")
	}

	tool, _ := parsed["tool"].(map[string]interface{})
	if tool == nil {
		t.Fatal("missing tool object")
	}
	if v, _ := tool["name"].(string); v != "vibe-harness" {
		t.Errorf("tool.name = %q, want %q", v, "vibe-harness")
	}
	if v, _ := tool["version"].(string); v != "0.1.0" {
		t.Errorf("tool.version = %q, want %q", v, "0.1.0")
	}
	if v, _ := tool["rules_hash"].(string); v != "sha256:abc123" {
		t.Errorf("tool.rules_hash = %q, want %q", v, "sha256:abc123")
	}

	if v, _ := parsed["target"].(string); v != "/path/to/dir" {
		t.Errorf("target = %q, want %q", v, "/path/to/dir")
	}

	stats, _ := parsed["stats"].(map[string]interface{})
	if stats == nil {
		t.Fatal("missing stats object")
	}
	if _, ok := stats["files_scanned"]; !ok {
		t.Error("missing stats.files_scanned")
	}
	if _, ok := stats["files_skipped"]; !ok {
		t.Error("missing stats.files_skipped")
	}
	if _, ok := stats["violations_by_rule"]; !ok {
		t.Error("missing stats.violations_by_rule")
	}
	if _, ok := stats["duration"]; !ok {
		t.Error("missing stats.duration")
	}

	results, _ := parsed["results"].([]interface{})
	if results == nil {
		t.Fatal("missing results array")
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}

	first, _ := results[0].(map[string]interface{})
	if v, _ := first["rule_id"].(string); v != "VH-G001" {
		t.Errorf("results[0].rule_id = %q, want %q", v, "VH-G001")
	}
	if v, _ := first["file"].(string); v != "src/main.go" {
		t.Errorf("results[0].file = %q, want %q", v, "src/main.go")
	}
	if v, _ := first["severity"].(string); v != "warning" {
		t.Errorf("results[0].severity = %q, want %q", v, "warning")
	}
}

func TestFormatJSON_EmptyViolations(t *testing.T) {
	result := &scanner.ScanResult{
		Tool: scanner.ToolInfo{
			Name:      "vibe-harness",
			Version:   "0.1.0",
			RulesHash: "sha256:def456",
		},
		Target:     "/empty/dir",
		Violations: nil,
		Stats: scanner.ScanStats{
			FilesScanned:    10,
			FilesSkipped:    0,
			ViolationsByRule: map[string]int{},
			Duration:         "0.5s",
		},
		ExitCode: 0,
	}

	data, err := FormatJSON(result)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	results, _ := parsed["results"].([]interface{})
	if results == nil {
		t.Fatal("results key missing or null")
	}
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}

func TestFormatJSON_SnakeCaseKeys(t *testing.T) {
	result := &scanner.ScanResult{
		Tool: scanner.ToolInfo{
			Name:      "vibe-harness",
			Version:   "0.1.0",
			RulesHash: "sha256:test",
		},
		Target:     "/test",
		Violations: []rules.Violation{{RuleID: "VH-G001", File: "a.go", Line: 1, Severity: "warning", Message: "msg"}},
		Stats: scanner.ScanStats{
			FilesScanned:    1,
			FilesSkipped:    0,
			ViolationsByRule: map[string]int{"VH-G001": 1},
			Duration:         "0s",
		},
	}

	data, err := FormatJSON(result)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	snakeCaseKeys := []string{"rules_hash", "files_scanned", "files_skipped", "violations_by_rule", "end_line", "rule_id"}

	toolRaw, _ := raw["tool"]
	statsRaw, _ := raw["stats"]
	resultsRaw, _ := raw["results"]

	var tool map[string]json.RawMessage
	json.Unmarshal(toolRaw, &tool)

	var stats map[string]json.RawMessage
	json.Unmarshal(statsRaw, &stats)

	var resultsList []map[string]json.RawMessage
	json.Unmarshal(resultsRaw, &resultsList)

	for _, key := range snakeCaseKeys {
		switch key {
		case "rules_hash":
			if _, ok := tool[key]; !ok {
				t.Errorf("missing snake_case key: tool.%s", key)
			}
		case "files_scanned", "files_skipped", "violations_by_rule":
			if _, ok := stats[key]; !ok {
				t.Errorf("missing snake_case key: stats.%s", key)
			}
		case "end_line", "rule_id":
			if len(resultsList) > 0 {
				if _, ok := resultsList[0][key]; !ok {
					t.Errorf("missing snake_case key: results[0].%s", key)
				}
			}
		}
	}
}