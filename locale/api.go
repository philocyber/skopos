package locale

func init() {
	Register("Interception policy is unavailable; tool execution is blocked", "Interception policy is unavailable; tool execution is blocked", "차단 정책을 불러올 수 없어 도구 실행이 차단되었습니다")
	Register("[tools] Could not read tool registry; tools unavailable: %v", "[tools] Could not read tool registry; tools unavailable: %v", "[tools] 도구 레지스트리를 읽을 수 없어 도구를 사용할 수 없습니다: %v")
	Register("A valid local setup token is required", "A valid local setup token is required", "유효한 로컬 설정 토큰이 필요합니다")
	Register("Too many authentication requests; try again shortly", "Too many authentication requests; try again shortly", "인증 요청이 너무 많습니다. 잠시 후 다시 시도해 주세요")
	Register("language must be en or ko", "language must be en or ko", "언어는 en 또는 ko여야 합니다")
	Register("task not found", "task not found", "작업을 찾을 수 없습니다")
	Register("no active task", "no active task", "활성 작업이 없습니다")
}

func init() {
	Register("Unauthorized", "Unauthorized", "인증이 필요합니다")
	Register("Invalid or expired token", "Invalid or expired token", "토큰이 잘못되었거나 만료되었습니다")
	Register("Password is already set", "Password is already set", "암호가 이미 설정되었습니다")
	Register("Password cannot be empty", "Password cannot be empty", "암호는 비워 둘 수 없습니다")
	Register("Password hashing failed", "Password hashing failed", "암호 해시 생성 실패")
	Register("Save failed: ", "Save failed: ", "저장 실패: ")
	Register("Token generation failed", "Token generation failed", "토큰 생성 실패")
	Register("Invalid request format", "Invalid request format", "요청 형식이 잘못되었습니다")
	Register("Authentication request is too large", "Authentication request is too large", "인증 요청이 너무 큽니다")
	Register("Authentication request timed out", "Authentication request timed out", "인증 요청 시간이 초과되었습니다")
	Register("Authentication request unavailable", "Authentication request unavailable", "인증 요청을 처리할 수 없습니다")
	Register("Password must be at most 72 bytes", "Password must be at most 72 bytes", "암호는 72바이트 이하여야 합니다")
	Register("New password cannot be empty", "New password cannot be empty", "새 암호는 비워 둘 수 없습니다")
	Register("Password is not initialized; set it first", "Password is not initialized; set it first", "암호가 초기화되지 않았습니다. 먼저 설정하세요")
	Register("Current password is incorrect", "Current password is incorrect", "현재 암호가 잘못되었습니다")
	Register("Incorrect username or password", "Incorrect username or password", "사용자 이름 또는 암호가 잘못되었습니다")
	Register("[auth] JWT key migrated from %s to %s (outside the browsable workspace)", "[auth] JWT key migrated from %s to %s (outside the browsable workspace)", "[auth] JWT 키를 %s에서 %s(탐색 가능한 작업 공간 밖)(으)로 이전했습니다")
	Register("[auth] New JWT key written to %s", "[auth] New JWT key written to %s", "[auth] 새 JWT 키를 %s에 저장했습니다")
}

func init() {
	Register("Preparing…", "Preparing…", "준비 중…")
	Register("Update failed", "Update failed", "업데이트 실패")
	Register("New version ready; restarting…", "New version ready; restarting…", "새 버전 준비 완료. 재시작 중…")
	Register("Current version %q is not a stable release; one-click updates are disabled", "Current version %q is not a stable release; one-click updates are disabled", "현재 버전 %q은(는) 정식 릴리스가 아니므로 원클릭 업데이트가 비활성화되었습니다")
	Register("Already running the latest version %s", "Already running the latest version %s", "이미 최신 버전 %s을(를) 실행 중입니다")
	Register("An update is already in progress", "An update is already in progress", "업데이트가 이미 진행 중입니다")
	Register("[update] Update failed: %v", "[update] Update failed: %v", "[update] 업데이트 실패: %v")
	Register("[update] %s -> %s staged; exiting to complete replacement", "[update] %s -> %s staged; exiting to complete replacement", "[update] %s -> %s 준비 완료. 교체를 완료하기 위해 종료합니다")
	Register("Cannot roll back while an update is in progress", "Cannot roll back while an update is in progress", "업데이트 진행 중에는 롤백할 수 없습니다")
	Register("[update] Manually rolled back to the previous version; exiting to complete the switch", "[update] Manually rolled back to the previous version; exiting to complete the switch", "[update] 이전 버전으로 수동 롤백했습니다. 전환을 완료하기 위해 종료합니다")
}
