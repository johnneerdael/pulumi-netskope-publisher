package provider

import (
	"testing"

	"github.com/blang/semver"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func TestAwsConstructCreatesEc2InstanceChild(t *testing.T) {
	createdTypes := constructAndCollectTypes(t, "netskope-publisher:index:AwsPublisher", property.NewMap(map[string]property.Value{
		"names":            property.New([]property.Value{property.New("pub-1")}),
		"registrations":    registrationMap("pub-1"),
		"subnetId":         property.New("subnet-123"),
		"securityGroupIds": property.New([]property.Value{property.New("sg-123")}),
		"amiId":            property.New("ami-123"),
	}))

	if !contains(createdTypes, "aws:ec2/instance:Instance") {
		t.Fatalf("expected AWS construct to create EC2 instance, got %v", createdTypes)
	}
}

func TestGcpConstructCreatesComputeInstanceChild(t *testing.T) {
	createdTypes := constructAndCollectTypes(t, "netskope-publisher:index:GcpPublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"project":       property.New("project"),
		"zone":          property.New("us-central1-a"),
		"network":       property.New("default"),
		"subnetwork":    property.New("default"),
		"image":         property.New("publisher-image"),
	}))

	if !contains(createdTypes, "gcp:compute/instance:Instance") {
		t.Fatalf("expected GCP construct to create compute instance, got %v", createdTypes)
	}
}

func TestAzureConstructCreatesVirtualMachineChild(t *testing.T) {
	createdTypes := constructAndCollectTypes(t, "netskope-publisher:index:AzurePublisher", property.NewMap(map[string]property.Value{
		"names":             property.New([]property.Value{property.New("pub-1")}),
		"registrations":     registrationMap("pub-1"),
		"resourceGroupName": property.New("rg"),
		"location":          property.New("westeurope"),
		"subnetId":          property.New("/subscriptions/000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/default"),
		"adminSshPublicKey": property.New("ssh-rsa AAA"),
		"imageId":           property.New("/subscriptions/000/resourceGroups/rg/providers/Microsoft.Compute/images/publisher"),
	}))

	if !contains(createdTypes, "azure-native:compute:VirtualMachine") {
		t.Fatalf("expected Azure construct to create virtual machine, got %v", createdTypes)
	}
}

func TestVsphereConstructCreatesVirtualMachineChild(t *testing.T) {
	createdTypes := constructAndCollectTypes(t, "netskope-publisher:index:VspherePublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"datacenter":    property.New("dc"),
		"cluster":       property.New("cluster"),
		"datastore":     property.New("datastore"),
		"networkName":   property.New("VM Network"),
		"templateName":  property.New("publisher-template"),
	}))

	if !contains(createdTypes, "vsphere:index/virtualMachine:VirtualMachine") {
		t.Fatalf("expected vSphere construct to create virtual machine, got %v", createdTypes)
	}
}

func constructAndCollectTypes(t *testing.T, token string, inputs property.Map) []string {
	t.Helper()

	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	var createdTypes []string
	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.1.0"),
		integration.WithProvider(provider),
		integration.WithMocks(&integration.MockResourceMonitor{
			NewResourceF: func(args integration.MockResourceArgs) (string, property.Map, error) {
				createdTypes = append(createdTypes, string(args.TypeToken))
				return args.Name + "-id", args.Inputs, nil
			},
			CallF: func(args integration.MockCallArgs) (property.Map, error) {
				return property.NewMap(map[string]property.Value{
					"id":                    property.New("lookup-id"),
					"resourcePoolId":        property.New("resource-pool-id"),
					"guestId":               property.New("otherLinux64Guest"),
					"networkInterfaceTypes": property.New([]property.Value{property.New("vmxnet3")}),
					"disks": property.New([]property.Value{property.New(map[string]property.Value{
						"size":            property.New(64.0),
						"eagerlyScrub":    property.New(false),
						"thinProvisioned": property.New(true),
					})}),
				}), nil
			},
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = server.Construct(p.ConstructRequest{
		Urn:    presource.URN("urn:pulumi:stack::project::" + token + "::publisher"),
		Inputs: inputs,
	})
	if err != nil {
		t.Fatal(err)
	}

	return createdTypes
}

func registrationMap(name string) property.Value {
	return property.New(map[string]property.Value{
		name: property.New(map[string]property.Value{
			"publisherId":       property.New(123.0),
			"registrationToken": property.New("token"),
		}),
	})
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
