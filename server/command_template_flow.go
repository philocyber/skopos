package server

import (
	"path"
	"strconv"
	"strings"

	"github.com/Philocyber/skopos/intercept"
	"mvdan.cc/sh/v3/syntax"
)

// Track possible parameter data through syntactic variable definitions. This
// deliberately merges scopes/branches and does not erase earlier definitions:
// a later literal assignment is not proof that every runtime path overwrites it.
// It is not a classifier for arbitrary programs or runtime command aliases.
func shellTemplateParameterFlow(file *syntax.File, containsParameter func(syntax.Node) bool) func(syntax.Node) bool {
	definitions := map[string][]syntax.Node{}
	owners := map[syntax.Node]string{}
	functions := map[string]bool{}
	literalShellPrograms := map[*syntax.Word]bool{}
	var calls []*syntax.CallExpr
	var ownerStack []string
	syntax.Walk(file, func(node syntax.Node) bool {
		if node == nil {
			ownerStack = ownerStack[:len(ownerStack)-1]
			return true
		}
		owner := ""
		if len(ownerStack) != 0 {
			owner = ownerStack[len(ownerStack)-1]
		}
		if function, ok := node.(*syntax.FuncDecl); ok {
			owner = function.Name.Value
			functions[owner] = false
		}
		owners[node] = owner
		ownerStack = append(ownerStack, owner)
		switch value := node.(type) {
		case *syntax.CallExpr:
			calls = append(calls, value)
			if target, ok := shellTemplatePrintfVariable(value); ok {
				definitions[target] = append(definitions[target], value)
			}
			if len(value.Args) != 0 {
				name, known := intercept.ShellTemplateLiteralWord(value.Args[0])
				if known && path.Base(name) != "eval" {
					words, _ := intercept.ShellTemplateProgramWords(value)
					childShell := false
					for _, word := range words {
						identity, fixed := intercept.ShellTemplateLiteralWord(word)
						switch path.Base(identity) {
						case "bash", "sh", "dash", "ksh":
							childShell = childShell || fixed
						}
						literalShellPrograms[word] = childShell
					}
				}
			}
		case *syntax.Assign:
			if value.Name != nil {
				definitions[value.Name.Value] = append(definitions[value.Name.Value], value)
			}
		case *syntax.WordIter:
			if value.Name != nil {
				for _, word := range value.Items {
					definitions[value.Name.Value] = append(definitions[value.Name.Value], word)
				}
			}
		}
		return true
	})
	tainted := map[string]bool{}
	var referencesTainted func(syntax.Node, int, string) bool
	referencesTainted = func(node syntax.Node, depth int, owner string) bool {
		found := false
		syntax.Walk(node, func(child syntax.Node) bool {
			if found {
				return false
			}
			if parameter, ok := child.(*syntax.ParamExp); ok && parameter.Param != nil {
				name := parameter.Param.Value
				functionOwner := owner
				if parsedOwner, exists := owners[child]; exists {
					functionOwner = parsedOwner
				}
				position, positionErr := strconv.Atoi(name)
				positional := name == "@" || name == "*" || positionErr == nil && position > 0
				if tainted[name] || functions[functionOwner] && positional || parameter.Excl && (len(tainted) != 0 || functions[functionOwner]) {
					found = true
				}
			}
			if word, ok := child.(*syntax.Word); ok && (len(tainted) != 0 || functions[owner]) {
				// A literal eval/-c/stdin program can contain expansions which the
				// outer parse correctly treats as text. Parse that fixed text only;
				// no supplied parameter bytes enter this second parser.
				if text, known := intercept.ShellTemplateLiteralWord(word); known && strings.Contains(text, "$") {
					if depth >= 16 {
						found = true
					} else if nested, err := syntax.NewParser().Parse(strings.NewReader(text), ""); err == nil {
						nestedOwner := owner
						if literalShellPrograms[word] {
							// A child shell owns its positional arguments. A fixed
							// program that only uses them as data is safe. If it
							// contains another program sink, conservatively retain
							// influence through any positional reference in its AST.
							programSink := len(intercept.ShellTemplateStdinSources(nested)) != 0
							syntax.Walk(nested, func(node syntax.Node) bool {
								if call, ok := node.(*syntax.CallExpr); ok {
									words, unknown := intercept.ShellTemplateProgramWords(call)
									programSink = programSink || unknown || len(words) != 0
									variableWords, variableUnknown := intercept.ShellTemplateVariableWords(call)
									if variableUnknown && referencesTainted(call, depth+1, owner) {
										programSink = true
									}
									for _, target := range variableWords {
										if referencesTainted(target, depth+1, owner) {
											programSink = true
										}
									}
								}
								return true
							})
							if !programSink {
								// This sentinel cannot be a parsed Bash function name
								// and is separate from the outer script's empty owner.
								nestedOwner = "\x00child-data"
							}
						}
						found = referencesTainted(nested, depth+1, nestedOwner)
					}
				}
			}
			return !found
		})
		return found
	}
	influenced := func(node syntax.Node) bool {
		if node == nil {
			return false
		}
		if containsParameter(node) {
			return true
		}
		return referencesTainted(node, 0, owners[node])
	}
	for changed := true; changed; {
		changed = false
		// Calls to parsed functions can deliver literal parameters to positional
		// variables. Keep that state local to the function, so a fixed bash -c
		// data program's $1 is not confused with the caller's parameter value.
		// Definitions/call sites are conservatively merged without assuming which
		// branch, recursive invocation or function redefinition executes.
		for _, call := range calls {
			if len(call.Args) == 0 {
				continue
			}
			name, known := intercept.ShellTemplateLiteralWord(call.Args[0])
			// set can replace positional data, and shift can move it. Never erase
			// influence on a later overwrite/shift: scopes and branches are merged.
			if known && name == "set" && !functions[owners[call]] {
				for _, argument := range call.Args[1:] {
					if influenced(argument) {
						functions[owners[call]] = true
						changed = true
						break
					}
				}
			}
			if current, exists := functions[name]; !known || !exists || current {
				continue
			}
			for _, argument := range call.Args[1:] {
				if influenced(argument) {
					functions[name] = true
					changed = true
					break
				}
			}
		}
		for name, sources := range definitions {
			if tainted[name] {
				continue
			}
			for _, source := range sources {
				if influenced(source) {
					tainted[name] = true
					changed = true
					break
				}
			}
		}
	}
	return influenced
}

// Bash printf -v writes a named variable without an assignment AST node. Keep
// the target fixed and merge array elements into their base variable. Dynamic
// targets and runtime aliases are separate unresolved channels.
func shellTemplatePrintfVariable(call *syntax.CallExpr) (string, bool) {
	if len(call.Args) < 2 {
		return "", false
	}
	offset := 0
	name, known := intercept.ShellTemplateLiteralWord(call.Args[offset])
	if known && name == "builtin" {
		offset++
		name, known = intercept.ShellTemplateLiteralWord(call.Args[offset])
	}
	if !known || name != "printf" || offset+1 >= len(call.Args) {
		return "", false
	}
	option, known := intercept.ShellTemplateLiteralWord(call.Args[offset+1])
	if !known || !strings.HasPrefix(option, "-v") {
		return "", false
	}
	target := strings.TrimPrefix(option, "-v")
	if target == "" {
		if offset+2 >= len(call.Args) {
			return "", false
		}
		target, known = intercept.ShellTemplateLiteralWord(call.Args[offset+2])
		if !known {
			return "", false
		}
	}
	if index := strings.IndexByte(target, '['); index >= 0 {
		target = target[:index]
	}
	return target, target != ""
}
