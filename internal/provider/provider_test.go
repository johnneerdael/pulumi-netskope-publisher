package provider

import (
	"encoding/base64"
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
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestNetskopeRegistrationCreatesMissingPublishersAndTokens(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if got := r.Header.Get("Authorization"); got != "Bearer api-token" {
			t.Fatalf("expected bearer authorization header, got %q", got)
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

func TestNetskopeRegistrationSupportsOAuth2ClientCredentials(t *testing.T) {
	var tokenRequests int
	var authorizationHeaders []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/oauth2/token":
			tokenRequests++
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if got := r.Form.Get("grant_type"); got != "client_credentials" {
				t.Fatalf("expected grant_type client_credentials, got %q", got)
			}
			if got := r.Form.Get("client_id"); got != "client-id" {
				t.Fatalf("expected client_id, got %q", got)
			}
			if got := r.Form.Get("client_secret"); got != "client-secret" {
				t.Fatalf("expected client_secret, got %q", got)
			}
			if got := r.Form.Get("scope"); got != "npa.publisher" {
				t.Fatalf("expected scope, got %q", got)
			}
			writeJSON(t, w, map[string]any{"access_token": "oauth-access-token"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/infrastructure/publishers":
			authorizationHeaders = append(authorizationHeaders, r.Header.Get("Authorization"))
			writeJSON(t, w, map[string]any{"data": map[string]any{"publishers": []map[string]any{}}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/infrastructure/publishers":
			authorizationHeaders = append(authorizationHeaders, r.Header.Get("Authorization"))
			writeJSON(t, w, map[string]any{"data": map[string]any{"id": 202}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/infrastructure/publishers/202/registration_token":
			authorizationHeaders = append(authorizationHeaders, r.Header.Get("Authorization"))
			writeJSON(t, w, map[string]any{"data": map[string]any{"token": "token-202"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	response := createRegistrationResource(t, property.NewMap(map[string]property.Value{
		"publisherNames": property.New([]property.Value{property.New("pub-a")}),
		"tenantUrl":      property.New(server.URL + "/"),
		"authMode":       property.New("oauth2"),
		"oauth2": property.New(map[string]property.Value{
			"tokenUrl":     property.New(server.URL + "/oauth2/token"),
			"clientId":     property.New("client-id"),
			"clientSecret": property.New("client-secret"),
			"scope":        property.New("npa.publisher"),
		}),
	}))

	pubA := response.Properties.Get("registrations").AsMap().Get("pub-a").AsMap()
	if got := pubA.Get("registrationToken").AsString(); got != "token-202" {
		t.Fatalf("expected registration token, got %q", got)
	}
	if tokenRequests != 1 {
		t.Fatalf("expected exactly one OAuth2 token request, got %d", tokenRequests)
	}
	for _, header := range authorizationHeaders {
		if header != "Bearer oauth-access-token" {
			t.Fatalf("expected OAuth2 bearer header, got %q", header)
		}
	}
}

func TestNetskopeClientReportsHTTPStatusAndBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"status":"error","message":"bad token"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newNetskopeClient(netskopeClientConfig{
		TenantURL:   server.URL,
		BearerToken: "bad-token",
		AuthMode:    "token",
		HTTPClient:  server.Client(),
	})

	var output map[string]any
	err := client.request(t.Context(), "Test operation", http.MethodGet, "/api/v2/test", nil, &output)
	if err == nil {
		t.Fatalf("expected request error")
	}
	if !strings.Contains(err.Error(), "Test operation failed (status=401)") {
		t.Fatalf("expected status in error, got %v", err)
	}
	if !strings.Contains(err.Error(), "bad token") {
		t.Fatalf("expected response body in error, got %v", err)
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

func TestKubernetesConstructCreatesNamespaceSecretsAndHelmReleaseChildren(t *testing.T) {
	createdTypes := constructAndCollectTypes(t, "netskope-publisher:index:KubernetesPublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"namespace":     property.New("netskope"),
	}))

	for _, expected := range []string{
		"kubernetes:core/v1:Namespace",
		"kubernetes:core/v1:Secret",
		"kubernetes:helm.sh/v3:Release",
	} {
		if !contains(createdTypes, expected) {
			t.Fatalf("expected Kubernetes construct to create %s child, got %v", expected, createdTypes)
		}
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

func TestRenderUserDataSupportsInstallUserOptions(t *testing.T) {
	password := "S3cret-Passw0rd!"
	gateway := "10.0.0.1"
	mtu := 1460
	userData := renderUserDataWithOptions("pub-1", "token-123", nil, cloudInitOptions{
		Bootstrap:                    true,
		BootstrapURL:                 defaultBootstrapURL,
		Nonat:                        true,
		InstallUser:                  "npa",
		InstallUserPassword:          &password,
		InstallUserSSHAuthorizedKeys: []string{"ssh-ed25519 AAAA fake-key"},
		DeleteDefaultUser:            true,
		GuestNetworkInterface: &GuestNetworkInterface{
			Name:        "ens160",
			Addresses:   []string{"10.0.0.10/24"},
			Gateway4:    &gateway,
			Nameservers: []string{"10.0.0.2"},
			MTU:         &mtu,
		},
	})

	for _, expected := range []string{
		"default_user:\n    name: npa",
		"lock_passwd: false",
		"ssh-ed25519 AAAA fake-key",
		"password: \"S3cret-Passw0rd!\"",
		"userdel -r ubuntu",
		"/home/npa/resources/.nonat",
		"su - npa -c 'curl -fsSL",
		"sudo /home/npa/npa_publisher_wizard -token \"token-123\"",
		"path: /etc/netplan/60-cloudinit-override.yaml",
		"ens160:",
		"mtu: 1460",
	} {
		if !strings.Contains(userData, expected) {
			t.Fatalf("expected user data to contain %q, got:\n%s", expected, userData)
		}
	}
}

func TestUserDataAdaptersRenderPlacement(t *testing.T) {
	payload := pulumi.String("#cloud-config").ToStringOutput()

	_ = plainUserData(payload)
	_ = base64UserData(payload)
	if got := metadataUserData(payload, "user-data"); got["user-data"] == nil {
		t.Fatalf("expected metadata user-data key")
	}
	if got := base64MetadataUserData(payload, "userData"); got["userData"] == nil {
		t.Fatalf("expected base64 metadata userData key")
	}
	if got := guestInfoUserData(payload); got["guestinfo.userdata.encoding"] == nil || got["guestinfo.userdata"] == nil {
		t.Fatalf("expected guestinfo userdata and encoding keys")
	}
	if got := scalewayUserData(payload); got["cloudInit"] == nil || got["userData"] == nil {
		t.Fatalf("expected Scaleway cloudInit and userData keys")
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

func TestAdditionalProviderConstructsCreateProviderChildren(t *testing.T) {
	cases := []struct {
		name     string
		token    string
		inputs   property.Map
		expected string
	}{
		{
			name:  "ESXi",
			token: "netskope-publisher:index:EsxiPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":          property.New([]property.Value{property.New("pub-1")}),
				"registrations":  registrationMap("pub-1"),
				"diskStore":      property.New("datastore1"),
				"virtualNetwork": property.New("VM Network"),
			}),
			expected: "esxi-native:index:VirtualMachine",
		},
		{
			name:  "Hcloud",
			token: "netskope-publisher:index:HcloudPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
			}),
			expected: "hcloud:index/server:Server",
		},
		{
			name:  "Nutanix",
			token: "netskope-publisher:index:NutanixPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"clusterUuid":   property.New("cluster-uuid"),
			}),
			expected: "nutanix:index/virtualMachine:VirtualMachine",
		},
		{
			name:  "OpenStack",
			token: "netskope-publisher:index:OpenstackPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"imageName":     property.New("Ubuntu 22.04"),
				"flavorName":    property.New("m1.medium"),
				"networkName":   property.New("private"),
			}),
			expected: "openstack:compute/instance:Instance",
		},
		{
			name:  "OVH",
			token: "netskope-publisher:index:OvhPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"serviceName":   property.New("project-id"),
				"region":        property.New("GRA11"),
				"imageId":       property.New("image-id"),
				"flavorId":      property.New("flavor-id"),
			}),
			expected: "ovh:CloudProject/instance:Instance",
		},
		{
			name:  "Scaleway",
			token: "netskope-publisher:index:ScalewayPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
			}),
			expected: "scaleway:instance/server:Server",
		},
		{
			name:  "OCI",
			token: "netskope-publisher:index:OciPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":              property.New([]property.Value{property.New("pub-1")}),
				"registrations":      registrationMap("pub-1"),
				"compartmentId":      property.New("ocid1.compartment.oc1..example"),
				"availabilityDomain": property.New("AD-1"),
				"subnetId":           property.New("ocid1.subnet.oc1..example"),
				"imageId":            property.New("ocid1.image.oc1..example"),
			}),
			expected: "oci:Core/instance:Instance",
		},
		{
			name:  "Alicloud",
			token: "netskope-publisher:index:AlicloudPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":            property.New([]property.Value{property.New("pub-1")}),
				"registrations":    registrationMap("pub-1"),
				"imageId":          property.New("ubuntu_22_04_x64_20G_alibase.vhd"),
				"vswitchId":        property.New("vsw-123"),
				"securityGroupIds": property.New([]property.Value{property.New("sg-123")}),
			}),
			expected: "alicloud:ecs/instance:Instance",
		},
		{
			name:  "Proxmox VE",
			token: "netskope-publisher:index:ProxmoxvePublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"nodeName":      property.New("pve-1"),
				"datastoreId":   property.New("local"),
				"templateVmId":  property.New(9000.0),
			}),
			expected: "proxmoxve:index/vmLegacy:VmLegacy",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			createdTypes := constructAndCollectTypes(t, tc.token, tc.inputs)
			if !contains(createdTypes, tc.expected) {
				t.Fatalf("expected %s construct to create %s child, got %v", tc.name, tc.expected, createdTypes)
			}
		})
	}
}

