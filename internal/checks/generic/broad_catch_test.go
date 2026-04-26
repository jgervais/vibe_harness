package generic

import (
	"os"
	"testing"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

func TestBroadCatchCheck_IDAndName(t *testing.T) {
	c := NewBroadCatchCheck()
	if c.ID() != "VH-G010" {
		t.Errorf("ID() = %q, want %q", c.ID(), "VH-G010")
	}
	if c.Name() != "Broad Exception Catching" {
		t.Errorf("Name() = %q, want %q", c.Name(), "Broad Exception Catching")
	}
}

func parseAndCheckAST(t *testing.T, check *BroadCatchCheck, fixtureRel string, language string) []rules.Violation {
	t.Helper()
	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	path := testdataPath(t, fixtureRel)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", fixtureRel, err)
	}

	parseResult, err := parser.ParseFile(language, content)
	if err != nil {
		t.Fatalf("parsing %s: %v", fixtureRel, err)
	}
	if parseResult == nil {
		t.Fatalf("parse result is nil for %s (language %s not supported?)", fixtureRel, language)
	}
	defer parseResult.Close()

	return check.CheckFileAST(path, content, language, &cfg, parseResult)
}

func TestBroadCatchCheck_Python_Clean(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	violations := parseAndCheckAST(t, c, "broad_catch/clean.py", "python")
	if len(violations) != 0 {
		t.Errorf("clean Python: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestBroadCatchCheck_Python_Violating(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	violations := parseAndCheckAST(t, c, "broad_catch/violating.py", "python")
	if len(violations) < 2 {
		t.Fatalf("violating Python: got %d violations, want >= 2", len(violations))
	}
	foundException := false
	foundBareExcept := false
	for _, v := range violations {
		if v.RuleID != "VH-G010" {
			t.Errorf("RuleID = %q, want VH-G010", v.RuleID)
		}
		if v.Severity != "warning" {
			t.Errorf("Severity = %q, want warning", v.Severity)
		}
		if v.Message == "Broad exception type 'Exception' caught at line 3" {
			foundException = true
		}
		if v.Message == "Broad exception type 'except' caught at line 8" {
			foundBareExcept = true
		}
	}
	if !foundException {
		t.Error("expected 'Exception' broad catch violation")
	}
	if !foundBareExcept {
		t.Error("expected bare 'except' broad catch violation")
	}
}

func TestBroadCatchCheck_TypeScript_Clean(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	violations := parseAndCheckAST(t, c, "broad_catch/clean.ts", "typescript")
	if len(violations) != 0 {
		t.Errorf("clean TypeScript: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestBroadCatchCheck_TypeScript_Violating(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	violations := parseAndCheckAST(t, c, "broad_catch/violating.ts", "typescript")
	if len(violations) == 0 {
		t.Fatal("violating TypeScript: got 0 violations, want > 0")
	}
	for _, v := range violations {
		if v.RuleID != "VH-G010" {
			t.Errorf("RuleID = %q, want VH-G010", v.RuleID)
		}
		if v.Severity != "warning" {
			t.Errorf("Severity = %q, want warning", v.Severity)
		}
	}
}

func TestBroadCatchCheck_Java_Clean(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	violations := parseAndCheckAST(t, c, "broad_catch/clean.java", "java")
	if len(violations) != 0 {
		t.Errorf("clean Java: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestBroadCatchCheck_Java_Violating(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	violations := parseAndCheckAST(t, c, "broad_catch/violating.java", "java")
	if len(violations) == 0 {
		t.Fatal("violating Java: got 0 violations, want > 0")
	}
	found := false
	for _, v := range violations {
		if v.RuleID != "VH-G010" {
			t.Errorf("RuleID = %q, want VH-G010", v.RuleID)
		}
		if v.Severity != "warning" {
			t.Errorf("Severity = %q, want warning", v.Severity)
		}
		if v.Message == "Broad exception type 'Exception' caught at line 5" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'Exception' broad catch violation in Java")
	}
}

func TestBroadCatchCheck_Ruby_Clean(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	violations := parseAndCheckAST(t, c, "broad_catch/clean.rb", "ruby")
	if len(violations) != 0 {
		t.Errorf("clean Ruby: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestBroadCatchCheck_Ruby_Violating(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	violations := parseAndCheckAST(t, c, "broad_catch/violating.rb", "ruby")
	if len(violations) == 0 {
		t.Fatal("violating Ruby: got 0 violations, want > 0")
	}
	found := false
	for _, v := range violations {
		if v.RuleID != "VH-G010" {
			t.Errorf("RuleID = %q, want VH-G010", v.RuleID)
		}
		if v.Severity != "warning" {
			t.Errorf("Severity = %q, want warning", v.Severity)
		}
		if v.Message == "Broad exception type 'Exception' caught at line 3" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'Exception' broad catch violation in Ruby")
	}
}

func TestBroadCatchCheck_Go_ZeroViolations(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	violations := parseAndCheckAST(t, c, "broad_catch/clean.go", "go")
	if len(violations) != 0 {
		t.Errorf("Go: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestBroadCatchCheck_Rust_ZeroViolations(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	violations := parseAndCheckAST(t, c, "broad_catch/clean.rs", "rust")
	if len(violations) != 0 {
		t.Errorf("Rust: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestBroadCatchCheck_UnsupportedLanguage(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	violations := c.CheckFileAST("test.sql", []byte("SELECT 1"), "sql", &cfg, nil)
	if len(violations) != 0 {
		t.Errorf("unsupported language: got %d violations, want 0", len(violations))
	}
}

func TestBroadCatchCheck_CheckFile_ReturnsNil(t *testing.T) {
	c := NewBroadCatchCheck()
	defer c.Close()
	cfg := config.DefaultConfig()
	violations := c.CheckFile("test.py", []byte("try:\n    pass\nexcept Exception:\n    pass"), "python", &cfg)
	if len(violations) != 0 {
		t.Errorf("CheckFile should always return nil, got %d violations", len(violations))
	}
}