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
