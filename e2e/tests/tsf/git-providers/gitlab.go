package gitproviders

import (
	"fmt"
	"strings"

	"github.com/konflux-ci/e2e-tests/pkg/framework"
)

type gitlabProvider struct {
	fw          *framework.Framework
	projectPath string
}

func (g *gitlabProvider) CreateBaseBranch(defaultBranch, gitRevision, newBranch string) error {
	return g.fw.AsKubeAdmin.CommonController.Gitlab.CreateGitlabNewBranch(g.projectPath, newBranch, gitRevision, defaultBranch)
}

func (g *gitlabProvider) WaitPaCInitChange(pacBranch string) (int, string, error) {
	mrs, err := g.fw.AsKubeAdmin.CommonController.Gitlab.GetMergeRequests(g.projectPath)
	if err != nil {
		return 0, "", err
	}
	for _, mr := range mrs {
		if mr.SourceBranch == pacBranch {
			return mr.IID, mr.SHA, nil
		}
	}
	branches := make([]string, 0, len(mrs))
	for _, mr := range mrs {
		branches = append(branches, mr.SourceBranch)
	}
	return 0, "", fmt.Errorf("no open MR with source branch %q (found %d open MR(s), source branches: %v); "+
		"if this stays empty, PaC is not pushing to GitLab—check pipelines-as-code-secret, token scopes, and GitLab integration",
		pacBranch, len(mrs), branches)
}

func (g *gitlabProvider) MergePaCChange(changeID int) (string, error) {
	mr, err := g.fw.AsKubeAdmin.CommonController.Gitlab.AcceptMergeRequest(g.projectPath, changeID)
	if err != nil {
		return "", err
	}
	if mr == nil {
		return "", fmt.Errorf("AcceptMergeRequest returned nil merge request")
	}
	// PaC labels push build PipelineRuns with the merge commit on the target branch (same as GitHub merge API SHA).
	// mr.SHA is the MR source-branch / diff HEAD and matches the init on-pull-request run—do not use it here.
	if mr.MergeCommitSHA != "" {
		return mr.MergeCommitSHA, nil
	}
	if mr.SquashCommitSHA != "" {
		return mr.SquashCommitSHA, nil
	}
	return mr.SHA, nil
}

func (g *gitlabProvider) DeleteRemoteBranch(branch string) error {
	err := g.fw.AsKubeAdmin.CommonController.Gitlab.DeleteBranch(g.projectPath, branch)
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "404") || strings.Contains(msg, "not found") {
		return nil
	}
	return err
}

func (g *gitlabProvider) CleanupClusterWebhooks() error {
	return g.fw.AsKubeAdmin.CommonController.Gitlab.DeleteWebhooks(g.projectPath, g.fw.ClusterAppDomain)
}
