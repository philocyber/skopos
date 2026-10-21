package locale

func init() {
	Register("HTTP execution requires a concrete HTTP or HTTPS destination", "HTTP execution requires a concrete HTTP or HTTPS destination", "HTTP 실행에는 구체적인 HTTP 또는 HTTPS 대상이 필요합니다")
	Register("Execution destination policy could not be loaded: %w", "Execution destination policy could not be loaded: %w", "실행 대상 정책을 불러오지 못했습니다: %w")
	Register("HTTP execution destination blocked by asset policy: %s", "HTTP execution destination blocked by asset policy: %s", "자산 정책에 의해 HTTP 실행 대상이 차단되었습니다: %s")
	Register("HTTP execution destination policy is unavailable", "HTTP execution destination policy is unavailable", "HTTP 실행 대상 정책을 사용할 수 없습니다")
}
