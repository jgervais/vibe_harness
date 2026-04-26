package ast

import (
	"errors"
	"sync"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Parser struct {
	parsers map[string]*tree_sitter.Parser
}

func NewParser() *Parser {
	return &Parser{
		parsers: make(map[string]*tree_sitter.Parser),
	}
}

func (p *Parser) getOrCreateParser(language string) (*tree_sitter.Parser, error) {
	if parser, ok := p.parsers[language]; ok {
		return parser, nil
	}

	ptr, ok := GetGrammar(language)
	if !ok {
		return nil, nil
	}

	tsLang := tree_sitter.NewLanguage(ptr)
	parser := tree_sitter.NewParser()
	parser.SetLanguage(tsLang)

	p.parsers[language] = parser
	return parser, nil
}

func (p *Parser) ParseFile(language string, content []byte) (*ParseResult, error) {
	if len(content) == 0 {
		return nil, errors.New("parse content is empty")
	}

	parser, err := p.getOrCreateParser(language)
	if err != nil {
		return nil, err
	}
	if parser == nil {
		return nil, nil
	}

	tree := parser.Parse(content, nil)
	hasError := tree.RootNode().HasError()

	return &ParseResult{
		tree:     tree,
		source:   content,
		language: language,
		hasError: hasError,
	}, nil
}

func (p *Parser) IsLanguageSupported(language string) bool {
	_, ok := GetGrammar(language)
	return ok
}

func (p *Parser) Close() {
	for lang, parser := range p.parsers {
		parser.Close()
		delete(p.parsers, lang)
	}
}

type ParseResult struct {
	tree     *tree_sitter.Tree
	source   []byte
	language string
	hasError bool
}

func (r *ParseResult) Tree() *tree_sitter.Tree {
	return r.tree
}

func (r *ParseResult) Source() []byte {
	return r.source
}

func (r *ParseResult) Language() string {
	return r.language
}

func (r *ParseResult) HasError() bool {
	return r.hasError
}

func (r *ParseResult) Close() {
	if r.tree != nil {
		r.tree.Close()
		r.tree = nil
	}
}

type QuerySet struct {
	queries    map[string]string
	compiled   map[string]*tree_sitter.Query
	compileMu  sync.Mutex
}

func NewQuerySet(queries map[string]string) *QuerySet {
	return &QuerySet{
		queries:  queries,
		compiled: make(map[string]*tree_sitter.Query),
	}
}

func (qs *QuerySet) GetQuery(language string) (string, bool) {
	q, ok := qs.queries[language]
	return q, ok
}

func (qs *QuerySet) Compile(language string, lang *tree_sitter.Language) (*tree_sitter.Query, error) {
	qs.compileMu.Lock()
	defer qs.compileMu.Unlock()

	if q, ok := qs.compiled[language]; ok {
		return q, nil
	}

	raw, ok := qs.queries[language]
	if !ok {
		return nil, errors.New("no query defined for language: " + language)
	}

	q, err := tree_sitter.NewQuery(lang, raw)
	if err != nil {
		return nil, err
	}

	qs.compiled[language] = q
	return q, nil
}

func (qs *QuerySet) Close() {
	qs.compileMu.Lock()
	defer qs.compileMu.Unlock()

	for lang, q := range qs.compiled {
		q.Close()
		delete(qs.compiled, lang)
	}
}