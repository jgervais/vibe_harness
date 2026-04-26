package generic

import (
	"fmt"
	"strings"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type MissingErrorHandlingCheck struct {
	querySet *ast.QuerySet
}

func NewMissingErrorHandlingCheck() *MissingErrorHandlingCheck {
	ioQueryPatterns := map[string]string{}
	for _, lang := range ast.SupportedLanguages() {
		patterns := ast.IOQueryPatterns(lang)
		if len(patterns) == 0 {
			continue
		}
		var parts []string
		for _, q := range patterns {
			parts = append(parts, q)
		}
		ioQueryPatterns[lang] = strings.Join(parts, "\n")
	}

	return &MissingErrorHandlingCheck{
		querySet: ast.NewQuerySet(ioQueryPatterns),
	}
}

func (c *MissingErrorHandlingCheck) ID() string   { return "VH-G009" }
func (c *MissingErrorHandlingCheck) Name() string { return "Missing Error Handling on I/O" }

func (c *MissingErrorHandlingCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	return nil
}

func (c *MissingErrorHandlingCheck) CheckFileAST(path string, content []byte, language string, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
	ioPatterns := ast.IOPatternsForLanguage(language)
	if len(ioPatterns) == 0 {
		return nil
	}

	_, ok := c.querySet.GetQuery(language)
	if !ok {
		return nil
	}

	grammarPtr, ok := ast.GetGrammar(language)
	if !ok {
		return nil
	}
	tsLang := tree_sitter.NewLanguage(grammarPtr)

	query, err := c.querySet.Compile(language, tsLang)
	if err != nil {
		return nil
	}

	tree := parseResult.Tree()
	if tree == nil {
		return nil
	}
	root := tree.RootNode()
	source := parseResult.Source()

	ioCallIdx := captureIdxForName(query, "io-call")

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	matches := cursor.Matches(query, root, source)

	var violations []rules.Violation

	for {
		match := matches.Next()
		if match == nil {
			break
		}

		for _, capture := range match.Captures {
			if capture.Index != ioCallIdx {
				continue
			}

			node := capture.Node
			callText := node.Utf8Text(source)

			if !matchesIOPattern(callText, ioPatterns) {
				continue
			}

			if isErrorHandled(&node, language, source) {
				continue
			}

			startPos := node.StartPosition()
			endPos := node.EndPosition()
			violations = append(violations, rules.Violation{
				RuleID:   "VH-G009",
				File:     path,
				Line:     int(startPos.Row) + 1,
				Column:   int(startPos.Column),
				EndLine:  int(endPos.Row) + 1,
				Message:  fmt.Sprintf("I/O call '%s' at line %d is not wrapped in error handling", callText, int(startPos.Row)+1),
				Severity: "error",
			})
		}
	}

	return violations
}

func isErrorHandled(node *tree_sitter.Node, language string, source []byte) bool {
	switch language {
	case "python", "typescript", "java":
		return isInsideTry(node)
	case "ruby":
		return isInsideRubyBegin(node)
	case "go":
		return isGoErrorHandled(node, source)
	case "rust":
		return isRustErrorHandled(node, source)
	}
	return false
}

func isInsideTry(node *tree_sitter.Node) bool {
	current := node.Parent()
	for current != nil {
		kind := current.Kind()
		if kind == "try_statement" {
			return true
		}
		current = current.Parent()
	}
	return false
}

func isInsideRubyBegin(node *tree_sitter.Node) bool {
	current := node.Parent()
	for current != nil {
		kind := current.Kind()
		if kind == "begin" || kind == "rescue_clause" {
			return true
		}
		current = current.Parent()
	}
	return false
}

func isGoErrorHandled(node *tree_sitter.Node, source []byte) bool {
	target := node
	for target != nil {
		if target.Kind() == "short_var_declaration" || target.Kind() == "var_declaration" || target.Kind() == "expression_statement" {
			break
		}
		target = target.Parent()
	}
	if target == nil {
		return false
	}

	nextSibling := target.NextNamedSibling()
	for nextSibling != nil {
		if nextSibling.Kind() == "if_statement" {
			for i := uint(0); i < nextSibling.ChildCount(); i++ {
				child := nextSibling.Child(i)
				childText := child.Utf8Text(source)
				if strings.Contains(childText, "err") && !strings.Contains(childText, "os.") {
					return true
				}
			}
		}
		break
	}

	return false
}

func isRustErrorHandled(node *tree_sitter.Node, source []byte) bool {
	remaining := string(source[node.EndByte():])
	if len(remaining) > 0 && remaining[0] == '?' {
		return true
	}

	parent := node.Parent()
	if parent != nil && parent.Kind() == "try_expression" {
		return true
	}

	current := node.Parent()
	for current != nil {
		kind := current.Kind()
		if kind == "match_expression" || kind == "try_expression" {
			return true
		}
		current = current.Parent()
	}

	return false
}

func (c *MissingErrorHandlingCheck) Close() {
	if c.querySet != nil {
		c.querySet.Close()
	}
}