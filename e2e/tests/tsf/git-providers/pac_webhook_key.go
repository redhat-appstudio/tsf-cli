package gitproviders

import (
	"regexp"
	"strings"
)

// Matches konflux-ci build-service getWebhookSecretKeyForComponent (each disallowed char -> '_').
var konfluxPaCWebhookSecretKeySanitizer = regexp.MustCompile(`[^-._a-zA-Z0-9]`)

// KonfluxPACWebhookSecretDataKey returns the Secret.Data key under pipelines-as-code-webhooks-secret
// for a component GitSource.URL (trim trailing ".git", then sanitize).
func KonfluxPACWebhookSecretDataKey(gitRepoURL string) string {
	s := strings.TrimSuffix(strings.TrimSpace(gitRepoURL), ".git")
	return konfluxPaCWebhookSecretKeySanitizer.ReplaceAllString(s, "_")
}
