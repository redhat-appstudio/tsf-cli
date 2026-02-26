#!/usr/bin/env bash
#
# Installs TSF on an OpenShift cluster. Assumes the TSF container image is
# already built and this script is running inside the container.
#
# Required environment variables:
#
#   From hack/private.env.template (copy to hack/private.env and fill in):
#     GITHUB__ORG         - GitHub organization name
#     QUAY__ORG           - Quay.io organization name
#     QUAY__API_TOKEN     - Quay.io API token
#     QUAY__URL           - Quay.io registry URL (e.g. https://quay.io)
#
#   Additional (from your GitHub App, not in the template):
#     GITHUB__APP_ID             - GitHub App numeric ID
#     GITHUB__APP_PRIVATE_KEY    - GitHub App private key (base64-encoded PEM)
#     GITHUB__APP_WEBHOOK_SECRET - GitHub App webhook secret
#

shopt -s inherit_errexit
set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

usage() {
    cat >&2 <<EOF
Usage: ${0##*/} [OPTIONS]

  Install and configure TSF on an OpenShift cluster.

Workflow:
  1. Validate required environment variables
  2. Validate existing OpenShift authentication context
  3. Create TSF configuration (tsf-config ConfigMap)
  4. Create GitHub App integration secret(Here we use the existing GitHub App)
  5. Register Quay.io integration
  6. Deploy TSF services
  7. Report deployed resource status

Required environment variables:

  From hack/private.env.template (copy to hack/private.env and fill in):
    GITHUB__ORG                  GitHub organization name
    QUAY__ORG                    Quay.io organization name
    QUAY__API_TOKEN              Quay.io API token
    QUAY__URL                    Quay.io registry URL (e.g. https://quay.io)

  Additional (from your GitHub App, not in the template):
    GITHUB__APP_ID               GitHub App numeric ID
    GITHUB__APP_PRIVATE_KEY      GitHub App private key (base64-encoded PEM)
    GITHUB__APP_WEBHOOK_SECRET   GitHub App webhook secret

  To base64-encode the private key:
    export GITHUB__APP_PRIVATE_KEY=\$(base64 < /path/to/app.private-key.pem)

Options:
  -d, --debug    Activate tracing/debug mode
  -h, --help     Display this message

Examples:
  ${0##*/}
  ${0##*/} --debug
EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
        -d | --debug)
            set -x
            DEBUG="--debug"
            export DEBUG
            info "Running script as: $(id)"
            ;;
        -h | --help)
            usage
            exit 0
            ;;
        *)
            fail "Unsupported argument: '$1'."
            ;;
        esac
        shift
    done
}

fail() {
    echo "# [ERROR] ${*}" >&2
    exit 1
}

info() {
    echo "# [INFO] ${*}"
}

#
# Functions
#

# Validates that all required environment variables are set and non-empty.
validate_env() {
    declare -r -a REQUIRED_VARS=(
        GITHUB__ORG
        GITHUB__APP_ID
        GITHUB__APP_PRIVATE_KEY
        GITHUB__APP_WEBHOOK_SECRET
        QUAY__ORG
        QUAY__API_TOKEN
        QUAY__URL
    )

    local missing=0
    for var in "${REQUIRED_VARS[@]}"; do
        if [[ -z "${!var:-}" ]]; then
            echo "# [ERROR] Required environment variable '${var}' is not set." >&2
            missing=1
        fi
    done

    if [[ "${missing}" -ne 0 ]]; then
        fail "Missing required environment variables. See hack/private.env.template."
    fi

    info "All required environment variables are set."
}

# Validates that we already have an authenticated OpenShift session.
ensure_ocp_auth() {
    info "Validating existing OpenShift authentication context..."
    if ! oc whoami &>/dev/null; then
        fail "OpenShift is not authenticated. Please log in before running this script."
    fi
    info "Using existing OpenShift session as '$(oc whoami)'."
}

# Creates the tsf-config ConfigMap in the tsf namespace. Uses --force so that
# the command succeeds even if the config already exists (idempotent).
create_tsf_config() {
    info "Creating TSF config..."
    tsf config --create --force
    info "TSF config created."
}

# Creates or updates the tsf-github-integration secret in the tsf namespace.
create_github_secret() {
    info "Creating GitHub App integration secret..."
    oc create secret generic tsf-github-integration \
        --from-literal=id="${GITHUB__APP_ID}" \
        --from-literal=pem="$(echo "${GITHUB__APP_PRIVATE_KEY}" | base64 -d)" \
        --from-literal=webhookSecret="${GITHUB__APP_WEBHOOK_SECRET}" \
        -n tsf \
        --dry-run=client -o yaml | oc apply -f -
    info "GitHub App integration secret created."
}

# Registers the Quay.io integration.
quay_integration() {
    info "Registering Quay integration (org='${QUAY__ORG}', url='${QUAY__URL}')..."
    tsf integration quay \
        --organization="${QUAY__ORG}" \
        --token="${QUAY__API_TOKEN}" \
        --url="${QUAY__URL}" \
        --force
    info "Quay integration registered."
}

# Deploys all TSF services.
deploy() {
    info "Deploying TSF..."
    tsf deploy
    info "TSF deploy command completed."
}

# Prints a summary of all deployed resources for CI log review.
report_deployment() {
    info "Generating deployment report..."

    local -r namespaces=(
        tsf tsf-keycloak tsf-tas tsf-tpa
        konflux-ui konflux-operator
        cert-manager-operator rhbk-operator rhtpa-operator
    )

    echo
    echo "============================================================"
    echo "  TSF Deployment Report"
    echo "============================================================"

    echo
    echo "## Pods"
    for ns in "${namespaces[@]}"; do
        local pods
        pods=$(oc get pods -n "${ns}" --no-headers \
            -o custom-columns='NAME:.metadata.name,STATUS:.status.phase,RESTARTS:.status.containerStatuses[0].restartCount' \
            2>/dev/null) || true
        if [[ -n "${pods}" ]]; then
            echo "  [${ns}]"
            echo "${pods}" | sed 's/^/    /'
        fi
    done

    echo
    echo "## Konflux Resources"
    oc get konflux -A --no-headers 2>/dev/null || echo "  (none found)"

    echo
    echo "============================================================"
}

#
# Main
#

main() {
    parse_args "$@"

    validate_env
    ensure_ocp_auth
    create_tsf_config
    create_github_secret
    quay_integration
    deploy
    report_deployment
}

if [ "${BASH_SOURCE[0]}" == "$0" ]; then
    main "$@"
    echo
    echo "Success"
fi
