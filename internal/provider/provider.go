package provider

import (
	"context"

	"github.com/blang/semver"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
)

const (
	Name              = "netskope-publisher"
	DisplayName       = "Netskope Publisher"
	Description       = "Pulumi components for provisioning Netskope Private Access Publishers on AWS, Azure, GCP, vSphere, and experimental Hyper-V."
	Publisher         = "johnneerdael"
	Homepage          = "https://johnneerdael.github.io/pulumi-netskope-publisher/"
	Repository        = "https://github.com/johnneerdael/pulumi-netskope-publisher"
	License           = "Apache-2.0"
	LogoURL           = "https://raw.githubusercontent.com/johnneerdael/pulumi-netskope-publisher/main/site/source/images/netskope-logo.png"
	PluginDownloadURL = "github://api.github.com/johnneerdael/pulumi-netskope-publisher"
)

func New() (p.Provider, error) {
	return infer.NewProviderBuilder().
		WithDisplayName(DisplayName).
		WithDescription(Description).
		WithPublisher(Publisher).
		WithHomepage(Homepage).
		WithRepository(Repository).
		WithLicense(License).
		WithLogoURL(LogoURL).
		WithPluginDownloadURL(PluginDownloadURL).
		WithKeywords("category/network", "kind/component", "pulumi", "netskope", "npa", "publisher", "aws", "azure", "gcp", "vsphere").
		WithResources(
			infer.Resource(&NetskopeRegistration{}),
		).
		WithComponents(
			infer.ComponentF(NewAwsPublisher),
			infer.ComponentF(NewAzurePublisher),
			infer.ComponentF(NewGcpPublisher),
			infer.ComponentF(NewVspherePublisher),
			infer.ComponentF(NewHypervPublisher),
		).
		WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{
			Name: "index",
		}).
		Build()
}

func Schema(ctx context.Context, version int) (string, error) {
	provider, err := New()
	if err != nil {
		return "", err
	}

	server, err := integration.NewServer(
		ctx,
		Name,
		semver.MustParse("0.1.5"),
		integration.WithProvider(provider),
	)
	if err != nil {
		return "", err
	}

	response, err := server.GetSchema(p.GetSchemaRequest{Version: version})
	if err != nil {
		return "", err
	}

	return response.Schema, nil
}
