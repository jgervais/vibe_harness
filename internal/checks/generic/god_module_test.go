package generic

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
)

func godModuleTestdataPath(t *testing.T, rel string) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	base := filepath.Dir(filepath.Dir(filepath.Dir(wd)))
	p := filepath.Join(base, "testdata", "god_module", rel)
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("fixture not found: %s (from wd=%s): %v", p, wd, err)
	}
	return p
}

func parseFixture(t *testing.T, parser *ast.Parser, lang string, fixturePath string) (*ast.ParseResult, []byte) {
	t.Helper()
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", fixturePath, err)
	}
	result, err := parser.ParseFile(lang, content)
	if err != nil {
		t.Fatalf("parsing fixture %s: %v", fixturePath, err)
	}
	if result == nil {
		t.Fatalf("parse result is nil for %s (language %s may not be supported)", fixturePath, lang)
	}
	return result, content
}

func TestGodModuleCheck_IDAndName(t *testing.T) {
	c := NewGodModuleCheck()
	if c.ID() != "VH-G012" {
		t.Errorf("ID() = %q, want %q", c.ID(), "VH-G012")
	}
	if c.Name() != "God Module" {
		t.Errorf("Name() = %q, want %q", c.Name(), "God Module")
	}
}

func TestGodModuleCheck_CheckFileReturnsNil(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	content := []byte("package main\nfunc main() {}")
	violations := c.CheckFile("test.go", content, "go", &cfg)
	if len(violations) != 0 {
		t.Errorf("CheckFile should return nil, got %d violations", len(violations))
	}
}

func TestGodModuleCheck_PythonClean(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "clean.py")
	result, content := parseFixture(t, parser, "python", path)
	defer result.Close()

	violations := c.CheckFileAST("clean.py", content, "python", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("clean python: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestGodModuleCheck_PythonViolating(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "violating.py")
	result, content := parseFixture(t, parser, "python", path)
	defer result.Close()

	violations := c.CheckFileAST("violating.py", content, "python", &cfg, result)
	if len(violations) == 0 {
		t.Fatal("violating python: got 0 violations, want > 0")
	}
	if violations[0].RuleID != "VH-G012" {
		t.Errorf("RuleID = %q, want VH-G012", violations[0].RuleID)
	}
	if violations[0].Severity != "warning" {
		t.Errorf("Severity = %q, want warning", violations[0].Severity)
	}
}

func TestGodModuleCheck_GoClean(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "clean.go")
	result, content := parseFixture(t, parser, "go", path)
	defer result.Close()

	violations := c.CheckFileAST("clean.go", content, "go", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("clean go: got %d violations, want 0", len(violations))
	}
}

func TestGodModuleCheck_GoViolating(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "violating.go")
	result, content := parseFixture(t, parser, "go", path)
	defer result.Close()

	violations := c.CheckFileAST("violating.go", content, "go", &cfg, result)
	if len(violations) == 0 {
		t.Fatal("violating go: got 0 violations, want > 0")
	}
	if violations[0].RuleID != "VH-G012" {
		t.Errorf("RuleID = %q, want VH-G012", violations[0].RuleID)
	}
}

func TestGodModuleCheck_TypeScriptClean(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "clean.ts")
	result, content := parseFixture(t, parser, "typescript", path)
	defer result.Close()

	violations := c.CheckFileAST("clean.ts", content, "typescript", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("clean typescript: got %d violations, want 0", len(violations))
	}
}

func TestGodModuleCheck_TypeScriptViolating(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "violating.ts")
	result, content := parseFixture(t, parser, "typescript", path)
	defer result.Close()

	violations := c.CheckFileAST("violating.ts", content, "typescript", &cfg, result)
	if len(violations) == 0 {
		t.Fatal("violating typescript: got 0 violations, want > 0")
	}
	if violations[0].RuleID != "VH-G012" {
		t.Errorf("RuleID = %q, want VH-G012", violations[0].RuleID)
	}
}

func TestGodModuleCheck_JavaClean(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "clean.java")
	result, content := parseFixture(t, parser, "java", path)
	defer result.Close()

	violations := c.CheckFileAST("clean.java", content, "java", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("clean java: got %d violations, want 0", len(violations))
	}
}

func TestGodModuleCheck_JavaViolating(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "violating.java")
	result, content := parseFixture(t, parser, "java", path)
	defer result.Close()

	violations := c.CheckFileAST("violating.java", content, "java", &cfg, result)
	if len(violations) == 0 {
		t.Fatal("violating java: got 0 violations, want > 0")
	}
	if violations[0].RuleID != "VH-G012" {
		t.Errorf("RuleID = %q, want VH-G012", violations[0].RuleID)
	}
}

