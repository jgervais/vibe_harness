module github.com/jgervais/vibe_harness

go 1.24.4

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/tree-sitter/go-tree-sitter v0.25.0
	github.com/tree-sitter/tree-sitter-go v0.25.0
	github.com/tree-sitter/tree-sitter-java v0.23.5
	github.com/tree-sitter/tree-sitter-python v0.25.0
	github.com/tree-sitter/tree-sitter-ruby v0.23.1
	github.com/tree-sitter/tree-sitter-rust v0.24.2
	github.com/tree-sitter/tree-sitter-typescript v0.23.2
)

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
)

replace (
	github.com/tree-sitter/tree-sitter-go/bindings/go => github.com/tree-sitter/tree-sitter-go v0.25.0
	github.com/tree-sitter/tree-sitter-java/bindings/go => github.com/tree-sitter/tree-sitter-java v0.23.5
	github.com/tree-sitter/tree-sitter-python/bindings/go => github.com/tree-sitter/tree-sitter-python v0.25.0
	github.com/tree-sitter/tree-sitter-ruby/bindings/go => github.com/tree-sitter/tree-sitter-ruby v0.23.1
	github.com/tree-sitter/tree-sitter-rust/bindings/go => github.com/tree-sitter/tree-sitter-rust v0.24.2
	github.com/tree-sitter/tree-sitter-typescript/bindings/go => github.com/tree-sitter/tree-sitter-typescript v0.23.2
)