func TestAdditionalProviderConstructsBootstrapWithRegistryFields(t *testing.T) {
	cases := []struct {
		name      string
		token     string
		inputs    property.Map
		childType string
		validate  func(*testing.T, property.Map)
	}{
		{
			name:  "ESXi",
			token: "netskope-publisher:index:EsxiPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":          property.New([]property.Value{property.New("pub-1")}),
				"registrations":  registrationMap("pub-1"),
				"diskStore":      property.New("datastore1"),
				"virtualNetwork": property.New("VM Network"),
			}),
			childType: "esxi-native:index:VirtualMachine",
			validate: func(t *testing.T, inputs property.Map) {
				info := inputs.Get("info").AsArray().AsSlice()
				assertKeyValueArrayHas(t, info, "guestinfo.userdata.encoding", "base64")
				userData := decodeRequiredBase64(t, keyValueArrayValue(t, info, "guestinfo.userdata"))
				assertBootstrapUserData(t, userData)
			},
		},
		{
			name:  "Hcloud",
			token: "netskope-publisher:index:HcloudPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
			}),
			childType: "hcloud:index/server:Server",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "Nutanix",
			token: "netskope-publisher:index:NutanixPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"clusterUuid":   property.New("cluster-uuid"),
			}),
			childType: "nutanix:index/virtualMachine:VirtualMachine",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, decodeRequiredBase64(t, inputs.Get("guestCustomizationCloudInitUserData").AsString()))
			},
		},
		{
			name:  "OpenStack",
			token: "netskope-publisher:index:OpenstackPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"imageName":     property.New("Ubuntu 22.04"),
				"flavorName":    property.New("m1.medium"),
				"networkName":   property.New("private"),
			}),
			childType: "openstack:compute/instance:Instance",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "OVH",
			token: "netskope-publisher:index:OvhPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"serviceName":   property.New("project-id"),
				"region":        property.New("GRA11"),
				"imageId":       property.New("image-id"),
				"flavorId":      property.New("flavor-id"),
			}),
			childType: "ovh:CloudProject/instance:Instance",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "Scaleway",
			token: "netskope-publisher:index:ScalewayPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
			}),
			childType: "scaleway:instance/server:Server",
			validate: func(t *testing.T, inputs property.Map) {
				userData := inputs.Get("userData").AsMap().Get("cloud-init").AsString()
				assertBootstrapUserData(t, userData)
			},
		},
		{
			name:  "OCI",
			token: "netskope-publisher:index:OciPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":              property.New([]property.Value{property.New("pub-1")}),
				"registrations":      registrationMap("pub-1"),
				"compartmentId":      property.New("ocid1.compartment.oc1..example"),
				"availabilityDomain": property.New("AD-1"),
				"subnetId":           property.New("ocid1.subnet.oc1..example"),
				"imageId":            property.New("ocid1.image.oc1..example"),
			}),
			childType: "oci:Core/instance:Instance",
			validate: func(t *testing.T, inputs property.Map) {
				userData := inputs.Get("metadata").AsMap().Get("userData").AsString()
				assertBootstrapUserData(t, decodeRequiredBase64(t, userData))
			},
		},
		{
			name:  "Alicloud",
			token: "netskope-publisher:index:AlicloudPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":            property.New([]property.Value{property.New("pub-1")}),
				"registrations":    registrationMap("pub-1"),
				"imageId":          property.New("ubuntu_22_04_x64_20G_alibase.vhd"),
				"vswitchId":        property.New("vsw-123"),
				"securityGroupIds": property.New([]property.Value{property.New("sg-123")}),
			}),
			childType: "alicloud:ecs/instance:Instance",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, decodeRequiredBase64(t, inputs.Get("userData").AsString()))
			},
		},
		{
			name:  "Proxmox VE",
			token: "netskope-publisher:index:ProxmoxvePublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"nodeName":      property.New("pve-1"),
				"datastoreId":   property.New("local"),
				"templateVmId":  property.New(9000.0),
			}),
			childType: "proxmoxve:index/fileLegacy:FileLegacy",
			validate: func(t *testing.T, inputs property.Map) {
				sourceRaw := inputs.Get("sourceRaw").AsMap()
				assertBootstrapUserData(t, sourceRaw.Get("data").AsString())
				if sourceRaw.Get("fileName").AsString() != "pub-1-user-data.yaml" {
					t.Fatalf("expected Proxmox VE cloud-init file name to be stable")
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resources := constructAndCollectResources(t, tc.token, tc.inputs)
			child := findResourceByType(t, resources, tc.childType)
			tc.validate(t, child.Inputs)
		})
	}
}

