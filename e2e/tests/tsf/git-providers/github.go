package gitproviders

import (
	"fmt"
	"strings"

	"github.com/konflux-ci/e2e-tests/pkg/framework"
	"github.com/konflux-ci/e2e-tests/pkg/utils/build"
)

type githubProvider struct {
	fw       *framework.Framework
	repoName string
}

func (g *githubProvider) CreateBaseBranch(defaultBranch, gitRevision, newBranch string) error {
	return g.fw.AsKubeAdmin.CommonController.Github.CreateRef(g.repoName, defaultBranch, gitRevision, newBranch)
}

func (g *githubProvider) WaitPaCInitChange(pacBranch string) (int, string, error) {
	prs, err := g.fw.AsKubeAdmin.CommonController.Github.ListPullRequests(g.repoName)
	if err != nil {
		return 0, "", err
	}
	for _, pr := range prs {
		if pr.Head.GetRef() == pacBranch {
			return pr.GetNumber(), pr.GetHead().GetSHA(), nil
		}
	}
	return 0, "", fmt.Errorf("could not get the expected PaC branch name %s", pacBranch)
}

func (g *githubProvider) MergePaCChange(changeID int) (string, error) {
	res, err := g.fw.AsKubeAdmin.CommonController.Github.MergePullRequest(g.repoName, changeID)
	if err != nil {
		return "", err
	}
	return res.GetSHA(), nil
}

func (g *githubProvider) DeleteRemoteBranch(branch string) error {
	err := g.fw.AsKubeAdmin.CommonController.Github.DeleteRef(g.repoName, branch)
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "Reference does not exist") {
		return nil
	}
	return err
}

func (g *githubProvider) CleanupClusterWebhooks() error {
	return build.CleanupWebhooks(g.fw, g.repoName)
}
