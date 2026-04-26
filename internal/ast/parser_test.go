package ast

import (
	"testing"
	"unsafe"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("NewParser() returned nil")
	}
	if p.parsers == nil {
		t.Fatal("NewParser() parser map is nil")
	}
	p.Close()
}

func TestParserParseFileUnsupportedLanguage(t *testing.T) {
	p := NewParser()
	defer p.Close()

	result, err := p.ParseFile("brainfuck", []byte("++++"))
	if result != nil {
		t.Errorf("ParseFile(unsupported) result = %v, want nil", result)
	}
	if err != nil {
		t.Errorf("ParseFile(unsupported) err = %v, want nil", err)
	}
}

func TestParserParseFileValidPython(t *testing.T) {
	p := NewParser()
	defer p.Close()

	src := []byte("def hello():\n    pass\n")
	result, err := p.ParseFile("python", src)
	if err != nil {
		t.Fatalf("ParseFile(python) err = %v", err)
	}
	if result == nil {
		t.Fatal("ParseFile(python) result is nil")
	}
	defer result.Close()

	if result.Language() != "python" {
		t.Errorf("Language() = %q, want %q", result.Language(), "python")
	}
	if result.HasError() {
		t.Error("HasError() = true, want false for valid Python")
	}
	if result.Tree() == nil {
		t.Error("Tree() returned nil")
	}
	root := result.Tree().RootNode()
	if root.Kind() != "module" {
		t.Errorf("Root kind = %q, want %q", root.Kind(), "module")
	}
}

func TestParserParseFileSyntaxError(t *testing.T) {
	p := NewParser()
	defer p.Close()

	src := []byte("def hello(\n")
	result, err := p.ParseFile("python", src)
	if err != nil {
		t.Fatalf("ParseFile(syntax error) err = %v", err)
	}
	if result == nil {
		t.Fatal("ParseFile(syntax error) result is nil")
	}
	defer result.Close()

	if !result.HasError() {
		t.Error("HasError() = false, want true for syntax error")
	}
}

func TestParserParseFileEmptyContent(t *testing.T) {
	p := NewParser()
	defer p.Close()

	result, err := p.ParseFile("python", []byte{})
	if result != nil {
		t.Errorf("ParseFile(empty) result = %v, want nil", result)
	}
	if err == nil {
		t.Error("ParseFile(empty) err = nil, want error")
	}

	result, err = p.ParseFile("python", nil)
	if result != nil {
		t.Errorf("ParseFile(nil) result = %v, want nil", result)
	}
	if err == nil {
		t.Error("ParseFile(nil) err = nil, want error")
	}
}

func TestParserClose(t *testing.T) {
	p := NewParser()
	_, _ = p.ParseFile("python", []byte("x = 1\n"))
	p.Close()

	if len(p.parsers) != 0 {
		t.Errorf("after Close(), parsers map has %d entries, want 0", len(p.parsers))
	}
}

func TestParserIsLanguageSupported(t *testing.T) {
	p := NewParser()
	defer p.Close()

	if !p.IsLanguageSupported("python") {
		t.Error("IsLanguageSupported(python) = false, want true")
	}
	if !p.IsLanguageSupported("go") {
		t.Error("IsLanguageSupported(go) = false, want true")
	}
	if p.IsLanguageSupported("brainfuck") {
		t.Error("IsLanguageSupported(brainfuck) = true, want false")
	}
	if p.IsLanguageSupported("") {
		t.Error("IsLanguageSupported('') = true, want false")
	}
}

func TestQuerySetGetQuery(t *testing.T) {
	qs := NewQuerySet(map[string]string{
		"python": "(function_definition) @func",
	})

	if q, ok := qs.GetQuery("python"); !ok || q != "(function_definition) @func" {
		t.Errorf("GetQuery(python) = (%q, %v), want (%q, true)", q, ok, "(function_definition) @func")
	}
	if _, ok := qs.GetQuery("cobol"); ok {
		t.Error("GetQuery(cobol) ok = true, want false")
	}
}

func TestQuerySetCompileCaching(t *testing.T) {
	ptr, ok := GetGrammar("python")
	if !ok {
		t.Fatal("GetGrammar(python) failed")
	}
	tsLang := tree_sitter.NewLanguage(ptr)

	qs := NewQuerySet(map[string]string{
		"python": "(function_definition) @func",
	})
	defer qs.Close()

	q1, err := qs.Compile("python", tsLang)
	if err != nil {
		t.Fatalf("Compile(python) err = %v", err)
	}
	if q1 == nil {
		t.Fatal("Compile(python) returned nil query")
	}

	q2, err := qs.Compile("python", tsLang)
	if err != nil {
		t.Fatalf("second Compile(python) err = %v", err)
	}
	if q2 != q1 {
		t.Error("second Compile(python) returned different query, expected cached result")
	}
}

func TestQuerySetCompileUnsupportedLanguage(t *testing.T) {
	ptr, _ := GetGrammar("python")
	tsLang := tree_sitter.NewLanguage(ptr)

	qs := NewQuerySet(map[string]string{})
	defer qs.Close()

	_, err := qs.Compile("cobol", tsLang)
	if err == nil {
		t.Error("Compile(unsupported) err = nil, want error")
	}
}

func TestQuerySetClose(t *testing.T) {
	ptr, _ := GetGrammar("python")
	_ = unsafe.Pointer(ptr)
	tsLang := tree_sitter.NewLanguage(ptr)

	qs := NewQuerySet(map[string]string{
		"python": "(function_definition) @func",
	})

	_, err := qs.Compile("python", tsLang)
	if err != nil {
		t.Fatalf("Compile err = %v", err)
	}

	qs.Close()

	if len(qs.compiled) != 0 {
		t.Errorf("after Close(), compiled map has %d entries, want 0", len(qs.compiled))
	}
}