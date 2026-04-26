package generic

import (
	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type ASTCheck interface {
	Check
	CheckFileAST(path string, content []byte, language string, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation
}

func captureIdxForName(q *tree_sitter.Query, name string) uint32 {
	names := q.CaptureNames()
	for i, n := range names {
		if n == name {
			return uint32(i)
		}
	}
	return 0
}