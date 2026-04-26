package scanner

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jgervais/vibe_harness/internal/ast"
	"github.com/jgervais/vibe_harness/internal/checks/generic"
	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/rules"
)

type ToolInfo struct {
	Name      string
	Version   string
	RulesHash string
}

type ScanStats struct {
	FilesScanned    int
	FilesSkipped    int
	ViolationsByRule map[string]int
	Duration         string
}

type ScanResult struct {
	Tool       ToolInfo
	Target     string
	Violations []rules.Violation
	Stats      ScanStats
	ExitCode   int
}

type CommentStyle struct {
	LinePrefixes []string
	BlockStart   string
	BlockEnd     string
}

var commentStyles = map[string]CommentStyle{
	"go":         {LinePrefixes: []string{"//"}, BlockStart: "/*", BlockEnd: "*/"},
	"java":       {LinePrefixes: []string{"//"}, BlockStart: "/*", BlockEnd: "*/"},
	"typescript":  {LinePrefixes: []string{"//"}, BlockStart: "/*", BlockEnd: "*/"},
	"rust":       {LinePrefixes: []string{"//"}, BlockStart: "/*", BlockEnd: "*/"},
	"javascript": {LinePrefixes: []string{"//"}, BlockStart: "/*", BlockEnd: "*/"},
	"python":     {LinePrefixes: []string{"#"}, BlockStart: `"""`, BlockEnd: `"""`},
	"ruby":       {LinePrefixes: []string{"#"}, BlockStart: "=begin", BlockEnd: "=end"},
	"sql":        {LinePrefixes: []string{"--"}, BlockStart: "/*", BlockEnd: "*/"},
}

func CommentStyleForLanguage(lang string) CommentStyle {
	style, ok := commentStyles[strings.ToLower(lang)]
	if !ok {
		return CommentStyle{}
	}
	return style
}

func ClassifyLine(line string, style CommentStyle, inBlockComment bool) (isComment bool, stillInBlock bool) {
	trimmed := strings.TrimLeft(line, " \t")

	if inBlockComment {
		isComment = true
		stillInBlock = true
		if strings.Contains(line, style.BlockEnd) {
			stillInBlock = false
		}
		return
	}

	for _, prefix := range style.LinePrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			isComment = true
			return
		}
	}

	if style.BlockStart != "" && strings.Contains(line, style.BlockStart) {
		isComment = true
		stillInBlock = true
		if strings.Contains(line, style.BlockEnd) {
			stillInBlock = false
		}
		return
	}

	isComment = false
	stillInBlock = false
	return
}

var skipDirs = map[string]bool{
	".git":         true,
	"vendor":       true,
	"node_modules": true,
}

func DiscoverFiles(root string, cfg *config.Config) ([]string, int, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, 0, err
	}
	if !info.IsDir() {
		return nil, 0, fs.ErrInvalid
	}

	var files []string
	skipped := 0
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, err)
			skipped++
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		relPath, _ := filepath.Rel(root, path)

		if d.IsDir() {
			if skipDirs[d.Name()] {
				return fs.SkipDir
			}
			if len(cfg.SourceDirs) > 0 && !cfg.IsSourceDirAncestor(relPath) {
				return fs.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(d.Name())
		if _, ok := cfg.Languages[ext]; !ok {
			return nil
		}

		if len(cfg.SourceDirs) > 0 && !cfg.IsInSourceDir(relPath) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot open %s: %v\n", path, err)
			skipped++
			return nil
		}
		defer f.Close()

		buf := make([]byte, 512)
		n, err := f.Read(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot read %s: %v\n", path, err)
			skipped++
			return nil
		}
		if bytes.IndexByte(buf[:n], 0x00) >= 0 {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	sort.Strings(files)
	return files, skipped, nil
}

func Scan(target string, cfg *config.Config, version string, rulesHash string) (*ScanResult, error) {
	start := time.Now()

	absTarget, err := filepath.Abs(target)
	if err != nil {
		absTarget = target
	}

	files, discoverSkipped, err := DiscoverFiles(target, cfg)
	if err != nil {
		return nil, fmt.Errorf("discovering files: %w", err)
	}

	astParser := ast.NewParser()
	defer astParser.Close()

	fl := generic.NewFileLengthCheck()
	hs := generic.NewHardcodedSecretsCheck()
	mv := generic.NewMagicValuesCheck()
	cr := generic.NewCommentRatioCheck()
	sf := generic.NewSecurityFeaturesCheck()
	dr := generic.NewDuplicationCheck()

	singleFileChecks := []generic.Check{fl, hs, cr, sf}

	var allViolations []rules.Violation
	var fileContents []generic.FileContent
	filesScanned := 0
	filesSkipped := discoverSkipped
	violationsByRule := map[string]int{}

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, err)
			filesSkipped++
			continue
		}

		ext := filepath.Ext(path)
		language := cfg.Languages[ext]

		var parseResult *ast.ParseResult
		if astParser.IsLanguageSupported(language) {
			parseResult, err = astParser.ParseFile(language, content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: AST parse failed for %s: %v\n", path, err)
				parseResult = nil
			}
		}

		for _, chk := range singleFileChecks {
			var violations []rules.Violation
			astCheck, ok := chk.(generic.ASTCheck)
			if ok && parseResult != nil {
				violations = astCheck.CheckFileAST(path, content, language, cfg, parseResult)
			} else {
				violations = chk.CheckFile(path, content, language, cfg)
			}
			for _, v := range violations {
				allViolations = append(allViolations, v)
				violationsByRule[v.RuleID]++
			}
		}

		if parseResult != nil {
			parseResult.Close()
		}

		fileContents = append(fileContents, generic.FileContent{
			Path:    path,
			Content: content,
		})
		filesScanned++
	}

	dupViolations := dr.CheckFiles(fileContents, cfg)
	for _, v := range dupViolations {
		allViolations = append(allViolations, v)
		violationsByRule[v.RuleID]++
	}

	mvViolations := mv.CheckFiles(fileContents, cfg)
	for _, v := range mvViolations {
		allViolations = append(allViolations, v)
		violationsByRule[v.RuleID]++
	}

	sort.Slice(allViolations, func(i, j int) bool {
		if allViolations[i].File != allViolations[j].File {
			return allViolations[i].File < allViolations[j].File
		}
		return allViolations[i].Line < allViolations[j].Line
	})

	duration := time.Since(start)
	exitCode := 0
	if len(allViolations) > 0 {
		exitCode = 1
	}

	return &ScanResult{
		Tool: ToolInfo{
			Name:      "vibe-harness",
			Version:   version,
			RulesHash: rulesHash,
		},
		Target:     absTarget,
		Violations: allViolations,
		Stats: ScanStats{
			FilesScanned:    filesScanned,
			FilesSkipped:    filesSkipped,
			ViolationsByRule: violationsByRule,
			Duration:         duration.String(),
		},
		ExitCode: exitCode,
	}, nil
}