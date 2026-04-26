package generic

import (
	"fmt"
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

type SwallowedErrorsCheck struct {
	parser *ast.Parser
	qs     *ast.QuerySet
}

func NewSwallowedErrorsCheck() *SwallowedErrorsCheck {
	return &SwallowedErrorsCheck{
		parser: ast.NewParser(),
		qs:     ast.NewQuerySet(g004Queries),
	}
}

func (c *SwallowedErrorsCheck) ID() string   { return "VH-G004" }
func (c *SwallowedErrorsCheck) Name() string { return "Swallowed Errors" }

func (c *SwallowedErrorsCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	return nil
}

func (c *SwallowedErrorsCheck) CheckFileAST(path string, content []byte, language string, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
	if parseResult == nil {
		return nil
	}

	switch language {
	case "python":
		return c.checkPython(path, content, cfg, parseResult)
	case "typescript":
		return c.checkTypeScript(path, content, cfg, parseResult)
	case "java":
		return c.checkJava(path, content, cfg, parseResult)
	case "ruby":
		return c.checkRuby(path, content, cfg, parseResult)
	case "go":
		return c.checkGo(path, content, cfg, parseResult)
	case "rust":
		return c.checkRust(path, content, cfg, parseResult)
	default:
		return nil
	}
}

func (c *SwallowedErrorsCheck) Close() {
	if c.parser != nil {
		c.parser.Close()
	}
	if c.qs != nil {
		c.qs.Close()
	}
}

var g004Queries = map[string]string{
	"python":     "(except_clause) @except",
	"typescript": "(catch_clause) @catch",
	"java":       "(catch_clause) @catch",
	"ruby":       "(rescue) @rescue",
	"go":         "(if_statement) @if_stmt",
	"rust":       "(call_expression function: (field_expression) @method) @call",
}

func (c *SwallowedErrorsCheck) checkPython(path string, content []byte, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
	var violations []rules.Violation
	tree := parseResult.Tree()
	root := tree.RootNode()
	source := parseResult.Source()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	lang, _ := ast.GetGrammar("python")
	tsLang := tree_sitter.NewLanguage(lang)
	query, err := c.qs.Compile("python", tsLang)
	if err != nil {
		return nil
	}

	exceptIdx := captureIdxForName(query, "except")

	matches := cursor.Matches(query, root, source)
	for {
		m := matches.Next()
		if m == nil {
			break
		}
		for _, cap := range m.Captures {
			if cap.Index != exceptIdx {
				continue
			}
			node := cap.Node
			if node.Kind() != "except_clause" {
				continue
			}
			line := int(node.StartPosition().Row) + 1
			if c.isPythonExceptSwallowed(&node, source, cfg) {
				violations = append(violations, rules.Violation{
					RuleID:   "VH-G004",
					File:     path,
					Line:     line,
					Column:   int(node.StartPosition().Column),
					EndLine:  line,
					Message:  fmt.Sprintf("Error is swallowed in catch/except block at line %d", line),
					Severity: "error",
				})
			}
		}
	}

	return violations
}

func (c *SwallowedErrorsCheck) isPythonExceptSwallowed(node *tree_sitter.Node, source []byte, cfg *config.Config) bool {
	blockNode := node.NamedChild(node.NamedChildCount() - 1)
	if blockNode == nil || blockNode.Kind() != "block" {
		return true
	}
	if blockNode == nil || blockNode.Kind() == "" {
		return true
	}
	if blockNode.NamedChildCount() == 0 {
		return true
	}
	for i := uint(0); i < blockNode.NamedChildCount(); i++ {
		child := blockNode.NamedChild(i)
		if child == nil {
			continue
		}
		if child.Kind() == "pass_statement" {
			continue
		}
		if child.Kind() == "raise_statement" {
			return false
		}
		if child.Kind() == "return_statement" {
			return false
		}
		if c.isLoggingCallNodePtr(child, source, cfg.MergedLoggingCalls("python")) {
			return false
		}
		return false
	}
	return true
}

func (c *SwallowedErrorsCheck) checkTypeScript(path string, content []byte, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
	var violations []rules.Violation
	tree := parseResult.Tree()
	root := tree.RootNode()
	source := parseResult.Source()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	lang, _ := ast.GetGrammar("typescript")
	tsLang := tree_sitter.NewLanguage(lang)
	query, err := c.qs.Compile("typescript", tsLang)
	if err != nil {
		return nil
	}

	catchIdx := captureIdxForName(query, "catch")

	matches := cursor.Matches(query, root, source)
	for {
		m := matches.Next()
		if m == nil {
			break
		}
		for _, cap := range m.Captures {
			if cap.Index != catchIdx {
				continue
			}
			node := cap.Node
			if node.Kind() != "catch_clause" {
				continue
			}
			line := int(node.StartPosition().Row) + 1
			if c.isTypeScriptCatchSwallowed(&node, source, cfg) {
				violations = append(violations, rules.Violation{
					RuleID:   "VH-G004",
					File:     path,
					Line:     line,
					Column:   int(node.StartPosition().Column),
					EndLine:  line,
					Message:  fmt.Sprintf("Error is swallowed in catch/except block at line %d", line),
					Severity: "error",
				})
			}
		}
	}

	return violations
}

func (c *SwallowedErrorsCheck) isTypeScriptCatchSwallowed(node *tree_sitter.Node, source []byte, cfg *config.Config) bool {
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil || bodyNode.Kind() == "" {
		return true
	}
	if bodyNode.NamedChildCount() == 0 {
		return true
	}
	for i := uint(0); i < bodyNode.NamedChildCount(); i++ {
		child := bodyNode.NamedChild(i)
		if child == nil {
			continue
		}
		if child.Kind() == "throw_statement" {
			return false
		}
		if child.Kind() == "return_statement" {
			return false
		}
		if c.isLoggingCallNodePtr(child, source, cfg.MergedLoggingCalls("typescript")) {
			return false
		}
	}
	return true
}

func (c *SwallowedErrorsCheck) checkJava(path string, content []byte, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
	var violations []rules.Violation
	tree := parseResult.Tree()
	root := tree.RootNode()
	source := parseResult.Source()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	lang, _ := ast.GetGrammar("java")
	tsLang := tree_sitter.NewLanguage(lang)
	query, err := c.qs.Compile("java", tsLang)
	if err != nil {
		return nil
	}

	catchIdx := captureIdxForName(query, "catch")

	matches := cursor.Matches(query, root, source)
	for {
		m := matches.Next()
		if m == nil {
			break
		}
		for _, cap := range m.Captures {
			if cap.Index != catchIdx {
				continue
			}
			node := cap.Node
			if node.Kind() != "catch_clause" {
				continue
			}
			line := int(node.StartPosition().Row) + 1
			if c.isJavaCatchSwallowed(&node, source, cfg) {
				violations = append(violations, rules.Violation{
					RuleID:   "VH-G004",
					File:     path,
					Line:     line,
					Column:   int(node.StartPosition().Column),
					EndLine:  line,
					Message:  fmt.Sprintf("Error is swallowed in catch/except block at line %d", line),
					Severity: "error",
				})
			}
		}
	}

	return violations
}

func (c *SwallowedErrorsCheck) isJavaCatchSwallowed(node *tree_sitter.Node, source []byte, cfg *config.Config) bool {
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil || bodyNode.Kind() == "" {
		return true
	}
	if bodyNode.NamedChildCount() == 0 {
		return true
	}
	for i := uint(0); i < bodyNode.NamedChildCount(); i++ {
		child := bodyNode.NamedChild(i)
		if child == nil {
			continue
		}
		if child.Kind() == "throw_statement" {
			return false
		}
		if child.Kind() == "return_statement" {
			return false
		}
		if c.isLoggingCallNodePtr(child, source, cfg.MergedLoggingCalls("java")) {
			return false
		}
	}
	return true
}

func (c *SwallowedErrorsCheck) checkRuby(path string, content []byte, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
	var violations []rules.Violation
	tree := parseResult.Tree()
	root := tree.RootNode()
	source := parseResult.Source()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	lang, _ := ast.GetGrammar("ruby")
	tsLang := tree_sitter.NewLanguage(lang)
	query, err := c.qs.Compile("ruby", tsLang)
	if err != nil {
		return nil
	}

	rescueIdx := captureIdxForName(query, "rescue")

	matches := cursor.Matches(query, root, source)
	for {
		m := matches.Next()
		if m == nil {
			break
		}
		for _, cap := range m.Captures {
			if cap.Index != rescueIdx {
				continue
			}
			node := cap.Node
			if node.Kind() != "rescue" {
				continue
			}
			line := int(node.StartPosition().Row) + 1
			if c.isRubyRescueSwallowed(&node, source, cfg) {
				violations = append(violations, rules.Violation{
					RuleID:   "VH-G004",
					File:     path,
					Line:     line,
					Column:   int(node.StartPosition().Column),
					EndLine:  line,
					Message:  fmt.Sprintf("Error is swallowed in catch/except block at line %d", line),
					Severity: "error",
				})
			}
		}
	}

	return violations
}

func (c *SwallowedErrorsCheck) isRubyRescueSwallowed(node *tree_sitter.Node, source []byte, cfg *config.Config) bool {
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil || bodyNode.Kind() == "" {
		return true
	}
	if bodyNode.NamedChildCount() == 0 {
		return true
	}
	for i := uint(0); i < bodyNode.NamedChildCount(); i++ {
		child := bodyNode.NamedChild(i)
		if child == nil {
			continue
		}
		text := string(source[child.StartByte():child.EndByte()])
		if text == "nil" && bodyNode.NamedChildCount() == 1 {
			return true
		}
		if child.Kind() == "identifier" {
			lowered := strings.ToLower(text)
			if lowered == "raise" || lowered == "fail" {
				return false
			}
		}
		if c.isLoggingCallNodePtr(child, source, cfg.MergedLoggingCalls("ruby")) {
			return false
		}
	}
	return true
}

func (c *SwallowedErrorsCheck) checkGo(path string, content []byte, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
	var violations []rules.Violation
	tree := parseResult.Tree()
	root := tree.RootNode()
	source := parseResult.Source()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	lang, _ := ast.GetGrammar("go")
	tsLang := tree_sitter.NewLanguage(lang)
	query, err := c.qs.Compile("go", tsLang)
	if err != nil {
		return nil
	}

	ifStmtIdx := captureIdxForName(query, "if_stmt")

	matches := cursor.Matches(query, root, source)
	for {
		m := matches.Next()
		if m == nil {
			break
		}
		for _, cap := range m.Captures {
			if cap.Index != ifStmtIdx {
				continue
			}
			node := cap.Node
			if node.Kind() != "if_statement" {
				continue
			}
			condNode := node.ChildByFieldName("condition")
			if condNode == nil || condNode.Kind() == "" {
				continue
			}
			condText := string(source[condNode.StartByte():condNode.EndByte()])
			if !strings.Contains(condText, "err") {
				continue
			}
			consequenceNode := node.ChildByFieldName("consequence")
			if consequenceNode == nil || consequenceNode.Kind() == "" {
				continue
			}
			if consequenceNode.NamedChildCount() == 0 {
				line := int(node.StartPosition().Row) + 1
				violations = append(violations, rules.Violation{
					RuleID:   "VH-G004",
					File:     path,
					Line:     line,
					Column:   int(node.StartPosition().Column),
					EndLine:  line,
					Message:  fmt.Sprintf("Error is swallowed in catch/except block at line %d", line),
					Severity: "error",
				})
			}
		}
	}

	c.findGoBlankIdentAssign(path, source, root, &violations)

	return violations
}

func (c *SwallowedErrorsCheck) findGoBlankIdentAssign(path string, source []byte, node *tree_sitter.Node, violations *[]rules.Violation) {
	if node.Kind() == "assignment_statement" {
		leftNode := node.ChildByFieldName("left")
		if leftNode != nil && leftNode.Kind() != "" {
			leftText := string(source[leftNode.StartByte():leftNode.EndByte()])
			rightNode := node.ChildByFieldName("right")
			rightText := ""
			if rightNode != nil && rightNode.Kind() != "" {
				rightText = string(source[rightNode.StartByte():rightNode.EndByte()])
			}
			if strings.TrimSpace(leftText) == "_" && strings.Contains(rightText, "err") {
				line := int(node.StartPosition().Row) + 1
				*violations = append(*violations, rules.Violation{
					RuleID:   "VH-G004",
					File:     path,
					Line:     line,
					Column:   int(node.StartPosition().Column),
					EndLine:  line,
					Message:  fmt.Sprintf("Error is swallowed in catch/except block at line %d", line),
					Severity: "error",
				})
				return
			}
		}
	}

	walkCursor := node.Walk()
	defer walkCursor.Close()

	if walkCursor.GotoFirstChild() {
		for {
			c.findGoBlankIdentAssign(path, source, walkCursor.Node(), violations)
			if !walkCursor.GotoNextSibling() {
				break
			}
		}
	}
}

func (c *SwallowedErrorsCheck) checkRust(path string, content []byte, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
	var violations []rules.Violation
	tree := parseResult.Tree()
	root := tree.RootNode()
	source := parseResult.Source()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	lang, _ := ast.GetGrammar("rust")
	tsLang := tree_sitter.NewLanguage(lang)
	query, err := c.qs.Compile("rust", tsLang)
	if err != nil {
		return nil
	}

	matches := cursor.Matches(query, root, source)
	for {
		m := matches.Next()
		if m == nil {
			break
		}
		for _, cap := range m.Captures {
			if cap.Node.Kind() == "field_expression" {
				for i := uint(0); i < cap.Node.NamedChildCount(); i++ {
					child := cap.Node.NamedChild(i)
					if child == nil || child.Kind() != "field_identifier" {
						continue
					}
					text := string(source[child.StartByte():child.EndByte()])
					if text == "unwrap" || text == "expect" {
						line := int(cap.Node.StartPosition().Row) + 1
						violations = append(violations, rules.Violation{
							RuleID:   "VH-G004",
							File:     path,
							Line:     line,
							Column:   int(cap.Node.StartPosition().Column),
							EndLine:  line,
							Message:  fmt.Sprintf("Error is swallowed in catch/except block at line %d", line),
							Severity: "error",
						})
					}
				}
			}
		}
	}

	return violations
}

func (c *SwallowedErrorsCheck) isLoggingCallNodePtr(node *tree_sitter.Node, source []byte, loggingCalls []string) bool {
	var callText string
	var callName string

	switch node.Kind() {
	case "expression_statement":
		for i := uint(0); i < node.NamedChildCount(); i++ {
			child := node.NamedChild(i)
			if child == nil {
				continue
			}
			if child.Kind() == "call" || child.Kind() == "call_expression" {
				return c.isLoggingCallNodePtr(child, source, loggingCalls)
			}
			if child.Kind() == "method_invocation" {
				return c.isLoggingCallNodePtr(child, source, loggingCalls)
			}
		}
		return false
	case "call", "call_expression":
		callText = string(source[node.StartByte():node.EndByte()])
		funcChild := node.ChildByFieldName("function")
		if funcChild != nil && funcChild.Kind() != "" {
			callName = string(source[funcChild.StartByte():funcChild.EndByte()])
		}
	case "method_invocation":
		objNode := node.ChildByFieldName("object")
		nameNode := node.ChildByFieldName("name")
		if objNode != nil && nameNode != nil && objNode.Kind() != "" && nameNode.Kind() != "" {
			objText := string(source[objNode.StartByte():objNode.EndByte()])
			nameText := string(source[nameNode.StartByte():nameNode.EndByte()])
			callName = objText + "." + nameText
			callText = callName
		}
	default:
		return false
	}

	if callName == "" {
		return false
	}

	for _, lc := range loggingCalls {
		if strings.HasPrefix(callName, lc) || strings.Contains(callText, lc) {
			return true
		}
	}
	return false
}