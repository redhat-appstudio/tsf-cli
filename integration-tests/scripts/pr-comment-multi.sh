#!/usr/bin/env bash
set -euo pipefail

# pr-comment-multi.sh
# Posts one consolidated PR comment with a table of multiple child PipelineRun outcomes.
#
# Expected environment variables:
#   GITHUB_TOKEN, JOB_SPEC, TEST_NAME, PIPELINE_RUN_AGGREGATE_STATUS,
#   COMMENT_TAG, INCLUDE_SUCCESS_COMMENT, KONFLUX_UI_URL,
#   ARTIFACT_BROWSER_BASE_URL, KONFLUX_NAMESPACE, KONFLUX_APPLICATION_NAME
#
# Arguments: --summaries-b64 <base64-entry> [<base64-entry> ...]

parse_sectioned_args() {
  local section=""
  for arg in "$@"; do
    case "${arg}" in
      --summaries-b64)
        section="${arg}"
        continue
        ;;
    esac

    case "${section}" in
      --summaries-b64) summaries+=("${arg}") ;;
    esac
  done
}

normalize_status_label() {
  local raw_status="$1"
  case "${raw_status}" in
    Succeeded|succeeded|True|true) echo "Succeeded" ;;
    Failed|failed|False|false) echo "Failed" ;;
    Cancelled|cancelled|Canceled|canceled) echo "Cancelled" ;;
    *) echo "Unknown" ;;
  esac
}

decode_base64() {
  local encoded="$1"
  local decoded=""
  if decoded="$(printf "%s" "${encoded}" | base64 -d 2>/dev/null)"; then
    printf "%s" "${decoded}"
    return 0
  fi
  if decoded="$(printf "%s" "${encoded}" | base64 -D 2>/dev/null)"; then
    printf "%s" "${decoded}"
    return 0
  fi
  return 1
}

require_tools() {
  local tool=""
  for tool in "$@"; do
    if ! command -v "${tool}" >/dev/null 2>&1; then
      echo "[ERROR] Required tool '${tool}' is not available in the image."
      exit 1
    fi
  done
}

parse_job_spec() {
  PR_AUTHOR="$(echo "${JOB_SPEC}" | jq -r '.git.pull_request_author')"
  GIT_ORG="$(echo "${JOB_SPEC}" | jq -r '.git.org')"
  GIT_REPO="$(echo "${JOB_SPEC}" | jq -r '.git.repo')"
  PR_NUMBER="$(echo "${JOB_SPEC}" | jq -r '.git.pull_request_number')"
  REPO_NAME="$(echo "${JOB_SPEC}" | jq -r '.git.repo // empty')"
  export PR_AUTHOR GIT_ORG GIT_REPO PR_NUMBER REPO_NAME
}

authenticate_github() {
  if [[ -z "${GITHUB_TOKEN:-}" ]]; then
    echo "[ERROR] GITHUB_TOKEN environment variable is not set."
    exit 1
  fi

  USER_LOGIN="$(curl -s -H "Authorization: token ${GITHUB_TOKEN}" "https://api.github.com/user" | jq -r '.login')"
  if [[ -z "${USER_LOGIN}" || "${USER_LOGIN}" == "null" ]]; then
    echo "[ERROR] Unable to retrieve user login."
    exit 1
  fi
}

initialize_comment_marker() {
  MARKER="<!-- konflux-multi-e2e:${COMMENT_TAG} -->"
}

