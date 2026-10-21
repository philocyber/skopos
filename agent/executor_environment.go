package agent

import (
	"os"
	"strings"
)

// These server-owned credentials are not implicit task credentials. Ordinary
// tool environment, proxy routing and explicit TOOL_* data channels are retained.
var executorCredentialKeys = []string{
	"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN",
	"SKOPOS_PG_DSN", "SKOPOS_LLM_PROXY", "PGPASSWORD", "POSTGRES_PASSWORD",
}

// ExecutorEnvironmentOverrides masks inherited SDK credentials without changing
// the parent process. Caller overrides cannot restore a server credential key.
func ExecutorEnvironmentOverrides(extra []string) []string {
	blocked := make(map[string]bool, len(executorCredentialKeys))
	for _, key := range executorCredentialKeys {
		blocked[key] = true
	}
	filtered := make([]string, 0, len(extra)+len(blocked))
	for _, entry := range extra {
		key, _, _ := strings.Cut(entry, "=")
		if !blocked[strings.ToUpper(key)] {
			filtered = append(filtered, entry)
		}
	}
	for _, key := range executorCredentialKeys {
		filtered = append(filtered, key+"=")
	}
	return filtered
}

// ExecutorEnvironment supplies the same filtered environment to native custom
// processes. Empty credentials are intentional masks over inherited values.
func ExecutorEnvironment(extra []string) []string {
	return append(os.Environ(), ExecutorEnvironmentOverrides(extra)...)
}
