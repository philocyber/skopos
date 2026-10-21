package locale

func init() {
	Register("HTTP execution requires a concrete HTTP or HTTPS destination", "HTTP execution requires a concrete HTTP or HTTPS destination")
	Register("Execution destination policy could not be loaded: %w", "Execution destination policy could not be loaded: %w")
	Register("HTTP execution destination blocked by asset policy: %s", "HTTP execution destination blocked by asset policy: %s")
	Register("HTTP execution destination policy is unavailable", "HTTP execution destination policy is unavailable")
}
