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

func TestMagicValuesCheck_CleanFile(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	content, err := os.ReadFile("../../../testdata/magic_values/clean.ts")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	violations := c.CheckFile("clean.ts", content, "typescript", &cfg)
	if len(violations) != 0 {
		for _, v := range violations {
			t.Errorf("unexpected violation: %s", v.Message)
		}
	}
}

func TestMagicValuesCheck_ViolatingFile(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	content, err := os.ReadFile("../../../testdata/magic_values/violating.ts")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	violations := c.CheckFile("violating.ts", content, "typescript", &cfg)

	hasMagicNumber := false
	hasMagicString := false
	for _, v := range violations {
		if v.RuleID != "VH-G006" {
			t.Errorf("RuleID = %q, want VH-G006", v.RuleID)
		}
		if v.Severity != "warning" {
			t.Errorf("Severity = %q, want warning", v.Severity)
		}
		if v.File != "violating.ts" {
			t.Errorf("File = %q, want violating.ts", v.File)
		}
		if strings.Contains(v.Message, "magic value:") {
			hasMagicNumber = true
		}
		if strings.Contains(v.Message, "magic string:") {
			hasMagicString = true
		}
	}
	if !hasMagicNumber {
		t.Error("expected magic numeric value violation")
	}
	if !hasMagicString {
		t.Error("expected magic string violation")
	}
}

func TestMagicValuesCheck_AllowedValues(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	code := []byte(`let a = 0;
let b = 1;
let c = 2;
let d = -1;
let e = true;
let f = false;
let g = null;
`)
	violations := c.CheckFile("test.ts", code, "typescript", &cfg)
	for _, v := range violations {
		t.Errorf("allowed value flagged as violation: %s", v.Message)
	}
}

func TestMagicValuesCheck_ConstNotFlagged(t *testing.T) {
	c := NewMagicValuesCheck()
	cfg := config.DefaultConfig()
	code := []byte(`const MAGIC = 42;
let x = 42;
let y = 42;
`)
	violations := c.CheckFile("test.ts", code, "typescript", &cfg)
	if len(violations) == 0 {
		t.Error("expected at least one violation for magic number 42 used inline")
	}
	for _, v := range violations {
		if v.Line == 1 {
			t.Errorf("constant line should not be flagged: %s", v.Message)
		}
	}
}