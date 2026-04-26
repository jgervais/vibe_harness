package generic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

func swallowedErrorsTestdata(t *testing.T, rel string) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	base := filepath.Dir(filepath.Dir(filepath.Dir(wd)))
	p := filepath.Join(base, "testdata", "swallowed_errors", rel)
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("fixture not found: %s (from wd=%s): %v", p, wd, err)
	}
	return p
}

func parseForTest(t *testing.T, language string, content []byte) *ast.ParseResult {
	t.Helper()
	p := ast.NewParser()
	defer p.Close()
	result, err := p.ParseFile(language, content)
	if err != nil {
		t.Fatalf("ParseFile(%s) err = %v", language, err)
	}
	return result
}

func TestSwallowedErrorsCheck_IDAndName(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	if c.ID() != "VH-G004" {
		t.Errorf("ID() = %q, want %q", c.ID(), "VH-G004")
	}
	if c.Name() != "Swallowed Errors" {
		t.Errorf("Name() = %q, want %q", c.Name(), "Swallowed Errors")
	}
}

func TestSwallowedErrorsCheck_CheckFileReturnsNil(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	content := []byte("try:\n    pass\nexcept:\n    pass\n")
	violations := c.CheckFile("test.py", content, "python", &cfg)
	if violations != nil {
		t.Errorf("CheckFile() = %v, want nil", violations)
	}
}

func TestSwallowedErrorsCheck_PythonViolating(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "violating.py")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "python", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "python", &cfg, pr)
	if len(violations) == 0 {
		t.Fatal("expected violations for violating Python file")
	}
	for _, v := range violations {
		if v.RuleID != "VH-G004" {
			t.Errorf("RuleID = %q, want VH-G004", v.RuleID)
		}
		if v.Severity != "error" {
			t.Errorf("Severity = %q, want error", v.Severity)
		}
	}
}

func TestSwallowedErrorsCheck_PythonClean(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "clean.py")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "python", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "python", &cfg, pr)
	if len(violations) != 0 {
		t.Errorf("clean Python file: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestSwallowedErrorsCheck_PythonBareExcept(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	content := []byte("try:\n    do_something()\nexcept:\n    pass\n")
	pr := parseForTest(t, "python", content)
	defer pr.Close()
	violations := c.CheckFileAST("test.py", content, "python", &cfg, pr)
	if len(violations) == 0 {
		t.Fatal("bare except should be flagged")
	}
}

func TestSwallowedErrorsCheck_TypeScriptViolating(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "violating.ts")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "typescript", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "typescript", &cfg, pr)
	if len(violations) == 0 {
		t.Fatal("expected violations for violating TypeScript file")
	}
	for _, v := range violations {
		if v.RuleID != "VH-G004" {
			t.Errorf("RuleID = %q, want VH-G004", v.RuleID)
		}
	}
}

func TestSwallowedErrorsCheck_TypeScriptClean(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "clean.ts")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "typescript", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "typescript", &cfg, pr)
	if len(violations) != 0 {
		t.Errorf("clean TypeScript file: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestSwallowedErrorsCheck_JavaViolating(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "violating.java")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "java", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "java", &cfg, pr)
	if len(violations) == 0 {
		t.Fatal("expected violations for violating Java file")
	}
	for _, v := range violations {
		if v.RuleID != "VH-G004" {
			t.Errorf("RuleID = %q, want VH-G004", v.RuleID)
		}
	}
}

func TestSwallowedErrorsCheck_JavaClean(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "clean.java")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "java", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "java", &cfg, pr)
	if len(violations) != 0 {
		t.Errorf("clean Java file: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestSwallowedErrorsCheck_RubyViolating(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "violating.rb")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "ruby", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "ruby", &cfg, pr)
	if len(violations) == 0 {
		t.Fatal("expected violations for violating Ruby file")
	}
	for _, v := range violations {
		if v.RuleID != "VH-G004" {
			t.Errorf("RuleID = %q, want VH-G004", v.RuleID)
		}
	}
}

func TestSwallowedErrorsCheck_RubyClean(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "clean.rb")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "ruby", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "ruby", &cfg, pr)
	if len(violations) != 0 {
		t.Errorf("clean Ruby file: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestSwallowedErrorsCheck_GoViolating(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "violating.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "go", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "go", &cfg, pr)
	if len(violations) == 0 {
		t.Fatal("expected violations for violating Go file")
	}
	foundEmptyIf := false
	foundBlankAssign := false
	for _, v := range violations {
		if v.RuleID != "VH-G004" {
			t.Errorf("RuleID = %q, want VH-G004", v.RuleID)
		}
	}
	for _, v := range violations {
		if v.Line == 4 {
			foundEmptyIf = true
		}
		if v.Line == 6 {
			foundBlankAssign = true
		}
	}
	if !foundEmptyIf {
		t.Error("expected violation for empty if err != nil {} block at line 4")
	}
	if !foundBlankAssign {
		t.Error("expected violation for _ = err at line 6")
	}
}

