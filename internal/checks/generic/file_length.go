package generic

import (
	"fmt"
	"strings"

	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

type Check interface {
	ID() string
	Name() string
	CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation
}

type FileLengthCheck struct{}

func NewFileLengthCheck() *FileLengthCheck {
	return &FileLengthCheck{}
}

func (c *FileLengthCheck) ID() string   { return "VH-G001" }
func (c *FileLengthCheck) Name() string { return "File Length" }

var lineCommentPrefixes = map[string]string{
	"go":         "//",
	"java":       "//",
	"typescript": "//",
	"javascript": "//",
	"rust":       "//",
	"python":     "#",
	"ruby":       "#",
	"sql":        "--",
}

func (c *FileLengthCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	lines := strings.Split(string(content), "\n")
	prefix := lineCommentPrefixes[language]
	inBlockComment := false
	codeLines := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if inBlockComment {
			if strings.Contains(trimmed, "*/") {
				inBlockComment = false
			}
			continue
		}

		if strings.HasPrefix(trimmed, "/*") {
			if !strings.Contains(trimmed[2:], "*/") {
				inBlockComment = true
			}
			continue
		}

		if prefix != "" && strings.HasPrefix(trimmed, prefix) {
			continue
		}

		codeLines++
	}

	if codeLines > 300 {
		return []rules.Violation{
			{
				RuleID:   "VH-G001",
				File:     path,
				Line:     1,
				Column:   0,
				EndLine:  0,
				Message:  fmt.Sprintf("file exceeds 300 non-blank, non-comment lines (%d)", codeLines),
				Severity: "warning",
			},
		}
	}

	return nil
}