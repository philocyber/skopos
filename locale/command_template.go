package locale

func init() {
	Register("Command parameters cannot be used as interpreter stdin programs; use fixed programs that read separate data", "Command parameters cannot be used as interpreter stdin programs; use fixed programs that read separate data", "명령 매개변수를 인터프리터의 표준 입력 프로그램으로 사용할 수 없습니다. 별도의 데이터를 읽는 고정 프로그램을 사용하세요")
	Register("Command parameters cannot be used to construct PowerShell script blocks; use fixed programs and separate data arguments", "Command parameters cannot be used to construct PowerShell script blocks; use fixed programs and separate data arguments", "명령 매개변수로 PowerShell 스크립트 블록을 구성할 수 없습니다. 고정 프로그램과 별도의 데이터 인수를 사용하세요")
	Register("Command parameters cannot be passed to expression-evaluating PowerShell commands; use fixed programs and separate data arguments", "Command parameters cannot be passed to expression-evaluating PowerShell commands; use fixed programs and separate data arguments", "명령 매개변수를 식을 실행하는 PowerShell 명령에 전달할 수 없습니다. 고정 프로그램과 별도의 데이터 인수를 사용하세요")
	Register("Command parameters cannot be used as shell variable names or assignment options; use fixed names and separate data arguments", "Command parameters cannot be used as shell variable names or assignment options; use fixed names and separate data arguments", "명령 매개변수를 셸 변수 이름이나 대입 옵션으로 사용할 수 없습니다. 고정 이름과 별도의 데이터 인수를 사용하세요")
	Register("Command tool parameters must be declared in its schema", "Command tool parameters must be declared in its schema", "명령 도구 매개변수는 스키마에 선언되어 있어야 합니다")
	Register("Command parameters cannot be interpolated into nested shell or Python programs or their execution options; pass separate data arguments", "Command parameters cannot be interpolated into nested shell or Python programs or their execution options; pass separate data arguments", "명령 매개변수를 중첩 셸 또는 Python 프로그램이나 실행 옵션에 삽입할 수 없습니다. 별도의 데이터 인수로 전달하세요")
	Register("Command parameters cannot be interpolated into CMD programs; use literal native arguments or a script data channel", "Command parameters cannot be interpolated into CMD programs; use literal native arguments or a script data channel", "명령 매개변수를 CMD 프로그램에 삽입할 수 없습니다. 네이티브 리터럴 인수 또는 스크립트 데이터 채널을 사용하세요")
}
