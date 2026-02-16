# Trusted Software Factory Installer

## Prerequisites

* An OpenShift cluster with admin access
* A GitHub organization (create a test organization if needed)
* A Quay.io account with access to an organization (currently only quay.io is supported; support for other Quay instances is planned for the future)

## Quickstart

1. Copy [private.env.template](../hack/private.env.template) as `tsf.env` and fill in the blanks. See the [Getting the information for tsf.env](#getting-the-information-for-tsfenv) section if you need help finding the required values.
2. Start a container with `podman run -it --rm --env-file tsf.env --entrypoint bash -p 8228:8228 --pull always quay.io/roming22-org/tsf:latest --login`.
3. Log in to the cluster with `oc login "$OCP__API_ENDPOINT" --username "$OCP__USERNAME" --password "$OCP__PASSWORD"`
4. Create the TSF config on the cluster with `tsf config --create`.
5. Check if the Red Hat Cert-Manager operator is already installed in the cluster. If it is, edit the `tsf-config` ConfigMap in the `tssc` namespace to set `manageSubscription: false` for the Cert-Manager product.
6. Create the github app integration with `tsf integration github --create --org "$GITHUB__ORG" "tsf-$(date +%m%d-%H%M)"`. Open the link to create the app, follow the instructions, and install the application to your GitHub organization.
7. Create the quay integration with `tsf integration quay --organization="$QUAY__ORG" --token="$QUAY__API_TOKEN" --url="$QUAY__URL"`. If you need information on how to generate the token, look further in this document.
8. Deploy all the services with `tsf deploy`.
9. Cluster users (including the admin user) should now be able to log in to the Konflux UI. You will find the URL in the logs of the deployment. You can also access it through the GitHub App (via the `Website` link on the application public page, or via the link displayed on the application configuration page which was the last page displayed after you installed the application).

## Getting the information for tsf.env

### GitHub

If you don't have one, create a test organization.
Use the name of that organization for `GITHUB__ORG`.

### OpenShift

* `OCP__API_ENDPOINT`: the full url of the cluster's api endpoint. Example: `https://api.example.com:6443`.
* `OCP__USERNAME`: User with admin privileges on the cluster. Example: `admin`.
* `OCP__PASSWORD`: Credential for the user.

### Quay

* `QUAY__URL`: full url of quay.io. Use: `https://quay.io`.
* `QUAY__API_TOKEN`: token giving access to an organization on quay.io.
* `QUAY__ORG`: the organization that the token gives access to.

To create the API token:
* Go to the Quay homepage.
* On the right-hand side, click the organization you wish to use. Using an organization is mandatory. If you do not have one, you can create one by clicking `Create New Organization`.
* On the left-hand side, click the `Applications` icon (second to last icon).
* On the right-hand side, click `Create New Application`, and enter the application name (e.g. `tsf`).
* Click the name of the newly created application.
* On the left-hand side, click the `Generate Token` icon (last icon).
* Select everything (TODO: find the minimum set of permissions required).
* Click `Generate Access Token`.
* Click `Authorize Application`.
* Copy the access token.

## Services

### Quay

When a new component is created in Konflux, a new repository is created, in the organization specified at install time.
If you are using a free quay.io account, the visibility of the repository should be changed to public manually because of the account limitations.
If you are using a paid quay.io account, or your own Quay instance, the repository visibility can remain private.

## Installer Troubleshooting

### tsf-subscription fails to install with "Error: upgrade failed"

#### Symptom

```
Error: upgrade failed: Unable to continue with update: Subscription "openshift-cert-manager-operator" in namespace "cert-manager-operator" exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"; annotation validation error: missing key "meta.helm.sh/release-name": must be set to "tsf-subscriptions"; annotation validation error: missing key "meta.helm.sh/release-namespace": must be set to "tsf"
```

#### Root cause

The subscription has already been installed by a third party. Helm does not want to take ownership of the resource.

#### Workaround

* Edit the `tsf-config` ConfigMap in the `tssc` namespace to set `manageSubscription: false` for the Cert-Manager product.

## Konflux Troubleshooting

This document is not the reference for Konflux troubleshooting, but lists a few common scenarios.

### No pull request is sent to the repository during onboarding

#### Symptom

After creating the component in Konflux and linking the repository, Konflux does not open the expected pull request to the repository.

#### Root cause

You did not install the GitHub App for the user/organization that has control of the repository.

#### Workaround

* Install the GitHub App for the user/organization that controls the repository, or move the repository to a user/organization for which the GitHub App has already been configured.
* Onboard the repository as a new component.

### The Konflux pull request does not trigger a PipelineRun

#### Symptom

When onboarding a repository to Konflux, the expected PR is opened to the repository, but CI is not triggered.

#### Root cause

The most likely root cause is that the certificate of the cluster is not valid.
This can be verified by going to the GitHub App settings, and look at `Recent Deliveries` under the `Advanced` tab.
Undelivered messages will appear with a yellow dot.

#### Workaround

Go to the `General` section of the GitHub App settings, disable `SSL verification` and save the new settings.
Go back to `Recent Deliveries` under the `Advanced` tab, locate the latest `pull_request.opened` event, click on the `...` and click on `Redeliver`.
You should see a new event being triggered, with a green check mark this time, and the PipelineRun should be executed.
