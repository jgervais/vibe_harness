package generic

import (
	"fmt"
	"strings"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

var g003FunctionQueries = map[string]string{
	"python": `
		(function_definition) @func-def
	`,
	"go": `
		(function_declaration) @func-def
		(method_declaration) @func-def
	`,
	"typescript": `
		(function_declaration) @func-def
		(method_definition) @func-def
		(arrow_function) @func-def
	`,
	"java": `
		(method_declaration) @func-def
		(constructor_declaration) @func-def
	`,
	"ruby": `
		(method) @func-def
		(singleton_method) @func-def
	`,
	"rust": `
		(function_item) @func-def
	`,
}

type MissingLoggingCheck struct {
	querySet *ast.QuerySet
}

func NewMissingLoggingCheck() *MissingLoggingCheck {
	return &MissingLoggingCheck{
		querySet: ast.NewQuerySet(g003FunctionQueries),
	}
}

func (c *MissingLoggingCheck) ID() string   { return "VH-G003" }
func (c *MissingLoggingCheck) Name() string { return "Missing Logging in I/O" }

func (c *MissingLoggingCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	return nil
}

func (c *MissingLoggingCheck) CheckFileAST(path string, content []byte, language string, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
	ioPatterns := ast.IOPatternsForLanguage(language)
	if len(ioPatterns) == 0 {
		return nil
	}

	grammarPtr, ok := ast.GetGrammar(language)
	if !ok {
		return nil
	}
	tsLang := tree_sitter.NewLanguage(grammarPtr)

	funcQuery, err := c.querySet.Compile(language, tsLang)
	if err != nil {
		return nil
	}

	tree := parseResult.Tree()
	if tree == nil {
		return nil
	}
	root := tree.RootNode()
	source := parseResult.Source()

	loggingCalls := cfg.MergedLoggingCalls(language)
	metricsCalls := cfg.MergedMetricsCalls(language)

	funcCursor := tree_sitter.NewQueryCursor()
	defer funcCursor.Close()

	funcMatches := funcCursor.Matches(funcQuery, root, source)

	captureNames := funcQuery.CaptureNames()
	funcDefIdx := uint32(0)
	for i, name := range captureNames {
		if name == "func-def" {
			funcDefIdx = uint32(i)
			break
		}
	}

	var violations []rules.Violation

	for {
		funcMatch := funcMatches.Next()
		if funcMatch == nil {
			break
		}

		var funcNode tree_sitter.Node
		found := false

		for _, capture := range funcMatch.Captures {
			if capture.Index == funcDefIdx {
				funcNode = capture.Node
				found = true
				break
			}
		}
		if !found {
			continue
		}

		ioCalls := findIOCallsInFunc(&funcNode, source, ioPatterns)
		if len(ioCalls) == 0 {
			continue
		}

		hasObservability := hasObservabilityCallInFunc(&funcNode, source, loggingCalls, metricsCalls)
		if hasObservability {
			continue
		}

		for _, callInfo := range ioCalls {
			violations = append(violations, rules.Violation{
				RuleID:   "VH-G003",
				File:     path,
				Line:     callInfo.line,
				Column:   callInfo.col,
				EndLine:  callInfo.line,
				Message:  fmt.Sprintf("I/O call '%s' at line %d is missing logging in the same scope", callInfo.name, callInfo.line),
				Severity: "warning",
			})
		}
	}

	return violations
}

type ioCallInfo struct {
	name string
	line int
	col  int
}

func findIOCallsInFunc(node *tree_sitter.Node, source []byte, ioPatterns []string) []ioCallInfo {
	var results []ioCallInfo
	collectIOCalls(node, source, ioPatterns, &results)
	return results
}

func collectIOCalls(node *tree_sitter.Node, source []byte, ioPatterns []string, results *[]ioCallInfo) {
	if isCallLikeKind(node.Kind()) {
		callText := node.Utf8Text(source)
		for _, pattern := range ioPatterns {
			if strings.Contains(callText, pattern) {
				point := node.StartPosition()
				*results = append(*results, ioCallInfo{
					name: callText,
					line: int(point.Row) + 1,
					col:  int(point.Column),
				})
				return
			}
		}
	}
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		collectIOCalls(child, source, ioPatterns, results)
	}
}

func hasObservabilityCallInFunc(node *tree_sitter.Node, source []byte, loggingCalls, metricsCalls []string) bool {
	allObs := make([]string, 0, len(loggingCalls)+len(metricsCalls))
	allObs = append(allObs, loggingCalls...)
	allObs = append(allObs, metricsCalls...)
	return findObservability(node, source, allObs)
}

func findObservability(node *tree_sitter.Node, source []byte, obsPatterns []string) bool {
	if isCallLikeKind(node.Kind()) {
		callText := node.Utf8Text(source)
		for _, pattern := range obsPatterns {
			if strings.Contains(callText, pattern) {
				return true
			}
		}
	}
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if findObservability(child, source, obsPatterns) {
			return true
		}
	}
	return false
}

func isCallLikeKind(kind string) bool {
	switch kind {
	case "call", "call_expression", "method_invocation", "method_call",
		"macro_invocation", "invoke":
		return true
	default:
		return false
	}
}

func matchesIOPattern(callText string, ioPatterns []string) bool {
	for _, pattern := range ioPatterns {
		if strings.Contains(callText, pattern) {
			return true
		}
	}
	return false
}

func (c *MissingLoggingCheck) Close() {
	c.querySet.Close()
}