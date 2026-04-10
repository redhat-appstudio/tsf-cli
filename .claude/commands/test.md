# Run E2E Tests

Run the TSF e2e tests with proper setup validation, logging, and result analysis.

## Steps

### 1. Validate Prerequisites

Check that all required environment variables and tools are available. For each missing item, ask the user to provide it before proceeding.

**Required environment variables** (check if set and non-empty):
- `KUBECONFIG` - path to kubeconfig file. Also verify the file actually exists and that `oc whoami` (or `kubectl cluster-info`) succeeds against it.
- `GITHUB_TOKEN` - GitHub personal access token
- `MY_GITHUB_ORG` - GitHub organization for test repos
- `E2E_APPLICATIONS_NAMESPACE` - namespace for test applications (default: `default-tenant`)
- `TSF_APPLICATION_NAME` - application name (must match what `setup-release.sh` was called with)
- `TSF_COMPONENT_NAME` - component name (must match what `setup-release.sh` was called with)
- `TSF_MANAGED_NAMESPACE` - managed namespace (must match what `setup-release.sh` was called with)

**Optional environment variables** (inform the user these exist but don't require them):
- `E2E_SKIP_CLEANUP` - set to `true` to keep test resources after run
- `SKIP_PAC_TESTS` - set to `true` to skip the entire test
- `KLOG_VERBOSITY` - klog verbosity level (default: 1)

**Required tools** (verify these are on PATH):
- `go`
- `oc` or `kubectl`
- `jq`

If `my-test.env` exists in the `e2e/` directory, source it first and then validate. If it doesn't exist, check each variable individually and ask for missing ones.

If any required variable is missing, use AskUserQuestion to ask the user to provide it. Do NOT proceed until all required variables are set.

### 2. Set Up Release Resources

Before running the tests, the release infrastructure must be set up on the cluster. Run `setup-release.sh` to create the managed namespace, EnterpriseContractPolicy, ImageRepositories, ReleasePlanAdmission, and ReleasePlan:

```
bash scripts/setup-release.sh \
  -t "${E2E_APPLICATIONS_NAMESPACE}" \
  -m "${TSF_MANAGED_NAMESPACE}" \
  -a "${TSF_APPLICATION_NAME}" \
  -c "${TSF_COMPONENT_NAME}" \
  -r "tsf-release"
```

If the script fails, report the error to the user and stop. Do NOT proceed to run the tests without successful release setup.

### 3. Run the Tests

Change to the `e2e/` directory.

Create the `e2e/logs/` directory if it doesn't exist.

Generate a log filename with the format: `e2e-<YYYY-MM-DD_HH-MM-SS>.log` (ISO-ish sortable timestamp).

Run the tests using `make test` and `tee` the output to the log file:
```
cd e2e && GOWORK=off make test 2>&1 | tee logs/<logfile>
```

Use a generous timeout (10 minutes) since e2e tests take a while. Run this in the foreground so we can see output as it happens.

### 4. Analyze Results

After the tests complete, analyze the log file:

1. **Check exit code** - did the test run succeed or fail?
2. **Count passed/failed/skipped** - look for Ginkgo's summary output (e.g., `Ran X of Y Specs`, `X Passed`, `X Failed`, `X Pending`, `X Skipped`)
3. **If there are failures**, extract and summarize:
   - Which test cases (`It` blocks) failed
   - The failure messages and relevant error output
   - Any timeout-related failures (common in e2e)
4. **Report the log file path** so the user can review the full output

Present a concise summary of the test run to the user.