func TestExpandedProviderConstructsCreateProviderChildren(t *testing.T) {
	cases := expandedProviderCases()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			createdTypes := constructAndCollectTypes(t, tc.token, tc.inputs)
			if !contains(createdTypes, tc.expected) {
				t.Fatalf("expected %s construct to create %s child, got %v", tc.name, tc.expected, createdTypes)
			}
		})
	}
}

func TestExpandedProviderConstructsBootstrapWithRegistryFields(t *testing.T) {
	cases := expandedProviderCases()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resources := constructAndCollectResources(t, tc.token, tc.inputs)
			child := findResourceByType(t, resources, tc.expected)
			tc.validate(t, child.Inputs)
		})
	}
}

type expandedProviderCase struct {
	name     string
	token    string
	inputs   property.Map
	expected string
	validate func(*testing.T, property.Map)
}

func expandedProviderCases() []expandedProviderCase {
	return []expandedProviderCase{
		{
			name:  "DigitalOcean",
			token: "netskope-publisher:index:DigitaloceanPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"region":        property.New("ams3"),
			}),
			expected: "digitalocean:index/droplet:Droplet",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "Vultr",
			token: "netskope-publisher:index:VultrPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"region":        property.New("ams"),
				"plan":          property.New("vc2-2c-4gb"),
				"osId":          property.New(1743.0),
			}),
			expected: "vultr:index/instance:Instance",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "Exoscale",
			token: "netskope-publisher:index:ExoscalePublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"zone":          property.New("ch-gva-2"),
				"type":          property.New("standard.medium"),
				"templateId":    property.New("template-id"),
				"diskSize":      property.New(50.0),
			}),
			expected: "exoscale:index/computeInstance:ComputeInstance",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "UpCloud",
			token: "netskope-publisher:index:UpcloudPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"zone":          property.New("nl-ams1"),
			}),
			expected: "upcloud:index/server:Server",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "Stackit",
			token: "netskope-publisher:index:StackitPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"projectId":     property.New("project-id"),
				"machineType":   property.New("g1.2"),
				"imageId":       property.New("image-id"),
			}),
			expected: "stackit:index/server:Server",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "Equinix",
			token: "netskope-publisher:index:EquinixPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"projectId":     property.New("project-id"),
				"metro":         property.New("AM"),
				"plan":          property.New("c3.small.x86"),
			}),
			expected: "equinix:metal/device:Device",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "Outscale",
			token: "netskope-publisher:index:OutscalePublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"imageId":       property.New("ami-123"),
			}),
			expected: "outscale:index/vm:Vm",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "OpenTelekomCloud",
			token: "netskope-publisher:index:OpentelekomcloudPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"networks": property.New([]property.Value{property.New(map[string]property.Value{
					"name": property.New("private"),
				})}),
			}),
			expected: "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userData").AsString())
			},
		},
		{
			name:  "TencentCloud",
			token: "netskope-publisher:index:TencentcloudPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":            property.New([]property.Value{property.New("pub-1")}),
				"registrations":    registrationMap("pub-1"),
				"availabilityZone": property.New("ap-guangzhou-6"),
				"imageId":          property.New("img-123"),
			}),
			expected: "tencentcloud:index/instance:Instance",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("userDataRaw").AsString())
			},
		},
		{
			name:  "Yandex",
			token: "netskope-publisher:index:YandexPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"imageId":       property.New("image-id"),
				"subnetId":      property.New("subnet-id"),
			}),
			expected: "yandex:index/computeInstance:ComputeInstance",
			validate: func(t *testing.T, inputs property.Map) {
				assertBootstrapUserData(t, inputs.Get("metadata").AsMap().Get("user-data").AsString())
			},
		},
	}
}

