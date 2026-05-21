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
	"github.com/pulumi/pulumi-go-provider/infer"
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

func TestPrivateAppCreateFailsWhenExistingWithoutAdopt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v2/steering/apps/private" {
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data": map[string]any{
					"private_apps": []map[string]any{{
						"app_id":   44,
						"app_name": "orders",
						"name":     "orders",
						"tags":     []map[string]any{},
					}},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	_, err := createPrivateAppResource(t, property.NewMap(map[string]property.Value{
		"tenantUrl":            property.New(server.URL),
		"bearerToken":          property.New("api-token"),
		"appName":              property.New("orders"),
		"appType":              property.New("client"),
		"host":                 property.New("orders.internal"),
		"protocols":            property.New([]property.Value{property.New(map[string]property.Value{"type": property.New("tcp"), "ports": property.New("443")})}),
		"clientlessAccess":     property.New(false),
		"isUserPortalApp":      property.New(false),
		"usePublisherDns":      property.New(false),
		"trustSelfSignedCerts": property.New(false),
	}))
	if err == nil {
		t.Fatalf("expected existing app error")
	}
	if !strings.Contains(err.Error(), "already exists") || !strings.Contains(err.Error(), "adoptExisting") {
		t.Fatalf("expected adopt hint, got %v", err)
	}
}

