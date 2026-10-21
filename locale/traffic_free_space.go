package locale

func init() {
	Register("Traffic write would consume reserved free space", "Traffic write would consume reserved free space")
	Register("%w (reserve=%d bytes)", "%w (reserve=%d bytes)")
	Register("SKOPOS_TRAFFIC_FREE_BYTES must be a positive byte count", "SKOPOS_TRAFFIC_FREE_BYTES must be a positive byte count")
	Register("Traffic free-space accounting unavailable on this platform", "Traffic free-space accounting unavailable on this platform")
}
