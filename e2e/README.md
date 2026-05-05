# TSF E2E Tests

End-to-end test suite for TSF (Trusted Software Factory) instances.

## Prerequisites

- **KUBECONFIG** pointing to a cluster with TSF installed
- **Git hosting** for the component under test (set **`GIT_PROVIDER`** to **`gitlab`** or **`gl`** for GitLab; use **`github`** or leave unset for GitHub — unknown values **fail fast** so you do not silently run the GitHub path):
  - **GitHub** (default): org with a fork or clone of [konflux-ci/testrepo](https://github.com/konflux-ci/testrepo), or your own repo; set `MY_GITHUB_ORG` and optionally `MY_GITHUB_REPO` (defaults to `testrepo`).
  - **GitLab**: set `GIT_PROVIDER=gitlab`, `MY_GITLAB_PROJECT_PATH` (URL path only, e.g. `group/subgroup/repo` on GitLab.com, not a numeric project id), and `GITLAB_BOT_TOKEN` or `GITLAB_TOKEN`. Optional `GITLAB_DEFAULT_BRANCH` defaults to `main`. **`GITLAB_SOURCE_REVISION` is optional**: if unset, the test creates the base branch from the **latest commit on the default branch** (use this when your GitLab copy does not contain the same commit SHA as the GitHub `GITHUB_SOURCE_REVISION` default). If set, it must be a commit that exists in the GitLab project. Optional `GITLAB_BASE_URL` defaults to `https://gitlab.com`; if `GITLAB_API_URL` is unset, the test derives `GITLAB_BASE_URL/api/v4` before the framework starts (override `GITLAB_API_URL` if your instance needs a different API base).
- **Token** with permission to manage the chosen repo (GitHub: `GITHUB_TOKEN`; GitLab: see above). For **GitLab PaC**, the test creates **`pipelines-as-code-secret`** ( **`password`** + **`provider.token`** only) in the applications namespace when absent; after the **component** exists, Konflux writes **`pipelines-as-code-webhooks-secret`**, and the test copies **`webhook.secret`** from the data key **`KonfluxPACWebhookSecretDataKey(GitSource.URL)`** (same rule as konflux build-service, e.g. `https___gitlab.com_group_repo` for `https://gitlab.com/group/repo.git`). The GitLab project hook token is **only** that value—no fallbacks. Ensure `GIT_PROVIDER=gitlab` is set—timeouts mentioning only short repo name `testrepo` usually mean the **GitHub** code path (default provider), not GitLab.
- **Quay org** accessible by the TSF instance's image-controller

## IDE Setup

The `e2e/` module is separate from the main CLI module. To get full IDE
navigation across both modules, create a `go.work` file in the repo root:

```
go work init . ./e2e
```

> **Note:** With `go.work` present, CLI builds will fail due to transitive
> dependency conflicts. Use `GOWORK=off make build` to build the CLI, or
> remove `go.work` when you don't need cross-module IDE support.

## Running tests

1. Build the test binary:
   ```
   make build
   ```

2. Set up your env file:
   ```
   cp my-test.env.template my-test.env
   # edit my-test.env and fill in the values
   ```

3. Source the env file and run:
   ```
   source my-test.env
   ./bin/tsf.test --ginkgo.v --ginkgo.label-filter="tsf-demo"
   ```
