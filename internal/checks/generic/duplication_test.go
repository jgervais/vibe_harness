package generic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jgervais/vibe_harness/internal/config"
)

func TestDuplicationCheck_ID(t *testing.T) {
	c := NewDuplicationCheck()
	if c.ID() != "VH-G007" {
		t.Errorf("ID() = %q, want %q", c.ID(), "VH-G007")
	}
}

func TestDuplicationCheck_Name(t *testing.T) {
	c := NewDuplicationCheck()
	if c.Name() != "Copy-Paste Duplication" {
		t.Errorf("Name() = %q, want %q", c.Name(), "Copy-Paste Duplication")
	}
}

func TestDuplicationCheck_CheckFileReturnsEmpty(t *testing.T) {
	c := NewDuplicationCheck()
	cfg := config.DefaultConfig()
	violations := c.CheckFile("test.go", []byte("package main"), "go", &cfg)
	if len(violations) != 0 {
		t.Errorf("CheckFile should return empty, got %d violations", len(violations))
	}
}

func TestDuplicationCheck_DuplicateBlocksDetected(t *testing.T) {
	c := NewDuplicationCheck()
	cfg := config.DefaultConfig()

	fileA, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "duplication", "file_a.go"))
	if err != nil {
		t.Fatalf("reading file_a.go: %v", err)
	}
	fileB, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "duplication", "file_b.go"))
	if err != nil {
		t.Fatalf("reading file_b.go: %v", err)
	}

	files := []FileContent{
		{Path: "testdata/duplication/file_a.go", Content: fileA},
		{Path: "testdata/duplication/file_b.go", Content: fileB},
	}

	violations := c.CheckFiles(files, &cfg)
	if len(violations) == 0 {
		t.Fatal("expected violations for duplicated blocks, got none")
	}

	for _, v := range violations {
		if v.RuleID != "VH-G007" {
			t.Errorf("RuleID = %q, want %q", v.RuleID, "VH-G007")
		}
		if v.Severity != "warning" {
			t.Errorf("Severity = %q, want %q", v.Severity, "warning")
		}
	}
}

func TestDuplicationCheck_ViolationMentionsBothFiles(t *testing.T) {
	c := NewDuplicationCheck()
	cfg := config.DefaultConfig()

	fileA, _ := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "duplication", "file_a.go"))
	fileB, _ := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "duplication", "file_b.go"))

	files := []FileContent{
		{Path: "testdata/duplication/file_a.go", Content: fileA},
		{Path: "testdata/duplication/file_b.go", Content: fileB},
	}

	violations := c.CheckFiles(files, &cfg)
	if len(violations) == 0 {
		t.Fatal("expected violations, got none")
	}

	foundMention := false
	for _, v := range violations {
		if contains(v.Message, "file_b.go") && contains(v.Message, "file_a.go") {
			foundMention = true
			break
		}
		if v.File == "testdata/duplication/file_a.go" && contains(v.Message, "file_b.go") {
			foundMention = true
			break
		}
		if v.File == "testdata/duplication/file_b.go" && contains(v.Message, "file_a.go") {
			foundMention = true
			break
		}
	}
	if !foundMention {
		t.Errorf("expected violation message to mention both files, got: %v", violations)
	}
}

func TestDuplicationCheck_NoDuplicates(t *testing.T) {
	c := NewDuplicationCheck()
	cfg := config.DefaultConfig()

	dir := t.TempDir()
	file1 := filepath.Join(dir, "unique1.go")
	file2 := filepath.Join(dir, "unique2.go")

	content1 := []byte(`package alpha
func Alpha(x int) int {
	return x + 1
}
func Beta(y int) string {
	return "hello"
}
`)

	content2 := []byte(`package beta
func Gamma(z float64) float64 {
	return z * 2.5
}
func Delta(s string) int {
	return len(s)
}
`)

	os.WriteFile(file1, content1, 0644)
	os.WriteFile(file2, content2, 0644)

	files := []FileContent{
		{Path: file1, Content: content1},
		{Path: file2, Content: content2},
	}

	violations := c.CheckFiles(files, &cfg)
	if len(violations) != 0 {
		t.Errorf("expected no violations for different files, got %d", len(violations))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}