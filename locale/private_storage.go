package locale

func init() {
	Register("Legacy signing key conflicts with the active key; resolve the obsolete workspace key before starting", "Legacy signing key conflicts with the active key; resolve the obsolete workspace key before starting", "이전 서명 키가 활성 키와 충돌합니다. 시작하기 전에 작업 공간의 이전 키를 확인하세요")
	Register("Cannot load private configuration file %s; check file permissions and JSON format", "Cannot load private configuration file %s; check file permissions and JSON format", "비공개 설정 파일 %s을(를) 불러올 수 없습니다. 파일 권한과 JSON 형식을 확인하세요")
	Register("Local signing key is invalid; refusing to replace it", "Local signing key is invalid; refusing to replace it", "로컬 서명 키가 유효하지 않아 교체하지 않습니다")
	Register("Private storage must not contain symbolic links: %s", "Private storage must not contain symbolic links: %s", "비공개 저장소에는 심볼릭 링크를 사용할 수 없습니다: %s")
	Register("Private storage contains unsupported file type: %s", "Private storage contains unsupported file type: %s", "비공개 저장소에 지원하지 않는 파일 형식이 있습니다: %s")
}