func TestPrivateAppAdoptsExistingByName(t *testing.T) {
	var patched bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/steering/apps/private":
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data": map[string]any{
					"private_apps": []map[string]any{{
						"app_id":   44,
						"app_name": "orders",
						"name":     "orders",
						"tags":     []map[string]any{{"tag_id": 7, "tag_name": "vpc-a"}},
					}},
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/api/v2/steering/apps/private/44":
			patched = true
			writeJSON(t, w, map[string]any{"status": "success", "data": map[string]any{"app_id": 44, "app_name": "orders", "name": "orders"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	response, err := createPrivateAppResource(t, property.NewMap(map[string]property.Value{
		"tenantUrl":            property.New(server.URL),
		"bearerToken":          property.New("api-token"),
		"appName":              property.New("orders"),
		"appType":              property.New("client"),
		"host":                 property.New("orders.internal"),
		"protocols":            property.New([]property.Value{property.New(map[string]property.Value{"type": property.New("tcp"), "ports": property.New("443")})}),
		"clientlessAccess":     property.New(false),
		"isUserPortalApp":      property.New(false),
		"usePublisherDns":      property.New(false),
		"trustSelfSignedCerts": property.New(false),
		"tags":                 property.New([]property.Value{property.New("vpc-a")}),
		"adoptExisting":        property.New(true),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "44" {
		t.Fatalf("expected adopted ID 44, got %q", response.ID)
	}
	if !patched {
		t.Fatalf("expected adopted app to be reconciled with PUT")
	}
}

func TestPrivateAppReadDropsResourceWhenRemoteAppIsMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v2/steering/apps/private/44" {
			http.NotFound(w, r)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	resource := PrivateApp{}
	response, err := resource.Read(t.Context(), infer.ReadRequest[PrivateAppArgs, PrivateAppOutputs]{
		ID: "44",
		Inputs: PrivateAppArgs{
			TenantURL:   server.URL,
			BearerToken: stringPtr("api-token"),
			AppName:     "orders",
		},
		State: PrivateAppOutputs{
			PrivateAppArgs: PrivateAppArgs{
				TenantURL:   server.URL,
				BearerToken: stringPtr("api-token"),
				AppName:     "orders",
			},
			AppID: 44,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "" {
		t.Fatalf("expected missing remote app to clear ID, got %q", response.ID)
	}
}

func TestPrivateAppRejectsHostsUntilApiShapeIsConfirmed(t *testing.T) {
	_, err := createPrivateAppResource(t, property.NewMap(map[string]property.Value{
		"tenantUrl":            property.New("https://tenant.example"),
		"bearerToken":          property.New("api-token"),
		"appName":              property.New("orders"),
		"appType":              property.New("client"),
		"host":                 property.New("orders.internal"),
		"hosts":                property.New([]property.Value{property.New("orders.internal"), property.New("orders-alt.internal")}),
		"protocols":            property.New([]property.Value{property.New(map[string]property.Value{"type": property.New("tcp"), "ports": property.New("443")})}),
		"clientlessAccess":     property.New(false),
		"isUserPortalApp":      property.New(false),
		"usePublisherDns":      property.New(false),
		"trustSelfSignedCerts": property.New(false),
	}))
	if err == nil {
		t.Fatalf("expected hosts validation error")
	}
	if !strings.Contains(err.Error(), "hosts is not supported by the documented private app API; use host") {
		t.Fatalf("expected hosts validation error, got %v", err)
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

func TestAwsConstructIncludesPublisherPlacementLabels(t *testing.T) {
	response := constructPublisherResource(t, "netskope-publisher:index:AwsPublisher", property.NewMap(map[string]property.Value{
		"names":            property.New([]property.Value{property.New("pub-1")}),
		"registrations":    registrationMap("pub-1"),
		"subnetId":         property.New("subnet-123"),
		"securityGroupIds": property.New([]property.Value{property.New("sg-123")}),
		"amiId":            property.New("ami-123"),
		"placementLabels":  property.New([]property.Value{property.New("vpc-a")}),
	}))

	publishers := response.State.Get("publishers").AsMap()
	pub := publishers.Get("pub-1").AsMap()
	labels := pub.Get("placementLabels").AsArray()
	if labels.Len() != 1 || labels.Get(0).AsString() != "vpc-a" {
		t.Fatalf("expected placementLabels [vpc-a], got %#v", labels)
	}
}

func TestTagPublisherAssignmentAddsAndRemovesOnlySelectedPublishers(t *testing.T) {
	var putBodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/steering/apps/private":
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data": map[string]any{
					"private_apps": []map[string]any{
						{
							"app_id":   10,
							"app_name": "orders",
							"tags":     []map[string]any{{"tag_name": "vpc-a"}},
							"service_publisher_assignments": []map[string]any{
								{"publisher_id": 99},
							},
						},
						{
							"app_id":   20,
							"app_name": "billing",
							"tags":     []map[string]any{{"tag_name": "vpc-b"}},
							"service_publisher_assignments": []map[string]any{
								{"publisher_id": 101},
								{"publisher_id": 99},
							},
						},
					},
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/api/v2/steering/apps/private/publishers":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			putBodies = append(putBodies, body)
			writeJSON(t, w, map[string]any{"status": "success", "data": body})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	response, err := createTagPublisherAssignmentResource(t, property.NewMap(map[string]property.Value{
		"tenantUrl":                property.New(server.URL),
		"bearerToken":              property.New("api-token"),
		"appTags":                  property.New([]property.Value{property.New("vpc-a")}),
		"publisherPlacementLabels": property.New([]property.Value{property.New("vpc-a")}),
		"publishers": property.New(map[string]property.Value{
			"pub-a": property.New(map[string]property.Value{
				"publisherId":     property.New(101.0),
				"placementLabels": property.New([]property.Value{property.New("vpc-a")}),
			}),
			"pub-b": property.New(map[string]property.Value{
				"publisherId":     property.New(202.0),
				"placementLabels": property.New([]property.Value{property.New("vpc-b")}),
			}),
		}),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(putBodies) != 2 {
		t.Fatalf("expected two publisher association updates, got %#v", putBodies)
	}
	firstBody := putBodies[0]
	if _, ok := firstBody["private_app_names"]; ok {
		t.Fatalf("did not expect private_app_names in publisher assignment body: %#v", firstBody)
	}
	if got := firstBody["private_app_ids"].([]any)[0]; got != "10" {
		t.Fatalf("expected first update to target private app ID 10, got %#v", firstBody)
	}
	if got := firstBody["publisher_ids"].([]any); len(got) != 2 || got[0] != "99" || got[1] != "101" {
		t.Fatalf("expected first update to keep 99 and add 101, got %#v", firstBody)
	}
	matchedApps := response.Properties.Get("matchedApps").AsArray()
	if matchedApps.Len() != 1 || matchedApps.Get(0).AsString() != "orders" {
		t.Fatalf("expected matched app orders, got %#v", matchedApps)
	}
}

func TestTagPublisherAssignmentFailsWhenPlacementLabelsSelectNoPublishers(t *testing.T) {
	_, err := createTagPublisherAssignmentResource(t, property.NewMap(map[string]property.Value{
		"tenantUrl":                property.New("https://tenant.example"),
		"bearerToken":              property.New("api-token"),
		"appTags":                  property.New([]property.Value{property.New("vpc-a")}),
		"publisherPlacementLabels": property.New([]property.Value{property.New("vpc-a")}),
		"publishers": property.New(map[string]property.Value{
			"pub-b": property.New(map[string]property.Value{
				"publisherId":     property.New(202.0),
				"placementLabels": property.New([]property.Value{property.New("vpc-b")}),
			}),
		}),
	}))
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), `publisherPlacementLabels [vpc-a] did not match any managed publishers`) {
		t.Fatalf("expected placement label validation error, got %v", err)
	}
}

func TestTagPublisherAssignmentDeleteRemovesSelectedPublishersFromMatchedApps(t *testing.T) {
	var deleteBodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/steering/apps/private":
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data": map[string]any{
					"private_apps": []map[string]any{{
						"app_id":   10,
						"app_name": "orders",
						"tags":     []map[string]any{{"tag_name": "vpc-a"}},
						"service_publisher_assignments": []map[string]any{
							{"publisher_id": 99},
							{"publisher_id": 101},
						},
					}},
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/steering/apps/private/publishers":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			deleteBodies = append(deleteBodies, body)
			writeJSON(t, w, map[string]any{"status": "success"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resource := TagPublisherAssignment{}
	_, err := resource.Delete(t.Context(), infer.DeleteRequest[TagPublisherAssignmentOutputs]{
		ID: "vpc-a",
		State: TagPublisherAssignmentOutputs{
			TagPublisherAssignmentArgs: TagPublisherAssignmentArgs{
				TenantURL:                server.URL,
				BearerToken:              stringPtr("api-token"),
				AppTags:                  []string{"vpc-a"},
				PublisherPlacementLabels: []string{"vpc-a"},
				Publishers: map[string]PublisherAssignmentInput{
					"pub-a": {PublisherID: 101, PlacementLabels: []string{"vpc-a"}},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(deleteBodies) != 1 {
		t.Fatalf("expected one delete request, got %#v", deleteBodies)
	}
	body := deleteBodies[0]
	if got := body["private_app_ids"].([]any)[0]; got != "10" {
		t.Fatalf("expected delete for private app ID 10, got %#v", body)
	}
	if got := body["publisher_ids"].([]any); len(got) != 1 || got[0] != "101" {
		t.Fatalf("expected delete for selected publisher 101 only, got %#v", body)
	}
}

func TestTagPublisherAssignmentReadRecomputesRemoteState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/steering/apps/private":
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data": map[string]any{
					"private_apps": []map[string]any{{
						"app_id":   10,
						"app_name": "orders",
						"tags":     []map[string]any{{"tag_name": "vpc-a"}},
						"service_publisher_assignments": []map[string]any{
							{"publisher_id": 101},
						},
					}},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resource := TagPublisherAssignment{}
	response, err := resource.Read(t.Context(), infer.ReadRequest[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs]{
		ID: "vpc-a",
		Inputs: TagPublisherAssignmentArgs{
			TenantURL:                server.URL,
			BearerToken:              stringPtr("api-token"),
			AppTags:                  []string{"vpc-a"},
			PublisherPlacementLabels: []string{"vpc-a"},
			Publishers: map[string]PublisherAssignmentInput{
				"pub-a": {PublisherID: 101, PlacementLabels: []string{"vpc-a"}},
			},
		},
		State: TagPublisherAssignmentOutputs{
			MatchedApps:        []string{"stale"},
			SelectedPublishers: []int{202},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "vpc-a" {
		t.Fatalf("expected read ID vpc-a, got %q", response.ID)
	}
	if len(response.State.MatchedApps) != 1 || response.State.MatchedApps[0] != "orders" {
		t.Fatalf("expected matched app orders after refresh, got %#v", response.State.MatchedApps)
	}
	if len(response.State.SelectedPublishers) != 1 || response.State.SelectedPublishers[0] != 101 {
		t.Fatalf("expected selected publisher 101 after refresh, got %#v", response.State.SelectedPublishers)
	}
}

func TestRealtimeProtectionPolicyCreatesRuleWithPolicyGroupReference(t *testing.T) {
	var created map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/policy/npa/policygroups":
			writeJSON(t, w, map[string]any{"status": "success", "data": []map[string]any{{"id": 12, "name": "default"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/policy/npa/rules":
			if err := json.NewDecoder(r.Body).Decode(&created); err != nil {
				t.Fatal(err)
			}
			writeJSON(t, w, map[string]any{
				"rule_id":   55,
				"rule_name": "orders-access",
				"rule_data": map[string]any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	response, err := createRealtimeProtectionPolicyResource(t, property.NewMap(map[string]property.Value{
		"tenantUrl":       property.New(server.URL),
		"bearerToken":     property.New("api-token"),
		"name":            property.New("orders-access"),
		"policyGroupName": property.New("default"),
		"appIds":          property.New([]property.Value{property.New(44.0)}),
		"appTags":         property.New([]property.Value{property.New("vpc-a")}),
		"users":           property.New([]property.Value{property.New("user@example.com")}),
		"groups":          property.New([]property.Value{property.New("CN=npa-users,OU=Groups,DC=example,DC=com")}),
		"action":          property.New("allow"),
		"enabled":         property.New(true),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "55" {
		t.Fatalf("expected policy ID 55, got %q", response.ID)
	}
	if created["rule_name"] != "orders-access" {
		t.Fatalf("expected rule_name in payload, got %#v", created)
	}
	if created["group_id"] != "12" {
		t.Fatalf("expected group_id string 12 in payload, got %#v", created)
	}
	if created["enabled"] != "1" {
		t.Fatalf("expected enabled string 1 in payload, got %#v", created)
	}
	ruleData := created["rule_data"].(map[string]any)
	if got := ruleData["privateApps"].([]any)[0]; got != "44" {
		t.Fatalf("expected privateApps [44] as strings, got %#v", ruleData["privateApps"])
	}
	if got := ruleData["privateAppTags"].([]any)[0]; got != "vpc-a" {
		t.Fatalf("expected privateAppTags [vpc-a], got %#v", ruleData["privateAppTags"])
	}
	if got := ruleData["users"].([]any)[0]; got != "user@example.com" {
		t.Fatalf("expected users in rule_data, got %#v", ruleData["users"])
	}
	if got := ruleData["userGroups"].([]any)[0]; got != "CN=npa-users,OU=Groups,DC=example,DC=com" {
		t.Fatalf("expected userGroups in rule_data, got %#v", ruleData["userGroups"])
	}
	action := ruleData["match_criteria_action"].(map[string]any)
	if got := action["action_name"]; got != "allow" {
		t.Fatalf("expected action_name allow, got %#v", action)
	}
}

func TestRealtimeProtectionPolicyDryRunDoesNotResolvePolicyGroupName(t *testing.T) {
	resource := RealtimeProtectionPolicy{}
	response, err := resource.Create(t.Context(), infer.CreateRequest[RealtimeProtectionPolicyArgs]{
		DryRun: true,
		Inputs: RealtimeProtectionPolicyArgs{
			TenantURL:       "https://unused.example",
			BearerToken:     stringPtr("api-token"),
			Name:            "orders-access",
			PolicyGroupName: stringPtr("default"),
			AppTags:         []string{"vpc-a"},
			Action:          "allow",
			Enabled:         true,
		},
	})
	if err != nil {
		t.Fatalf("expected dry-run create without live lookup, got %v", err)
	}
	if response.ID != "orders-access" {
		t.Fatalf("expected dry-run ID to use policy name, got %q", response.ID)
	}
	if response.Output.ResolvedPolicyGroupID != 0 {
		t.Fatalf("expected unresolved policy group during dry-run, got %d", response.Output.ResolvedPolicyGroupID)
	}
}

func TestRealtimeProtectionPolicyReadParsesPolicyEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/policy/npa/rules/55":
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data": map[string]any{
					"rule_id":   55,
					"rule_name": "orders-access",
					"rule_data": map[string]any{},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resource := RealtimeProtectionPolicy{}
	response, err := resource.Read(t.Context(), infer.ReadRequest[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{
		ID: "55",
		Inputs: RealtimeProtectionPolicyArgs{
			TenantURL:   server.URL,
			BearerToken: stringPtr("api-token"),
			Name:        "orders-access",
			Action:      "allow",
			Enabled:     true,
		},
		State: RealtimeProtectionPolicyOutputs{
			RealtimeProtectionPolicyArgs: RealtimeProtectionPolicyArgs{
				TenantURL:   server.URL,
				BearerToken: stringPtr("api-token"),
				Name:        "orders-access",
				Action:      "allow",
				Enabled:     true,
			},
			PolicyID: 55,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "55" {
		t.Fatalf("expected read ID 55, got %q", response.ID)
	}
	if response.State.PolicyID != 55 {
		t.Fatalf("expected state policy ID 55, got %d", response.State.PolicyID)
	}
}

func TestRealtimeProtectionPolicyUpdateParsesPolicyEnvelope(t *testing.T) {
	var patched map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v2/policy/npa/rules/55":
			if err := json.NewDecoder(r.Body).Decode(&patched); err != nil {
				t.Fatal(err)
			}
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data": map[string]any{
					"rule_id":   55,
					"rule_name": "orders-access",
					"rule_data": map[string]any{},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resource := RealtimeProtectionPolicy{}
	response, err := resource.Update(t.Context(), infer.UpdateRequest[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{
		ID: "55",
		Inputs: RealtimeProtectionPolicyArgs{
			TenantURL:     server.URL,
			BearerToken:   stringPtr("api-token"),
			Name:          "orders-access",
			PolicyGroupID: intPtr(12),
			AppIDs:        []int{44},
			Action:        "block",
			Enabled:       false,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Output.PolicyID != 55 {
		t.Fatalf("expected updated policy ID 55, got %d", response.Output.PolicyID)
	}
	if patched["group_id"] != "12" {
		t.Fatalf("expected group_id string 12 in update payload, got %#v", patched)
	}
	if patched["enabled"] != "0" {
		t.Fatalf("expected enabled string 0 in update payload, got %#v", patched)
	}
}

func TestRealtimeProtectionPolicyReadDropsResourceWhenRemotePolicyIsMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v2/policy/npa/rules/55" {
			http.NotFound(w, r)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	resource := RealtimeProtectionPolicy{}
	response, err := resource.Read(t.Context(), infer.ReadRequest[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{
		ID: "55",
		Inputs: RealtimeProtectionPolicyArgs{
			TenantURL:   server.URL,
			BearerToken: stringPtr("api-token"),
			Name:        "orders-access",
			Action:      "allow",
			Enabled:     true,
		},
		State: RealtimeProtectionPolicyOutputs{
			RealtimeProtectionPolicyArgs: RealtimeProtectionPolicyArgs{
				TenantURL:   server.URL,
				BearerToken: stringPtr("api-token"),
				Name:        "orders-access",
				Action:      "allow",
				Enabled:     true,
			},
			PolicyID: 55,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "" {
		t.Fatalf("expected missing remote policy to clear ID, got %q", response.ID)
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

func TestVultrConstructRejectsMissingImageChoice(t *testing.T) {
	err := constructPublisherResourceError(t, "netskope-publisher:index:VultrPublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"region":        property.New("ams"),
		"plan":          property.New("vc2-2c-4gb"),
	}))
	if err == nil || !strings.Contains(err.Error(), "VultrPublisher requires one of: osId, imageId") {
		t.Fatalf("expected Vultr missing image choice error, got %v", err)
	}
}

func TestVultrConstructRejectsConflictingImageChoices(t *testing.T) {
	err := constructPublisherResourceError(t, "netskope-publisher:index:VultrPublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"region":        property.New("ams"),
		"plan":          property.New("vc2-2c-4gb"),
		"osId":          property.New(1743.0),
		"imageId":       property.New("img-123"),
	}))
	if err == nil || !strings.Contains(err.Error(), "VultrPublisher accepts only one of: osId, imageId") {
		t.Fatalf("expected Vultr conflicting image choice error, got %v", err)
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

func constructPublisherResource(t *testing.T, token string, inputs property.Map) p.ConstructResponse {
	t.Helper()

	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.2.0"),
		integration.WithProvider(provider),
		integration.WithMocks(&integration.MockResourceMonitor{
			NewResourceF: func(args integration.MockResourceArgs) (string, property.Map, error) {
				if string(args.TypeToken) == "netskope-publisher:index:NetskopeRegistration" {
					return args.Name + "-id", property.NewMap(map[string]property.Value{
						"registrations": registrationMap("pub-1"),
					}), nil
				}
				return args.Name + "-id", args.Inputs, nil
			},
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	response, err := server.Construct(p.ConstructRequest{
		Urn:    presource.URN("urn:pulumi:stack::project::" + token + "::publisher"),
		Inputs: inputs,
	})
	if err != nil {
		t.Fatal(err)
	}

	return response
}

func constructPublisherResourceError(t *testing.T, token string, inputs property.Map) error {
	t.Helper()

	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.3.0"),
		integration.WithProvider(provider),
		integration.WithMocks(&integration.MockResourceMonitor{
			NewResourceF: func(args integration.MockResourceArgs) (string, property.Map, error) {
				return args.Name + "-id", args.Inputs, nil
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
	return err
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

func createPrivateAppResource(t *testing.T, inputs property.Map) (p.CreateResponse, error) {
	t.Helper()

	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.2.0"),
		integration.WithProvider(provider),
	)
	if err != nil {
		t.Fatal(err)
	}

	return server.Create(p.CreateRequest{
		Urn:        presource.URN("urn:pulumi:stack::project::netskope-publisher:index:PrivateApp::app"),
		Properties: inputs,
	})
}

func createTagPublisherAssignmentResource(t *testing.T, inputs property.Map) (p.CreateResponse, error) {
	t.Helper()

	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.2.0"),
		integration.WithProvider(provider),
	)
	if err != nil {
		t.Fatal(err)
	}

	return server.Create(p.CreateRequest{
		Urn:        presource.URN("urn:pulumi:stack::project::netskope-publisher:index:TagPublisherAssignment::assignment"),
		Properties: inputs,
	})
}

func createRealtimeProtectionPolicyResource(t *testing.T, inputs property.Map) (p.CreateResponse, error) {
	t.Helper()

	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.2.0"),
		integration.WithProvider(provider),
	)
	if err != nil {
		t.Fatal(err)
	}

	return server.Create(p.CreateRequest{
		Urn:        presource.URN("urn:pulumi:stack::project::netskope-publisher:index:RealtimeProtectionPolicy::policy"),
		Properties: inputs,
	})
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

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
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
