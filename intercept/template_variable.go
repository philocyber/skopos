package intercept

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// ShellTemplateVariableWords finds Bash builtin operands that name variables.
// Array subscripts in these names can execute arithmetic/command expansions even
// when the operand was quoted as a literal argument by the caller.
func ShellTemplateVariableWords(call *syntax.CallExpr) (protected []*syntax.Word, unknown bool) {
	if len(call.Args) == 0 {
		return nil, false
	}
	offset := 0
	name, fixed := staticWord(call.Args[offset])
	if !fixed {
		return nil, false
	}
	// command preserves builtin execution while bypassing function lookup.
	// Keep identity queries as data and refuse unknown wrapper options.
	for name == "command" {
		offset++
		for offset < len(call.Args) {
			option, known := staticWord(call.Args[offset])
			if !known {
				return nil, true
			}
			if option == "-v" || option == "-V" {
				return nil, false
			}
			if option == "--" {
				offset++
				break
			}
			if option == "-p" {
				offset++
				continue
			}
			if strings.HasPrefix(option, "-") {
				return nil, true
			}
			break
		}
		if offset >= len(call.Args) {
			return nil, false
		}
		name, fixed = staticWord(call.Args[offset])
		if !fixed {
			return nil, true
		}
	}
	args := call.Args[offset+1:]
	switch name {
	case "printf":
		if len(args) == 0 {
			return nil, false
		}
		option, known := staticWord(args[0])
		// A dynamic first operand can select -v instead of a format string.
		protected = append(protected, args[0])
		if !known {
			return protected, true
		}
		if option == "-v" {
			if len(args) < 2 {
				return protected, true
			}
			protected = append(protected, args[1])
		}
		return protected, false
	case "unset":
		// Function/nameref flags are kept fixed as well. A literal name does not
		// depend on knowing which runtime array/nameref binding already exists.
		return args, false
	case "read":
		for i := 0; i < len(args); i++ {
			option, known := staticWord(args[i])
			if !known {
				return append(protected, args[i:]...), true
			}
			if option == "--" {
				return append(protected, args[i:]...), false
			}
			if !strings.HasPrefix(option, "-") || option == "-" {
				return append(protected, args[i:]...), false
			}
			protected = append(protected, args[i])
			for flag := 1; flag < len(option); flag++ {
				switch option[flag] {
				case 'r', 's', 'e', 'E':
					continue
				case 'a', 'd', 'i', 'n', 'N', 'p', 't', 'u':
					// A fixed prompt/delimiter/default flag makes its attached
					// suffix data too. Unknown/dynamic flag spellings were refused
					// above; parameter bytes cannot change this fixed prefix.
					dataOption := option[flag] == 'd' || option[flag] == 'i' || option[flag] == 'p'
					if flag+1 < len(option) && dataOption {
						protected = protected[:len(protected)-1]
					}
					if flag+1 == len(option) {
						i++
						if i >= len(args) {
							return protected, true
						}
						if !dataOption {
							protected = append(protected, args[i])
						}
					}
					flag = len(option)
				default:
					return protected, true
				}
			}
		}
		return protected, false
	default:
		return nil, false
	}
}
