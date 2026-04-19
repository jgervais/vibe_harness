package output

import (
	"encoding/json"
	"testing"

	"github.com/jgervais/vibe_harness/internal/rules"
	"github.com/jgervais/vibe_harness/internal/scanner"
)

func TestFormatSARIF_WithViolations(t *testing.T) {
	result := &scanner.ScanResult{
		Tool: scanner.ToolInfo{
			Name:    "vibe-harness",
			Version: "0.1.0",
		},
		Violations: []rules.Violation{
			{
				RuleID:   "VH-G001",
				File:     "src/main.go",
				Line:     1,
				Message:  "file exceeds 300 non-blank, non-comment lines (412)",
				Severity: "warning",
			},
			{
				RuleID:   "VH-G005",
				File:     "src/auth.go",
				Line:     42,
				Message:  "hardcoded secret detected",
				Severity: "error",
			},
		},
	}

	data, err := FormatSARIF(result)
	if err != nil {
		t.Fatalf("FormatSARIF returned error: %v", err)
	}

	if !json.Valid(data) {
		t.Fatal("output is not valid JSON")
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if _, ok := doc["$schema"]; !ok {
		t.Error("missing $schema field")
	}
	if v := doc["version"]; v != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %v", v)
	}
	runs, ok := doc["runs"].([]interface{})
	if !ok {
		t.Fatal("missing or invalid runs array")
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}

	run := runs[0].(map[string]interface{})
	driver := run["tool"].(map[string]interface{})["driver"].(map[string]interface{})

	if driver["name"] != "vibe-harness" {
		t.Errorf("expected driver name 'vibe-harness', got %v", driver["name"])
	}

	sarifRules := driver["rules"].([]interface{})
	if len(sarifRules) != 6 {
		t.Errorf("expected 6 rules, got %d", len(sarifRules))
	}

	sarifResults := run["results"].([]interface{})
	if len(sarifResults) != 2 {
		t.Errorf("expected 2 results, got %d", len(sarifResults))
	}

	for i, r := range sarifResults {
		res := r.(map[string]interface{})
		if _, ok := res["ruleId"]; !ok {
			t.Errorf("result %d missing ruleId", i)
		}
		if _, ok := res["level"]; !ok {
			t.Errorf("result %d missing level", i)
		}
		msg := res["message"].(map[string]interface{})
		if _, ok := msg["text"]; !ok {
			t.Errorf("result %d missing message.text", i)
		}
		locs := res["locations"].([]interface{})
		if len(locs) == 0 {
			t.Errorf("result %d has no locations", i)
			continue
		}
		loc := locs[0].(map[string]interface{})
		if _, ok := loc["physicalLocation"]; !ok {
			t.Errorf("result %d missing physicalLocation in location", i)
		}
	}
}

func TestFormatSARIF_EmptyViolations(t *testing.T) {
	result := &scanner.ScanResult{
		Tool: scanner.ToolInfo{
			Name:    "vibe-harness",
			Version: "0.1.0",
		},
		Violations: []rules.Violation{},
	}

	data, err := FormatSARIF(result)
	if err != nil {
		t.Fatalf("FormatSARIF returned error: %v", err)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	runs := doc["runs"].([]interface{})
	results := runs[0].(map[string]interface{})["results"].([]interface{})
	if len(results) != 0 {
		t.Errorf("expected empty results array, got %d items", len(results))
	}
}