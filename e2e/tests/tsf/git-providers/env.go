package gitproviders

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/konflux-ci/e2e-tests/pkg/constants"
)

// Environment variable names for TSF demo e2e (in addition to konflux e2e-tests vars).
const (
	EnvGitProvider           = "GIT_PROVIDER"
	EnvMyGithubRepo          = "MY_GITHUB_REPO"
	EnvGithubSourceRevision  = "GITHUB_SOURCE_REVISION"
	EnvGitlabBaseURL         = "GITLAB_BASE_URL"
	EnvMyGitlabProjectPath   = "MY_GITLAB_PROJECT_PATH"
	EnvGitlabSourceRevision  = "GITLAB_SOURCE_REVISION"
	EnvGitlabDefaultBranch   = "GITLAB_DEFAULT_BRANCH"
	EnvGitlabTokenAlias      = "GITLAB_TOKEN" // copied to GITLAB_BOT_TOKEN before framework init if needed
	defaultGithubRepo        = "testrepo"
	defaultGitlabBaseURL     = "https://gitlab.com"
	defaultGitlabBranch      = "main"
	defaultGithubSrcRevision = "1255dc36534b9db7b99efbc281117435ea03255f"
)

// Kind selects which SCM API the demo test uses.
type Kind string

const (
	KindGitHub Kind = "github"
	KindGitLab Kind = "gitlab"
)

// ParseGitProvider returns KindGitHub when raw is empty.
// Accepted non-empty values: github, gitlab, gl (case-insensitive, UTF-8 BOM trimmed).
// Any other value returns an error so typos do not silently use GitHub while you expect GitLab.
func ParseGitProvider(raw string) (Kind, error) {
	s := strings.TrimPrefix(strings.TrimSpace(raw), "\ufeff")
	if s == "" {
		return KindGitHub, nil
	}
	switch strings.ToLower(s) {
	case string(KindGitHub):
		return KindGitHub, nil
	case string(KindGitLab), "gl":
		return KindGitLab, nil
	default:
		return KindGitHub, fmt.Errorf("%s=%q is invalid: use github, gitlab, or gl (omit for github)", EnvGitProvider, raw)
	}
}

// PrepareProcessEnvForGitLab sets konflux framework variables from TSF-friendly
// aliases before framework.NewFramework. Call only when Kind == KindGitLab.
func PrepareProcessEnvForGitLab() {
	if os.Getenv(constants.GITLAB_BOT_TOKEN_ENV) == "" {
		if t := strings.TrimSpace(os.Getenv(EnvGitlabTokenAlias)); t != "" {
			_ = os.Setenv(constants.GITLAB_BOT_TOKEN_ENV, t)
		}
	}
	if strings.TrimSpace(os.Getenv(constants.GITLAB_API_URL_ENV)) == "" {
		base := normalizedGitlabOriginURL(os.Getenv(EnvGitlabBaseURL))
		_ = os.Setenv(constants.GITLAB_API_URL_ENV, strings.TrimSuffix(base, "/")+"/api/v4")
	}
}

// normalizedGitlabOriginURL returns a URL with scheme suitable for joining paths (e.g. https://gitlab.com).
func normalizedGitlabOriginURL(raw string) string {
	base := strings.TrimSuffix(strings.TrimSpace(raw), "/")
	if base == "" {
		base = defaultGitlabBaseURL
	}
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	return base
}

// GitlabHostForSCMLabel returns the hostname from GITLAB_BASE_URL for appstudio.redhat.com/scm.host (must match clone URL host).
func GitlabHostForSCMLabel() (string, error) {
	u, err := url.Parse(normalizedGitlabOriginURL(os.Getenv(EnvGitlabBaseURL)))
	if err != nil {
		return "", fmt.Errorf("parse GITLAB_BASE_URL: %w", err)
	}
	h := u.Hostname()
	if h == "" {
		return "", fmt.Errorf("GITLAB_BASE_URL has no host (set e.g. https://gitlab.com)")
	}
	return h, nil
}

// GithubComponentURL returns https clone URL for org/repo.
func GithubComponentURL(org, repo string) string {
	org = strings.TrimSpace(org)
	repo = strings.TrimSpace(repo)
	return fmt.Sprintf("https://github.com/%s/%s", org, repo)
}

// GitlabComponentURL returns https clone URL for a project path on a GitLab host.
func GitlabComponentURL(baseURL, projectPath string) string {
	baseURL = strings.TrimSuffix(normalizedGitlabOriginURL(baseURL), "/")
	projectPath = strings.Trim(strings.TrimSpace(projectPath), "/")
	return fmt.Sprintf("%s/%s.git", baseURL, projectPath)
}

// RequiredGithubRepo returns MY_GITHUB_REPO or default testrepo.
func RequiredGithubRepo() string {
	if r := strings.TrimSpace(os.Getenv(EnvMyGithubRepo)); r != "" {
		return r
	}
	return defaultGithubRepo
}

// GithubSourceRevision returns GITHUB_SOURCE_REVISION or the historical default SHA.
func GithubSourceRevision() string {
	if r := strings.TrimSpace(os.Getenv(EnvGithubSourceRevision)); r != "" {
		return r
	}
	return defaultGithubSrcRevision
}

// GitlabBaseURL returns GITLAB_BASE_URL or https://gitlab.com (trimmed, no trailing slash).
func GitlabBaseURL() string {
	return strings.TrimSuffix(normalizedGitlabOriginURL(os.Getenv(EnvGitlabBaseURL)), "/")
}

// RequiredGitlabProjectPath returns MY_GITLAB_PROJECT_PATH (e.g. group/sub/repo).
func RequiredGitlabProjectPath() (string, error) {
	p := strings.Trim(strings.TrimSpace(os.Getenv(EnvMyGitlabProjectPath)), "/")
	if p == "" {
		return "", fmt.Errorf("%s is not set", EnvMyGitlabProjectPath)
	}
	return p, nil
}

// GitlabSourceRevision returns GITLAB_SOURCE_REVISION, or empty to mean "tip of the default branch".
// Empty is recommended when the GitHub default SHA (GITHUB_SOURCE_REVISION) is not present in your
// GitLab project; konflux CreateGitlabNewBranch then resolves the latest commit on the default branch.
func GitlabSourceRevision() string {
	return strings.TrimSpace(os.Getenv(EnvGitlabSourceRevision))
}

// GitlabDefaultBranch returns GITLAB_DEFAULT_BRANCH or "main".
func GitlabDefaultBranch() string {
	if b := strings.TrimSpace(os.Getenv(EnvGitlabDefaultBranch)); b != "" {
		return b
	}
	return defaultGitlabBranch
}
