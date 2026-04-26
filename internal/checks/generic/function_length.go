package generic

import (
	"fmt"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

const functionLengthThreshold = 50

var g002Queries = map[string]string{
	"python": `
		(function_definition name: (identifier) @func-name body: (block) @func-body)
	`,
	"go": `
		(function_declaration name: (identifier) @func-name body: (block) @func-body)
		(method_declaration name: (field_identifier) @func-name body: (block) @func-body)
	`,
	"typescript": `
		(function_declaration name: (identifier) @func-name body: (statement_block) @func-body)
		(method_definition name: (property_identifier) @func-name body: (statement_block) @func-body)
	`,
	"java": `
		(method_declaration (identifier) @func-name body: (block) @func-body)
		(constructor_declaration (identifier) @func-name body: (constructor_body) @func-body)
	`,
	"ruby": `
		(method name: (identifier) @func-name body: (_) @func-body)
		(singleton_method name: (identifier) @func-name body: (_) @func-body)
	`,
	"rust": `
		(function_item name: (identifier) @func-name body: (block) @func-body)
	`,
}

type FunctionLengthCheck struct {
	querySet *ast.QuerySet
}

func NewFunctionLengthCheck() *FunctionLengthCheck {
	return &FunctionLengthCheck{
		querySet: ast.NewQuerySet(g002Queries),
	}
}

func (c *FunctionLengthCheck) ID() string   { return "VH-G002" }
func (c *FunctionLengthCheck) Name() string { return "Function Length" }

func (c *FunctionLengthCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	return nil
}

func (c *FunctionLengthCheck) CheckFileAST(path string, content []byte, language string, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
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

	captureNames := query.CaptureNames()
	var funcNameIdx uint32
	var funcBodyIdx uint32
	for i, name := range captureNames {
		switch name {
		case "func-name":
			funcNameIdx = uint32(i)
		case "func-body":
			funcBodyIdx = uint32(i)
		}
	}

	var violations []rules.Violation

	for {
		match := matches.Next()
		if match == nil {
			break
		}

		var funcName string
		var bodyNode *tree_sitter.Node

		for _, capture := range match.Captures {
			if capture.Index == funcNameIdx {
				funcName = capture.Node.Utf8Text(parseResult.Source())
			}
			if capture.Index == funcBodyIdx {
				node := capture.Node
				bodyNode = &node
			}
		}

		if bodyNode == nil || funcName == "" {
			continue
		}

		stmtCount := countStatements(bodyNode)

		if stmtCount > functionLengthThreshold {
			startPos := bodyNode.StartPosition()
			endPos := bodyNode.EndPosition()
			violations = append(violations, rules.Violation{
				RuleID:   "VH-G002",
				File:     path,
				Line:     int(startPos.Row) + 1,
				Column:   int(startPos.Column),
				EndLine:  int(endPos.Row) + 1,
				Message:  fmt.Sprintf("Function '%s' has %d statements (threshold: %d)", funcName, stmtCount, functionLengthThreshold),
				Severity: "warning",
			})
		}
	}

	return violations
}

func countStatements(bodyNode *tree_sitter.Node) int {
	var listNode *tree_sitter.Node
	for i := uint(0); i < bodyNode.ChildCount(); i++ {
		child := bodyNode.Child(i)
		if child != nil && child.IsNamed() && child.Kind() == "statement_list" {
			node := child
			listNode = node
			break
		}
	}
	target := bodyNode
	if listNode != nil {
		target = listNode
	}

	count := 0
	childCount := target.ChildCount()
	for i := uint(0); i < childCount; i++ {
		child := target.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() {
			continue
		}
		kind := child.Kind()
		if isComment(kind) {
			continue
		}
		count++
	}
	return count
}

func isComment(kind string) bool {
	switch kind {
	case "comment", "line_comment", "block_comment",
		"python_comment", "comment_block",
		"single_line_comment", "multi_line_comment",
		"traditional_comment", "line_comment_java":
		return true
	}
	return false
}

func (c *FunctionLengthCheck) Close() {
	c.querySet.Close()
}