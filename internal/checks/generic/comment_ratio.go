package generic

import (
	"fmt"
	"strings"

	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

type CommentRatioCheck struct{}

func NewCommentRatioCheck() *CommentRatioCheck {
	return &CommentRatioCheck{}
}

func (c *CommentRatioCheck) ID() string   { return "VH-G008" }
func (c *CommentRatioCheck) Name() string { return "Comment-to-Code Ratio" }

type commentConfig struct {
	linePrefix  string
	blockStart  string
	blockEnd    string
	tripleQuote bool
}

var languageCommentCfg = map[string]commentConfig{
	"go":         {linePrefix: "//", blockStart: "/*", blockEnd: "*/"},
	"java":       {linePrefix: "//", blockStart: "/*", blockEnd: "*/"},
	"typescript": {linePrefix: "//", blockStart: "/*", blockEnd: "*/"},
	"javascript": {linePrefix: "//", blockStart: "/*", blockEnd: "*/"},
	"rust":       {linePrefix: "//", blockStart: "/*", blockEnd: "*/"},
	"python":     {linePrefix: "#", tripleQuote: true},
	"ruby":       {linePrefix: "#"},
	"sql":        {linePrefix: "--"},
}

func (c *CommentRatioCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	lines := strings.Split(string(content), "\n")
	lcfg, ok := languageCommentCfg[language]
	if !ok {
		return nil
	}

	inBlockComment := false
	inTripleQuote := false
	tripleQuoteDelim := ""

	totalNonBlank := 0
	commentLines := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		totalNonBlank++

		if lcfg.tripleQuote {
			handled, isComment := handlePythonLine(trimmed, &inTripleQuote, &tripleQuoteDelim)
			if handled {
				if isComment {
					commentLines++
				}
				continue
			}
		}

		if inBlockComment {
			if strings.Contains(trimmed, lcfg.blockEnd) {
				inBlockComment = false
			}
			commentLines++
			continue
		}

		if lcfg.blockStart != "" && strings.HasPrefix(trimmed, lcfg.blockStart) {
			if strings.Contains(trimmed[2:], lcfg.blockEnd) {
				commentLines++
			} else {
				inBlockComment = true
				commentLines++
			}
			continue
		}

		if lcfg.linePrefix != "" && strings.HasPrefix(trimmed, lcfg.linePrefix) {
			commentLines++
			continue
		}
	}

	if totalNonBlank == 0 {
		return nil
	}

	codeLines := totalNonBlank - commentLines
	ratio := float64(commentLines) / float64(totalNonBlank)

	if ratio > 0.25 {
		return []rules.Violation{
			{
				RuleID:   "VH-G008",
				File:     path,
				Line:     1,
				Column:   0,
				EndLine:  0,
				Message:  fmt.Sprintf("excessive comment-to-code ratio (%d:%d, exceeds 1:3)", commentLines, codeLines),
				Severity: "note",
			},
		}
	}

	return nil
}

func handlePythonLine(trimmed string, inTripleQuote *bool, tripleQuoteDelim *string) (handled bool, isComment bool) {
	if *inTripleQuote {
		if strings.Contains(trimmed, *tripleQuoteDelim) {
			*inTripleQuote = false
			*tripleQuoteDelim = ""
		}
		return true, true
	}

	for _, delim := range []string{`"""`, `'''`} {
		if strings.HasPrefix(trimmed, delim) {
			rest := trimmed[len(delim):]
			if strings.Contains(rest, delim) {
				return true, true
			}
			*inTripleQuote = true
			*tripleQuoteDelim = delim
			return true, true
		}
	}

	return false, false
}