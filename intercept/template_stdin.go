package intercept

import (
	"path"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// ShellTemplateStdinSources returns syntactic inputs which a recognized shell or Python
// may parse as a program. It tracks redirection order and pipeline/group inputs,
// without reading files or executing producers. It is not variable-flow analysis
// or an attestation of runtime aliases, functions or arbitrary interpreters.
func ShellTemplateStdinSources(file *syntax.File) []syntax.Node {
	inputs := make(map[*syntax.Stmt]map[string][]syntax.Node)
	pending := make(map[*syntax.Stmt][]syntax.Node)
	producers := make(map[*syntax.Stmt]*syntax.Stmt)
	var stack []syntax.Node
	var sources []syntax.Node
	syntax.Walk(file, func(node syntax.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if stmt, ok := node.(*syntax.Stmt); ok {
			fds := make(map[string][]syntax.Node)
			for i := len(stack) - 1; i >= 0; i-- {
				if parent, ok := stack[i].(*syntax.Stmt); ok {
					for fd, input := range inputs[parent] {
						fds[fd] = input
					}
					break
				}
			}
			if producer, piped := producers[stmt]; piped {
				// Walk visits the producer first, so its ordered descriptor
				// overrides are available before binding the consumer's stdin.
				fds["0"] = templateProducerSources(producer, inputs[producer])
			} else if input, overridden := pending[stmt]; overridden {
				fds["0"] = input
			}
			for _, redir := range stmt.Redirs {
				fd := "0"
				if redir.Op == syntax.DplOut || redir.Op == syntax.RdrOut || redir.Op == syntax.AppOut || redir.Op == syntax.ClbOut {
					fd = "1"
				}
				if redir.N != nil {
					fd = redir.N.Value
				}
				switch redir.Op {
				case syntax.DplIn, syntax.DplOut:
					var input []syntax.Node
					from, known := staticWord(redir.Word)
					if known {
						input = append(input, fds[strings.TrimSuffix(from, "-")]...)
					} else {
						// A generated selector can copy any inherited descriptor.
						for _, prior := range fds {
							input = append(input, prior...)
						}
					}
					fds[fd] = append(input, redir.Word)
					if known && len(from) > 1 && strings.HasSuffix(from, "-") {
						fds[strings.TrimSuffix(from, "-")] = nil
					}
				case syntax.RdrIn, syntax.RdrInOut, syntax.WordHdoc:
					fds[fd] = []syntax.Node{redir.Word}
				case syntax.RdrOut, syntax.AppOut, syntax.ClbOut:
					// An output-only file descriptor cannot supply shell input.
					fds[fd] = nil
				case syntax.RdrAll, syntax.AppAll:
					fds["1"], fds["2"] = nil, nil
				case syntax.Hdoc, syntax.DashHdoc:
					fds[fd] = nil
					if redir.Hdoc != nil {
						fds[fd] = []syntax.Node{redir.Hdoc}
					}
				}
			}
			inputs[stmt] = fds
			if pipe, ok := stmt.Cmd.(*syntax.BinaryCmd); ok && (pipe.Op == syntax.Pipe || pipe.Op == syntax.PipeAll) {
				pending[pipe.X] = fds["0"]
				producers[pipe.Y] = pipe.X
			}
			if call, ok := stmt.Cmd.(*syntax.CallExpr); ok && templateMayReadShellStdin(call) {
				sources = append(sources, fds["0"]...)
			}
		}
		stack = append(stack, node)
		return true
	})
	return sources
}

func templateProducerSources(stmt *syntax.Stmt, fds map[string][]syntax.Node) []syntax.Node {
	if call, ok := stmt.Cmd.(*syntax.CallExpr); ok {
		words, known := literalCallWords(call)
		if known && len(words) > 0 {
			switch path.Base(words[0]) {
			case "printf", "echo":
				// Literal output does not consume inherited stdin.
				return []syntax.Node{call}
			case "cat":
				if len(words) == 1 {
					return append([]syntax.Node{call}, fds["0"]...)
				}
			}
		}
	}
	// An opaque producer can consume any readable inherited descriptor.
	sources := []syntax.Node{stmt}
	for _, input := range fds {
		sources = append(sources, input...)
	}
	return sources
}

func templateMayReadShellStdin(call *syntax.CallExpr) bool {
	if len(call.Args) == 0 {
		return false
	}
	words := make([]string, len(call.Args))
	for i, word := range call.Args {
		value, known := staticWord(word)
		if !known {
			value = "<dynamic>"
		}
		words[i] = value
	}
	for len(words) > 0 {
		name := path.Base(words[0])
		tail, wrapper := wrappedCommand(name, words[1:])
		if !wrapper {
			break
		}
		if len(tail) == 0 {
			if name == "command" && len(words) > 1 && (words[1] == "-v" || words[1] == "-V") {
				return false
			}
			// Unsupported wrapper options can choose a stdin interpreter.
			return true
		}
		words = tail
	}
	if len(words) == 0 {
		return false
	}
	name := path.Base(words[0])
	if templatePythonName.MatchString(name) {
		return PythonTemplateReadsStdin(words[1:])
	}
	if name != "bash" && name != "sh" && name != "dash" && name != "ksh" {
		return false
	}
	source := shellProgramSource(name, words[1:])
	return source.kind == "stdin" || source.kind == "unknown"
}
