package gitproviders

import (
	"fmt"

	"github.com/konflux-ci/e2e-tests/pkg/framework"
	gitlab "github.com/xanzy/go-gitlab"
)

// DeleteAllGitlabProjectWebhooks removes every hook on the GitLab project (project path or ID).
// Use in test cleanup when Konflux/PaC or other automation may leave multiple webhooks.
func DeleteAllGitlabProjectWebhooks(fw *framework.Framework, projectPath string) error {
	if fw == nil {
		return fmt.Errorf("framework is nil")
	}
	gl := fw.AsKubeAdmin.CommonController.Gitlab
	if gl == nil {
		return fmt.Errorf("GitLab client is nil")
	}
	if projectPath == "" {
		return fmt.Errorf("project path is empty")
	}
	c := gl.GetClient()
	for {
		hooks, _, err := c.Projects.ListProjectHooks(projectPath, &gitlab.ListProjectHooksOptions{PerPage: 100})
		if err != nil {
			return fmt.Errorf("list project hooks for %q: %w", projectPath, err)
		}
		if len(hooks) == 0 {
			return nil
		}
		for _, h := range hooks {
			if _, err := c.Projects.DeleteProjectHook(projectPath, h.ID); err != nil {
				return fmt.Errorf("delete hook id %d for %q: %w", h.ID, projectPath, err)
			}
		}
	}
}
