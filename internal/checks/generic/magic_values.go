package generic

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

type MagicValuesCheck struct{}

func NewMagicValuesCheck() *MagicValuesCheck {
	return &MagicValuesCheck{}
}

func (c *MagicValuesCheck) ID() string   { return "VH-G006" }
func (c *MagicValuesCheck) Name() string { return "Magic Values" }

var numericLiteralRe = regexp.MustCompile(`(?:-?\d+\.\d+|-?\d+)`)
var stringLiteralRe = regexp.MustCompile(`"([^"\\]|\\.)*"|'([^'\\]|\\.)*'`)
var importLineRe = regexp.MustCompile(`^\s*(?:import\s|require\s*\(|from\s+['"]|require\s+['"])`)
var allCapsIdentRe = regexp.MustCompile(`[A-Z][A-Z0-9_]*\s*(?::=|=|:)\s`)
var pascalCaseIdentRe = regexp.MustCompile(`[A-Z][a-zA-Z0-9]+\s*(?::=|=|:)\s`)
var constKeywordRe = regexp.MustCompile(`\bconst\b`)

var allowedNumericValues = map[string]bool{
	"0":  true,
	"1":  true,
	"-1": true,
	"2":  true,
}

var allowedStringValues = map[string]bool{
	`""`:          true,
	`''`:          true,
	`"true"`:      true,
	`"false"`:     true,
	`"null"`:      true,
	`"nil"`:       true,
	`"None"`:      true,
	`"undefined"`: true,
	`'true'`:      true,
	`'false'`:     true,
	`'null'`:      true,
	`'nil'`:       true,
	`'None'`:      true,
	`'undefined'`: true,
}

var emptyCollectionRe = regexp.MustCompile(`^\s*(?:let|var|const)\s+\w+\s*=\s*(?:\[\]|\{\})`)

func isConstantLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if constKeywordRe.MatchString(trimmed) {
		return true
	}
	if allCapsIdentRe.MatchString(trimmed) && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") {
		return true
	}
	if pascalCaseIdentRe.MatchString(trimmed) && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") {
		return true
	}
	return false
}

func isImportLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return importLineRe.MatchString(trimmed)
}

func formatCount(n int) string {
	return strconv.Itoa(n)
}

func (c *MagicValuesCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	lines := strings.Split(string(content), "\n")
	var violations []rules.Violation

	numericCounts := map[string]int{}
	numericFirstLine := map[string]int{}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isImportLine(trimmed) {
			continue
		}
		if isConstantLine(trimmed) {
			continue
		}
		for _, match := range numericLiteralRe.FindAllString(trimmed, -1) {
			if allowedNumericValues[match] {
				continue
			}
			numericCounts[match]++
			if _, exists := numericFirstLine[match]; !exists {
				numericFirstLine[match] = i + 1
			}
		}
	}

	reportedNumerics := map[string]bool{}
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isImportLine(trimmed) {
			continue
		}
		isConst := isConstantLine(trimmed)

		if !isConst {
			for _, match := range numericLiteralRe.FindAllString(trimmed, -1) {
				if allowedNumericValues[match] {
					continue
				}
				if numericCounts[match] < 2 {
					continue
				}
				if reportedNumerics[match] {
					continue
				}
				reportedNumerics[match] = true
				violations = append(violations, rules.Violation{
					RuleID:   "VH-G006",
					File:     path,
					Line:     numericFirstLine[match],
					Column:   0,
					EndLine:  0,
					Message:  "magic value: " + match + " used inline (appears " + formatCount(numericCounts[match]) + " times)",
					Severity: "warning",
				})
			}

			if !emptyCollectionRe.MatchString(trimmed) {
				for _, match := range stringLiteralRe.FindAllString(trimmed, -1) {
					if allowedStringValues[match] {
						continue
					}
					inner := match
					if len(inner) >= 2 {
						inner = inner[1 : len(inner)-1]
					}
					if len(inner) >= 20 {
						display := match
						if len(display) > 40 {
							display = display[:37] + "..."
						}
						violations = append(violations, rules.Violation{
							RuleID:   "VH-G006",
							File:     path,
							Line:     lineNum,
							Column:   0,
							EndLine:  0,
							Message:  "magic string: " + display + " used inline (20+ chars)",
							Severity: "warning",
						})
					}
				}
			}
		}
	}

	if violations == nil {
		return nil
	}
	return violations
}