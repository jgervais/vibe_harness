package generic

import (
	"regexp"

	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

type SecurityFeaturesCheck struct{}

func NewSecurityFeaturesCheck() *SecurityFeaturesCheck {
	return &SecurityFeaturesCheck{}
}

func (c *SecurityFeaturesCheck) ID() string   { return "VH-G011" }
func (c *SecurityFeaturesCheck) Name() string { return "Disabled Security Features" }

type securityPattern struct {
	regex   *regexp.Regexp
	message string
}

var securityPatterns = []securityPattern{
	{
		regex:   regexp.MustCompile(`verify\s*=\s*False`),
		message: "disabled security verification: verify=False",
	},
	{
		regex:   regexp.MustCompile(`InsecureSkipVerify\s*:\s*true`),
		message: "disabled security verification: InsecureSkipVerify=true",
	},
	{
		regex:   regexp.MustCompile(`rejectUnauthorized\s*:\s*false`),
		message: "disabled security verification: rejectUnauthorized=false",
	},
	{
		regex:   regexp.MustCompile(`--no-verify-ssl`),
		message: "disabled security verification: --no-verify-ssl",
	},
	{
		regex:   regexp.MustCompile(`ssl_verify\s*:\s*false`),
		message: "disabled security verification: ssl_verify=false",
	},
	{
		regex:   regexp.MustCompile(`CURLOPT_SSL_VERIFYPEER\s*[,=]\s*false`),
		message: "disabled security verification: CURLOPT_SSL_VERIFYPEER=false",
	},
}

func (c *SecurityFeaturesCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	var violations []rules.Violation
	lines := splitLines(content)
	for lineNum, line := range lines {
		for _, pat := range securityPatterns {
			if pat.regex.Match(line) {
				violations = append(violations, rules.Violation{
					RuleID:   "VH-G011",
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