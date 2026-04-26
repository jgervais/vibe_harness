package generic

import (
	"os"
	"testing"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
)

func functionLengthFixturePath(t *testing.T, rel string) string {
	t.Helper()
	return testdataPath(t, "function_length/"+rel)
}

func parseFunctionLengthFixture(t *testing.T, rel, language string) (*ast.ParseResult, []byte) {
	t.Helper()
	path := functionLengthFixturePath(t, rel)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	parser := ast.NewParser()
	defer parser.Close()
	result, err := parser.ParseFile(language, content)
	if err != nil {
		t.Fatalf("failed to parse %s: %v", path, err)
	}
	if result == nil {
		t.Fatalf("parse result is nil for %s (language %q unsupported?)", path, language)
	}
	return result, content
}

func TestFunctionLengthCheck_ID(t *testing.T) {
	c := NewFunctionLengthCheck()
	defer c.Close()
	if c.ID() != "VH-G002" {
		t.Errorf("ID() = %q, want %q", c.ID(), "VH-G002")
	}
}

func TestFunctionLengthCheck_Name(t *testing.T) {
	c := NewFunctionLengthCheck()
	defer c.Close()
	if c.Name() != "Function Length" {
		t.Errorf("Name() = %q, want %q", c.Name(), "Function Length")
	}
}

func TestFunctionLengthCheck_UnsupportedLanguage(t *testing.T) {
	c := NewFunctionLengthCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()
	content := []byte("def foo(): pass\n")
	result, err := parser.ParseFile("cobol", content)
	if err != nil || result != nil {
		return
	}
	violations := c.CheckFileAST("test.cbl", content, "cobol", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for unsupported language, got %d", len(violations))
	}
}

func TestFunctionLengthCheck_CommentOnlyBody(t *testing.T) {
	c := NewFunctionLengthCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	pySrc := []byte("def comment_only():\n    # just a comment\n    # another comment\n")
	parser := ast.NewParser()
	defer parser.Close()
	result, err := parser.ParseFile("python", pySrc)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if result == nil {
		t.Fatal("parse result is nil")
	}
	defer result.Close()
	violations := c.CheckFileAST("comment_only.py", pySrc, "python", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for comment-only body, got %d", len(violations))
	}
}

func TestFunctionLengthCheck_LanguageFixtures(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name         string
		cleanFile    string
		violatingFile string
		language     string
		violatingFunc string
	}{
		{
			name:          "Python",
			cleanFile:     "clean.py",
			violatingFile: "violating.py",
			language:      "python",
			violatingFunc: "long_function",
		},
		{
			name:          "Go",
			cleanFile:     "clean.go",
			violatingFile: "violating.go",
			language:      "go",
			violatingFunc: "longFunction",
		},
		{
			name:          "TypeScript",
			cleanFile:     "clean.ts",
			violatingFile: "violating.ts",
			language:      "typescript",
			violatingFunc: "longFunction",
		},
		{
			name:          "Java",
			cleanFile:     "clean.java",
			violatingFile: "violating.java",
			language:      "java",
			violatingFunc: "longMethod",
		},
		{
			name:          "Ruby",
			cleanFile:     "clean.rb",
			violatingFile: "violating.rb",
			language:      "ruby",
			violatingFunc: "long_method",
		},
		{
			name:          "Rust",
			cleanFile:     "clean.rs",
			violatingFile: "violating.rs",
			language:      "rust",
			violatingFunc: "long_function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_clean", func(t *testing.T) {
			c := NewFunctionLengthCheck()
			defer c.Close()

			result, content := parseFunctionLengthFixture(t, tt.cleanFile, tt.language)
			defer result.Close()

			violations := c.CheckFileAST(tt.cleanFile, content, tt.language, &cfg, result)
			if len(violations) != 0 {
				t.Errorf("expected 0 violations for clean file, got %d", len(violations))
				for _, v := range violations {
					t.Logf("  violation: %s", v.Message)
				}
			}
		})

		t.Run(tt.name+"_violating", func(t *testing.T) {
			c := NewFunctionLengthCheck()
			defer c.Close()

			result, content := parseFunctionLengthFixture(t, tt.violatingFile, tt.language)
			defer result.Close()

			violations := c.CheckFileAST(tt.violatingFile, content, tt.language, &cfg, result)
			if len(violations) == 0 {
				t.Fatal("expected at least 1 violation for violating file, got 0")
			}

			v := violations[0]
			if v.RuleID != "VH-G002" {
				t.Errorf("RuleID = %q, want %q", v.RuleID, "VH-G002")
			}
			if v.File != tt.violatingFile {
				t.Errorf("File = %q, want %q", v.File, tt.violatingFile)
			}
			if v.Line < 1 {
				t.Errorf("Line = %d, want >= 1", v.Line)
			}
			if v.Severity != "warning" {
				t.Errorf("Severity = %q, want %q", v.Severity, "warning")
			}
		})
	}
}