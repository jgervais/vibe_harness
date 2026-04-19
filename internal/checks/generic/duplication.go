package generic

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"

	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

type FileContent struct {
	Path    string
	Content []byte
}

type DuplicationCheck struct{}

func NewDuplicationCheck() *DuplicationCheck {
	return &DuplicationCheck{}
}

func (c *DuplicationCheck) ID() string   { return "VH-G007" }
func (c *DuplicationCheck) Name() string { return "Copy-Paste Duplication" }

func (c *DuplicationCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	return nil
}

var identifierRe = regexp.MustCompile(`^[a-zA-Z_]\w*$`)

func normalizeLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}
	fields := strings.Fields(trimmed)
	var parts []string
	for _, f := range fields {
		if identifierRe.MatchString(f) {
			parts = append(parts, "ID")
		} else {
			parts = append(parts, f)
		}
	}
	return strings.Join(parts, " ")
}

const blockSize = 6

type block struct {
	path      string
	startLine int
	endLine   int
	hash      uint64
}

func computeBlockHash(lines []string) uint64 {
	h := fnv.New64a()
	for _, l := range lines {
		norm := normalizeLine(l)
		h.Write([]byte(norm))
		h.Write([]byte{0})
	}
	return h.Sum64()
}

func extractBlocks(fc FileContent) []block {
	rawLines := strings.Split(string(fc.Content), "\n")
	var codeLines []struct {
		text string
		num  int
	}
	for i, l := range rawLines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			codeLines = append(codeLines, struct {
				text string
				num  int
			}{text: l, num: i + 1})
		}
	}

	var blocks []block
	if len(codeLines) < blockSize {
		return blocks
	}

	for i := 0; i <= len(codeLines)-blockSize; i++ {
		var lines []string
		for j := 0; j < blockSize; j++ {
			lines = append(lines, codeLines[i+j].text)
		}
		blocks = append(blocks, block{
			path:      fc.Path,
			startLine: codeLines[i].num,
			endLine:   codeLines[i+blockSize-1].num,
			hash:      computeBlockHash(lines),
		})
	}
	return blocks
}

func (c *DuplicationCheck) CheckFiles(files []FileContent, cfg *config.Config) []rules.Violation {
	var allBlocks []block
	for _, fc := range files {
		allBlocks = append(allBlocks, extractBlocks(fc)...)
	}

	hashMap := make(map[uint64][]block)
	for _, b := range allBlocks {
		hashMap[b.hash] = append(hashMap[b.hash], b)
	}

	var violations []rules.Violation
	seen := make(map[string]bool)

	for _, blocks := range hashMap {
		if len(blocks) < 2 {
			continue
		}
		for i := 0; i < len(blocks); i++ {
			for j := i + 1; j < len(blocks); j++ {
				a, b := blocks[i], blocks[j]
				if a.path == b.path {
					continue
				}
				key := fmt.Sprintf("%s:%d-%s:%d", a.path, a.startLine, b.path, b.startLine)
				if seen[key] {
					continue
				}
				seen[key] = true
				violations = append(violations, rules.Violation{
					RuleID:   "VH-G007",
					File:     a.path,
					Line:     a.startLine,
					Column:   0,
					EndLine:  a.endLine,
					Message:  fmt.Sprintf("duplicate code block (6+ lines, 80%%+ similarity) also in %s:%d", b.path, b.startLine),
					Severity: "warning",
				})
			}
		}
	}

	return violations
}