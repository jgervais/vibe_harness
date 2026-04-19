package output

import (
	"fmt"
	"io"

	"github.com/jgervais/vibe_harness/internal/scanner"
)

func FormatHuman(w io.Writer, result *scanner.ScanResult) {
	files := map[string]bool{}
	for _, v := range result.Violations {
		files[v.File] = true
		fmt.Fprintf(w, "%s:%d:%s — %s\n", v.File, v.Line, v.RuleID, v.Message)
	}
	fmt.Fprintf(w, "\n%d violation(s) found in %d file(s)\n", len(result.Violations), len(files))
}