package generic

import (
	"regexp"

	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

type HardcodedSecretsCheck struct{}

func NewHardcodedSecretsCheck() *HardcodedSecretsCheck {
	return &HardcodedSecretsCheck{}
}

func (c *HardcodedSecretsCheck) ID() string { return "VH-G005" }
func (c *HardcodedSecretsCheck) Name() string { return "Hardcoded Secrets" }

type secretPattern struct {
	regex   *regexp.Regexp
	message string
}

var secretPatterns = []secretPattern{
	{
		regex:   regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		message: "hardcoded secret: AWS access key pattern",
	},
	{
		regex:   regexp.MustCompile(`(?i)secret[_\s]*[=:]\s*["'][A-Za-z0-9+/]{40,}["']`),
		message: "hardcoded secret: AWS secret key pattern",
	},
	{
		regex:   regexp.MustCompile(`(?i)(api[_\s]?key|apikey|API_KEY)\s*[=:]\s*["'][^"']{20,}["']`),
		message: "hardcoded secret: API key assignment",
	},
	{
		regex:   regexp.MustCompile(`(?i)(mongodb|postgres|mysql)://[^:@\s]+:[^@\s]+@`),
		message: "hardcoded secret: credential connection string",
	},
	{
		regex:   regexp.MustCompile(`(?i)jdbc:[^\s"']+`),
		message: "hardcoded secret: JDBC connection string",
	},
	{
		regex:   regexp.MustCompile(`(?i)connection_string\s*[=:]\s*["'][^"']+["']`),
		message: "hardcoded secret: connection string assignment",
	},
	{
		regex:   regexp.MustCompile(`Bearer\s+[A-Za-z0-9\-_.~+/]+=*`),
		message: "hardcoded secret: Bearer token",
	},
	{
		regex:   regexp.MustCompile(`-----BEGIN RSA PRIVATE KEY-----`),
		message: "hardcoded secret: RSA private key marker",
	},
	{
		regex:   regexp.MustCompile(`-----BEGIN PRIVATE KEY-----`),
		message: "hardcoded secret: private key marker",
	},
}

func (c *HardcodedSecretsCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	var violations []rules.Violation
	lines := splitLines(content)
	for lineNum, line := range lines {
		for _, pat := range secretPatterns {
			if pat.regex.Match(line) {
				violations = append(violations, rules.Violation{
					RuleID:   "VH-G005",
					File:     path,
					Line:     lineNum + 1,
					Column:   0,
					EndLine:  0,
					Message:  pat.message,
					Severity: "error",
				})
			}
		}
	}
	return violations
}

func splitLines(content []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range content {
		if b == '\n' {
			lines = append(lines, content[start:i])
			start = i + 1
		}
	}
	if start < len(content) {
		lines = append(lines, content[start:])
	}
	return lines
}