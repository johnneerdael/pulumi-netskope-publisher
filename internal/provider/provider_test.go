package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blang/semver"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func TestNetskopeRegistrationCreatesMissingPublishersAndTokens(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if got := r.Header.Get("Netskope-Api-Token"); got != "api-token" {
			t.Fatalf("expected Netskope API token header, got %q", got)
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/infrastructure/publishers":
			writeJSON(t, w, map[string]any{
				"data": map[string]any{
					"publishers": []map[string]any{{
						"publisher_name": "pub-a",
						"publisher_id":   "101",
					}},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/infrastructure/publishers":
			writeJSON(t, w, map[string]any{"data": map[string]any{"id": 202}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/infrastructure/publishers/101/registration_token":
			writeJSON(t, w, map[string]any{"data": map[string]any{"token": "token-101"}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/infrastructure/publishers/202/registration_token":
			writeJSON(t, w, map[string]any{"data": map[string]any{"token": "token-202"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	response := createRegistrationResource(t, property.NewMap(map[string]property.Value{
		"publisherNames": property.New([]property.Value{property.New("pub-a"), property.New("pub-b")}),
		"tenantUrl":      property.New(server.URL + "/"),
		"apiToken":       property.New("api-token"),
	}))

	registrations := response.Properties.Get("registrations").AsMap()
	pubA := registrations.Get("pub-a").AsMap()
	pubB := registrations.Get("pub-b").AsMap()

	if got := int(pubA.Get("publisherId").AsNumber()); got != 101 {
		t.Fatalf("expected existing publisher ID 101, got %d", got)
	}
	if got := pubA.Get("registrationToken").AsString(); got != "token-101" {
		t.Fatalf("expected existing publisher token, got %q", got)
	}
	if got := pubA.Get("existedBefore").AsBool(); !got {
		t.Fatalf("expected pub-a to be marked as existing")
	}
	if got := int(pubB.Get("publisherId").AsNumber()); got != 202 {
		t.Fatalf("expected created publisher ID 202, got %d", got)
	}
	if got := pubB.Get("registrationToken").AsString(); got != "token-202" {
		t.Fatalf("expected created publisher token, got %q", got)
	}
	if got := pubB.Get("existedBefore").AsBool(); got {
		t.Fatalf("expected pub-b to be marked as newly created")
	}

	expectedRequests := []string{
		"GET /api/v2/infrastructure/publishers",
		"POST /api/v2/infrastructure/publishers/101/registration_token",
		"POST /api/v2/infrastructure/publishers",
		"POST /api/v2/infrastructure/publishers/202/registration_token",
	}
	if !equalStrings(requests, expectedRequests) {
		t.Fatalf("expected requests %v, got %v", expectedRequests, requests)
	}
}

func TestAwsConstructCreatesRegistrationChildWhenRegistrationsOmitted(t *testing.T) {
	createdTypes := constructAndCollectTypes(t, "netskope-publisher:index:AwsPublisher", property.NewMap(map[string]property.Value{
		"names":            property.New([]property.Value{property.New("pub-1")}),
		"tenantUrl":        property.New("https://tenant.goskope.com"),
		"apiToken":         property.New("api-token"),
		"subnetId":         property.New("subnet-123"),
		"securityGroupIds": property.New([]property.Value{property.New("sg-123")}),
		"amiId":            property.New("ami-123"),
	}))

	if !contains(createdTypes, "netskope-publisher:index:NetskopeRegistration") {
		t.Fatalf("expected AWS construct to create NetskopeRegistration child, got %v", createdTypes)
	}
	if !contains(createdTypes, "aws:ec2/instance:Instance") {
		t.Fatalf("expected AWS construct to still create EC2 instance, got %v", createdTypes)
	}
}

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

func TestRenderUserDataBootstrapsPublisherByDefault(t *testing.T) {
	userData := renderUserData("pub-1", "token-123", nil)

	for _, expected := range []string{
		"system_info:\n  default_user:\n    name: ubuntu",
		"install -d -o ubuntu -g ubuntu -m 0755 /home/ubuntu/resources",
		"install -o ubuntu -g ubuntu -m 0644 /dev/null /home/ubuntu/resources/.nonat",
		"curl -fsSL https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/generic/bootstrap.sh | sudo bash",
		"sudo /home/ubuntu/npa_publisher_wizard -token \"token-123\"",
	} {
		if !strings.Contains(userData, expected) {
			t.Fatalf("expected user data to contain %q, got:\n%s", expected, userData)
		}
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
				if string(args.TypeToken) == "netskope-publisher:index:NetskopeRegistration" {
					return args.Name + "-id", property.NewMap(map[string]property.Value{
						"registrations": registrationMap("pub-1"),
					}), nil
				}
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

func createRegistrationResource(t *testing.T, inputs property.Map) p.CreateResponse {
	t.Helper()

	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.1.0"),
		integration.WithProvider(provider),
	)
	if err != nil {
		t.Fatal(err)
	}

	response, err := server.Create(p.CreateRequest{
		Urn:        presource.URN("urn:pulumi:stack::project::netskope-publisher:index:NetskopeRegistration::registration"),
		Properties: inputs,
	})
	if err != nil {
		t.Fatal(err)
	}

	return response
}

func registrationMap(name string) property.Value {
	return property.New(map[string]property.Value{
		name: property.New(map[string]property.Value{
			"publisherId":       property.New(123.0),
			"registrationToken": property.New("token"),
			"existedBefore":     property.New(true),
		}),
	})
}

func writeJSON(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatal(err)
	}
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func equalStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
