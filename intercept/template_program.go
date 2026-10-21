package intercept

import (
	"path"

	"mvdan.cc/sh/v3/syntax"
)

// ShellTemplateLiteralWord decodes a syntactically literal shell word without
// executing expansions or consulting runtime command/environment state.
func ShellTemplateLiteralWord(word *syntax.Word) (string, bool) { return staticWord(word) }

// ShellTemplateProgramWords identifies words controlling a recognized nested
// shell or Python program. It performs no execution, filesystem or environment lookup.
// Unknown means the recognized invocation cannot be classified by this grammar;
// callers must not assume its placeholder arguments are ordinary data.
// This is not a classifier for arbitrary interpreters or runtime-defined commands.
func ShellTemplateProgramWords(call *syntax.CallExpr) (protected []*syntax.Word, unknown bool) {
	if len(call.Args) == 0 {
		return nil, false
	}
	words := make([]string, len(call.Args))
	known := make([]bool, len(words))
	for i, word := range call.Args {
		words[i], known[i] = staticWord(word)
	}
	offset := 0
	for known[offset] {
		name := path.Base(words[offset])
		tail, wrapper := wrappedCommand(name, words[offset+1:])
		if !wrapper {
			break
		}
		if len(tail) == 0 {
			// Identity queries never execute their argument.
			if name == "command" && len(words) > offset+1 && known[offset+1] && (words[offset+1] == "-v" || words[offset+1] == "-V") {
				return nil, false
			}
			return nil, true
		}
		next := len(words) - len(tail)
		for i := offset; i < next; i++ {
			if !known[i] || !singleRequestArgument(call.Args[i]) {
				return nil, true
			}
		}
		protected = append(protected, call.Args[offset:next]...)
		offset = next
	}
	if !known[offset] {
		return nil, len(protected) != 0
	}
	name := path.Base(words[offset])
	if templatePythonName.MatchString(name) {
		return pythonTemplateProgramWords(call, offset, words, known, protected)
	}
	if name == "eval" {
		// eval joins every argument and parses the result again.
		return append(protected, call.Args[offset:]...), false
	}
	if name != "bash" && name != "sh" && name != "dash" && name != "ksh" {
		if len(protected) != 0 {
			// Changing a wrapped command identity can select an interpreter.
			return append(protected, call.Args[offset]), false
		}
		return nil, false
	}
	source := shellProgramSource(name, words[offset+1:])
	if source.kind == "unknown" {
		return nil, true
	}
	end := offset + 1 + source.index
	if end > len(words) {
		end = len(words)
	}
	for i := offset; i < end; i++ {
		if !known[i] || !singleRequestArgument(call.Args[i]) {
			return nil, true
		}
	}
	// Option words must remain fixed: interpolation could select -c or change
	// noexec mode. Script filenames and positional arguments are separate data.
	protected = append(protected, call.Args[offset:end]...)
	if source.kind == "command" && end < len(words) {
		protected = append(protected, call.Args[end])
	}
	if !source.stopped && !source.command && end < len(words) {
		// A presumed script filename can become an option after substitution.
		protected = append(protected, call.Args[end])
	}
	return protected, false
}
