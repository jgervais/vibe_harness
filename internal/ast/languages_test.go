package ast

import (
	"testing"
	"unsafe"
)

func TestLanguageForExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
		ok       bool
	}{
		{".py", "python", true},
		{".go", "go", true},
		{".ts", "typescript", true},
		{".tsx", "typescript", true},
		{".js", "javascript", true},
		{".java", "java", true},
		{".rb", "ruby", true},
		{".rs", "rust", true},
		{".xyz", "", false},
		{".cpp", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		gotLang, gotOk := LanguageForExtension(tt.ext)
		if gotLang != tt.expected || gotOk != tt.ok {
			t.Errorf("LanguageForExtension(%q) = (%q, %v), want (%q, %v)",
				tt.ext, gotLang, gotOk, tt.expected, tt.ok)
		}
	}
}

func TestSupportedLanguages(t *testing.T) {
	langs := SupportedLanguages()

	expected := []string{"go", "java", "javascript", "python", "ruby", "rust", "tsx", "typescript"}

	if len(langs) != len(expected) {
		t.Fatalf("SupportedLanguages() returned %d languages, want %d", len(langs), len(expected))
	}

	for i, lang := range langs {
		if lang != expected[i] {
			t.Errorf("SupportedLanguages()[%d] = %q, want %q", i, lang, expected[i])
		}
	}
}

func TestGetGrammarSupported(t *testing.T) {
	langsWithGrammar := []string{"python", "go", "typescript", "tsx", "javascript", "java", "ruby", "rust"}

	for _, lang := range langsWithGrammar {
		ptr, ok := GetGrammar(lang)
		if !ok {
			t.Errorf("GetGrammar(%q) returned ok=false, want true", lang)
		}
		if ptr == unsafe.Pointer(nil) {
			t.Errorf("GetGrammar(%q) returned nil pointer", lang)
		}
	}
}

func TestGetGrammarUnsupported(t *testing.T) {
	ptr, ok := GetGrammar("cobol")
	if ok {
		t.Errorf("GetGrammar(\"cobol\") returned ok=true, want false")
	}
	if ptr != nil {
		t.Errorf("GetGrammar(\"cobol\") returned non-nil pointer, want nil")
	}
}

func TestGetGrammarCaching(t *testing.T) {
	ptr1, _ := GetGrammar("python")
	ptr2, _ := GetGrammar("python")
	if ptr1 != ptr2 {
		t.Errorf("GetGrammar(\"python\") returned different pointers on successive calls")
	}
}