func TestProxmoxveConstructCreatesSnippetBackedVmClone(t *testing.T) {
	resources := constructAndCollectResources(t, "netskope-publisher:index:ProxmoxvePublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"nodeName":      property.New("pve-1"),
		"datastoreId":   property.New("local"),
		"templateVmId":  property.New(9000.0),
		"vmId":          property.New(101.0),
		"networkBridge": property.New("vmbr1"),
		"ipAddress":     property.New("10.10.0.50/24"),
		"gateway":       property.New("10.10.0.1"),
	}))

	file := findResourceByType(t, resources, "proxmoxve:index/fileLegacy:FileLegacy")
	if file.Inputs.Get("contentType").AsString() != "snippets" {
		t.Fatalf("expected Proxmox VE user data to be uploaded as a snippet")
	}

	vm := findResourceByType(t, resources, "proxmoxve:index/vmLegacy:VmLegacy")
	if vm.Inputs.Get("clone").AsMap().Get("vmId").AsNumber() != 9000 {
		t.Fatalf("expected Proxmox VE VM to clone from template VM 9000")
	}
	if vm.Inputs.Get("initialization").AsMap().Get("userDataFileId").IsNull() {
		t.Fatalf("expected Proxmox VE VM to reference the cloud-init snippet")
	}
	ipConfig := vm.Inputs.Get("initialization").AsMap().Get("ipConfigs").AsArray().AsSlice()[0].AsMap()
	if ipConfig.Get("ipv4").AsMap().Get("address").AsString() != "10.10.0.50/24" {
		t.Fatalf("expected Proxmox VE VM to set cloud-init IPv4 config")
	}
	if vm.Inputs.Get("networkDevices").AsArray().AsSlice()[0].AsMap().Get("bridge").AsString() != "vmbr1" {
		t.Fatalf("expected Proxmox VE VM to use requested bridge")
	}
}

