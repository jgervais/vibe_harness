package output

import (
	"encoding/json"

	"github.com/jgervais/vibe_harness/internal/rules"
	"github.com/jgervais/vibe_harness/internal/scanner"
)

type sarifSchema struct {
	Schema  string       `json:"$schema"`
	Version string       `json:"version"`
	Runs    []sarifRun   `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool    `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Rules   []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID                   string               `json:"id"`
	Name                 string               `json:"name"`
	ShortDescription     sarifMessageText     `json:"shortDescription"`
	DefaultConfiguration sarifConfiguration  `json:"defaultConfiguration"`
}

type sarifMessageText struct {
	Text string `json:"text"`
}

type sarifConfiguration struct {
	Level string `json:"level"`
}

type sarifResult struct {
	RuleId    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessageText `json:"message"`
	Locations []sarifLocation  `json:"locations"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion            `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
}

func FormatSARIF(result *scanner.ScanResult) ([]byte, error) {
	checks := rules.Checks()
	sarifRules := make([]sarifRule, len(checks))
	for i, c := range checks {
		sarifRules[i] = sarifRule{
			ID:   c.ID,
			Name: c.Name,
			ShortDescription: sarifMessageText{
				Text: c.Description,
			},
			DefaultConfiguration: sarifConfiguration{
				Level: c.Severity,
			},
		}
	}

	sarifResults := make([]sarifResult, len(result.Violations))
	for i, v := range result.Violations {
		sarifResults[i] = sarifResult{
			RuleId: v.RuleID,
			Level:  v.Severity,
			Message: sarifMessageText{
				Text: v.Message,
			},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: v.File,
						},
						Region: sarifRegion{
							StartLine: v.Line,
						},
					},
				},
			},
		}
	}

	if sarifResults == nil {
		sarifResults = []sarifResult{}
	}

	doc := sarifSchema{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    result.Tool.Name,
						Version: result.Tool.Version,
						Rules:   sarifRules,
					},
				},
				Results: sarifResults,
			},
		},
	}

	return json.MarshalIndent(doc, "", "  ")
}