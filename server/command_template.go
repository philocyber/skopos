package server

import (
	"errors"
	"regexp"
	"strings"

	"github.com/Philocyber/skopos/intercept"
	"github.com/Philocyber/skopos/locale"
	"mvdan.cc/sh/v3/syntax"
)

var commandPlaceholder = regexp.MustCompile(`\{([^{}]+)\}`)

type templateLiteral struct {
	start, end int
	quote      byte
}

// Keep the operator's shell program unchanged; insert values only in parsed
// literal data positions, with escaping for that position's actual quote mode.
func renderBashCommandTemplate(program string, params map[string]any) (string, error) {
	file, err := syntax.NewParser().Parse(strings.NewReader(program), "")
	if err != nil {
		return "", errors.New("command template has invalid Bash syntax")
	}
	containsParameter := func(node syntax.Node) bool {
		text := program[int(node.Pos().Offset()):int(node.End().Offset())]
		for _, match := range commandPlaceholder.FindAllStringSubmatch(text, -1) {
			if _, exists := params[match[1]]; exists {
				return true
			}
		}
		return false
	}
	parameterFlow := shellTemplateParameterFlow(file, containsParameter)
	var programParameter, variableParameter bool
	for _, source := range intercept.ShellTemplateStdinSources(file) {
		if parameterFlow(source) {
			programParameter = true
		}
	}
	syntax.Walk(file, func(node syntax.Node) bool {
		if call, ok := node.(*syntax.CallExpr); ok {
			words, unknown := intercept.ShellTemplateProgramWords(call)
			if unknown && parameterFlow(call) {
				programParameter = true
			}
			for _, word := range words {
				if parameterFlow(word) {
					programParameter = true
				}
			}
			variableWords, variableUnknown := intercept.ShellTemplateVariableWords(call)
			if variableUnknown && parameterFlow(call) {
				variableParameter = true
			}
			for _, word := range variableWords {
				if parameterFlow(word) {
					variableParameter = true
				}
			}
		}
		return true
	})
	if programParameter {
		return "", locale.NewError("Command parameters cannot be interpolated into nested shell or Python programs or their execution options; pass separate data arguments")
	}
	if variableParameter {
		return "", locale.NewError("Command parameters cannot be used as shell variable names or assignment options; use fixed names and separate data arguments")
	}
	var spans []templateLiteral
	var stack []syntax.Node
	syntax.Walk(file, func(node syntax.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		var parent syntax.Node
		if len(stack) > 0 {
			parent = stack[len(stack)-1]
		}
		stack = append(stack, node)
		// Here-document expansion and delimiters need their own data protocol.
		for _, ancestor := range stack {
			switch value := ancestor.(type) {
			case *syntax.ParamExp, *syntax.ArithmExp, *syntax.ArithmCmd:
				return true
			case *syntax.CmdSubst:
				if value.Backquotes {
					return true
				}
			}
			if redirect, ok := ancestor.(*syntax.Redirect); ok && redirect.Hdoc != nil && node.Pos().Offset() >= redirect.Hdoc.Pos().Offset() {
				return true
			}
		}
		switch value := node.(type) {
		case *syntax.Lit:
			quote := byte(0)
			if _, ok := parent.(*syntax.DblQuoted); ok {
				quote = '"'
			}
			spans = append(spans, templateLiteral{int(value.Pos().Offset()), int(value.End().Offset()), quote})
		case *syntax.SglQuoted:
			quote := byte('\'')
			start := int(value.Pos().Offset()) + 1
			if value.Dollar {
				quote = '$'
				start++
			}
			spans = append(spans, templateLiteral{start, int(value.End().Offset()) - 1, quote})
		}
		return true
	})
	var out strings.Builder
	offset := 0
	for _, match := range commandPlaceholder.FindAllStringSubmatchIndex(program, -1) {
		value, exists := params[program[match[2]:match[3]]]
		if !exists {
			continue
		}
		start, end := match[0], match[1]
		found := false
		quote := byte(0)
		for _, span := range spans {
			if start >= span.start && end <= span.end {
				found = true
				quote = span.quote
				break
			}
		}
		if !found {
			return "", errors.New("command placeholder must be in literal shell argument data, not shell expansion or here-document syntax")
		}
		if quote != '\'' {
			slashes := 0
			for n := start - 1; n >= 0 && program[n] == '\\'; n-- {
				slashes++
			}
			if slashes%2 != 0 {
				return "", errors.New("command placeholder cannot follow an unmatched shell escape")
			}
		}
		text := valToStr(value)
		if strings.ContainsRune(text, 0) {
			return "", errors.New("command parameter contains a NUL byte")
		}
		var replacement string
		switch quote {
		case '\'':
			replacement = strings.ReplaceAll(text, "'", `'\''`)
		case '$':
			replacement = strings.NewReplacer(`\`, `\\`, "'", `\'`).Replace(text)
		case '"':
			replacement = strings.NewReplacer(`\`, `\\`, `"`, `\"`, `$`, `\$`, "`", "\\`").Replace(text)
		default:
			replacement = shellQuote(text)
		}
		out.WriteString(program[offset:start])
		out.WriteString(replacement)
		offset = end
	}
	out.WriteString(program[offset:])
	return out.String(), nil
}
