package rules

import (
	"fmt"
	"regexp"
)

var ruleIDPattern = regexp.MustCompile(`^VH-G\d{3}$`)

type Violation struct {
	RuleID   string
	File     string
	Line     int
	Column   int
	EndLine int
	Message  string
	Severity string
}

type Check struct {
	ID          string
	Name        string
	Description string
	Severity    string
	RequiresAST bool
	Threshold   string
}

func Checks() []Check {
	return []Check{
		{
			ID:          "VH-G001",
			Name:        "File Length",
			Description: "Files must not exceed 300 non-blank, non-comment lines",
			Severity:    "warning",
			RequiresAST: false,
			Threshold:   "300 lines",
		},
		{
			ID:          "VH-G002",
			Name:        "Function Length",
			Description: "Functions must not exceed 50 statements",
			Severity:    "warning",
			RequiresAST: true,
			Threshold:   "50 statements",
		},
		{
			ID:          "VH-G003",
			Name:        "Missing Logging in I/O",
			Description: "I/O calls should have logging in the same scope",
			Severity:    "warning",
			RequiresAST: true,
			Threshold:   "Pattern-based",
		},
		{
			ID:          "VH-G004",
			Name:        "Swallowed Errors",
			Description: "Catch/except blocks must handle errors (re-raise, return, or log)",
			Severity:    "error",
			RequiresAST: true,
			Threshold:   "Pattern-based",
		},
		{
			ID:          "VH-G005",
			Name:        "Hardcoded Secrets",
			Description: "Detects hardcoded secrets and credentials",
			Severity:    "error",
			RequiresAST: false,
			Threshold:   "Pattern match",
		},
		{
			ID:          "VH-G006",
			Name:        "Magic Values",
			Description: "Detects magic numbers and inline strings repeated across the codebase",
			Severity:    "error",
			RequiresAST: false,
			Threshold:   "3+ occurrences across codebase",
		},
		{
			ID:          "VH-G007",
			Name:        "Copy-Paste Duplication",
			Description: "Detects duplicated code blocks across files",
			Severity:    "warning",
			RequiresAST: false,
			Threshold:   "6 lines, 80% similarity",
		},
		{
			ID:          "VH-G008",
			Name:        "Comment-to-Code Ratio",
			Description: "Flags files where comments exceed 1:3 ratio",
			Severity:    "note",
			RequiresAST: false,
			Threshold:   "1:3 ratio",
		},
		{
			ID:          "VH-G009",
			Name:        "Missing Error Handling on I/O",
			Description: "I/O calls must be wrapped in error-handling constructs",
			Severity:    "error",
			RequiresAST: true,
			Threshold:   "Pattern-based",
		},
		{
			ID:          "VH-G010",
			Name:        "Broad Exception Catching",
			Description: "Catching root exception types (Exception, Throwable, bare except)",
			Severity:    "warning",
			RequiresAST: true,
			Threshold:   "Pattern-based",
		},
		{
			ID:          "VH-G011",
			Name:        "Disabled Security Features",
			Description: "Detects disabled security verification",
			Severity:    "error",
			RequiresAST: false,
			Threshold:   "Pattern match",
		},
		{
			ID:          "VH-G012",
			Name:        "God Module",
			Description: "Files must not exceed 20 public exports",
			Severity:    "warning",
			RequiresAST: true,
			Threshold:   "20 exports",
		},
	}
}

var validSeverities = map[string]bool{
	"error":   true,
	"warning": true,
	"note":    true,
}

func (v Violation) Validate() error {
	if !ruleIDPattern.MatchString(v.RuleID) {
		return fmt.Errorf("invalid RuleID %q: must match VH-G\\d{3}", v.RuleID)
	}
	if v.File == "" {
		return fmt.Errorf("File must not be empty")
	}
	if v.Line < 1 {
		return fmt.Errorf("Line must be >= 1, got %d", v.Line)
	}
	if !validSeverities[v.Severity] {
		return fmt.Errorf("invalid Severity %q: must be error, warning, or note", v.Severity)
	}
	return nil
}