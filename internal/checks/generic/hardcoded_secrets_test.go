package generic

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jgervais/vibe_harness/internal/config"
)

func fixturePath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	return filepath.Join(root, "testdata", "hardcoded_secrets", name)
}

func TestHardcodedSecretsCheck_IDAndName(t *testing.T) {
	c := NewHardcodedSecretsCheck()
	if c.ID() != "VH-G005" {
		t.Errorf("ID() = %q, want %q", c.ID(), "VH-G005")
	}
	if c.Name() != "Hardcoded Secrets" {
		t.Errorf("Name() = %q, want %q", c.Name(), "Hardcoded Secrets")
	}
}

func TestHardcodedSecretsCheck_CleanFile(t *testing.T) {
	c := NewHardcodedSecretsCheck()
	cfg := config.DefaultConfig()
	content, err := os.ReadFile(fixturePath("clean.py"))
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

func TestHardcodedSecretsCheck_ViolatingFile(t *testing.T) {
	c := NewHardcodedSecretsCheck()
	cfg := config.DefaultConfig()
	content, err := os.ReadFile(fixturePath("violating.py"))
	if err != nil {
		t.Fatalf("reading violating fixture: %v", err)
	}
	violations := c.CheckFile("violating.py", content, "python", &cfg)
	if len(violations) == 0 {
		t.Fatal("violating file: got 0 violations, want > 0")
	}
	ruleIDs := map[string]int{}
	for _, v := range violations {
		ruleIDs[v.RuleID]++
		if v.RuleID != "VH-G005" {
			t.Errorf("RuleID = %q, want VH-G005", v.RuleID)
		}
		if v.Severity != "error" {
			t.Errorf("Severity = %q, want error", v.Severity)
		}
	}
	if ruleIDs["VH-G005"] < 3 {
		t.Errorf("expected at least 3 VH-G005 violations, got %d", ruleIDs["VH-G005"])
	}
}

func TestHardcodedSecretsCheck_AWSAccessKey(t *testing.T) {
	c := NewHardcodedSecretsCheck()
	cfg := config.DefaultConfig()
	input := []byte(`AWS_KEY = "AKIAIOSFODNN7EXAMPLE"`)
	violations := c.CheckFile("test.py", input, "python", &cfg)
	found := false
	for _, v := range violations {
		if v.Message == "hardcoded secret: AWS access key pattern" {
			found = true
			break
		}
	}
	if !found {
		t.Error("AWS access key pattern not detected")
	}
}

func TestHardcodedSecretsCheck_PrivateKeyMarker(t *testing.T) {
	c := NewHardcodedSecretsCheck()
	cfg := config.DefaultConfig()
	input := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIABAQCAVEA`)
	violations := c.CheckFile("test.pem", input, "", &cfg)
	found := false
	for _, v := range violations {
		if v.Message == "hardcoded secret: RSA private key marker" {
			found = true
			break
		}
	}
	if !found {
		t.Error("RSA private key marker not detected")
	}
}