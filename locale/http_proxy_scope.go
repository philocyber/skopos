package locale

func init() {
	Register("Recording proxy destination binding is unavailable", "Recording proxy destination binding is unavailable", "기록 프록시 대상 연결 바인딩을 사용할 수 없습니다")
	Register("HTTP proxy tunnel could not be established", "HTTP proxy tunnel could not be established", "HTTP 프록시 터널을 설정할 수 없습니다")
	Register("HTTP proxy target resolution failed", "HTTP proxy target resolution failed", "HTTP 프록시 대상 해석에 실패했습니다")
	Register("HTTP proxy target resolved to no approved addresses", "HTTP proxy target resolved to no approved addresses", "HTTP 프록시 대상에서 승인된 주소를 찾지 못했습니다")
	Register("Proxy-side DNS cannot enforce IP asset restrictions; use a direct request or a literal IP destination", "Proxy-side DNS cannot enforce IP asset restrictions; use a direct request or a literal IP destination", "프록시 측 DNS는 IP 자산 제한을 적용할 수 없습니다. 직접 요청하거나 IP 주소 목적지를 사용하세요")
}
