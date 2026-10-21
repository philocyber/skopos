package locale

func init() {
	Register("[traffic] Capture skipped; cannot generate exchange ID: %v", "[traffic] Capture skipped; cannot generate exchange ID: %v", "[traffic] 수집 생략; 교환 ID 생성 실패: %v")
	Register("Traffic capture exceeds exchange or host limit", "Traffic capture exceeds exchange or host limit", "트래픽 수집이 교환 또는 호스트 제한을 초과합니다")
	Register("Traffic retained exchange limit reached", "Traffic retained exchange limit reached", "저장된 트래픽 교환 수 한도에 도달했습니다")
	Register("SKOPOS_TRAFFIC_MAX_EXCHANGES must be a positive exchange count", "SKOPOS_TRAFFIC_MAX_EXCHANGES must be a positive exchange count", "SKOPOS_TRAFFIC_MAX_EXCHANGES는 양수 교환 수여야 합니다")
	Register("%s must be a positive byte count", "%s must be a positive byte count", "%s는 양수 byte 수여야 합니다")
	Register("[traffic] Capture skipped; capture quota admission failed: %v", "[traffic] Capture skipped; capture quota admission failed: %v", "[traffic] 수집 생략; 수집 할당량 승인 실패: %v")
}
