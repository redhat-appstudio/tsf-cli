package gitproviders

import (
	"fmt"

	"github.com/konflux-ci/e2e-tests/pkg/framework"
)

// Provider abstracts GitHub PR and GitLab MR flows used by the TSF demo test.
type Provider interface {
	// CreateBaseBranch points newBranch at gitRevision (commit SHA) from defaultBranch context.
	CreateBaseBranch(defaultBranch, gitRevision, newBranch string) error
	// WaitPaCInitChange waits until PaC creates an open PR/MR from pacBranch; returns API id and head SHA.
	WaitPaCInitChange(pacBranch string) (changeID int, headSHA string, err error)
	// MergePaCChange merges PR/MR changeID; returns merge commit SHA.
	MergePaCChange(changeID int) (mergeSHA string, err error)
	// DeleteRemoteBranch removes a branch/ref; missing branch is not an error.
	DeleteRemoteBranch(branch string) error
	// CleanupClusterWebhooks removes hooks whose URL contains the cluster app domain.
	CleanupClusterWebhooks() error
}

// New returns a Provider for the given kind. repoKey is GitHub short repo name or GitLab path (group/project).
func New(kind Kind, fw *framework.Framework, repoKey string) (Provider, error) {
	switch kind {
	case KindGitHub:
		return &githubProvider{fw: fw, repoName: repoKey}, nil
	case KindGitLab:
		return &gitlabProvider{fw: fw, projectPath: repoKey}, nil
	default:
		return nil, fmt.Errorf("unsupported git provider kind %q", kind)
	}
}
