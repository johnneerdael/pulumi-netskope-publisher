package provider

import (
	"context"
	"encoding/json"

	"github.com/blang/semver"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
)

const (
	Name              = "netskope-publisher"
	DisplayName       = "Netskope Publisher"
	Description       = "Pulumi components for provisioning Netskope Private Access Publishers on AWS, Azure, GCP, Kubernetes, vSphere, ESXi, Hcloud, Nutanix, OpenStack, OVH, Scaleway, OCI, Alicloud, Proxmox VE, DigitalOcean, Vultr, Exoscale, UpCloud, Stackit, Equinix Metal, Outscale, OpenTelekomCloud, TencentCloud, Yandex, and experimental Hyper-V."
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
		WithKeywords("category/network", "kind/component", "pulumi", "netskope", "npa", "publisher", "aws", "azure", "gcp", "kubernetes", "vsphere", "esxi", "hcloud", "nutanix", "openstack", "ovh", "scaleway", "oci", "alicloud", "proxmoxve", "proxmox", "digitalocean", "vultr", "exoscale", "upcloud", "stackit", "equinix", "outscale", "opentelekomcloud", "tencentcloud", "yandex").
		WithResources(
			infer.Resource(&NetskopeRegistration{}),
			infer.Resource(&PrivateApp{}),
		).
		WithComponents(
			infer.ComponentF(NewAwsPublisher),
			infer.ComponentF(NewAzurePublisher),
			infer.ComponentF(NewGcpPublisher),
			infer.ComponentF(NewKubernetesPublisher),
			infer.ComponentF(NewVspherePublisher),
			infer.ComponentF(NewEsxiPublisher),
			infer.ComponentF(NewHcloudPublisher),
			infer.ComponentF(NewNutanixPublisher),
			infer.ComponentF(NewOpenstackPublisher),
			infer.ComponentF(NewOvhPublisher),
			infer.ComponentF(NewScalewayPublisher),
			infer.ComponentF(NewOciPublisher),
			infer.ComponentF(NewAlicloudPublisher),
			infer.ComponentF(NewProxmoxvePublisher),
			infer.ComponentF(NewDigitaloceanPublisher),
			infer.ComponentF(NewVultrPublisher),
			infer.ComponentF(NewExoscalePublisher),
			infer.ComponentF(NewUpcloudPublisher),
			infer.ComponentF(NewStackitPublisher),
			infer.ComponentF(NewEquinixPublisher),
			infer.ComponentF(NewOutscalePublisher),
			infer.ComponentF(NewOpentelekomcloudPublisher),
			infer.ComponentF(NewTencentcloudPublisher),
			infer.ComponentF(NewYandexPublisher),
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
		semver.MustParse("0.2.2"),
		integration.WithProvider(provider),
	)
	if err != nil {
		return "", err
	}

	response, err := server.GetSchema(p.GetSchemaRequest{Version: version})
	if err != nil {
		return "", err
	}

	return normalizeSchema(response.Schema)
}

func normalizeSchema(raw string) (string, error) {
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		return "", err
	}
	schema["language"] = map[string]interface{}{
		"nodejs": map[string]interface{}{
			"packageName":          "@johninnl/pulumi-netskope-publisher",
			"respectSchemaVersion": true,
		},
		"python": map[string]interface{}{
			"packageName":          "pulumi_netskope_publisher",
			"pythonRequires":       ">=3.9",
			"requires":             map[string]interface{}{"pulumi": ">=3.0.0,<4.0.0"},
			"respectSchemaVersion": true,
			"pyproject":            map[string]interface{}{"enabled": true},
		},
		"csharp": map[string]interface{}{
			"packageName":          "JohninNL.Pulumi.NetskopePublisher",
			"packageReferences":    map[string]interface{}{"Pulumi": "3.*"},
			"respectSchemaVersion": true,
		},
		"go": map[string]interface{}{
			"importBasePath":       "github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher",
			"pulumiSDKVersion":     3,
			"respectSchemaVersion": true,
		},
		"java": map[string]interface{}{
			"basePackage": "com.pulumi",
			"buildFiles":  "gradle",
		},
	}
	normalized, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}
	return string(normalized), nil
}
