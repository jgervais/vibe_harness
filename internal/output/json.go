package output

import (
	"encoding/json"

	"github.com/jgervais/vibe_harness/internal/scanner"
)

type JSONTool struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	RulesHash string `json:"rules_hash"`
}

type JSONStats struct {
	FilesScanned    int            `json:"files_scanned"`
	FilesSkipped    int            `json:"files_skipped"`
	ViolationsByRule map[string]int `json:"violations_by_rule"`
	Duration         string         `json:"duration"`
}

type JSONViolation struct {
	RuleID   string `json:"rule_id"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	EndLine  int    `json:"end_line"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type JSONOutput struct {
	Version string         `json:"version"`
	Tool    JSONTool        `json:"tool"`
	Target  string          `json:"target"`
	Stats   JSONStats       `json:"stats"`
	Results []JSONViolation `json:"results"`
}

func FormatJSON(result *scanner.ScanResult) ([]byte, error) {
	violations := make([]JSONViolation, len(result.Violations))
	for i, v := range result.Violations {
		violations[i] = JSONViolation{
			RuleID:   v.RuleID,
			File:     v.File,
			Line:     v.Line,
			Column:   v.Column,
			EndLine:  v.EndLine,
			Message:  v.Message,
			Severity: v.Severity,
		}
	}

	if violations == nil {
		violations = []JSONViolation{}
	}

	out := JSONOutput{
		Version: "1.0",
		Tool: JSONTool{
			Name:      result.Tool.Name,
			Version:   result.Tool.Version,
			RulesHash: result.Tool.RulesHash,
		},
		Target: result.Target,
		Stats: JSONStats{
			FilesScanned:    result.Stats.FilesScanned,
			FilesSkipped:    result.Stats.FilesSkipped,
			ViolationsByRule: result.Stats.ViolationsByRule,
			Duration:         result.Stats.Duration,
		},
		Results: violations,
	}

	return json.MarshalIndent(out, "", "  ")
}