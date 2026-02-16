package main

import (
	"context"
	"fmt"

	"github.com/redhat-appstudio/helmet/api"
)

// CustomURLProvider implements integrations.URLProvider by building
// GitHub App URLs from the cluster's OpenShift ingress domain.
type CustomURLProvider struct{}

// GetCallbackURL is not used in this example.
func (CustomURLProvider) GetCallbackURL(_ context.Context, _ api.IntegrationContext) (string, error) {
	return "", nil
}

// GetHomepageURL showcases how to derive the homepage URL from the product configuration and the ingress domain.
func (CustomURLProvider) GetHomepageURL(ctx context.Context, ic api.IntegrationContext) (string, error) {
	ingressDomain, err := ic.GetOpenShiftIngressDomain(ctx)
	if err != nil {
		return "", fmt.Errorf("ingress domain unavailable (non-OpenShift cluster); "+
			"provide --homepage-url explicitly: %w", err)
	}
	namespace, err := ic.GetProductNamespace("Konflux")
	if err != nil {
		return "", fmt.Errorf("product unavailable: %w", err)
	}
	return fmt.Sprintf("https://konflux-ui-%s.%s", namespace, ingressDomain), nil
}

// GetWebhookURL showcases how to derive the webhook URL from the ingress domain.
func (CustomURLProvider) GetWebhookURL(ctx context.Context, ic api.IntegrationContext) (string, error) {
	ingressDomain, err := ic.GetOpenShiftIngressDomain(ctx)
	if err != nil {
		return "", fmt.Errorf("ingress domain unavailable (non-OpenShift cluster); "+
			"provide --webhook-url explicitly: %w", err)
	}
	return fmt.Sprintf(
		"https://pipelines-as-code-controller-openshift-pipelines.%s",
		ingressDomain,
	), nil
}
