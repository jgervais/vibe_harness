package generic

import (
	"fmt"
	"regexp"
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

func (c *MagicValuesCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	return nil
}

type numericOccurrence struct {
	value string
	file  string
	line  int
}

type stringOccurrence struct {
	value string
	file  string
	line  int
}

var numericLiteralRe = regexp.MustCompile(`(?:-?\d+\.\d+|-?\d+)`)
var stringLiteralRe = regexp.MustCompile(`"([^"\\]|\\.)*"|'([^'\\]|\\.)*'`)
var importLineRe = regexp.MustCompile(`^\s*(?:import\s|require\s*\(|from\s+['"]|require\s+['"])`)
var allCapsIdentRe = regexp.MustCompile(`[A-Z][A-Z0-9_]*\s*(?::=|=|:)\s`)
var pascalCaseIdentRe = regexp.MustCompile(`[A-Z][a-zA-Z0-9]+\s*(?::=|=|:)\s`)
var constKeywordRe = regexp.MustCompile(`\bconst\b`)
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

func isSingleDigit(s string) bool {
	abs := strings.TrimPrefix(s, "-")
	if len(abs) == 1 && abs[0] >= '0' && abs[0] <= '9' {
		return true
	}
	return false
}

func (c *MagicValuesCheck) CheckFiles(files []FileContent, cfg *config.Config) []rules.Violation {
	var violations []rules.Violation

	numericCounts := map[string]int{}
	numericFirst := map[string]numericOccurrence{}
	stringCounts := map[string]int{}
	stringFirst := map[string]stringOccurrence{}

	for _, f := range files {
		if cfg.IsTestFile(f.Path) {
			continue
		}
		lines := strings.Split(string(f.Content), "\n")
		for i, line := range lines {
			lineNum := i + 1
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || isImportLine(trimmed) || isConstantLine(trimmed) {
				continue
			}
			for _, match := range numericLiteralRe.FindAllString(trimmed, -1) {
				if isSingleDigit(match) {
					continue
				}
				numericCounts[match]++
				if _, exists := numericFirst[match]; !exists {
					numericFirst[match] = numericOccurrence{value: match, file: f.Path, line: lineNum}
				}
			}
			if emptyCollectionRe.MatchString(trimmed) {
				continue
			}
			for _, match := range stringLiteralRe.FindAllString(trimmed, -1) {
				inner := match
				if len(inner) >= 2 {
					inner = inner[1 : len(inner)-1]
				}
				if len(inner) < 20 {
					continue
				}
				stringCounts[match]++
				if _, exists := stringFirst[match]; !exists {
					stringFirst[match] = stringOccurrence{value: match, file: f.Path, line: lineNum}
				}
			}
		}
	}

	for val, count := range numericCounts {
		if count < 3 {
			continue
		}
		first := numericFirst[val]
		display := first.value
		violations = append(violations, rules.Violation{
			RuleID:   "VH-G006",
			File:     first.file,
			Line:     first.line,
			Column:   0,
			EndLine:  0,
			Message:  fmt.Sprintf("magic value: %s used inline (appears %d times across codebase)", display, count),
			Severity: "error",
		})
	}

	for val, count := range stringCounts {
		if count < 3 {
			continue
		}
		first := stringFirst[val]
		display := first.value
		if len(display) > 40 {
			display = display[:37] + "..."
		}
		violations = append(violations, rules.Violation{
			RuleID:   "VH-G006",
			File:     first.file,
			Line:     first.line,
			Column:   0,
			EndLine:  0,
			Message:  fmt.Sprintf("magic string: %s used inline (appears %d times across codebase)", display, count),
			Severity: "error",
		})
	}

	if violations == nil {
		return nil
	}
	return violations
}