func TestOpenstackConstructAssociatesFloatingIP(t *testing.T) {
	resources := constructAndCollectResources(t, "netskope-publisher:index:OpenstackPublisher", property.NewMap(map[string]property.Value{
		"names":            property.New([]property.Value{property.New("pub-1")}),
		"registrations":    registrationMap("pub-1"),
		"imageName":        property.New("Ubuntu 22.04"),
		"flavorName":       property.New("m1.medium"),
		"networkName":      property.New("private"),
		"assignFloatingIp": property.New(true),
		"floatingIpPool":   property.New("public"),
	}))

	if findResourceByType(t, resources, "openstack:networking/floatingIp:FloatingIp").Inputs.Get("pool").AsString() != "public" {
		t.Fatalf("expected OpenStack floating IP to use requested pool")
	}

	association := findResourceByType(t, resources, "openstack:networking/floatingIpAssociate:FloatingIpAssociate")
	if _, ok := association.Inputs.GetOk("floatingIp"); !ok {
		t.Fatalf("expected OpenStack floating IP association to set floatingIp")
	}
	if _, ok := association.Inputs.GetOk("portId"); !ok {
		t.Fatalf("expected OpenStack floating IP association to set portId")
	}
}

type capturedResource struct {
	Type   string
	Name   string
	Inputs property.Map
}

