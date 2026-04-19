package generic

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

func commentRatioFixturePath(t *testing.T, filename string) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	for base := wd; base != "/"; base = filepath.Dir(base) {
		candidate := filepath.Join(base, "testdata", "comment_ratio", filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	t.Fatalf("fixture not found: %s (from wd=%s)", filename, wd)
	return ""
}

func TestCommentRatioCheck_IDAndName(t *testing.T) {
	c := NewCommentRatioCheck()
	if c.ID() != "VH-G008" {
		t.Errorf("ID() = %q, want %q", c.ID(), "VH-G008")
	}
	if c.Name() != "Comment-to-Code Ratio" {
		t.Errorf("Name() = %q, want %q", c.Name(), "Comment-to-Code Ratio")
	}
}

func TestCommentRatioCheck_CleanGoFile(t *testing.T) {
	c := NewCommentRatioCheck()
	cfg := config.DefaultConfig()
	content, err := os.ReadFile(commentRatioFixturePath(t, "clean.go"))
	if err != nil {
		t.Fatalf("reading clean fixture: %v", err)
	}
	violations := c.CheckFile("clean.go", content, "go", &cfg)
	if len(violations) != 0 {
		t.Errorf("clean file: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s: %s", v.RuleID, v.Message)
		}
	}
}

func TestCommentRatioCheck_ViolatingRubyFile(t *testing.T) {
	c := NewCommentRatioCheck()
	cfg := config.DefaultConfig()
	content, err := os.ReadFile(commentRatioFixturePath(t, "violating.rb"))
	if err != nil {
		t.Fatalf("reading violating fixture: %v", err)
	}
	violations := c.CheckFile("violating.rb", content, "ruby", &cfg)
	if len(violations) != 1 {
		t.Fatalf("violating file: got %d violations, want 1", len(violations))
	}
	v := violations[0]
	if v.RuleID != "VH-G008" {
		t.Errorf("RuleID = %q, want VH-G008", v.RuleID)
	}
	if v.File != "violating.rb" {
		t.Errorf("File = %q, want %q", v.File, "violating.rb")
	}
	if v.Line != 1 {
		t.Errorf("Line = %d, want 1", v.Line)
	}
	if v.Severity != "note" {
		t.Errorf("Severity = %q, want note", v.Severity)
	}
	if !strings.Contains(v.Message, "exceeds 1:3") {
		t.Errorf("Message = %q, want to contain 'exceeds 1:3'", v.Message)
	}
}

func TestCommentRatioCheck_BlankLinesNotCounted(t *testing.T) {
	c := NewCommentRatioCheck()
	cfg := config.DefaultConfig()
	input := []byte("x = 1\n\n\n\ny = 2\n\n\n\nz = 3\n")
	violations := c.CheckFile("test.py", input, "python", &cfg)
	if len(violations) != 0 {
		t.Errorf("code-only with blank lines: got %d violations, want 0", len(violations))
	}
}

func TestCommentRatioCheck_BlockCommentTracking(t *testing.T) {
	c := NewCommentRatioCheck()
	cfg := config.DefaultConfig()
	input := []byte("/* line 1\n   line 2\n   line 3 */\nx = 1\ny = 2\nz = 3\nw = 4\n")
	violations := c.CheckFile("test.go", input, "go", &cfg)
	if len(violations) != 1 {
		t.Fatalf("block comment test: got %d violations, want 1", len(violations))
	}
	v := violations[0]
	if v.RuleID != "VH-G008" {
		t.Errorf("RuleID = %q, want VH-G008", v.RuleID)
	}
}

func TestCommentRatioCheck_ViolationFields(t *testing.T) {
	c := NewCommentRatioCheck()
	cfg := config.DefaultConfig()
	input := []byte("# comment 1\n# comment 2\nx = 1\n")
	violations := c.CheckFile("test.rb", input, "ruby", &cfg)
	if len(violations) != 1 {
		t.Fatalf("got %d violations, want 1", len(violations))
	}
	v := violations[0]
	expected := rules.Violation{
		RuleID:   "VH-G008",
		File:     "test.rb",
		Line:     1,
		Column:   0,
		EndLine:  0,
		Severity: "note",
	}
	if v.RuleID != expected.RuleID || v.File != expected.File || v.Line != expected.Line || v.Column != expected.Column || v.EndLine != expected.EndLine || v.Severity != expected.Severity {
		t.Errorf("violation fields = %+v, want %+v", v, expected)
	}
}

func TestCommentRatioCheck_UnknownLanguage(t *testing.T) {
	c := NewCommentRatioCheck()
	cfg := config.DefaultConfig()
	input := []byte("some content\n")
	violations := c.CheckFile("test.xyz", input, "unknown", &cfg)
	if len(violations) != 0 {
		t.Errorf("unknown language: got %d violations, want 0", len(violations))
	}
}