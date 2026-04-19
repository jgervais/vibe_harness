package generic

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jgervais/vibe_harness/internal/config"
)

func buildMultiViolationContent() []byte {
	var lines []string
	lines = append(lines, "package main")
	lines = append(lines, "")
	lines = append(lines, `key := "AKIAIOSFODNN7EXAMPLE"`)
	lines = append(lines, "tls.Config{InsecureSkipVerify: true}")
	for i := 0; i < 310; i++ {
		lines = append(lines, fmt.Sprintf("x%d := 42", i))
	}
	return []byte(strings.Join(lines, "\n"))
}

func TestMultiViolation_EachCheckIndividually(t *testing.T) {
	cfg := config.DefaultConfig()
	content := buildMultiViolationContent()
	path := "multi.go"
	language := "go"

	t.Run("VH-G001_FileLength", func(t *testing.T) {
		c := NewFileLengthCheck()
		violations := c.CheckFile(path, content, language, &cfg)
		if len(violations) == 0 {
			t.Error("FileLengthCheck should flag file exceeding 300 code lines")
		}
		for _, v := range violations {
			if v.RuleID != "VH-G001" {
				t.Errorf("expected VH-G001, got %s", v.RuleID)
			}
		}
	})

	t.Run("VH-G005_HardcodedSecrets", func(t *testing.T) {
		c := NewHardcodedSecretsCheck()
		violations := c.CheckFile(path, content, language, &cfg)
		if len(violations) == 0 {
			t.Error("HardcodedSecretsCheck should flag AWS access key")
		}
		found := false
		for _, v := range violations {
			if v.RuleID == "VH-G005" && strings.Contains(v.Message, "AWS access key") {
				found = true
			}
		}
		if !found {
			t.Error("expected AWS access key violation from VH-G005")
		}
	})

	t.Run("VH-G006_MagicValues", func(t *testing.T) {
		c := NewMagicValuesCheck()
		violations := c.CheckFile(path, content, language, &cfg)
		if len(violations) == 0 {
			t.Error("MagicValuesCheck should flag repeated magic number 42")
		}
		found := false
		for _, v := range violations {
			if v.RuleID == "VH-G006" && strings.Contains(v.Message, "42") {
				found = true
			}
		}
		if !found {
			t.Error("expected magic value 42 violation from VH-G006")
		}
	})

	t.Run("VH-G011_SecurityFeatures", func(t *testing.T) {
		c := NewSecurityFeaturesCheck()
		violations := c.CheckFile(path, content, language, &cfg)
		if len(violations) == 0 {
			t.Error("SecurityFeaturesCheck should flag InsecureSkipVerify: true")
		}
		found := false
		for _, v := range violations {
			if v.RuleID == "VH-G011" && strings.Contains(v.Message, "InsecureSkipVerify") {
				found = true
			}
		}
		if !found {
			t.Error("expected InsecureSkipVerify violation from VH-G011")
		}
	})
}