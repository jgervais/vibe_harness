package generic

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jgervais/vibe_harness/internal/config"
)

func TestFileLengthCheck_ID(t *testing.T) {
	c := NewFileLengthCheck()
	if c.ID() != "VH-G001" {
		t.Errorf("ID() = %q, want %q", c.ID(), "VH-G001")
	}
}

func TestFileLengthCheck_Name(t *testing.T) {
	c := NewFileLengthCheck()
	if c.Name() != "File Length" {
		t.Errorf("Name() = %q, want %q", c.Name(), "File Length")
	}
}

func TestFileLengthCheck_CleanFile(t *testing.T) {
	c := NewFileLengthCheck()
	cfg := config.DefaultConfig()
	content := []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	violations := c.CheckFile("clean.go", content, "go", &cfg)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(violations))
	}
}

func TestFileLengthCheck_ViolatingFile(t *testing.T) {
	c := NewFileLengthCheck()
	cfg := config.DefaultConfig()

	var lines []string
	for i := 0; i < 310; i++ {
		lines = append(lines, fmt.Sprintf("\tx = %d", i))
	}
	content := []byte(strings.Join(lines, "\n"))

	violations := c.CheckFile("violating.go", content, "go", &cfg)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.RuleID != "VH-G001" {
		t.Errorf("RuleID = %q, want %q", v.RuleID, "VH-G001")
	}
	if v.File != "violating.go" {
		t.Errorf("File = %q, want %q", v.File, "violating.go")
	}
	if v.Line != 1 {
		t.Errorf("Line = %d, want 1", v.Line)
	}
	if v.Column != 0 {
		t.Errorf("Column = %d, want 0", v.Column)
	}
	if v.EndLine != 0 {
		t.Errorf("EndLine = %d, want 0", v.EndLine)
	}
	if v.Severity != "warning" {
		t.Errorf("Severity = %q, want %q", v.Severity, "warning")
	}
	expectedMsg := "file exceeds 300 non-blank, non-comment lines (310)"
	if v.Message != expectedMsg {
		t.Errorf("Message = %q, want %q", v.Message, expectedMsg)
	}
}

func TestFileLengthCheck_BlankAndCommentLinesNotCounted(t *testing.T) {
	c := NewFileLengthCheck()
	cfg := config.DefaultConfig()

	var lines []string
	for i := 0; i < 305; i++ {
		lines = append(lines, fmt.Sprintf("\tx = %d", i))
	}
	for i := 0; i < 100; i++ {
		lines = append(lines, "")       // blank lines
		lines = append(lines, "// comment") // comment lines
	}
	content := []byte(strings.Join(lines, "\n"))

	violations := c.CheckFile("mixed.go", content, "go", &cfg)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Message != "file exceeds 300 non-blank, non-comment lines (305)" {
		t.Errorf("Message = %q, want count of 305", violations[0].Message)
	}
}

func TestFileLengthCheck_PythonHashComments(t *testing.T) {
	c := NewFileLengthCheck()
	cfg := config.DefaultConfig()

	var lines []string
	for i := 0; i < 305; i++ {
		lines = append(lines, fmt.Sprintf("x = %d", i))
	}
	for i := 0; i < 50; i++ {
		lines = append(lines, "# this is a comment")
	}
	content := []byte(strings.Join(lines, "\n"))

	violations := c.CheckFile("violating.py", content, "python", &cfg)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Message != "file exceeds 300 non-blank, non-comment lines (305)" {
		t.Errorf("Message = %q, want count of 305", violations[0].Message)
	}
}

func TestFileLengthCheck_BlockComments(t *testing.T) {
	c := NewFileLengthCheck()
	cfg := config.DefaultConfig()

	var lines []string
	for i := 0; i < 305; i++ {
		lines = append(lines, fmt.Sprintf("\tx = %d", i))
	}
	lines = append(lines, "/*")
	lines = append(lines, " * multi-line block comment")
	lines = append(lines, " * more comment text")
	lines = append(lines, " */")
	content := []byte(strings.Join(lines, "\n"))

	violations := c.CheckFile("block.go", content, "go", &cfg)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Message != "file exceeds 300 non-blank, non-comment lines (305)" {
		t.Errorf("Message = %q, want count of 305", violations[0].Message)
	}
}