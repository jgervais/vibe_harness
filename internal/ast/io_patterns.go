package ast

var ioPatterns = map[string][]string{
	"python":     {"open", "requests.", "socket.", "urllib.", "conn.", "cursor.", "read", "write", "send", "recv", "connect"},
	"typescript": {"fetch", "fs.", "http.", "axios.", "readFile", "writeFile", "createReadStream", "createWriteStream"},
	"go":         {"os.", "net.", "http.", "sql.", "io.", "ReadFile", "WriteFile", "Open", "Dial", "Listen"},
	"java":       {"Files.", "Socket", "PreparedStatement", "Connection", "readLine", "read", "write", "connect"},
	"ruby":       {"open", "IO.", "Net::", "File.", "read", "write", "send", "recv", "connect"},
	"rust":       {"std::fs::", "std::net::", "std::io::", "reqwest::", "read_to_string", "write", "read"},
}

func IOPatternsForLanguage(language string) []string {
	patterns, ok := ioPatterns[language]
	if !ok {
		return nil
	}
	result := make([]string, len(patterns))
	copy(result, patterns)
	return result
}

var ioQueryPatterns = map[string]map[string]string{
	"python": {
		"call_identifier": "(call_expression function: (identifier) @io-call)",
		"call_attribute":  "(call_expression function: (attribute) @io-call)",
	},
	"typescript": {
		"call_identifier":    "(call_expression function: (identifier) @io-call)",
		"call_member":        "(call_expression function: (member_expression) @io-call)",
	},
	"go": {
		"call_identifier": "(call_expression function: (identifier) @io-call)",
		"call_selector":   "(call_expression function: (selector_expression) @io-call)",
	},
	"java": {
		"method_name":   "(method_invocation name: (identifier) @io-call)",
		"method_object": "(method_invocation object: (identifier) @io-call)",
	},
	"ruby": {
		"call_method": "(call method: (identifier) @io-call)",
		"send_method": "(send receiver: (identifier) method: (identifier) @io-call)",
	},
	"rust": {
		"call_identifier":     "(call_expression function: (identifier) @io-call)",
		"call_scoped":          "(call_expression function: (scoped_identifier) @io-call)",
	},
}

func IOQueryPatterns(language string) map[string]string {
	queries, ok := ioQueryPatterns[language]
	if !ok {
		return nil
	}
	result := make(map[string]string, len(queries))
	for k, v := range queries {
		result[k] = v
	}
	return result
}

var errorHandlingPatterns = map[string][]string{
	"python":     {"try", "except", "with"},
	"typescript": {"try", "catch", "finally"},
	"go":         {"if err", "return err"},
	"java":       {"try", "catch", "finally"},
	"ruby":       {"begin", "rescue", "ensure"},
	"rust":       {"?", "match", "Result"},
}

func ErrorHandlingPatterns(language string) []string {
	patterns, ok := errorHandlingPatterns[language]
	if !ok {
		return nil
	}
	result := make([]string, len(patterns))
	copy(result, patterns)
	return result
}