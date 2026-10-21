package locale

func init() {
	Register("Traffic capture paused until WAL recovery succeeds", "Traffic capture paused until WAL recovery succeeds", "WAL 복구가 성공할 때까지 트래픽 수집이 일시 중지됩니다")
	Register("Traffic recovery could not acquire rollback journal", "Traffic recovery could not acquire rollback journal", "트래픽 복구가 롤백 저널을 확보하지 못했습니다")
	Register("Traffic recovery could not restore WAL", "Traffic recovery could not restore WAL", "트래픽 복구가 WAL을 복원하지 못했습니다")
	Register("Traffic WAL write estimate exceeds storage admission limit", "Traffic WAL write estimate exceeds storage admission limit", "트래픽 WAL 쓰기 예상 크기가 저장 허용 한도를 초과했습니다")
	Register("Traffic WAL storage admission blocked by an active snapshot", "Traffic WAL storage admission blocked by an active snapshot", "활성 스냅샷으로 인해 트래픽 WAL 저장 쓰기가 차단되었습니다")
	Register("%w (threshold=%d bytes)", "%w (threshold=%d bytes)", "%w (임계값=%d bytes)")
	Register("SKOPOS_TRAFFIC_WAL_BYTES must be at least 65536 bytes", "SKOPOS_TRAFFIC_WAL_BYTES must be at least 65536 bytes", "SKOPOS_TRAFFIC_WAL_BYTES는 최소 65536 bytes여야 합니다")
	Register("Traffic index storage limit reached", "Traffic index storage limit reached", "트래픽 인덱스 저장 한도에 도달했습니다")
	Register("%w (limit=%d bytes)", "%w (limit=%d bytes)", "%w (제한=%d bytes)")
	Register("SKOPOS_TRAFFIC_INDEX_BYTES must be at least 262144 bytes", "SKOPOS_TRAFFIC_INDEX_BYTES must be at least 262144 bytes", "SKOPOS_TRAFFIC_INDEX_BYTES는 최소 262144 bytes여야 합니다")
	Register("[traffic] Capture skipped; index storage admission failed: %v", "[traffic] Capture skipped; index storage admission failed: %v", "[traffic] 인덱스 저장 한도로 인해 수집을 건너뜁니다: %v")
	Register("Traffic blob storage limit reached", "Traffic blob storage limit reached", "트래픽 blob 저장 한도에 도달했습니다")
	Register("%w (limit=%d, used=%d bytes)", "%w (limit=%d, used=%d bytes)", "%w (제한=%d, 사용=%d bytes)")
	Register("SKOPOS_TRAFFIC_BLOB_BYTES must be a positive byte count", "SKOPOS_TRAFFIC_BLOB_BYTES must be a positive byte count", "SKOPOS_TRAFFIC_BLOB_BYTES는 양수 byte 수여야 합니다")
	Register("[traffic] Capture skipped; body storage failed: %v", "[traffic] Capture skipped; body storage failed: %v", "[traffic] 본문 저장에 실패하여 수집을 건너뜁니다: %v")
}
