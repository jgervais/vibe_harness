package output

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/jgervais/vibe_harness/internal/rules"
	"github.com/jgervais/vibe_harness/internal/scanner"
)

func TestFormatHumanZeroViolations(t *testing.T) {
	result := &scanner.ScanResult{Violations: []rules.Violation{}}
	var buf bytes.Buffer
	FormatHuman(&buf, result)

	got := buf.String()
	if !strings.Contains(got, "0 violation(s) found in 0 file(s)") {
		t.Errorf("expected zero summary, got:\n%s", got)
	}
	violationPattern := fmt.Sprintf("%s:%d:%s", "x", 0, "VH-G000")
	if strings.Contains(got, violationPattern) || strings.Contains(got, " — ") {
		t.Errorf("expected no violation lines for zero violations, got:\n%s", got)
	}
}

func TestFormatHumanMultipleViolations(t *testing.T) {
	result := &scanner.ScanResult{
		Violations: []rules.Violation{
			{File: "src/main.go", Line: 42, RuleID: "VH-G001", Message: "file exceeds 300 non-blank, non-comment lines (412)"},
			{File: "src/config.py", Line: 7, RuleID: "VH-G005", Message: `hardcoded secret: AWS access key pattern "AKIA..."`},
			{File: "src/main.go", Line: 15, RuleID: "VH-G006", Message: "magic number 42"},
		},
	}
	var buf bytes.Buffer
	FormatHuman(&buf, result)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines (3 violations + blank + summary), got %d: %v", len(lines), lines)
	}

	for i, v := range result.Violations {
		expected := fmt.Sprintf("%s:%d:%s — %s", v.File, v.Line, v.RuleID, v.Message)
		if lines[i] != expected {
			t.Errorf("line %d: expected %q, got %q", i, expected, lines[i])
		}
	}

	if lines[3] != "" {
		t.Errorf("expected blank line at index 3, got %q", lines[3])
	}

	if lines[4] != "3 violation(s) found in 2 file(s)" {
		t.Errorf("expected summary '3 violation(s) found in 2 file(s)', got %q", lines[4])
	}
}

func TestFormatHumanWriterOutput(t *testing.T) {
	result := &scanner.ScanResult{
		Violations: []rules.Violation{
			{File: "a.go", Line: 1, RuleID: "VH-G001", Message: "msg"},
		},
	}
	var buf bytes.Buffer
	FormatHuman(&buf, result)

	if buf.Len() == 0 {
		t.Error("expected output written to writer, got empty buffer")
	}
}

func TestFormatHumanFormatString(t *testing.T) {
	v := rules.Violation{File: "src/main.go", Line: 42, RuleID: "VH-G001", Message: "file exceeds 300 non-blank, non-comment lines (412)"}
	result := &scanner.ScanResult{Violations: []rules.Violation{v}}
	var buf bytes.Buffer
	FormatHuman(&buf, result)

	line := strings.Split(buf.String(), "\n")[0]
	expected := "src/main.go:42:VH-G001 — file exceeds 300 non-blank, non-comment lines (412)"
	if line != expected {
		t.Errorf("expected %q, got %q", expected, line)
	}
}