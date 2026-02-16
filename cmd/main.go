package main

import (
	"fmt"
	"os"

	"github.com/redhat-appstudio/helmet/api"
	"github.com/redhat-appstudio/helmet/framework"
	"github.com/redhat-appstudio/tsf-cli/installer"
)

var (
	// Build-time variables set via ldflags
	version  = "v0.0.0-SNAPSHOT"
	commitID = ""
)

func main() {
	// 1. Create application context with metadata
	appCtx := createAppContext()

	// 2. Get current working directory for local overrides
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 3. Build MCP image reference
	mcpImage := buildMCPImage()

	// 4. Create application with framework options (GitHub URLs via CustomURLProvider; see framework/integrations.go)
	appIntegrations := framework.StandardIntegrations()
	appIntegrations = framework.WithURLProvider(appIntegrations, CustomURLProvider{})
	app, err := framework.NewAppFromTarball(
		appCtx,
		installer.InstallerTarball,
		cwd,
		framework.WithIntegrations(appIntegrations...),
		framework.WithMCPImage(mcpImage),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 5. Run the application
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// createAppContext initializes the application context with build metadata
// and descriptive information about the example application.
func createAppContext() *api.AppContext {
	return api.NewAppContext(
		"tsf",
		api.WithVersion(version),
		api.WithCommitID(commitID),
		api.WithNamespace("tsf"),
		api.WithShortDescription("Trusted Software Factory Installer"),
		api.WithLongDescription(`Trusted Software Factory Installer

This application allows you to install the Trusted Software Factory on your cluster. This includes the following components:
	- Cert-Manager
	- Keycloak
	- Konflux
	- OpenShift Pipelines
	- Trusted Artifact Signer
	- Trusted Profile Analyzer
`),
	)
}

// buildMCPImage constructs the container image reference for the MCP server.
// Uses the commit ID for versioning when available, falls back to 'latest'.
func buildMCPImage() string {
	mcpImage := "quay.io/redhat-appstudio/tsf-cli"
	if commitID != "" && commitID != "unknown" {
		return fmt.Sprintf("%s:%s", mcpImage, commitID)
	}
	return fmt.Sprintf("%s:latest", mcpImage)
}
