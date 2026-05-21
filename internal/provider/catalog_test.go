package provider

import "testing"

func TestProviderCatalogIncludesCurrentComponents(t *testing.T) {
	required := []string{
		"AwsPublisher",
		"AzurePublisher",
		"GcpPublisher",
		"KubernetesPublisher",
		"VspherePublisher",
		"EsxiPublisher",
		"HcloudPublisher",
		"NutanixPublisher",
		"OpenstackPublisher",
		"OvhPublisher",
		"ScalewayPublisher",
		"OciPublisher",
		"AlicloudPublisher",
		"ProxmoxvePublisher",
		"DigitaloceanPublisher",
		"VultrPublisher",
		"ExoscalePublisher",
		"UpcloudPublisher",
		"StackitPublisher",
		"EquinixPublisher",
		"OutscalePublisher",
		"OpentelekomcloudPublisher",
		"TencentcloudPublisher",
		"YandexPublisher",
		"HypervPublisher",
	}

	for _, name := range required {
		entry, ok := providerCatalog[name]
		if !ok {
			t.Fatalf("%s missing from Go provider catalog", name)
		}
		if entry.Token != "netskope-publisher:index:"+name {
			t.Fatalf("%s token mismatch: %s", name, entry.Token)
		}
	}
}

func TestProviderCatalogValidationMetadata(t *testing.T) {
	entry := providerCatalog["DigitaloceanPublisher"]
	if len(entry.RequiredInputs) != 1 || entry.RequiredInputs[0] != "region" {
		t.Fatalf("DigitaloceanPublisher required inputs mismatch: %#v", entry.RequiredInputs)
	}

	hyperv := providerCatalog["HypervPublisher"]
	if hyperv.ExperimentalOptInField != "enableExperimentalHyperv" {
		t.Fatalf("HypervPublisher missing experimental opt-in metadata")
	}
}

func TestProviderCatalogOneOfValidationMetadata(t *testing.T) {
	vultr := providerCatalog["VultrPublisher"]
	if len(vultr.RequiredOneOf) != 1 || len(vultr.RequiredOneOf[0]) != 2 {
		t.Fatalf("VultrPublisher missing required-one-of metadata: %#v", vultr.RequiredOneOf)
	}
	if vultr.RequiredOneOf[0][0] != "osId" || vultr.RequiredOneOf[0][1] != "imageId" {
		t.Fatalf("VultrPublisher required-one-of mismatch: %#v", vultr.RequiredOneOf)
	}
	if len(vultr.MutuallyExclusive) != 1 || len(vultr.MutuallyExclusive[0]) != 2 {
		t.Fatalf("VultrPublisher missing mutually-exclusive metadata: %#v", vultr.MutuallyExclusive)
	}
	if vultr.MutuallyExclusive[0][0] != "osId" || vultr.MutuallyExclusive[0][1] != "imageId" {
		t.Fatalf("VultrPublisher mutually-exclusive mismatch: %#v", vultr.MutuallyExclusive)
	}
}
