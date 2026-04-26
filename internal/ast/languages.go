package ast

import (
	"sort"
	"sync"
	"unsafe"

	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

var extensionToLanguage = map[string]string{
	".py":   "python",
	".ts":   "typescript",
	".tsx":  "typescript",
	".js":   "javascript",
	".go":   "go",
	".java": "java",
	".rb":   "ruby",
	".rs":   "rust",
}

var languageToGrammar = map[string]func() unsafe.Pointer{
	"python":     tree_sitter_python.Language,
	"go":         tree_sitter_go.Language,
	"typescript": tree_sitter_typescript.LanguageTypescript,
	"tsx":        tree_sitter_typescript.LanguageTSX,
	"javascript": tree_sitter_typescript.LanguageTypescript,
	"java":       tree_sitter_java.Language,
	"ruby":       tree_sitter_ruby.Language,
	"rust":       tree_sitter_rust.Language,
}

var (
	grammarCache   map[string]unsafe.Pointer
	grammarCacheMu sync.Mutex
)

func SupportedLanguages() []string {
	names := make([]string, 0, len(languageToGrammar))
	for lang := range languageToGrammar {
		names = append(names, lang)
	}
	sort.Strings(names)
	return names
}

func LanguageForExtension(ext string) (string, bool) {
	lang, ok := extensionToLanguage[ext]
	return lang, ok
}

func GetGrammar(language string) (unsafe.Pointer, bool) {
	grammarFn, ok := languageToGrammar[language]
	if !ok {
		return nil, false
	}

	grammarCacheMu.Lock()
	defer grammarCacheMu.Unlock()

	if grammarCache == nil {
		grammarCache = make(map[string]unsafe.Pointer)
	}

	if ptr, cached := grammarCache[language]; cached {
		return ptr, true
	}

	ptr := grammarFn()
	grammarCache[language] = ptr
	return ptr, true
}