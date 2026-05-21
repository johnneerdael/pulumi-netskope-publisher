package provider

type providerCatalogEntry struct {
	DisplayName            string
	ComponentName          string
	Token                  string
	Implementation         string
	UserDataMode           string
	RequiredInputs         []string
	RequiredOneOf          [][]string
	MutuallyExclusive      [][]string
	ExperimentalOptInField string
}

var providerCatalog = map[string]providerCatalogEntry{
	"AwsPublisher":          providerEntry("AWS", "AwsPublisher", "bespoke", "base64", "subnetId", "securityGroupIds"),
	"AzurePublisher":        providerEntry("Azure", "AzurePublisher", "bespoke", "customData", "resourceGroupName", "location", "subnetId", "adminSshPublicKey"),
	"GcpPublisher":          providerEntry("GCP", "GcpPublisher", "bespoke", "metadata", "project", "zone", "network", "subnetwork", "image"),
	"KubernetesPublisher":   providerEntry("Kubernetes", "KubernetesPublisher", "bespoke", "none"),
	"VspherePublisher":      providerEntry("vSphere", "VspherePublisher", "bespoke", "guestInfo", "datacenter", "datastore", "networkName", "templateName"),
	"EsxiPublisher":         providerEntry("ESXi Native", "EsxiPublisher", "bespoke", "guestInfo", "diskStore", "virtualNetwork"),
	"HcloudPublisher":       providerEntry("Hcloud", "HcloudPublisher", "catalogRawVm", "plain"),
	"NutanixPublisher":      providerEntry("Nutanix", "NutanixPublisher", "catalogRawVm", "base64", "clusterUuid"),
	"OpenstackPublisher":    providerEntry("OpenStack", "OpenstackPublisher", "catalogRawVm", "plain", "imageName", "flavorName", "networkName"),
	"OvhPublisher":          providerEntry("OVH", "OvhPublisher", "catalogRawVm", "plain", "serviceName", "region", "imageId", "flavorId"),
	"ScalewayPublisher":     providerEntry("Scaleway", "ScalewayPublisher", "catalogRawVm", "scalewayDual"),
	"OciPublisher":          providerEntry("OCI", "OciPublisher", "catalogRawVm", "ociMetadata", "compartmentId", "availabilityDomain", "subnetId", "imageId"),
	"AlicloudPublisher":     providerEntry("Alicloud", "AlicloudPublisher", "catalogRawVm", "base64", "imageId", "vswitchId", "securityGroupIds"),
	"ProxmoxvePublisher":    providerEntry("Proxmox VE", "ProxmoxvePublisher", "catalogSpecializedVm", "proxmoxSnippet", "nodeName", "datastoreId", "templateVmId"),
	"DigitaloceanPublisher": providerEntry("DigitalOcean", "DigitaloceanPublisher", "catalogRawVm", "plain", "region"),
	"VultrPublisher": {
		DisplayName:       "Vultr",
		ComponentName:     "VultrPublisher",
		Token:             "netskope-publisher:index:VultrPublisher",
		Implementation:    "catalogRawVm",
		UserDataMode:      "plain",
		RequiredInputs:    []string{"region", "plan"},
		RequiredOneOf:     [][]string{{"osId", "imageId"}},
		MutuallyExclusive: [][]string{{"osId", "imageId"}},
	},
	"ExoscalePublisher": providerEntry("Exoscale", "ExoscalePublisher", "catalogRawVm", "plain", "zone", "type", "templateId", "diskSize"),
	"UpcloudPublisher":  providerEntry("UpCloud", "UpcloudPublisher", "catalogRawVm", "plain", "zone"),
	"StackitPublisher":  providerEntry("Stackit", "StackitPublisher", "catalogRawVm", "plain", "projectId", "machineType", "imageId"),
	"EquinixPublisher":  providerEntry("Equinix Metal", "EquinixPublisher", "catalogRawVm", "plain", "projectId", "metro", "plan"),
	"OutscalePublisher": providerEntry("Outscale", "OutscalePublisher", "catalogRawVm", "plain", "imageId"),
	"OpentelekomcloudPublisher": {
		DisplayName:       "OpenTelekomCloud",
		ComponentName:     "OpentelekomcloudPublisher",
		Token:             "netskope-publisher:index:OpentelekomcloudPublisher",
		Implementation:    "catalogRawVm",
		UserDataMode:      "plain",
		RequiredInputs:    []string{"networks"},
		MutuallyExclusive: [][]string{{"imageName", "imageId"}, {"flavorName", "flavorId"}},
	},
	"TencentcloudPublisher": providerEntry("TencentCloud", "TencentcloudPublisher", "catalogRawVm", "raw", "availabilityZone", "imageId"),
	"YandexPublisher":       providerEntry("Yandex Cloud", "YandexPublisher", "catalogRawVm", "metadata", "imageId", "subnetId"),
	"NetskopeRegistration":  providerEntry("Netskope Registration", "NetskopeRegistration", "resource", "none", "publisherNames", "tenantUrl"),
	"HypervPublisher": {
		DisplayName:            "Hyper-V",
		ComponentName:          "HypervPublisher",
		Token:                  "netskope-publisher:index:HypervPublisher",
		Implementation:         "bespoke",
		UserDataMode:           "none",
		RequiredInputs:         []string{"switchName", "hardDrives"},
		ExperimentalOptInField: "enableExperimentalHyperv",
	},
}

func providerEntry(displayName string, componentName string, implementation string, userDataMode string, requiredInputs ...string) providerCatalogEntry {
	return providerCatalogEntry{
		DisplayName:    displayName,
		ComponentName:  componentName,
		Token:          "netskope-publisher:index:" + componentName,
		Implementation: implementation,
		UserDataMode:   userDataMode,
		RequiredInputs: requiredInputs,
	}
}
