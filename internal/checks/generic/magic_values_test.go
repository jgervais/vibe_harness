package generic

import (
	"os"
	"strings"
	"testing"

	"github.com/jgervais/vibe_harness/internal/config"
)

func TestMagicValuesCheck_IDAndName(t *testing.T) {
	c := NewMagicValuesCheck()
	if c.ID() != "VH-G006" {
		t.Errorf("ID() = %q, want %q", c.ID(), "VH-G006")
	}
	if c.Name() != "Magic Values" {
		t.Errorf("Name() = %q, want %q", c.Name(), "Magic Values")
	}
}

func TestMagicValuesCheck_CheckFileReturnsEmpty(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	violations := c.CheckFile("test.ts", []byte("let x = 1;"), "typescript", &cfg)
	if len(violations) != 0 {
		t.Errorf("CheckFile should return empty, got %d violations", len(violations))
	}
}

func TestMagicValuesCheck_CleanFile(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	content, err := os.ReadFile("../../../testdata/magic_values/clean.ts")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	violations := c.CheckFiles([]FileContent{{Path: "clean.ts", Content: content}}, &cfg)
	if len(violations) != 0 {
		for _, v := range violations {
			t.Errorf("unexpected violation: %s", v.Message)
		}
	}
}

func TestMagicValuesCheck_ViolatingMagicNumber(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	code := []byte(`let a = 42;
let b = 42;
let c = 42;
`)
	violations := c.CheckFiles([]FileContent{{Path: "test.ts", Content: code}}, &cfg)
	if len(violations) == 0 {
		t.Fatal("expected magic number violation for 42 appearing 3 times")
	}
	found := false
	for _, v := range violations {
		if v.RuleID == "VH-G006" && strings.Contains(v.Message, "magic value") && strings.Contains(v.Message, "42") {
			if v.Severity != "error" {
				t.Errorf("Severity = %q, want error", v.Severity)
			}
			found = true
		}
	}
	if !found {
		t.Error("expected magic value violation for 42")
	}
}

func TestMagicValuesCheck_ViolatingMagicString(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	label := `"this is a very long magic string value"`
	code := []byte("let a = " + label + "\nlet b = " + label + "\nlet c = " + label + "\n")
	violations := c.CheckFiles([]FileContent{{Path: "test.ts", Content: code}}, &cfg)
	if len(violations) == 0 {
		t.Fatal("expected magic string violation for repeated 20+ char string")
	}
	found := false
	for _, v := range violations {
		if v.RuleID == "VH-G006" && strings.Contains(v.Message, "magic string") {
			if v.Severity != "error" {
				t.Errorf("Severity = %q, want error", v.Severity)
			}
			found = true
		}
	}
	if !found {
		t.Error("expected magic string violation")
	}
}

func TestMagicValuesCheck_SingleDigitNotFlagged(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	code := []byte(`let a = 5;
let b = 5;
let c = 5;
`)
	violations := c.CheckFiles([]FileContent{{Path: "test.ts", Content: code}}, &cfg)
	for _, v := range violations {
		if strings.Contains(v.Message, "5") {
			t.Errorf("single-digit number should not be flagged: %s", v.Message)
		}
	}
}

func TestMagicValuesCheck_TwoOccurrencesNotFlagged(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	code := []byte(`let a = 42;
let b = 42;
`)
	violations := c.CheckFiles([]FileContent{{Path: "test.ts", Content: code}}, &cfg)
	for _, v := range violations {
		t.Errorf("2 occurrences should not be flagged: %s", v.Message)
	}
}

func TestMagicValuesCheck_UniqueStringNotFlagged(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	code := []byte(`let a = "this is a very long unique string one";
let b = "this is a very long unique string two";
`)
	violations := c.CheckFiles([]FileContent{{Path: "test.ts", Content: code}}, &cfg)
	for _, v := range violations {
		if strings.Contains(v.Message, "magic string") {
			t.Errorf("unique string should not be flagged: %s", v.Message)
		}
	}
}

func TestMagicValuesCheck_ConstNotFlagged(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	code := []byte(`const MAGIC = 42;
let x = 42;
let y = 42;
`)
	violations := c.CheckFiles([]FileContent{{Path: "test.ts", Content: code}}, &cfg)
	for _, v := range violations {
		if v.Line == 1 {
			t.Errorf("constant line should not be flagged: %s", v.Message)
		}
	}
}

func TestMagicValuesCheck_TestFilesSkipped(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	code := []byte(`let a = 42;
let b = 42;
let c = 42;
`)
	violations := c.CheckFiles([]FileContent{{Path: "foo_test.ts", Content: code}}, &cfg)
	if len(violations) != 0 {
		t.Errorf("test files should be skipped, got %d violations", len(violations))
	}
}