func TestSwallowedErrorsCheck_GoClean(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "clean.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "go", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "go", &cfg, pr)
	if len(violations) != 0 {
		t.Errorf("clean Go file: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestSwallowedErrorsCheck_RustViolating(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "violating.rs")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "rust", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "rust", &cfg, pr)
	if len(violations) == 0 {
		t.Fatal("expected violations for violating Rust file")
	}
	for _, v := range violations {
		if v.RuleID != "VH-G004" {
			t.Errorf("RuleID = %q, want VH-G004", v.RuleID)
		}
	}
}

func TestSwallowedErrorsCheck_RustClean(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	path := swallowedErrorsTestdata(t, "clean.rs")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	pr := parseForTest(t, "rust", content)
	defer pr.Close()
	violations := c.CheckFileAST(path, content, "rust", &cfg, pr)
	if len(violations) != 0 {
		t.Errorf("clean Rust file: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestSwallowedErrorsCheck_UnsupportedLanguage(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	content := []byte("some code\n")
	pr := parseForTest(t, "python", content)
	defer pr.Close()
	violations := c.CheckFileAST("test.txt", content, "brainfuck", &cfg, pr)
	if len(violations) != 0 {
		t.Errorf("unsupported language: got %d violations, want 0", len(violations))
	}
}

func TestSwallowedErrorsCheck_NilParseResult(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	content := []byte("try:\n    pass\nexcept:\n    pass\n")
	violations := c.CheckFileAST("test.py", content, "python", &cfg, nil)
	if len(violations) != 0 {
		t.Errorf("nil parse result: got %d violations, want 0", len(violations))
	}
}

func TestSwallowedErrorsCheck_ViolationValidate(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	content := []byte("try:\n    do_something()\nexcept Exception as e:\n    pass\n")
	pr := parseForTest(t, "python", content)
	defer pr.Close()
	violations := c.CheckFileAST("test.py", content, "python", &cfg, pr)
	if len(violations) == 0 {
		t.Fatal("expected at least one violation")
	}
	for _, v := range violations {
		if err := v.Validate(); err != nil {
			t.Errorf("violation validation failed: %v", err)
		}
	}
}

func TestSwallowedErrorsCheck_ImplementsASTCheck(t *testing.T) {
	var _ ASTCheck = NewSwallowedErrorsCheck()
}

func TestSwallowedErrorsCheck_ImplementsCheck(t *testing.T) {
	var _ Check = NewSwallowedErrorsCheck()
}

func TestSwallowedErrorsCheck_GoReturnErrNotFlagged(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	content := []byte("package main\nfunc example() {\n    _, err := doSomething()\n    if err != nil {\n        return err\n    }\n}\n")
	pr := parseForTest(t, "go", content)
	defer pr.Close()
	violations := c.CheckFileAST("test.go", content, "go", &cfg, pr)
	for _, v := range violations {
		if v.Line == 4 {
			t.Errorf("if err != nil { return err } should not be flagged, got violation at line %d: %s", v.Line, v.Message)
		}
	}
}

func TestSwallowedErrorsCheck_GoBlankIdentErrFlagged(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	content := []byte("package main\nfunc example() {\n    _ = err\n}\n")
	pr := parseForTest(t, "go", content)
	defer pr.Close()
	violations := c.CheckFileAST("test.go", content, "go", &cfg, pr)
	if len(violations) == 0 {
		t.Fatal("_ = err should be flagged")
	}
}

func TestSwallowedErrorsCheck_RustExpectFlagged(t *testing.T) {
	c := NewSwallowedErrorsCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	content := []byte("fn main() {\n    do_something().expect(\"msg\");\n}\n")
	pr := parseForTest(t, "rust", content)
	defer pr.Close()
	violations := c.CheckFileAST("test.rs", content, "rust", &cfg, pr)
	if len(violations) == 0 {
		t.Fatal(".expect() should be flagged")
	}
}

func verifySwallowedErrorsViolation(t *testing.T, v rules.Violation) {
	t.Helper()
	if v.RuleID != "VH-G004" {
		t.Errorf("RuleID = %q, want VH-G004", v.RuleID)
	}
	if v.Severity != "error" {
		t.Errorf("Severity = %q, want error", v.Severity)
	}
	if v.Line < 1 {
		t.Errorf("Line = %d, want >= 1", v.Line)
	}
}