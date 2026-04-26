package generic

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

const godModuleThreshold = 20

var g012Queries = map[string]string{
	"python":     "(function_definition name: (identifier) @name) (class_definition name: (identifier) @name)",
	"typescript": "(export_statement) @export",
	"go":         "(function_declaration name: (identifier) @name) (method_declaration name: (field_identifier) @name)",
	"java":       "(method_declaration) @method (class_declaration) @class",
	"ruby":       "(method name: (identifier) @name) (singleton_method name: (identifier) @name)",
	"rust":       "(function_item) @fn (struct_item) @struct (enum_item) @enum",
}

type GodModuleCheck struct {
	querySet *ast.QuerySet
}

func NewGodModuleCheck() *GodModuleCheck {
	return &GodModuleCheck{
		querySet: ast.NewQuerySet(g012Queries),
	}
}

func (c *GodModuleCheck) ID() string   { return "VH-G012" }
func (c *GodModuleCheck) Name() string { return "God Module" }

func (c *GodModuleCheck) CheckFile(path string, content []byte, language string, cfg *config.Config) []rules.Violation {
	return nil
}

func (c *GodModuleCheck) CheckFileAST(path string, content []byte, language string, cfg *config.Config, parseResult *ast.ParseResult) []rules.Violation {
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

	count := 0

	switch language {
	case "python":
		nameIdx, _ := query.CaptureIndexForName("name")
		nameIdx32 := uint32(nameIdx)
		for {
			match := matches.Next()
			if match == nil {
				break
			}
			for _, capture := range match.Captures {
				if capture.Index == nameIdx32 {
					text := capture.Node.Utf8Text(parseResult.Source())
					if !strings.HasPrefix(text, "_") {
						count++
					}
				}
			}
		}
	case "typescript":
		exportIdx, _ := query.CaptureIndexForName("export")
		exportIdx32 := uint32(exportIdx)
		for {
			match := matches.Next()
			if match == nil {
				break
			}
			for _, capture := range match.Captures {
				if capture.Index == exportIdx32 {
					count++
				}
			}
		}
	case "go":
		nameIdx, _ := query.CaptureIndexForName("name")
		nameIdx32 := uint32(nameIdx)
		for {
			match := matches.Next()
			if match == nil {
				break
			}
			for _, capture := range match.Captures {
				if capture.Index == nameIdx32 {
					text := capture.Node.Utf8Text(parseResult.Source())
					if len(text) > 0 && unicode.IsUpper(rune(text[0])) {
						count++
					}
				}
			}
		}
	case "java":
		methodIdx, _ := query.CaptureIndexForName("method")
		methodIdx32 := uint32(methodIdx)
		classIdx, _ := query.CaptureIndexForName("class")
		classIdx32 := uint32(classIdx)
		for {
			match := matches.Next()
			if match == nil {
				break
			}
			for _, capture := range match.Captures {
				if capture.Index == methodIdx32 {
					if hasJavaPublicModifier(&capture.Node, parseResult.Source()) {
						count++
					}
				} else if capture.Index == classIdx32 {
					if hasJavaPublicModifier(&capture.Node, parseResult.Source()) {
						count++
					}
				}
			}
		}
	case "ruby":
		nameIdx, _ := query.CaptureIndexForName("name")
		nameIdx32 := uint32(nameIdx)
		for {
			match := matches.Next()
			if match == nil {
				break
			}
			for _, capture := range match.Captures {
				if capture.Index == nameIdx32 {
					text := capture.Node.Utf8Text(parseResult.Source())
					if !strings.HasPrefix(text, "_") {
						count++
					}
				}
			}
		}
	case "rust":
		fnIdx, _ := query.CaptureIndexForName("fn")
		fnIdx32 := uint32(fnIdx)
		structIdx, _ := query.CaptureIndexForName("struct")
		structIdx32 := uint32(structIdx)
		enumIdx, _ := query.CaptureIndexForName("enum")
		enumIdx32 := uint32(enumIdx)
		for {
			match := matches.Next()
			if match == nil {
				break
			}
			for _, capture := range match.Captures {
				if capture.Index == fnIdx32 || capture.Index == structIdx32 || capture.Index == enumIdx32 {
					if hasRustPublicVisibility(&capture.Node, parseResult.Source()) {
						count++
					}
				}
			}
		}
	}

	if count > godModuleThreshold {
		return []rules.Violation{
			{
				RuleID:   "VH-G012",
				File:     path,
				Line:     1,
				Column:   0,
				EndLine:  0,
				Message:  fmt.Sprintf("File has %d public exports (threshold: %d)", count, godModuleThreshold),
				Severity: "warning",
			},
		}
	}

	return nil
}

func hasJavaPublicModifier(node *tree_sitter.Node, source []byte) bool {
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Kind() == "modifiers" {
			for j := uint(0); j < child.ChildCount(); j++ {
				modChild := child.Child(j)
				if modChild.Utf8Text(source) == "public" {
					return true
				}
			}
			return false
		}
	}
	return false
}

func hasRustPublicVisibility(node *tree_sitter.Node, source []byte) bool {
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Kind() == "visibility_modifier" {
			return isRustPublic(child.Utf8Text(source))
		}
	}
	return false
}

func isRustPublic(vis string) bool {
	if vis == "pub" {
		return true
	}
	if strings.HasPrefix(vis, "pub(") {
		inner := strings.TrimPrefix(vis, "pub(")
		inner = strings.TrimSuffix(inner, ")")
		if inner == "crate" || inner == "super" || strings.HasPrefix(inner, "in ") {
			return false
		}
		return true
	}
	return false
}

func (c *GodModuleCheck) Close() {
	c.querySet.Close()
}