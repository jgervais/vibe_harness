package generic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jgervais/vibe_harness/internal/config"
)

func testdataPath(t *testing.T, rel string) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	base := filepath.Dir(filepath.Dir(filepath.Dir(wd)))
	p := filepath.Join(base, "testdata", rel)
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("fixture not found: %s (from wd=%s): %v", p, wd, err)
	}
	return p
}

func TestSecurityFeaturesCheck_IDAndName(t *testing.T) {
	c := NewSecurityFeaturesCheck()
	if c.ID() != "VH-G011" {
		t.Errorf("ID() = %q, want %q", c.ID(), "VH-G011")
	}
	if c.Name() != "Disabled Security Features" {
		t.Errorf("Name() = %q, want %q", c.Name(), "Disabled Security Features")
	}
}

func TestSecurityFeaturesCheck_CleanFile(t *testing.T) {
	c := NewSecurityFeaturesCheck()
	cfg := config.DefaultConfig()
	content, err := os.ReadFile(testdataPath(t, "security_features/clean.py"))
	if err != nil {
		t.Fatalf("reading clean fixture: %v", err)
	}
	violations := c.CheckFile("clean.py", content, "python", &cfg)
	if len(violations) != 0 {
		t.Errorf("clean file: got %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  unexpected: %s line %d: %s", v.RuleID, v.Line, v.Message)
		}
	}
}

func TestSecurityFeaturesCheck_ViolatingFile(t *testing.T) {
	c := NewSecurityFeaturesCheck()
	cfg := config.DefaultConfig()
	content, err := os.ReadFile(testdataPath(t, "security_features/violating.go"))
	if err != nil {
		t.Fatalf("reading violating fixture: %v", err)
	}
	violations := c.CheckFile("violating.go", content, "go", &cfg)
	if len(violations) == 0 {
		t.Fatal("violating file: got 0 violations, want > 0")
	}
	count := 0
	for _, v := range violations {
		if v.RuleID != "VH-G011" {
			t.Errorf("RuleID = %q, want VH-G011", v.RuleID)
		}
		if v.Severity != "error" {
			t.Errorf("Severity = %q, want error", v.Severity)
		}
		if v.Message == "disabled security verification: InsecureSkipVerify=true" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 InsecureSkipVerify violation, got %d", count)
	}
}

func TestSecurityFeaturesCheck_VerifyTrueNotFlagged(t *testing.T) {
	c := NewSecurityFeaturesCheck()
	cfg := config.DefaultConfig()
	input := []byte(`resp = requests.get("https://example.com", verify=True)`)
	violations := c.CheckFile("test.py", input, "python", &cfg)
	if len(violations) != 0 {
		t.Errorf("verify=True should not be flagged, got %d violations", len(violations))
	}
}

func TestSecurityFeaturesCheck_InsecureSkipVerifyFalseNotFlagged(t *testing.T) {
	c := NewSecurityFeaturesCheck()
	cfg := config.DefaultConfig()
	input := []byte(`tls.Config{InsecureSkipVerify: false}`)
	violations := c.CheckFile("test.go", input, "go", &cfg)
	if len(violations) != 0 {
		t.Errorf("InsecureSkipVerify: false should not be flagged, got %d violations", len(violations))
	}
}

func TestSecurityFeaturesCheck_RejectUnauthorizedFalse(t *testing.T) {
	c := NewSecurityFeaturesCheck()
	cfg := config.DefaultConfig()
	input := []byte(`const options = { rejectUnauthorized: false };`)
	violations := c.CheckFile("test.js", input, "javascript", &cfg)
	if len(violations) == 0 {
		t.Fatal("rejectUnauthorized: false should be flagged")
	}
	found := false
	for _, v := range violations {
		if v.Message == "disabled security verification: rejectUnauthorized=false" {
			found = true
		}
	}
	if !found {
		t.Error("expected rejectUnauthorized=false violation message")
	}
}