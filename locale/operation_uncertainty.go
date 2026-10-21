package locale

func init() {
	Register("Shell execution is not confined by input rules; human approval is required", "Shell execution is not confined by input rules; human approval is required", "입력 규칙은 셸 실행을 격리하지 않습니다. 사람의 승인이 필요합니다")
	Register("The shell program cannot be inspected before execution; human approval is required", "The shell program cannot be inspected before execution; human approval is required", "실행 전에 셸 프로그램을 검사할 수 없으므로 사람의 승인이 필요합니다")
	Register("Request method depends on runtime arguments or configuration; human approval is required", "Request method depends on runtime arguments or configuration; human approval is required", "요청 메서드가 런타임 인수나 설정에 따라 달라지므로 사람의 승인이 필요합니다")
	Register("The command selects its executable at runtime; human approval is required", "The command selects its executable at runtime; human approval is required", "명령이 실행 시 실행 파일을 선택하므로 사람의 승인이 필요합니다")
}
