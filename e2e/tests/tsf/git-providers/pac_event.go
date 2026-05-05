package gitproviders

// PaCEventTypePullRequest returns the value of label pipelinesascode.tekton.dev/event-type
// for a PaC PipelineRun triggered by a PR/MR (after kubernetes label cleaning).
func PaCEventTypePullRequest(k Kind) string {
	if k == KindGitLab {
		// PaC GitLab: X-Gitlab-Event "Merge Request Hook" → "Merge Request" → formatting.CleanValueKubernetes → "Merge_Request"
		return "Merge_Request"
	}
	return "pull_request"
}

// PaCEventTypePush returns the value of label pipelinesascode.tekton.dev/event-type for a push build.
// PaC sets lowercase "push" on labels for both GitHub and GitLab (see pipelinesascode.tekton.dev/event-type).
func PaCEventTypePush(_ Kind) string {
	return "push"
}
