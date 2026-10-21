package locale

func init() {
	Register("Review context is incomplete; human confirmation is required: ", "Review context is incomplete; human confirmation is required: ", "검토 문맥이 불완전하여 사용자 확인이 필요합니다: ")
	Register("Model review failed; applying failure policy: ", "Model review failed; applying failure policy: ", "모델 검토 실패. 실패 정책 적용: ")
	Register("Model output could not be parsed; applying failure policy", "Model output could not be parsed; applying failure policy", "모델 출력을 분석할 수 없어 실패 정책을 적용합니다")
	Register("[Model] ", "[Model] ", "[모델] ")
	Register("[Model]", "[Model]", "[모델]")
	Register("Allow", "Allow", "허용")
	Register("Block", "Block", "차단")
	Register("Request human approval", "Request human approval", "사용자 승인 요청")
	Register("Interception rule [", "Interception rule [", "차단 규칙 [")
	Register("] prohibits this tool call", "] prohibits this tool call", "]이(가) 이 도구 호출을 금지합니다")
	Register("] requires user approval; please wait", "] requires user approval; please wait", "]에 사용자 승인이 필요합니다. 기다려 주세요")
	Register("Tool %s requests approval (#%d)", "Tool %s requests approval (#%d)", "도구 %s 승인 요청(#%d)")
	Register("Work was cancelled", "Work was cancelled", "작업이 취소되었습니다")
	Register("Work was cancelled before execution", "Work was cancelled before execution", "실행 전에 작업이 취소되었습니다")
	Register("Approval timed out; applying timeout policy", "Approval timed out; applying timeout policy", "승인 시간이 초과되어 시간 초과 정책을 적용합니다")
	Register("Approval was already resolved or does not exist; refresh the record", "Approval was already resolved or does not exist; refresh the record", "승인이 이미 처리되었거나 없습니다. 기록을 새로 고치세요")
	Register("Human denied execution", "Human denied execution", "사용자가 실행을 거부했습니다")
	Register("Human allowed execution", "Human allowed execution", "사용자가 실행을 허용했습니다")
	Register("Tool arguments are not valid JSON", "Tool arguments are not valid JSON", "도구 인수가 유효한 JSON이 아닙니다")
	Register("Execution ended without a tool result", "Execution ended without a tool result", "도구 결과 없이 실행이 종료되었습니다")
}