decode_child_summaries() {
  local max_len="${#summaries[@]}"
  if (( max_len == 0 )); then
    echo "[ERROR] No child pipeline data provided."
    exit 1
  fi

  total_runs_count=0
  failed_runs_count=0
  failed_rows=""

  for ((idx=0; idx<max_len; idx++)); do
    local run_name=""
    local summary_item_b64="${summaries[$idx]}"
    local summary_item
    if ! summary_item="$(decode_base64 "${summary_item_b64}")"; then
      echo "[WARN] Invalid base64 child summary at index ${idx}; skipping entry."
      continue
    fi
    if ! printf "%s" "${summary_item}" | jq -e . >/dev/null 2>&1; then
      echo "[WARN] Invalid decoded JSON child summary at index ${idx}; skipping entry."
      continue
    fi

    run_name="$(printf "%s" "${summary_item}" | jq -r '.run // empty')"
    local raw_status
    raw_status="$(printf "%s" "${summary_item}" | jq -r '.status // empty')"

    if [[ -z "${run_name}" || -z "${raw_status}" ]]; then
      echo "[WARN] Missing run/status in child summary at index ${idx}; skipping entry."
      continue
    fi

    local result_label
    result_label="$(normalize_status_label "${raw_status}")"
    if [[ "${result_label}" == "Unknown" ]]; then
      echo "[WARN] Unknown status '${raw_status}' for run ${run_name}."
    fi

    total_runs_count=$((total_runs_count + 1))

    local build_log_url test_log_url
    if [[ -z "${KONFLUX_UI_URL}" || -z "${KONFLUX_NAMESPACE}" || -z "${KONFLUX_APPLICATION_NAME}" ]]; then
      build_log_url="#"
      test_log_url="#"
    else
      build_log_url="${KONFLUX_UI_URL}/ns/${KONFLUX_NAMESPACE}/applications/${KONFLUX_APPLICATION_NAME}/pipelineruns/${run_name}/logs"
      test_log_url="${build_log_url}"
      if [[ -n "${ARTIFACT_BROWSER_BASE_URL}" && -n "${REPO_NAME}" ]]; then
        test_log_url="${ARTIFACT_BROWSER_BASE_URL%/}/${REPO_NAME}/${run_name}"
      fi
    fi
    if [[ "${result_label}" != "Succeeded" ]]; then
      failed_runs_count=$((failed_runs_count + 1))
      failed_rows="${failed_rows}| \`${run_name}\` | **${result_label}** | \`/retest\` | [View Pipeline Log](${build_log_url}) | [View Test Logs](${test_log_url}) |\n"
    fi
  done

  # Safety net: if all summaries failed to decode, force a degraded comment
  # instead of silently skipping.
  if (( total_runs_count == 0 && max_len > 0 )); then
    echo "[ERROR] All ${max_len} child pipeline summaries failed to decode. Posting degraded comment."
    failed_runs_count="${max_len}"
    total_runs_count="${max_len}"
    failed_rows="| (all summaries failed to decode) | **DecodeError** | \`/retest\` | # | # |\n"
  fi
}

build_comment_body() {
  local failed_details_section="No failed child pipelines."
  if (( failed_runs_count > 0 )); then
    failed_details_section="$(cat <<EOF
| PipelineRun Name | Status | Rerun command | Build Log | Test Log |
|------------------|--------|---------------|-----------|----------|
$(printf "%b" "${failed_rows}")
EOF
    )"
  fi

  # Always comment when there are failed nested runs, even if main pipeline is Succeeded.
  if (( failed_runs_count == 0 )) && [[ "${PIPELINE_RUN_AGGREGATE_STATUS}" == "Succeeded" && "${INCLUDE_SUCCESS_COMMENT}" != "true" ]]; then
    echo "[INFO] Aggregate status is Succeeded, no failed nested runs, and include-success-comment is false; skipping comment."
    exit 0
  fi

  PR_COMMENT="$(cat <<EOF
@${PR_AUTHOR}: The following matrix E2E test has ${PIPELINE_RUN_AGGREGATE_STATUS}, say **/retest** to rerun failed tests.

Total child pipelines: **${total_runs_count}**
Failed child pipelines: **${failed_runs_count}**

### Failed child pipelines
${failed_details_section}

${MARKER}
EOF
  )"
}

post_comment() {
  local request_data
  request_data="$(jq -n --arg body "${PR_COMMENT}" '{body: $body}')"
  local response
  response="$(curl -s -H "Authorization: token ${GITHUB_TOKEN}" -H "Content-Type: application/json" -X POST -d "${request_data}" "https://api.github.com/repos/${GIT_ORG}/${GIT_REPO}/issues/${PR_NUMBER}/comments")"

  if echo "${response}" | jq -e '.id' >/dev/null 2>&1; then
    echo "[INFO] Consolidated matrix comment posted successfully."
  else
    echo "[ERROR] Failed to post consolidated comment. Response: ${response}"
    exit 1
  fi
}

main() {
  require_tools jq curl base64

  export TEST_NAME PIPELINE_RUN_AGGREGATE_STATUS COMMENT_TAG INCLUDE_SUCCESS_COMMENT KONFLUX_UI_URL ARTIFACT_BROWSER_BASE_URL

  if [[ "${PIPELINE_RUN_AGGREGATE_STATUS}" == "None" ]]; then
    echo "[INFO] Aggregate status is 'None' (pipeline may have been cancelled before tasks ran). Skipping comment."
    exit 0
  fi

  parse_job_spec
  authenticate_github
  initialize_comment_marker

  summaries=()
  parse_sectioned_args "$@"

  decode_child_summaries
  build_comment_body
  post_comment
}

main "$@"
