// Package treesitter provides tree-sitter grammar bindings for AST-powered checks.
package treesitter

import (
	_ "github.com/tree-sitter/go-tree-sitter"

	_ "github.com/tree-sitter/tree-sitter-go/bindings/go"
	_ "github.com/tree-sitter/tree-sitter-java/bindings/go"
	_ "github.com/tree-sitter/tree-sitter-python/bindings/go"
	_ "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	_ "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	_ "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)