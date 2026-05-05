package gitproviders

import (
	"fmt"
	"strings"

	"github.com/konflux-ci/e2e-tests/pkg/framework"
	corev1 "k8s.io/api/core/v1"
)

const pacSecretKeyProviderToken = "provider.token"
const pacSecretKeyWebhookSecret = "webhook.secret"
const pacSecretNameDefault = "pipelines-as-code-secret"

// Konflux build-service stores per-repo webhook material here (see pipelinesAsCodeWebhooksSecretName).
const pacWebhooksSecretName = "pipelines-as-code-webhooks-secret"

// CreateGitlabPACSecret creates pipelines-as-code SCM credentials for GitLab (password + provider.token only).
// webhook.secret is synced later via EnsureGitlabPACSecretShape once Konflux has written pipelines-as-code-webhooks-secret.
func CreateGitlabPACSecret(f *framework.Framework, secretName, token string) error {
	if token == "" {
		return fmt.Errorf("GitLab token is empty")
	}
	host, err := GitlabHostForSCMLabel()
	if err != nil {
		return err
	}
	buildSecret := corev1.Secret{}
	buildSecret.Name = secretName
	buildSecret.Labels = map[string]string{
		"appstudio.redhat.com/credentials": "scm",
		"appstudio.redhat.com/scm.host":    host,
	}
	buildSecret.Type = corev1.SecretTypeBasicAuth
	buildSecret.StringData = map[string]string{
		"password":                token,
		pacSecretKeyProviderToken: token,
	}
	_, err = f.AsKubeAdmin.CommonController.CreateSecret(f.UserNamespace, &buildSecret)
	if err != nil {
		return fmt.Errorf("create GitLab PAC secret: %w", err)
	}
	return nil
}

// KonfluxPACWebhookSecretValue returns the only supported webhook HMAC secret for GitLab hooks:
// tenantNamespace/pipelines-as-code-webhooks-secret, data key = KonfluxPACWebhookSecretDataKey(gitRepoURL).
func KonfluxPACWebhookSecretValue(f *framework.Framework, tenantNamespace, gitRepoURL string) (string, error) {
	if f == nil {
		return "", fmt.Errorf("framework is nil")
	}
	if strings.TrimSpace(tenantNamespace) == "" {
		return "", fmt.Errorf("tenant namespace is empty")
	}
	key := KonfluxPACWebhookSecretDataKey(gitRepoURL)
	if key == "" {
		return "", fmt.Errorf("empty webhook secret data key for git URL %q", gitRepoURL)
	}
	sec, err := f.AsKubeAdmin.CommonController.GetSecret(tenantNamespace, pacWebhooksSecretName)
	if err != nil {
		return "", fmt.Errorf("get Secret %s/%s: %w", tenantNamespace, pacWebhooksSecretName, err)
	}
	raw, ok := sec.Data[key]
	if !ok || len(raw) == 0 {
		return "", fmt.Errorf("secret %s/%s has no non-empty data key %q (expected Konflux key for GitSource.URL %q)",
			tenantNamespace, pacWebhooksSecretName, key, gitRepoURL)
	}
	return string(raw), nil
}

// EnsureGitlabPACSecretShape updates pipelines-as-code-secret: webhook.secret is always synced from
// tenant pipelines-as-code-webhooks-secret (KonfluxPACWebhookSecretDataKey(gitRepoURL)); provider.token and password are set if missing.
func EnsureGitlabPACSecretShape(f *framework.Framework, secretName, apiToken, gitRepoURL string) error {
	if secretName == "" {
		secretName = pacSecretNameDefault
	}
	if apiToken == "" {
		return fmt.Errorf("GitLab token is empty")
	}
	if strings.TrimSpace(gitRepoURL) == "" {
		return fmt.Errorf("gitRepoURL is empty")
	}
	sec, err := f.AsKubeAdmin.CommonController.GetSecret(f.UserNamespace, secretName)
	if err != nil {
		return err
	}
	if sec.Data == nil {
		sec.Data = map[string][]byte{}
	}
	changed := false
	wh, err := KonfluxPACWebhookSecretValue(f, f.UserNamespace, gitRepoURL)
	if err != nil {
		return err
	}
	if string(sec.Data[pacSecretKeyWebhookSecret]) != wh {
		sec.Data[pacSecretKeyWebhookSecret] = []byte(wh)
		changed = true
	}
	if len(sec.Data[pacSecretKeyProviderToken]) == 0 {
		sec.Data[pacSecretKeyProviderToken] = []byte(apiToken)
		changed = true
	}
	if len(sec.Data["password"]) == 0 {
		sec.Data["password"] = []byte(apiToken)
		changed = true
	}
	if !changed {
		return nil
	}
	_, err = f.AsKubeAdmin.CommonController.UpdateSecret(f.UserNamespace, sec)
	if err != nil {
		return fmt.Errorf("update GitLab PAC secret %q: %w", secretName, err)
	}
	return nil
}
