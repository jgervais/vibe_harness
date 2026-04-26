package generic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/config"
)

func missingLoggingFixturePath(t *testing.T, name string) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	base := filepath.Dir(filepath.Dir(filepath.Dir(wd)))
	p := filepath.Join(base, "testdata", "missing_logging", name)
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("fixture not found: %s: %v", p, err)
	}
	return p
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(missingLoggingFixturePath(t, name))
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	return data
}

func TestMissingLogging_Python_Clean(t *testing.T) {
	content := readFixture(t, "clean.py")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("python", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil (unsupported language?)")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("clean.py", content, "python", &cfg, parseResult)
	if len(violations) > 0 {
		t.Errorf("expected 0 violations for clean.py, got %d: %v", len(violations), violations)
	}
}

func TestMissingLogging_Python_Violating(t *testing.T) {
	content := readFixture(t, "violating.py")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("python", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("violating.py", content, "python", &cfg, parseResult)
	if len(violations) == 0 {
		t.Error("expected at least 1 violation for violating.py, got 0")
	}
}

func TestMissingLogging_Go_Clean(t *testing.T) {
	content := readFixture(t, "clean.go")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("go", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("clean.go", content, "go", &cfg, parseResult)
	if len(violations) > 0 {
		t.Errorf("expected 0 violations for clean.go, got %d: %v", len(violations), violations)
	}
}

func TestMissingLogging_Go_Violating(t *testing.T) {
	content := readFixture(t, "violating.go")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("go", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("violating.go", content, "go", &cfg, parseResult)
	if len(violations) == 0 {
		t.Error("expected at least 1 violation for violating.go, got 0")
	}
}

func TestMissingLogging_TypeScript_Clean(t *testing.T) {
	content := readFixture(t, "clean.ts")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("typescript", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("clean.ts", content, "typescript", &cfg, parseResult)
	if len(violations) > 0 {
		t.Errorf("expected 0 violations for clean.ts, got %d: %v", len(violations), violations)
	}
}

func TestMissingLogging_TypeScript_Violating(t *testing.T) {
	content := readFixture(t, "violating.ts")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("typescript", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("violating.ts", content, "typescript", &cfg, parseResult)
	if len(violations) == 0 {
		t.Error("expected at least 1 violation for violating.ts, got 0")
	}
}

func TestMissingLogging_Java_Clean(t *testing.T) {
	content := readFixture(t, "clean.java")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("java", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("clean.java", content, "java", &cfg, parseResult)
	if len(violations) > 0 {
		t.Errorf("expected 0 violations for clean.java, got %d: %v", len(violations), violations)
	}
}

func TestMissingLogging_Java_Violating(t *testing.T) {
	content := readFixture(t, "violating.java")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("java", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("violating.java", content, "java", &cfg, parseResult)
	if len(violations) == 0 {
		t.Error("expected at least 1 violation for violating.java, got 0")
	}
}

func TestMissingLogging_Ruby_Clean(t *testing.T) {
	content := readFixture(t, "clean.rb")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("ruby", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("clean.rb", content, "ruby", &cfg, parseResult)
	if len(violations) > 0 {
		t.Errorf("expected 0 violations for clean.rb, got %d: %v", len(violations), violations)
	}
}

func TestMissingLogging_Ruby_Violating(t *testing.T) {
	content := readFixture(t, "violating.rb")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("ruby", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("violating.rb", content, "ruby", &cfg, parseResult)
	if len(violations) == 0 {
		t.Error("expected at least 1 violation for violating.rb, got 0")
	}
}

func TestMissingLogging_Rust_Clean(t *testing.T) {
	content := readFixture(t, "clean.rs")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("rust", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("clean.rs", content, "rust", &cfg, parseResult)
	if len(violations) > 0 {
		t.Errorf("expected 0 violations for clean.rs, got %d: %v", len(violations), violations)
	}
}

func TestMissingLogging_Rust_Violating(t *testing.T) {
	content := readFixture(t, "violating.rs")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	parseResult, err := parser.ParseFile("rust", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("violating.rs", content, "rust", &cfg, parseResult)
	if len(violations) == 0 {
		t.Error("expected at least 1 violation for violating.rs, got 0")
	}
}

func TestMissingLogging_UnsupportedLanguage(t *testing.T) {
	check := NewMissingLoggingCheck()
	defer check.Close()

	cfg := config.DefaultConfig()
	violations := check.CheckFileAST("test.txt", []byte("hello"), "brainfuck", &cfg, nil)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for unsupported language, got %d", len(violations))
	}
}

func TestMissingLogging_ConfigProvidedLoggingCalls(t *testing.T) {
	content := readFixture(t, "violating.py")
	check := NewMissingLoggingCheck()
	defer check.Close()

	parser := ast.NewParser()
	defer parser.Close()

	cfg := config.DefaultConfig()
	cfg.Observability.LoggingCalls = []string{"custom_logger"}
	parseResult, err := parser.ParseFile("python", content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parseResult == nil {
		t.Fatal("parseResult is nil")
	}
	defer parseResult.Close()

	violations := check.CheckFileAST("violating.py", content, "python", &cfg, parseResult)
	if len(violations) == 0 {
		t.Error("expected at least 1 violation even with custom logging calls, since violating.py has no custom_logger")
	}
}