func TestGodModuleCheck_RubyClean(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "clean.rb")
	result, content := parseFixture(t, parser, "ruby", path)
	defer result.Close()

	violations := c.CheckFileAST("clean.rb", content, "ruby", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("clean ruby: got %d violations, want 0", len(violations))
	}
}

func TestGodModuleCheck_RubyViolating(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "violating.rb")
	result, content := parseFixture(t, parser, "ruby", path)
	defer result.Close()

	violations := c.CheckFileAST("violating.rb", content, "ruby", &cfg, result)
	if len(violations) == 0 {
		t.Fatal("violating ruby: got 0 violations, want > 0")
	}
	if violations[0].RuleID != "VH-G012" {
		t.Errorf("RuleID = %q, want VH-G012", violations[0].RuleID)
	}
}

func TestGodModuleCheck_RustClean(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "clean.rs")
	result, content := parseFixture(t, parser, "rust", path)
	defer result.Close()

	violations := c.CheckFileAST("clean.rs", content, "rust", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("clean rust: got %d violations, want 0", len(violations))
	}
}

func TestGodModuleCheck_RustViolating(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	path := godModuleTestdataPath(t, "violating.rs")
	result, content := parseFixture(t, parser, "rust", path)
	defer result.Close()

	violations := c.CheckFileAST("violating.rs", content, "rust", &cfg, result)
	if len(violations) == 0 {
		t.Fatal("violating rust: got 0 violations, want > 0")
	}
	if violations[0].RuleID != "VH-G012" {
		t.Errorf("RuleID = %q, want VH-G012", violations[0].RuleID)
	}
}

func TestGodModuleCheck_ZeroExportsNotFlagged(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	content := []byte("def _private(): pass\n")
	result, err := parser.ParseFile("python", content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if result == nil {
		t.Skip("python not supported")
	}
	defer result.Close()

	violations := c.CheckFileAST("empty.py", content, "python", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("file with zero public exports should not be flagged, got %d violations", len(violations))
	}
}

func TestGodModuleCheck_PythonPrivateNotCounted(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	var lines []string
	for i := 0; i < 25; i++ {
		lines = append(lines, fmt.Sprintf("def _private%d(): pass", i))
	}
	lines = append(lines, "def public_one(): pass")
	content := []byte(fmt.Sprintf("%s\n", strings.Join(lines, "\n")))

	result, err := parser.ParseFile("python", content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if result == nil {
		t.Skip("python not supported")
	}
	defer result.Close()

	violations := c.CheckFileAST("mostly_private.py", content, "python", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("file with only 1 public export should not be flagged, got %d violations", len(violations))
	}
}

func TestGodModuleCheck_GoUnexportedNotCounted(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	content := []byte("package main\n")
	for i := 0; i < 25; i++ {
		content = append(content, []byte(fmt.Sprintf("func private%d() {}\n", i))...)
	}
	content = append(content, []byte("func PublicOne() {}\n")...)

	result, err := parser.ParseFile("go", content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if result == nil {
		t.Skip("go not supported")
	}
	defer result.Close()

	violations := c.CheckFileAST("mostly_unexported.go", content, "go", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("file with only 1 exported func should not be flagged, got %d violations", len(violations))
	}
}

func TestGodModuleCheck_RustPubCrateNotCounted(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	content := []byte("pub(crate) fn internal1() {}\npub(crate) fn internal2() {}\npub fn public_one() {}\n")

	result, err := parser.ParseFile("rust", content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if result == nil {
		t.Skip("rust not supported")
	}
	defer result.Close()

	violations := c.CheckFileAST("restricted.rs", content, "rust", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("file with pub(crate) functions should not be flagged as god module, got %d violations", len(violations))
	}
}

func TestGodModuleCheck_UnsupportedLanguage(t *testing.T) {
	c := NewGodModuleCheck()
	cfg := config.DefaultConfig()
	parser := ast.NewParser()
	defer parser.Close()

	content := []byte("SELECT * FROM users;")
	result, err := parser.ParseFile("sql", content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if result == nil {
		violations := c.CheckFileAST("query.sql", content, "sql", &cfg, nil)
		if len(violations) != 0 {
			t.Errorf("unsupported language should produce 0 violations, got %d", len(violations))
		}
		return
	}
	defer result.Close()

	violations := c.CheckFileAST("query.sql", content, "sql", &cfg, result)
	if len(violations) != 0 {
		t.Errorf("unsupported language should produce 0 violations, got %d", len(violations))
	}
}