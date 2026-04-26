package generic

import (
	"fmt"
	"strings"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

var g010Queries = map[string]string{
	"python":     "(except_clause) @except",
	"typescript": "(catch_clause) @catch",
	"java":       "(catch_formal_parameter) @catch_param",
	"ruby":       "(rescue) @rescue",
}

var g010BroadTypes = map[string]map[string]bool{
	"python": {"Exception": true, "BaseException": true},
	"java":   {"Exception": true, "Throwable": true, "RuntimeException": true},
	"ruby":   {"Exception": true, "StandardError": true},
}

type BroadCatchCheck struct {
	querySet *ast.QuerySet
}

func NewBroadCatchCheck() *BroadCatchCheck {
	return &BroadCatchCheck{
		querySet: ast.NewQuerySet(g010Queries),
	}
}

func (c *BroadCatchCheck) ID() string   { return "VH-G010" }
func (c *BroadCatchCheck) Name() string { return "Broad Exception Catching" }

func (c *BroadCatchCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	return nil
}

func (c *BroadCatchCheck) CheckFileAST(path string, content []byte, language string, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
	if language == "go" || language == "rust" {
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

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	matches := cursor.Matches(query, root, parseResult.Source())

	var violations []rules.Violation

	for {
		match := matches.Next()
		if match == nil {
			break
		}

		for _, capture := range match.Captures {
			node := capture.Node
			broadType := c.detectBroadCatch(language, &node, parseResult.Source())
			if broadType != "" {
				startPos := node.StartPosition()
				violations = append(violations, rules.Violation{
					RuleID:   "VH-G010",
					File:     path,
					Line:     int(startPos.Row) + 1,
					Column:   int(startPos.Column),
					EndLine:  int(node.EndPosition().Row) + 1,
					Message:  fmt.Sprintf("Broad exception type '%s' caught at line %d", broadType, int(startPos.Row)+1),
					Severity: "warning",
				})
			}
		}
	}

	return violations
}

func (c *BroadCatchCheck) detectBroadCatch(language string, node *tree_sitter.Node, source []byte) string {
	broadTypes, ok := g010BroadTypes[language]
	if !ok {
		if language == "typescript" {
			return "any"
		}
		return ""
	}

	switch language {
	case "python":
		return c.detectPythonBroadCatch(node, source, broadTypes)
	case "java":
		return c.detectJavaBroadCatch(node, source, broadTypes)
	case "ruby":
		return c.detectRubyBroadCatch(node, source, broadTypes)
	}

	return ""
}

func (c *BroadCatchCheck) detectPythonBroadCatch(node *tree_sitter.Node, source []byte, broadTypes map[string]bool) string {
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Kind() == "as_pattern" {
			for j := uint(0); j < child.ChildCount(); j++ {
				grandchild := child.Child(j)
				if grandchild.Kind() == "identifier" {
					typeText := grandchild.Utf8Text(source)
					if broadTypes[typeText] {
						return typeText
					}
					return ""
				}
			}
			return ""
		}
	}

	return "except"
}

func (c *BroadCatchCheck) detectJavaBroadCatch(node *tree_sitter.Node, source []byte, broadTypes map[string]bool) string {
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Kind() == "catch_type" {
			for j := uint(0); j < child.ChildCount(); j++ {
				grandchild := child.Child(j)
				if grandchild.IsNamed() {
					typeText := grandchild.Utf8Text(source)
					if broadTypes[typeText] {
						return typeText
					}
					return ""
				}
			}
			return ""
		}
	}
	return ""
}

func (c *BroadCatchCheck) detectRubyBroadCatch(node *tree_sitter.Node, source []byte, broadTypes map[string]bool) string {
	hasExceptions := false
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Kind() == "exceptions" {
			hasExceptions = true
			for j := uint(0); j < child.ChildCount(); j++ {
				grandchild := child.Child(j)
				if grandchild.IsNamed() {
					typeText := strings.TrimSpace(grandchild.Utf8Text(source))
					if broadTypes[typeText] {
						return typeText
					}
				}
			}
		}
	}

	if !hasExceptions {
		return "rescue"
	}

	return ""
}

func (c *BroadCatchCheck) Close() {
	c.querySet.Close()
}