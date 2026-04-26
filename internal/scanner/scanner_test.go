package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jgervais/vibe_harness/internal/config"
)

func testConfig() config.Config {
	return config.Config{
		SourceDirs: []string{"**"},
		Languages: map[string]string{
			".go": "go",
			".py": "python",
			".ts": "typescript",
		},
		TestFilePattern: []string{"_test.", "testdata"},
	}
}

func TestScan(t *testing.T) {
	cfg := testConfig()
	testdataDir := filepath.Join("..", "..", "testdata")

	result, err := Scan(testdataDir, &cfg, "dev", "unknown")
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Scan returned nil result")
	}

	if len(result.Violations) == 0 {
		t.Error("expected some violations in testdata, got none")
	}

	if result.ExitCode != 1 {
		t.Errorf("expected ExitCode=1 when violations exist, got %d", result.ExitCode)
	}

	if result.Stats.FilesScanned == 0 {
		t.Error("expected FilesScanned > 0")
	}

	if len(result.Stats.ViolationsByRule) == 0 {
		t.Error("expected ViolationsByRule to have entries")
	}

	if result.Tool.Name != "vibe-harness" {
		t.Errorf("expected Tool.Name=vibe-harness, got %s", result.Tool.Name)
	}

	if result.Stats.Duration == "" {
		t.Error("expected Duration to be set")
	}
}

func TestScan_EmptyDir(t *testing.T) {
	cfg := testConfig()
	tmp := t.TempDir()

	result, err := Scan(tmp, &cfg, "dev", "unknown")
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected ExitCode=0 with no violations, got %d", result.ExitCode)
	}

	if len(result.Violations) != 0 {
		t.Errorf("expected no violations, got %d", len(result.Violations))
	}
}

func TestScan_SortedViolations(t *testing.T) {
	cfg := testConfig()
	tmp := t.TempDir()

	os.WriteFile(filepath.Join(tmp, "a.go"), []byte("package a\nAPI_KEY = \"sk-1234567890abcdef12345678\"\n"), 0644)
	os.WriteFile(filepath.Join(tmp, "b.go"), []byte("package b\nAPI_KEY = \"sk-1234567890abcdef12345678\"\n"), 0644)

	result, err := Scan(tmp, &cfg, "dev", "unknown")
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	for i := 1; i < len(result.Violations); i++ {
		prev := result.Violations[i-1]
		cur := result.Violations[i]
		if prev.File > cur.File || (prev.File == cur.File && prev.Line > cur.Line) {
			t.Errorf("violations not sorted: %v before %v", prev, cur)
		}
	}
}

func TestDiscoverFiles(t *testing.T) {
	cfg := testConfig()
	tmp := t.TempDir()

	os.MkdirAll(filepath.Join(tmp, ".git"), 0755)
	os.WriteFile(filepath.Join(tmp, ".git", "config"), []byte("gitdir: foo"), 0644)

	os.MkdirAll(filepath.Join(tmp, "vendor"), 0755)
	os.WriteFile(filepath.Join(tmp, "vendor", "pkg.go"), []byte("package vendor"), 0644)

	os.MkdirAll(filepath.Join(tmp, "node_modules"), 0755)
	os.WriteFile(filepath.Join(tmp, "node_modules", "index.js"), []byte("module.exports = {}"), 0644)

	os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main\nfunc main() {}"), 0644)
	os.WriteFile(filepath.Join(tmp, "app.py"), []byte("print('hello')"), 0644)
	os.WriteFile(filepath.Join(tmp, "readme.txt"), []byte("hello world"), 0644)

	binaryData := []byte("package main\nfunc main() {}")
	binaryData = append(binaryData, 0x00)
	binaryData = append(binaryData, []byte("binary")...)
	os.WriteFile(filepath.Join(tmp, "binary.go"), binaryData, 0644)

	files, _, err := DiscoverFiles(tmp, &cfg)
	if err != nil {
		t.Fatalf("DiscoverFiles returned error: %v", err)
	}

	expectedFiles := map[string]bool{
		filepath.Join(tmp, "app.py"):  true,
		filepath.Join(tmp, "main.go"): true,
	}

	if len(files) != len(expectedFiles) {
		t.Errorf("expected %d files, got %d: %v", len(expectedFiles), len(files), files)
	}

	for _, f := range files {
		if !expectedFiles[f] {
			t.Errorf("unexpected file: %s", f)
		}
	}
}