func constructAndCollectTypes(t *testing.T, token string, inputs property.Map) []string {
	t.Helper()
	resources := constructAndCollectResources(t, token, inputs)
	createdTypes := make([]string, 0, len(resources))
	for _, resource := range resources {
		createdTypes = append(createdTypes, resource.Type)
	}
	return createdTypes
}

func constructAndCollectResources(t *testing.T, token string, inputs property.Map) []capturedResource {
	t.Helper()

	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	var createdResources []capturedResource
	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.1.0"),
		integration.WithProvider(provider),
		integration.WithMocks(&integration.MockResourceMonitor{
			NewResourceF: func(args integration.MockResourceArgs) (string, property.Map, error) {
				createdResources = append(createdResources, capturedResource{
					Type:   string(args.TypeToken),
					Name:   args.Name,
					Inputs: args.Inputs,
				})
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

	return createdResources
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

func findResourceByType(t *testing.T, resources []capturedResource, expectedType string) capturedResource {
	t.Helper()
	for _, resource := range resources {
		if resource.Type == expectedType {
			return resource
		}
	}
	t.Fatalf("expected construct to create %s, got %v", expectedType, resourceTypes(resources))
	return capturedResource{}
}

func resourceTypes(resources []capturedResource) []string {
	types := make([]string, 0, len(resources))
	for _, resource := range resources {
		types = append(types, resource.Type)
	}
	return types
}

func assertBootstrapUserData(t *testing.T, userData string) {
	t.Helper()
	for _, expected := range []string{
		"#cloud-config",
		"curl -fsSL https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/generic/bootstrap.sh | sudo bash",
		"sudo /home/ubuntu/npa_publisher_wizard -token \"token\"",
	} {
		if !strings.Contains(userData, expected) {
			t.Fatalf("expected bootstrap user data to contain %q, got:\n%s", expected, userData)
		}
	}
}

func decodeRequiredBase64(t *testing.T, value string) string {
	t.Helper()
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		t.Fatalf("expected base64-encoded value, got %q: %v", value, err)
	}
	return string(decoded)
}

func assertKeyValueArrayHas(t *testing.T, values []property.Value, expectedKey string, expectedValue string) {
	t.Helper()
	value := keyValueArrayValue(t, values, expectedKey)
	if value != expectedValue {
		t.Fatalf("expected %s to be %q, got %q", expectedKey, expectedValue, value)
	}
}

func keyValueArrayValue(t *testing.T, values []property.Value, expectedKey string) string {
	t.Helper()
	for _, item := range values {
		itemMap := item.AsMap()
		if itemMap.Get("key").AsString() == expectedKey {
			return itemMap.Get("value").AsString()
		}
	}
	t.Fatalf("expected key/value array to contain %s", expectedKey)
	return ""
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
