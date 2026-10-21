package locale

func init() {
	Register("Traffic write would consume reserved free space", "Traffic write would consume reserved free space", "트래픽 쓰기가 예약된 여유 공간을 소진할 수 있습니다")
	Register("%w (reserve=%d bytes)", "%w (reserve=%d bytes)", "%w (예약=%d bytes)")
	Register("SKOPOS_TRAFFIC_FREE_BYTES must be a positive byte count", "SKOPOS_TRAFFIC_FREE_BYTES must be a positive byte count", "SKOPOS_TRAFFIC_FREE_BYTES는 양수 byte 수여야 합니다")
	Register("Traffic free-space accounting unavailable on this platform", "Traffic free-space accounting unavailable on this platform", "이 플랫폼에서 트래픽 여유 공간을 계산할 수 없습니다")
}