func TestDiscoverFiles_SourceDirs(t *testing.T) {
	cfg := testConfig()
	cfg.SourceDirs = []string{"src/**"}

	tmp := t.TempDir()

	os.MkdirAll(filepath.Join(tmp, "src"), 0755)
	os.WriteFile(filepath.Join(tmp, "src", "main.go"), []byte("package main"), 0644)

	os.MkdirAll(filepath.Join(tmp, "src", "sub"), 0755)
	os.WriteFile(filepath.Join(tmp, "src", "sub", "lib.go"), []byte("package sub"), 0644)

	os.MkdirAll(filepath.Join(tmp, "vendor"), 0755)
	os.WriteFile(filepath.Join(tmp, "vendor", "pkg.go"), []byte("package vendor"), 0644)

	os.MkdirAll(filepath.Join(tmp, "internal"), 0755)
	os.WriteFile(filepath.Join(tmp, "internal", "secret.go"), []byte("package internal"), 0644)

	os.WriteFile(filepath.Join(tmp, "root.go"), []byte("package main"), 0644)

	files, _, err := DiscoverFiles(tmp, &cfg)
	if err != nil {
		t.Fatalf("DiscoverFiles returned error: %v", err)
	}

	expected := []string{
		filepath.Join(tmp, "src", "main.go"),
		filepath.Join(tmp, "src", "sub", "lib.go"),
	}
	if len(files) != len(expected) {
		t.Errorf("expected %d files, got %d: %v", len(expected), len(files), files)
	}
	for _, f := range expected {
		found := false
		for _, f2 := range files {
			if f == f2 {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected file not found: %s", f)
		}
	}
}

func TestDiscoverFiles_SourceDirs_AncestorSkipping(t *testing.T) {
	cfg := testConfig()
	cfg.SourceDirs = []string{"internal/**"}

	tmp := t.TempDir()

	os.MkdirAll(filepath.Join(tmp, "internal", "checks"), 0755)
	os.WriteFile(filepath.Join(tmp, "internal", "checks", "check.go"), []byte("package checks"), 0644)

	os.MkdirAll(filepath.Join(tmp, "vendor", "pkg"), 0755)
	os.WriteFile(filepath.Join(tmp, "vendor", "pkg", "lib.go"), []byte("package pkg"), 0644)

	files, _, err := DiscoverFiles(tmp, &cfg)
	if err != nil {
		t.Fatalf("DiscoverFiles returned error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file (inside internal/**), got %d: %v", len(files), files)
	}
	if len(files) > 0 && !strings.Contains(files[0], "internal/checks/check.go") {
		t.Errorf("expected internal/checks/check.go, got %s", files[0])
	}
}

func TestDiscoverFiles_SourceDirs_NoMatch(t *testing.T) {
	cfg := testConfig()
	cfg.SourceDirs = []string{"nonexistent/**"}

	tmp := t.TempDir()

	os.MkdirAll(filepath.Join(tmp, "src"), 0755)
	os.WriteFile(filepath.Join(tmp, "src", "main.go"), []byte("package main"), 0644)

	files, _, err := DiscoverFiles(tmp, &cfg)
	if err != nil {
		t.Fatalf("DiscoverFiles returned error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for non-matching source_directories, got %d", len(files))
	}
}

func TestClassifyLine_GoLineComment(t *testing.T) {
	style := CommentStyleForLanguage("go")
	isComment, _ := ClassifyLine("// this is a comment", style, false)
	if !isComment {
		t.Error("expected line comment to be classified as comment")
	}
}

func TestClassifyLine_GoCode(t *testing.T) {
	style := CommentStyleForLanguage("go")
	isComment, _ := ClassifyLine("x := 1", style, false)
	if isComment {
		t.Error("expected code line to not be classified as comment")
	}
}

func TestClassifyLine_PythonLineComment(t *testing.T) {
	style := CommentStyleForLanguage("python")
	isComment, _ := ClassifyLine("# this is a comment", style, false)
	if !isComment {
		t.Error("expected python line comment to be classified as comment")
	}
}

func TestClassifyLine_PythonCode(t *testing.T) {
	style := CommentStyleForLanguage("python")
	isComment, _ := ClassifyLine("x = 1", style, false)
	if isComment {
		t.Error("expected python code line to not be classified as comment")
	}
}

func TestClassifyLine_RubyLineComment(t *testing.T) {
	style := CommentStyleForLanguage("ruby")
	isComment, _ := ClassifyLine("# this is a comment", style, false)
	if !isComment {
		t.Error("expected ruby line comment to be classified as comment")
	}
}

func TestClassifyLine_BlockCommentTransitions(t *testing.T) {
	style := CommentStyleForLanguage("go")

	isComment, inBlock := ClassifyLine("/* start", style, false)
	if !isComment || !inBlock {
		t.Errorf("expected block start: isComment=%v, inBlock=%v", isComment, inBlock)
	}

	isComment, inBlock = ClassifyLine("middle of block", style, true)
	if !isComment || !inBlock {
		t.Errorf("expected middle block line: isComment=%v, inBlock=%v", isComment, inBlock)
	}

	isComment, inBlock = ClassifyLine("end */", style, true)
	if !isComment || inBlock {
		t.Errorf("expected block end: isComment=%v, inBlock=%v", isComment, inBlock)
	}
}

func TestClassifyLine_SingleLineBlockComment(t *testing.T) {
	style := CommentStyleForLanguage("go")
	isComment, inBlock := ClassifyLine("/* inline */", style, false)
	if !isComment || inBlock {
		t.Errorf("expected single-line block: isComment=%v, inBlock=%v", isComment, inBlock)
	}
}

func TestCommentStyleForLanguage_AllSupported(t *testing.T) {
	langs := []string{"go", "java", "typescript", "rust", "javascript", "python", "ruby", "sql"}
	for _, lang := range langs {
		style := CommentStyleForLanguage(lang)
		if len(style.LinePrefixes) == 0 {
			t.Errorf("expected LinePrefixes for %s", lang)
		}
		if style.BlockStart == "" {
			t.Errorf("expected BlockStart for %s", lang)
		}
		if style.BlockEnd == "" {
			t.Errorf("expected BlockEnd for %s", lang)
		}
	}
}

func TestCommentStyleForLanguage_Unknown(t *testing.T) {
	style := CommentStyleForLanguage("cobol")
	if len(style.LinePrefixes) != 0 || style.BlockStart != "" || style.BlockEnd != "" {
		t.Error("expected zero-value CommentStyle for unknown language")
	}
}

func TestClassifyLine_SqlLineComment(t *testing.T) {
	style := CommentStyleForLanguage("sql")
	isComment, _ := ClassifyLine("-- select query", style, false)
	if !isComment {
		t.Error("expected SQL line comment to be classified as comment")
	}
}

func TestMultiViolation_ScannerIntegration(t *testing.T) {
	cfg := testConfig()
	tmp := t.TempDir()

	var lines []string
	lines = append(lines, "package main")
	lines = append(lines, "")
	lines = append(lines, `key := "AKIAIOSFODNN7EXAMPLE"`)
	lines = append(lines, "tls.Config{InsecureSkipVerify: true}")
	for i := 0; i < 310; i++ {
		lines = append(lines, fmt.Sprintf("x%d := 42", i))
	}
	content := []byte(strings.Join(lines, "\n"))

	multiFile := filepath.Join(tmp, "multi.go")
	if err := os.WriteFile(multiFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := Scan(tmp, &cfg, "dev", "unknown")
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if len(result.Violations) == 0 {
		t.Fatal("expected violations from scanner, got none")
	}

	ruleSet := map[string]bool{}
	for _, v := range result.Violations {
		ruleSet[v.RuleID] = true
	}

	expectedRules := []string{"VH-G001", "VH-G005", "VH-G006", "VH-G011"}
	for _, ruleID := range expectedRules {
		if !ruleSet[ruleID] {
			t.Errorf("expected violation from %s in scanner results, rules present: %v", ruleID, ruleSet)
		}
	}

	if len(ruleSet) < 4 {
		t.Errorf("expected violations from at least 4 different rules, got %d: %v", len(ruleSet), ruleSet)
	}
}

func TestDiscoverFiles_NonExistentRoot(t *testing.T) {
	cfg := testConfig()
	_, _, err := DiscoverFiles("/non/existent/path", &cfg)
	if err == nil {
		t.Fatal("expected error for non-existent root, got nil")
	}
}

func TestEdge_UnreadableFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping: running as root, permission denial not testable")
	}

	cfg := testConfig()
	tmp := t.TempDir()

	secretPath := filepath.Join(tmp, "secret.go")
	if err := os.WriteFile(secretPath, []byte("package main\nAPI_KEY = \"sk-1234567890abcdef12345678\"\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	goodPath := filepath.Join(tmp, "good.go")
	if err := os.WriteFile(goodPath, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if err := os.Chmod(secretPath, 0000); err != nil {
		t.Fatalf("failed to chmod file: %v", err)
	}
	defer os.Chmod(secretPath, 0644)

	result, err := Scan(tmp, &cfg, "dev", "unknown")
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if result.Stats.FilesSkipped == 0 {
		t.Error("expected FilesSkipped > 0 for unreadable file")
	}

	if result.Stats.FilesScanned == 0 {
		t.Error("expected FilesScanned > 0 for readable files")
	}
}

func TestEdge_BinaryFileContent(t *testing.T) {
	cfg := testConfig()
	tmp := t.TempDir()

	binaryData := []byte{0x70, 0x61, 0x63, 0x6B, 0x61, 0x67, 0x65, 0x00, 0x6D, 0x61, 0x69, 0x6E}
	if err := os.WriteFile(filepath.Join(tmp, "binary.go"), binaryData, 0644); err != nil {
		t.Fatalf("failed to write binary file: %v", err)
	}

	goodPath := filepath.Join(tmp, "good.go")
	if err := os.WriteFile(goodPath, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("failed to write good file: %v", err)
	}

	files, _, err := DiscoverFiles(tmp, &cfg)
	if err != nil {
		t.Fatalf("DiscoverFiles returned error: %v", err)
	}

	for _, f := range files {
		if filepath.Base(f) == "binary.go" {
			t.Error("binary file with null bytes should be skipped")
		}
	}

	found := false
	for _, f := range files {
		if filepath.Base(f) == "good.go" {
			found = true
		}
	}
	if !found {
		t.Error("expected good.go to be discovered")
	}
}

func TestEdge_EmptyDirectory(t *testing.T) {
	cfg := testConfig()
	tmp := t.TempDir()

	result, err := Scan(tmp, &cfg, "dev", "unknown")
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected ExitCode=0 for empty directory, got %d", result.ExitCode)
	}

	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations for empty directory, got %d", len(result.Violations))
	}

	if result.Stats.FilesScanned != 0 {
		t.Errorf("expected FilesScanned=0 for empty directory, got %d", result.Stats.FilesScanned)
	}
}

func TestEdge_SymlinkCycle(t *testing.T) {
	tmp := t.TempDir()

	subdir := filepath.Join(tmp, "subdir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	linkPath := filepath.Join(subdir, "cycle")
	if err := os.Symlink("..", linkPath); err != nil {
		t.Skipf("skipping: cannot create symlink: %v", err)
	}

	if err := os.WriteFile(filepath.Join(subdir, "real.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cfg := testConfig()

	done := make(chan struct{})
	var files []string
	var err error
	go func() {
		files, _, err = DiscoverFiles(tmp, &cfg)
		close(done)
	}()

	select {
	case <-done:
		if err != nil {
			t.Fatalf("DiscoverFiles returned error: %v", err)
		}
		for _, f := range files {
			if filepath.Base(f) == "cycle" {
				t.Error("symlink target should be skipped")
			}
		}
	case <-time.After(5 * time.Second):
		t.Fatal("DiscoverFiles did not complete within timeout, possible symlink cycle loop")
	}
}
