package locale

func init() {
	Register("Command parameters cannot be used as interpreter stdin programs; use fixed programs that read separate data", "Command parameters cannot be used as interpreter stdin programs; use fixed programs that read separate data")
	Register("Command parameters cannot be used to construct PowerShell script blocks; use fixed programs and separate data arguments", "Command parameters cannot be used to construct PowerShell script blocks; use fixed programs and separate data arguments")
	Register("Command parameters cannot be passed to expression-evaluating PowerShell commands; use fixed programs and separate data arguments", "Command parameters cannot be passed to expression-evaluating PowerShell commands; use fixed programs and separate data arguments")
	Register("Command parameters cannot be used as shell variable names or assignment options; use fixed names and separate data arguments", "Command parameters cannot be used as shell variable names or assignment options; use fixed names and separate data arguments")
	Register("Command tool parameters must be declared in its schema", "Command tool parameters must be declared in its schema")
	Register("Command parameters cannot be interpolated into nested shell or Python programs or their execution options; pass separate data arguments", "Command parameters cannot be interpolated into nested shell or Python programs or their execution options; pass separate data arguments")
	Register("Command parameters cannot be interpolated into CMD programs; use literal native arguments or a script data channel", "Command parameters cannot be interpolated into CMD programs; use literal native arguments or a script data channel")
}
