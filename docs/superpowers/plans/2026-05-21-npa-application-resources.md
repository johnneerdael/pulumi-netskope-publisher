# NPA Application Resources Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add focused Netskope Private Access application resources to the existing Pulumi provider: private apps, realtime protection policies, publisher placement labels, and tag-based publisher assignment reconciliation.

**Architecture:** Keep the Go `pulumi-go-provider` implementation as the schema source. Extract the existing registration HTTP code into a shared endpoint-specific Netskope client, then add one focused resource file per new NPA concept. Publisher components remain components, but gain Pulumi-side `placementLabels` metadata used by the assignment reconciler.

**Tech Stack:** Go, `pulumi-go-provider`, Pulumi component resources, `httptest`, generated Pulumi schema/SDKs, TypeScript examples/docs.

---

## File Structure

- Create `internal/provider/netskope_client.go`: shared auth, request, response parsing, publisher registration endpoints, private app endpoints, tag endpoints, publisher association endpoints, policy group lookup, and policy rule endpoints.
- Modify `internal/provider/registration.go`: keep `NetskopeRegistration` resource and registration orchestration; remove duplicated HTTP client internals after moving them to `netskope_client.go`.
- Create `internal/provider/private_app.go`: `PrivateApp` resource args, outputs, create/read/update/delete, explicit adoption behavior, dry-run behavior.
- Create `internal/provider/realtime_protection_policy.go`: `RealtimeProtectionPolicy` resource args, outputs, policy group resolution, create/read/update/delete.
- Create `internal/provider/tag_publisher_assignment.go`: `TagPublisherAssignment` resource args, outputs, reconciliation algorithm.
- Modify `internal/provider/types.go`: add shared NPA structs and `PlacementLabels` to `CommonPublisherArgs`, `PublisherOutput`, and any small output structs needed by assignment.
- Modify `internal/provider/components.go`: add `placementLabels` to every publisher arg struct, propagate through `common()` helpers, and include labels in each publisher output.
- Modify `internal/provider/provider.go`: register `PrivateApp`, `RealtimeProtectionPolicy`, and `TagPublisherAssignment`.
- Modify `internal/provider/provider_test.go`: add focused unit tests with `httptest`.
- Modify `README.md`, `site/source/admin/component/index.md`, and related docs pages after schema generation to describe the new NPA deployment flow.
- Regenerate `schema.json` and SDKs using existing scripts after Go resources compile.

The first implementation will not add a separate `PrivateAppTagAttachment` resource. Inline `PrivateApp` tags and `TagPublisherAssignment` cover the approved deployment flow.

---

### Task 1: Extract Shared Netskope Client

**Files:**
- Create: `internal/provider/netskope_client.go`
- Modify: `internal/provider/registration.go`
- Test: `internal/provider/provider_test.go`

- [ ] **Step 1: Write the failing regression test for existing registration behavior**

Add this test near the existing OAuth2 registration test in `internal/provider/provider_test.go`:

```go
func TestNetskopeClientReportsHTTPStatusAndBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"status":"error","message":"bad token"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newNetskopeClient(netskopeClientConfig{
		TenantURL:   server.URL,
		BearerToken: "bad-token",
		AuthMode:   "token",
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
```

- [ ] **Step 2: Run the focused test and verify it fails**

Run:

```bash
go test ./internal/provider -run TestNetskopeClientReportsHTTPStatusAndBody -count=1
```

Expected: compile failure because `netskopeClientConfig` does not exist or runtime failure because the current client does not preserve the response body.

- [ ] **Step 3: Create the shared client**

Create `internal/provider/netskope_client.go` with this structure, moving the existing request/auth helpers out of `registration.go`:

```go
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type netskopeClientConfig struct {
	TenantURL   string
	APIToken    *string
	BearerToken string
	AuthMode    string
	OAuth2      *NetskopeOAuth2Args
	HTTPClient  *http.Client
}

type netskopeClient struct {
	tenantURL   string
	bearerToken string
	authMode    string
	oauth2      *NetskopeOAuth2Args
	accessToken string
	httpClient  *http.Client
}

func newNetskopeClient(config netskopeClientConfig) netskopeClient {
	token := config.BearerToken
	if token == "" {
		token = stringValue(config.APIToken)
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return netskopeClient{
		tenantURL:    strings.TrimRight(config.TenantURL, "/"),
		bearerToken:  token,
		authMode:     defaultString(&config.AuthMode, "token"),
		oauth2:       config.OAuth2,
		httpClient:   httpClient,
	}
}

func (client *netskopeClient) endpoint(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return client.tenantURL + path
}

func (client *netskopeClient) request(ctx context.Context, operation string, method string, path string, body any, output any) error {
	token, err := client.resolveAccessToken(ctx)
	if err != nil {
		return err
	}

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}

	request, err := http.NewRequestWithContext(ctx, method, client.endpoint(path), reader)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("%s failed: %w", operation, err)
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("%s failed reading response: %w", operation, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s failed (status=%d): %s", operation, response.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	if output == nil || len(bodyBytes) == 0 {
		return nil
	}
	if err := json.Unmarshal(bodyBytes, output); err != nil {
		return fmt.Errorf("%s returned invalid JSON: %w", operation, err)
	}
	return nil
}
```

Also move the existing `resolveAccessToken` and `fetchOAuth2AccessToken` methods unchanged, except update field names from `client.client` to `client.httpClient`.

- [ ] **Step 4: Move publisher registration endpoints into the shared client**

Add these methods to `internal/provider/netskope_client.go`:

```go
func (client *netskopeClient) listPublishers(ctx context.Context) (map[string]int, error) {
	var response struct {
		Data struct {
			Publishers []struct {
				Name string      `json:"publisher_name"`
				ID   interface{} `json:"publisher_id"`
			} `json:"publishers"`
		} `json:"data"`
	}
	if err := client.request(ctx, "List publishers", http.MethodGet, "/api/v2/infrastructure/publishers", nil, &response); err != nil {
		return nil, err
	}
	publishers := make(map[string]int, len(response.Data.Publishers))
	for _, publisher := range response.Data.Publishers {
		id, err := parsePublisherID(publisher.ID)
		if err != nil {
			return nil, fmt.Errorf("List publishers returned invalid publisher ID for %s: %w", publisher.Name, err)
		}
		publishers[publisher.Name] = id
	}
	return publishers, nil
}

func (client *netskopeClient) createPublisher(ctx context.Context, name string) (int, error) {
	var response struct {
		Data struct {
			ID interface{} `json:"id"`
		} `json:"data"`
	}
	if err := client.request(ctx, "Create publisher "+name, http.MethodPost, "/api/v2/infrastructure/publishers", map[string]string{"name": name}, &response); err != nil {
		return 0, err
	}
	return parsePublisherID(response.Data.ID)
}

func (client *netskopeClient) generateRegistrationToken(ctx context.Context, publisherID int) (string, error) {
	var response struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	path := fmt.Sprintf("/api/v2/infrastructure/publishers/%d/registration_token", publisherID)
	if err := client.request(ctx, fmt.Sprintf("Generate registration token for publisher %d", publisherID), http.MethodPost, path, nil, &response); err != nil {
		return "", err
	}
	return response.Data.Token, nil
}
```

- [ ] **Step 5: Update registration to use the new client config**

In `internal/provider/registration.go`, replace the old `newNetskopeClient(...)` call with:

```go
netskopeClient := newNetskopeClient(netskopeClientConfig{
	TenantURL:    args.TenantURL,
	APIToken:     args.APIToken,
	BearerToken:  stringValue(args.BearerToken),
	AuthMode:     defaultString(args.AuthMode, "token"),
	OAuth2:       args.OAuth2,
	HTTPClient:   client,
})
```

Remove the old `netskopeClient` type, `newNetskopeClient` function, `listPublishers`, `createPublisher`, `generateRegistrationToken`, `request`, `resolveAccessToken`, and `fetchOAuth2AccessToken` from `registration.go`. Keep `parsePublisherID` and shared small helpers wherever they currently live unless moving them is needed for compilation.

- [ ] **Step 6: Run registration tests**

Run:

```bash
go test ./internal/provider -run 'TestNetskopeRegistration|TestNetskopeClient' -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/provider/netskope_client.go internal/provider/registration.go internal/provider/provider_test.go
git commit -m "refactor: share netskope api client"
```

---

### Task 2: Add Private App Client Methods and Resource

**Files:**
- Modify: `internal/provider/netskope_client.go`
- Create: `internal/provider/private_app.go`
- Modify: `internal/provider/provider.go`
- Test: `internal/provider/provider_test.go`

- [ ] **Step 1: Write failing tests for create and adoption**

Add these tests to `internal/provider/provider_test.go`:

```go
func TestPrivateAppCreateFailsWhenExistingWithoutAdopt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v2/steering/apps/private" {
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data": []map[string]any{{"app_id": 44, "app_name": "orders", "name": "orders", "tags": []map[string]any{}}},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	_, err := createPrivateAppResource(t, property.NewMap(map[string]property.Value{
		"tenantUrl":           property.New(server.URL),
		"bearerToken":         property.New("api-token"),
		"appName":             property.New("orders"),
		"appType":             property.New("client"),
		"host":                property.New("orders.internal"),
		"protocols":           property.New([]property.Value{property.New(map[string]property.Value{"type": property.New("tcp"), "ports": property.New("443")})}),
		"clientlessAccess":    property.New(false),
		"isUserPortalApp":     property.New(false),
		"usePublisherDns":     property.New(false),
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
				"data": []map[string]any{{"app_id": 44, "app_name": "orders", "name": "orders", "tags": []map[string]any{{"tag_id": 7, "tag_name": "vpc-a"}}}},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v2/steering/apps/private/44":
			patched = true
			writeJSON(t, w, map[string]any{"status": "success", "data": map[string]any{"app_id": 44, "app_name": "orders", "name": "orders"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	response, err := createPrivateAppResource(t, property.NewMap(map[string]property.Value{
		"tenantUrl":           property.New(server.URL),
		"bearerToken":         property.New("api-token"),
		"appName":             property.New("orders"),
		"appType":             property.New("client"),
		"host":                property.New("orders.internal"),
		"protocols":           property.New([]property.Value{property.New(map[string]property.Value{"type": property.New("tcp"), "ports": property.New("443")})}),
		"clientlessAccess":    property.New(false),
		"isUserPortalApp":     property.New(false),
		"usePublisherDns":     property.New(false),
		"trustSelfSignedCerts": property.New(false),
		"tags":                property.New([]property.Value{property.New("vpc-a")}),
		"adoptExisting":       property.New(true),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "44" {
		t.Fatalf("expected adopted ID 44, got %q", response.ID)
	}
	if !patched {
		t.Fatalf("expected adopted app to be reconciled with PATCH")
	}
}
```

Add helper:

```go
func createPrivateAppResource(t *testing.T, inputs property.Map) (p.CreateResponse, error) {
	t.Helper()
	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}
	server, err := integration.NewServer(t.Context(), Name, semver.MustParse("0.2.0"), integration.WithProvider(provider))
	if err != nil {
		t.Fatal(err)
	}
	return server.Create(p.CreateRequest{
		Urn:        presource.NewURN("stack", "project", "", "netskope-publisher:index:PrivateApp", "app"),
		Properties: inputs,
	})
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run:

```bash
go test ./internal/provider -run 'TestPrivateApp' -count=1
```

Expected: FAIL because `PrivateApp` is not registered.

- [ ] **Step 3: Add private app client types and methods**

Add to `internal/provider/netskope_client.go`:

```go
type privateAppProtocol struct {
	Type  string `json:"type"`
	Ports string `json:"ports,omitempty"`
	Port  string `json:"port,omitempty"`
}

type privateAppTag struct {
	TagID   int    `json:"tag_id,omitempty"`
	TagName string `json:"tag_name"`
}

type privateAppPayload struct {
	AppName              string               `json:"app_name"`
	AppType              string               `json:"app_type,omitempty"`
	Host                 any                  `json:"host"`
	ClientlessAccess     bool                 `json:"clientless_access"`
	IsUserPortalApp      bool                 `json:"is_user_portal_app"`
	Protocols            []privateAppProtocol `json:"protocols"`
	TrustSelfSignedCerts bool                 `json:"trust_self_signed_certs"`
	UsePublisherDNS      bool                 `json:"use_publisher_dns"`
	PrivateAppTags       []privateAppTag      `json:"private_app_tags,omitempty"`
	Tags                 []privateAppTag      `json:"tags,omitempty"`
}

type privateAppRecord struct {
	AppID   int             `json:"app_id"`
	ID      int             `json:"id"`
	AppName string          `json:"app_name"`
	Name    string          `json:"name"`
	Host    any             `json:"host"`
	Tags    []privateAppTag `json:"tags"`
}

func (app privateAppRecord) resourceID() int {
	if app.AppID != 0 {
		return app.AppID
	}
	return app.ID
}

func (client *netskopeClient) listPrivateApps(ctx context.Context) ([]privateAppRecord, error) {
	var response struct {
		Status string             `json:"status"`
		Data   []privateAppRecord `json:"data"`
	}
	if err := client.request(ctx, "List private apps", http.MethodGet, "/api/v2/steering/apps/private", nil, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (client *netskopeClient) findPrivateAppByName(ctx context.Context, name string) (*privateAppRecord, error) {
	apps, err := client.listPrivateApps(ctx)
	if err != nil {
		return nil, err
	}
	for _, app := range apps {
		if app.AppName == name || app.Name == name {
			return &app, nil
		}
	}
	return nil, nil
}

func (client *netskopeClient) createPrivateApp(ctx context.Context, payload privateAppPayload) (privateAppRecord, error) {
	var response struct {
		Status string           `json:"status"`
		Data   privateAppRecord `json:"data"`
	}
	if err := client.request(ctx, "Create private app "+payload.AppName, http.MethodPost, "/api/v2/steering/apps/private", payload, &response); err != nil {
		return privateAppRecord{}, err
	}
	return response.Data, nil
}

func (client *netskopeClient) updatePrivateApp(ctx context.Context, id int, payload privateAppPayload) (privateAppRecord, error) {
	var response struct {
		Status string           `json:"status"`
		Data   privateAppRecord `json:"data"`
	}
	path := fmt.Sprintf("/api/v2/steering/apps/private/%d", id)
	if err := client.request(ctx, "Update private app "+payload.AppName, http.MethodPatch, path, payload, &response); err != nil {
		return privateAppRecord{}, err
	}
	return response.Data, nil
}

func (client *netskopeClient) deletePrivateApp(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v2/steering/apps/private/%d", id)
	return client.request(ctx, fmt.Sprintf("Delete private app %d", id), http.MethodDelete, path, nil, nil)
}
```

- [ ] **Step 4: Add the PrivateApp resource**

Create `internal/provider/private_app.go`:

```go
package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

type PrivateApp struct{}

type PrivateAppProtocol struct {
	Type  string `pulumi:"type"`
	Ports string `pulumi:"ports"`
}

type PrivateAppArgs struct {
	TenantURL            string              `pulumi:"tenantUrl"`
	APIToken             *string             `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken          *string             `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode             *string             `pulumi:"authMode,optional"`
	OAuth2               *NetskopeOAuth2Args `pulumi:"oauth2,optional"`
	AppName              string              `pulumi:"appName"`
	AppType              *string             `pulumi:"appType,optional"`
	Host                 string              `pulumi:"host"`
	Hosts                []string            `pulumi:"hosts,optional"`
	Protocols            []PrivateAppProtocol `pulumi:"protocols"`
	ClientlessAccess     bool                `pulumi:"clientlessAccess"`
	IsUserPortalApp      bool                `pulumi:"isUserPortalApp"`
	UsePublisherDNS      bool                `pulumi:"usePublisherDns"`
	TrustSelfSignedCerts bool                `pulumi:"trustSelfSignedCerts"`
	Tags                 []string            `pulumi:"tags,optional"`
	AdoptExisting        *bool               `pulumi:"adoptExisting,optional"`
}

type PrivateAppOutputs struct {
	PrivateAppArgs
	AppID int `pulumi:"appId"`
}

func (*PrivateApp) Annotate(a infer.Annotator) {
	a.SetToken("index", "PrivateApp")
}

func (*PrivateApp) Create(ctx context.Context, req infer.CreateRequest[PrivateAppArgs]) (infer.CreateResponse[PrivateAppOutputs], error) {
	output := PrivateAppOutputs{PrivateAppArgs: req.Inputs}
	if req.DryRun {
		return infer.CreateResponse[PrivateAppOutputs]{ID: req.Inputs.AppName, Output: output}, nil
	}
	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	existing, err := client.findPrivateAppByName(ctx, req.Inputs.AppName)
	if err != nil {
		return infer.CreateResponse[PrivateAppOutputs]{}, err
	}
	payload := privateAppPayloadFromArgs(req.Inputs)
	if existing != nil {
		if !defaultBool(req.Inputs.AdoptExisting, false) {
			return infer.CreateResponse[PrivateAppOutputs]{}, fmt.Errorf("private app %q already exists; import it or set adoptExisting: true to manage it", req.Inputs.AppName)
		}
		updated, err := client.updatePrivateApp(ctx, existing.resourceID(), payload)
		if err != nil {
			return infer.CreateResponse[PrivateAppOutputs]{}, err
		}
		output.AppID = updated.resourceID()
		return infer.CreateResponse[PrivateAppOutputs]{ID: strconv.Itoa(output.AppID), Output: output}, nil
	}
	created, err := client.createPrivateApp(ctx, payload)
	if err != nil {
		return infer.CreateResponse[PrivateAppOutputs]{}, err
	}
	output.AppID = created.resourceID()
	return infer.CreateResponse[PrivateAppOutputs]{ID: strconv.Itoa(output.AppID), Output: output}, nil
}
```

Add `Read`, `Update`, and `Delete` methods in the same file:

```go
func (*PrivateApp) Read(ctx context.Context, req infer.ReadRequest[PrivateAppArgs, PrivateAppOutputs]) (infer.ReadResponse[PrivateAppArgs, PrivateAppOutputs], error) {
	return infer.ReadResponse[PrivateAppArgs, PrivateAppOutputs]{ID: req.ID, Inputs: req.Inputs, State: req.State}, nil
}

func (*PrivateApp) Update(ctx context.Context, req infer.UpdateRequest[PrivateAppArgs, PrivateAppOutputs]) (infer.UpdateResponse[PrivateAppOutputs], error) {
	output := PrivateAppOutputs{PrivateAppArgs: req.Inputs}
	appID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.UpdateResponse[PrivateAppOutputs]{}, fmt.Errorf("invalid private app ID %q: %w", req.ID, err)
	}
	output.AppID = appID
	if req.DryRun {
		return infer.UpdateResponse[PrivateAppOutputs]{Output: output}, nil
	}
	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	updated, err := client.updatePrivateApp(ctx, appID, privateAppPayloadFromArgs(req.Inputs))
	if err != nil {
		return infer.UpdateResponse[PrivateAppOutputs]{}, err
	}
	output.AppID = updated.resourceID()
	return infer.UpdateResponse[PrivateAppOutputs]{Output: output}, nil
}

func (*PrivateApp) Delete(ctx context.Context, req infer.DeleteRequest[PrivateAppOutputs]) (infer.DeleteResponse, error) {
	appID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid private app ID %q: %w", req.ID, err)
	}
	client := newResourceClient(req.State.TenantURL, req.State.APIToken, req.State.BearerToken, req.State.AuthMode, req.State.OAuth2, http.DefaultClient)
	return infer.DeleteResponse{}, client.deletePrivateApp(ctx, appID)
}
```

Add helpers:

```go
func newResourceClient(tenantURL string, apiToken *string, bearerToken *string, authMode *string, oauth2 *NetskopeOAuth2Args, httpClient *http.Client) netskopeClient {
	return newNetskopeClient(netskopeClientConfig{
		TenantURL:    tenantURL,
		APIToken:     apiToken,
		BearerToken:  stringValue(bearerToken),
		AuthMode:     defaultString(authMode, "token"),
		OAuth2:       oauth2,
		HTTPClient:   httpClient,
	})
}

func privateAppPayloadFromArgs(args PrivateAppArgs) privateAppPayload {
	host := any(args.Host)
	if len(args.Hosts) > 0 {
		host = args.Hosts
	}
	protocols := make([]privateAppProtocol, 0, len(args.Protocols))
	for _, protocol := range args.Protocols {
		protocols = append(protocols, privateAppProtocol{Type: protocol.Type, Port: protocol.Ports, Ports: protocol.Ports})
	}
	tags := make([]privateAppTag, 0, len(args.Tags))
	for _, tag := range args.Tags {
		tags = append(tags, privateAppTag{TagName: tag})
	}
	return privateAppPayload{
		AppName:              args.AppName,
		AppType:              defaultString(args.AppType, "client"),
		Host:                 host,
		ClientlessAccess:     args.ClientlessAccess,
		IsUserPortalApp:      args.IsUserPortalApp,
		Protocols:            protocols,
		TrustSelfSignedCerts: args.TrustSelfSignedCerts,
		UsePublisherDNS:      args.UsePublisherDNS,
		PrivateAppTags:       tags,
		Tags:                 tags,
	}
}
```

- [ ] **Step 5: Register the resource**

Modify `internal/provider/provider.go`:

```go
WithResources(
	infer.Resource(&NetskopeRegistration{}),
	infer.Resource(&PrivateApp{}),
)
```

- [ ] **Step 6: Run focused tests**

Run:

```bash
go test ./internal/provider -run 'TestPrivateApp' -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/provider/netskope_client.go internal/provider/private_app.go internal/provider/provider.go internal/provider/provider_test.go
git commit -m "feat: add private app resource"
```

---

### Task 3: Add Publisher Placement Labels

**Files:**
- Modify: `internal/provider/types.go`
- Modify: `internal/provider/components.go`
- Test: `internal/provider/provider_test.go`

- [ ] **Step 1: Write failing placement label output test**

Add to `internal/provider/provider_test.go`:

```go
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
	if len(labels) != 1 || labels[0].AsString() != "vpc-a" {
		t.Fatalf("expected placementLabels [vpc-a], got %#v", labels)
	}
}
```

Add this helper near `constructAndCollectTypes`:

```go
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
```

- [ ] **Step 2: Run the test and verify it fails**

Run:

```bash
go test ./internal/provider -run TestAwsConstructIncludesPublisherPlacementLabels -count=1
```

Expected: FAIL because `placementLabels` is not in the schema/output.

- [ ] **Step 3: Add shared placement fields**

Modify `internal/provider/types.go`:

```go
type CommonPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	// existing fields stay below
}

type PublisherOutput struct {
	PublisherID       pulumi.IntOutput    `pulumi:"publisherId"`
	RegistrationToken pulumi.StringOutput `pulumi:"registrationToken" provider:"secret"`
	VMID              pulumi.StringOutput `pulumi:"vmId"`
	PrivateIP         pulumi.StringOutput `pulumi:"privateIp"`
	PublicIP          pulumi.StringOutput `pulumi:"publicIp"`
	PlacementLabels   pulumi.StringArrayOutput `pulumi:"placementLabels"`
}
```

- [ ] **Step 4: Update publisher output helper**

Find `publisherOutput` in `internal/provider/components.go` and change its signature to accept labels:

```go
func publisherOutput(registration PublisherRegistrationInput, vmID pulumi.StringOutput, privateIP pulumi.StringOutput, publicIP pulumi.StringOutput, placementLabels []string) pulumi.Map {
	return pulumi.Map{
		"publisherId":       pulumi.Int(registration.PublisherID),
		"registrationToken": pulumi.ToSecret(pulumi.String(registration.RegistrationToken)),
		"vmId":              vmID,
		"privateIp":         privateIP,
		"publicIp":          publicIP,
		"placementLabels":   toStringArray(placementLabels),
	}
}
```

Update each call from:

```go
publisherOutput(registration, instance.ID().ToStringOutput(), instance.PrivateIp, instance.PublicIp)
```

to:

```go
publisherOutput(registration, instance.ID().ToStringOutput(), instance.PrivateIp, instance.PublicIp, args.PlacementLabels)
```

For helpers that only have `common CommonPublisherArgs`, pass `common.PlacementLabels`.

- [ ] **Step 5: Add placementLabels to all publisher args**

In `internal/provider/components.go`, add this field to every publisher args struct:

```go
PlacementLabels []string `pulumi:"placementLabels,optional"`
```

Add it near `Replicas` in each struct. Update every `common()` helper to set:

```go
PlacementLabels: args.PlacementLabels,
```

Update `commonFromExpandedArgs`:

```go
PlacementLabels: fieldValue[[]string](value, "PlacementLabels"),
```

- [ ] **Step 6: Run focused and full provider tests**

Run:

```bash
go test ./internal/provider -run TestAwsConstructIncludesPublisherPlacementLabels -count=1
go test ./internal/provider -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/provider/types.go internal/provider/components.go internal/provider/provider_test.go
git commit -m "feat: expose publisher placement labels"
```

---

### Task 4: Add Tag Publisher Assignment Reconciler

**Files:**
- Modify: `internal/provider/netskope_client.go`
- Create: `internal/provider/tag_publisher_assignment.go`
- Modify: `internal/provider/provider.go`
- Test: `internal/provider/provider_test.go`

- [ ] **Step 1: Write failing reconciliation test**

Add to `internal/provider/provider_test.go`:

```go
func TestTagPublisherAssignmentAddsAndRemovesOnlySelectedPublishers(t *testing.T) {
	var putBodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/steering/apps/private":
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data": []map[string]any{
					{"app_id": 10, "app_name": "orders", "tags": []map[string]any{{"tag_name": "vpc-a"}}, "service_publisher_assignments": []map[string]any{{"publisher_id": 99}}},
					{"app_id": 20, "app_name": "billing", "tags": []map[string]any{{"tag_name": "vpc-b"}}, "service_publisher_assignments": []map[string]any{{"publisher_id": 101}, {"publisher_id": 99}}},
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

	_, err := createTagPublisherAssignmentResource(t, property.NewMap(map[string]property.Value{
		"tenantUrl":                property.New(server.URL),
		"bearerToken":              property.New("api-token"),
		"appTags":                  property.New([]property.Value{property.New("vpc-a")}),
		"publisherPlacementLabels": property.New([]property.Value{property.New("vpc-a")}),
		"publishers": property.New(map[string]property.Value{
			"pub-a": property.New(map[string]property.Value{
				"publisherId":      property.New(101.0),
				"placementLabels":  property.New([]property.Value{property.New("vpc-a")}),
			}),
			"pub-b": property.New(map[string]property.Value{
				"publisherId":      property.New(202.0),
				"placementLabels":  property.New([]property.Value{property.New("vpc-b")}),
			}),
		}),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(putBodies) != 2 {
		t.Fatalf("expected two publisher association updates, got %#v", putBodies)
	}
}
```

Add this helper near `createPrivateAppResource`:

```go
func createTagPublisherAssignmentResource(t *testing.T, inputs property.Map) (p.CreateResponse, error) {
	t.Helper()
	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}
	server, err := integration.NewServer(t.Context(), Name, semver.MustParse("0.2.0"), integration.WithProvider(provider))
	if err != nil {
		t.Fatal(err)
	}
	return server.Create(p.CreateRequest{
		Urn:        presource.URN("urn:pulumi:stack::project::netskope-publisher:index:TagPublisherAssignment::assignment"),
		Properties: inputs,
	})
}
```

- [ ] **Step 2: Run the test and verify it fails**

Run:

```bash
go test ./internal/provider -run TestTagPublisherAssignmentAddsAndRemovesOnlySelectedPublishers -count=1
```

Expected: FAIL because resource and association client methods do not exist.

- [ ] **Step 3: Add client association types and methods**

Add to `internal/provider/netskope_client.go`:

```go
type privateAppPublisherAssignment struct {
	PublisherID int `json:"publisher_id"`
}

type privateAppRecordWithPublishers struct {
	privateAppRecord
	ServicePublisherAssignments []privateAppPublisherAssignment `json:"service_publisher_assignments"`
}

func (client *netskopeClient) listPrivateAppsWithPublishers(ctx context.Context) ([]privateAppRecordWithPublishers, error) {
	var response struct {
		Status string                           `json:"status"`
		Data   []privateAppRecordWithPublishers `json:"data"`
	}
	if err := client.request(ctx, "List private apps", http.MethodGet, "/api/v2/steering/apps/private", nil, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (client *netskopeClient) replacePrivateAppPublishers(ctx context.Context, appNames []string, publisherIDs []int) error {
	ids := make([]string, 0, len(publisherIDs))
	for _, id := range publisherIDs {
		ids = append(ids, strconv.Itoa(id))
	}
	body := map[string]any{
		"private_app_names": appNames,
		"publisher_ids":     ids,
	}
	return client.request(ctx, "Replace private app publishers", http.MethodPut, "/api/v2/steering/apps/private/publishers", body, nil)
}
```

Add `strconv` to the imports if not already present.

- [ ] **Step 4: Add resource args and reconcile implementation**

Create `internal/provider/tag_publisher_assignment.go`:

```go
package provider

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"
)

type TagPublisherAssignment struct{}

type PublisherAssignmentInput struct {
	PublisherID      int      `pulumi:"publisherId"`
	PlacementLabels  []string `pulumi:"placementLabels,optional"`
}

type TagPublisherAssignmentArgs struct {
	TenantURL                string                              `pulumi:"tenantUrl"`
	APIToken                 *string                             `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken              *string                             `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode                 *string                             `pulumi:"authMode,optional"`
	OAuth2                   *NetskopeOAuth2Args                 `pulumi:"oauth2,optional"`
	AppTags                  []string                            `pulumi:"appTags"`
	PublisherPlacementLabels []string                            `pulumi:"publisherPlacementLabels"`
	Publishers               map[string]PublisherAssignmentInput `pulumi:"publishers"`
	MatchMode                *string                             `pulumi:"matchMode,optional"`
}

type TagPublisherAssignmentOutputs struct {
	TagPublisherAssignmentArgs
	MatchedApps      []string `pulumi:"matchedApps"`
	SelectedPublishers []int  `pulumi:"selectedPublishers"`
}

func (*TagPublisherAssignment) Annotate(a infer.Annotator) {
	a.SetToken("index", "TagPublisherAssignment")
}

func (*TagPublisherAssignment) Create(ctx context.Context, req infer.CreateRequest[TagPublisherAssignmentArgs]) (infer.CreateResponse[TagPublisherAssignmentOutputs], error) {
	output, err := reconcileTagPublisherAssignment(ctx, req.Inputs, req.DryRun)
	if err != nil {
		return infer.CreateResponse[TagPublisherAssignmentOutputs]{}, err
	}
	return infer.CreateResponse[TagPublisherAssignmentOutputs]{ID: strings.Join(req.Inputs.AppTags, ","), Output: output}, nil
}

func (*TagPublisherAssignment) Update(ctx context.Context, req infer.UpdateRequest[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs]) (infer.UpdateResponse[TagPublisherAssignmentOutputs], error) {
	output, err := reconcileTagPublisherAssignment(ctx, req.Inputs, req.DryRun)
	if err != nil {
		return infer.UpdateResponse[TagPublisherAssignmentOutputs]{}, err
	}
	return infer.UpdateResponse[TagPublisherAssignmentOutputs]{Output: output}, nil
}
```

Add delete/read:

```go
func (*TagPublisherAssignment) Read(ctx context.Context, req infer.ReadRequest[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs]) (infer.ReadResponse[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs], error) {
	return infer.ReadResponse[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs]{ID: req.ID, Inputs: req.Inputs, State: req.State}, nil
}

func (*TagPublisherAssignment) Delete(ctx context.Context, req infer.DeleteRequest[TagPublisherAssignmentOutputs]) (infer.DeleteResponse, error) {
	return infer.DeleteResponse{}, nil
}
```

Add reconciliation helpers:

```go
func reconcileTagPublisherAssignment(ctx context.Context, args TagPublisherAssignmentArgs, dryRun bool) (TagPublisherAssignmentOutputs, error) {
	selected := selectPublishersByPlacement(args.Publishers, args.PublisherPlacementLabels)
	output := TagPublisherAssignmentOutputs{TagPublisherAssignmentArgs: args, SelectedPublishers: selected}
	if dryRun {
		return output, nil
	}
	client := newResourceClient(args.TenantURL, args.APIToken, args.BearerToken, args.AuthMode, args.OAuth2, http.DefaultClient)
	apps, err := client.listPrivateAppsWithPublishers(ctx)
	if err != nil {
		return output, err
	}
	selectedSet := intSet(selected)
	for _, app := range apps {
		matches := appMatchesTags(app.Tags, args.AppTags, defaultString(args.MatchMode, "any"))
		current := currentPublisherIDs(app.ServicePublisherAssignments)
		next := reconcilePublisherIDs(current, selectedSet, matches)
		if !sameInts(current, next) {
			if err := client.replacePrivateAppPublishers(ctx, []string{app.AppName}, next); err != nil {
				return output, err
			}
		}
		if matches {
			output.MatchedApps = append(output.MatchedApps, app.AppName)
		}
	}
	sort.Strings(output.MatchedApps)
	return output, nil
}
```

Add deterministic set helpers in the same file:

```go
func selectPublishersByPlacement(publishers map[string]PublisherAssignmentInput, labels []string) []int {
	labelSet := stringSet(labels)
	var selected []int
	for _, publisher := range publishers {
		if intersectsStringSet(publisher.PlacementLabels, labelSet) {
			selected = append(selected, publisher.PublisherID)
		}
	}
	sort.Ints(selected)
	return selected
}

func appMatchesTags(tags []privateAppTag, desired []string, mode string) bool {
	actual := map[string]bool{}
	for _, tag := range tags {
		actual[tag.TagName] = true
	}
	if mode == "all" {
		for _, tag := range desired {
			if !actual[tag] {
				return false
			}
		}
		return len(desired) > 0
	}
	for _, tag := range desired {
		if actual[tag] {
			return true
		}
	}
	return false
}

func currentPublisherIDs(assignments []privateAppPublisherAssignment) []int {
	ids := make([]int, 0, len(assignments))
	for _, assignment := range assignments {
		ids = append(ids, assignment.PublisherID)
	}
	sort.Ints(ids)
	return ids
}

func reconcilePublisherIDs(current []int, selected map[int]bool, matches bool) []int {
	nextSet := intSet(current)
	for id := range selected {
		if matches {
			nextSet[id] = true
		} else {
			delete(nextSet, id)
		}
	}
	return sortedIntSet(nextSet)
}
```

- [ ] **Step 5: Register the resource**

Modify `internal/provider/provider.go`:

```go
WithResources(
	infer.Resource(&NetskopeRegistration{}),
	infer.Resource(&PrivateApp{}),
	infer.Resource(&TagPublisherAssignment{}),
)
```

- [ ] **Step 6: Run focused tests**

Run:

```bash
go test ./internal/provider -run TestTagPublisherAssignmentAddsAndRemovesOnlySelectedPublishers -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/provider/netskope_client.go internal/provider/tag_publisher_assignment.go internal/provider/provider.go internal/provider/provider_test.go
git commit -m "feat: reconcile app tags to publisher pools"
```

---

### Task 5: Add Realtime Protection Policy Resource

**Files:**
- Modify: `internal/provider/netskope_client.go`
- Create: `internal/provider/realtime_protection_policy.go`
- Modify: `internal/provider/provider.go`
- Test: `internal/provider/provider_test.go`

- [ ] **Step 1: Write failing policy create/update/delete test**

Add to `internal/provider/provider_test.go`:

```go
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
			writeJSON(t, w, map[string]any{"status": "success", "data": map[string]any{"id": 55, "name": "orders-access"}})
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
		"appTags":         property.New([]property.Value{property.New("vpc-a")}),
		"users":           property.New([]property.Value{property.New("user@example.com")}),
		"action":          property.New("allow"),
		"enabled":         property.New(true),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "55" {
		t.Fatalf("expected policy ID 55, got %q", response.ID)
	}
	if created["name"] != "orders-access" {
		t.Fatalf("expected policy name in payload, got %#v", created)
	}
}
```

Add this helper near `createPrivateAppResource`:

```go
func createRealtimeProtectionPolicyResource(t *testing.T, inputs property.Map) (p.CreateResponse, error) {
	t.Helper()
	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}
	server, err := integration.NewServer(t.Context(), Name, semver.MustParse("0.2.0"), integration.WithProvider(provider))
	if err != nil {
		t.Fatal(err)
	}
	return server.Create(p.CreateRequest{
		Urn:        presource.URN("urn:pulumi:stack::project::netskope-publisher:index:RealtimeProtectionPolicy::policy"),
		Properties: inputs,
	})
}
```

- [ ] **Step 2: Run the test and verify it fails**

Run:

```bash
go test ./internal/provider -run TestRealtimeProtectionPolicyCreatesRuleWithPolicyGroupReference -count=1
```

Expected: FAIL because the resource and client methods do not exist.

- [ ] **Step 3: Add policy client methods**

Add to `internal/provider/netskope_client.go`:

```go
type policyGroupRecord struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type realtimePolicyPayload struct {
	Name          string   `json:"name"`
	PolicyGroupID int     `json:"policy_group_id,omitempty"`
	AppIDs        []int    `json:"private_app_ids,omitempty"`
	AppTags       []string `json:"private_app_tags,omitempty"`
	Users         []string `json:"users,omitempty"`
	Groups        []string `json:"groups,omitempty"`
	Action        string   `json:"action"`
	Enabled       bool     `json:"enabled"`
}

type realtimePolicyRecord struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (client *netskopeClient) findPolicyGroupByName(ctx context.Context, name string) (*policyGroupRecord, error) {
	var response struct {
		Status string              `json:"status"`
		Data   []policyGroupRecord `json:"data"`
	}
	if err := client.request(ctx, "List policy groups", http.MethodGet, "/api/v2/policy/npa/policygroups", nil, &response); err != nil {
		return nil, err
	}
	for _, group := range response.Data {
		if group.Name == name {
			return &group, nil
		}
	}
	return nil, nil
}

func (client *netskopeClient) createRealtimePolicy(ctx context.Context, payload realtimePolicyPayload) (realtimePolicyRecord, error) {
	var response struct {
		Status string               `json:"status"`
		Data   realtimePolicyRecord `json:"data"`
	}
	if err := client.request(ctx, "Create realtime protection policy "+payload.Name, http.MethodPost, "/api/v2/policy/npa/rules", payload, &response); err != nil {
		return realtimePolicyRecord{}, err
	}
	return response.Data, nil
}

func (client *netskopeClient) updateRealtimePolicy(ctx context.Context, id int, payload realtimePolicyPayload) (realtimePolicyRecord, error) {
	var response struct {
		Status string               `json:"status"`
		Data   realtimePolicyRecord `json:"data"`
	}
	path := fmt.Sprintf("/api/v2/policy/npa/rules/%d", id)
	if err := client.request(ctx, "Update realtime protection policy "+payload.Name, http.MethodPatch, path, payload, &response); err != nil {
		return realtimePolicyRecord{}, err
	}
	return response.Data, nil
}

func (client *netskopeClient) deleteRealtimePolicy(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v2/policy/npa/rules/%d", id)
	return client.request(ctx, fmt.Sprintf("Delete realtime protection policy %d", id), http.MethodDelete, path, nil, nil)
}
```

- [ ] **Step 4: Add resource implementation**

Create `internal/provider/realtime_protection_policy.go`:

```go
package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

type RealtimeProtectionPolicy struct{}

type RealtimeProtectionPolicyArgs struct {
	TenantURL       string              `pulumi:"tenantUrl"`
	APIToken        *string             `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string             `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string             `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args `pulumi:"oauth2,optional"`
	Name            string              `pulumi:"name"`
	PolicyGroupID   *int                `pulumi:"policyGroupId,optional"`
	PolicyGroupName *string             `pulumi:"policyGroupName,optional"`
	AppIDs          []int               `pulumi:"appIds,optional"`
	AppTags         []string            `pulumi:"appTags,optional"`
	Users           []string            `pulumi:"users,optional"`
	Groups          []string            `pulumi:"groups,optional"`
	Action          string              `pulumi:"action"`
	Enabled         bool                `pulumi:"enabled"`
}

type RealtimeProtectionPolicyOutputs struct {
	RealtimeProtectionPolicyArgs
	PolicyID int `pulumi:"policyId"`
	ResolvedPolicyGroupID int `pulumi:"resolvedPolicyGroupId"`
}

func (*RealtimeProtectionPolicy) Annotate(a infer.Annotator) {
	a.SetToken("index", "RealtimeProtectionPolicy")
}
```

Add create/update/delete:

```go
func (*RealtimeProtectionPolicy) Create(ctx context.Context, req infer.CreateRequest[RealtimeProtectionPolicyArgs]) (infer.CreateResponse[RealtimeProtectionPolicyOutputs], error) {
	output, payload, err := realtimePolicyPayloadFromArgs(ctx, req.Inputs)
	if err != nil {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	if req.DryRun {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{ID: req.Inputs.Name, Output: output}, nil
	}
	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	created, err := client.createRealtimePolicy(ctx, payload)
	if err != nil {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	output.PolicyID = created.ID
	return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{ID: strconv.Itoa(created.ID), Output: output}, nil
}

func (*RealtimeProtectionPolicy) Update(ctx context.Context, req infer.UpdateRequest[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]) (infer.UpdateResponse[RealtimeProtectionPolicyOutputs], error) {
	output, payload, err := realtimePolicyPayloadFromArgs(ctx, req.Inputs)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	policyID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, fmt.Errorf("invalid realtime policy ID %q: %w", req.ID, err)
	}
	output.PolicyID = policyID
	if req.DryRun {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{Output: output}, nil
	}
	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	updated, err := client.updateRealtimePolicy(ctx, policyID, payload)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	output.PolicyID = updated.ID
	return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{Output: output}, nil
}

func (*RealtimeProtectionPolicy) Delete(ctx context.Context, req infer.DeleteRequest[RealtimeProtectionPolicyOutputs]) (infer.DeleteResponse, error) {
	policyID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid realtime policy ID %q: %w", req.ID, err)
	}
	client := newResourceClient(req.State.TenantURL, req.State.APIToken, req.State.BearerToken, req.State.AuthMode, req.State.OAuth2, http.DefaultClient)
	return infer.DeleteResponse{}, client.deleteRealtimePolicy(ctx, policyID)
}
```

Add payload resolver:

```go
func realtimePolicyPayloadFromArgs(ctx context.Context, args RealtimeProtectionPolicyArgs) (RealtimeProtectionPolicyOutputs, realtimePolicyPayload, error) {
	output := RealtimeProtectionPolicyOutputs{RealtimeProtectionPolicyArgs: args}
	groupID := 0
	if args.PolicyGroupID != nil {
		groupID = *args.PolicyGroupID
	}
	if groupID == 0 && args.PolicyGroupName != nil && *args.PolicyGroupName != "" {
		client := newResourceClient(args.TenantURL, args.APIToken, args.BearerToken, args.AuthMode, args.OAuth2, http.DefaultClient)
		group, err := client.findPolicyGroupByName(ctx, *args.PolicyGroupName)
		if err != nil {
			return output, realtimePolicyPayload{}, err
		}
		if group == nil {
			return output, realtimePolicyPayload{}, fmt.Errorf("policy group %q not found", *args.PolicyGroupName)
		}
		groupID = group.ID
	}
	output.ResolvedPolicyGroupID = groupID
	return output, realtimePolicyPayload{
		Name:          args.Name,
		PolicyGroupID: groupID,
		AppIDs:        args.AppIDs,
		AppTags:       args.AppTags,
		Users:         args.Users,
		Groups:        args.Groups,
		Action:        args.Action,
		Enabled:       args.Enabled,
	}, nil
}
```

- [ ] **Step 5: Register the resource**

Modify `internal/provider/provider.go`:

```go
WithResources(
	infer.Resource(&NetskopeRegistration{}),
	infer.Resource(&PrivateApp{}),
	infer.Resource(&TagPublisherAssignment{}),
	infer.Resource(&RealtimeProtectionPolicy{}),
)
```

- [ ] **Step 6: Run focused tests**

Run:

```bash
go test ./internal/provider -run TestRealtimeProtectionPolicyCreatesRuleWithPolicyGroupReference -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/provider/netskope_client.go internal/provider/realtime_protection_policy.go internal/provider/provider.go internal/provider/provider_test.go
git commit -m "feat: add realtime protection policy resource"
```

---

### Task 6: Schema, Docs, Examples, and Validation

**Files:**
- Modify: `schema.json`
- Modify: `README.md`
- Modify: `site/source/admin/component/index.md`
- Create: `examples/npa-application/Pulumi.yaml`
- Create: `examples/npa-application/package.json`
- Create: `examples/npa-application/index.ts`
- Create: `examples/npa-application/README.md`
- Generated SDK files under `sdk/`

- [ ] **Step 1: Run full Go tests**

Run:

```bash
npm run go:test
```

Expected: PASS.

- [ ] **Step 2: Regenerate schema and SDKs**

Run:

```bash
npm run sdk:gen
```

Expected: `schema.json` and SDK files update with `PrivateApp`, `TagPublisherAssignment`, `RealtimeProtectionPolicy`, and `placementLabels`.

- [ ] **Step 3: Add TypeScript example**

Create `examples/npa-application/index.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import {
  AwsPublisher,
  PrivateApp,
  RealtimeProtectionPolicy,
  TagPublisherAssignment,
} from "@johninnl/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publishers = new AwsPublisher("vpc-a-publishers", {
  names: ["vpc-a-pub-1"],
  placementLabels: ["vpc-a"],
  tenantUrl: config.require("tenantUrl"),
  bearerToken: config.requireSecret("bearerToken"),
  subnetId: config.require("subnetId"),
  securityGroupIds: config.requireObject<string[]>("securityGroupIds"),
  amiId: config.require("amiId"),
});

const app = new PrivateApp("orders", {
  tenantUrl: config.require("tenantUrl"),
  bearerToken: config.requireSecret("bearerToken"),
  appName: "orders",
  appType: "client",
  host: "orders.internal",
  protocols: [{ type: "tcp", ports: "443" }],
  clientlessAccess: false,
  isUserPortalApp: false,
  usePublisherDns: false,
  trustSelfSignedCerts: false,
  tags: ["vpc-a"],
});

const assignment = new TagPublisherAssignment("vpc-a-access", {
  tenantUrl: config.require("tenantUrl"),
  bearerToken: config.requireSecret("bearerToken"),
  appTags: ["vpc-a"],
  publisherPlacementLabels: ["vpc-a"],
  publishers: publishers.publishers,
});

const policy = new RealtimeProtectionPolicy("orders-access", {
  tenantUrl: config.require("tenantUrl"),
  bearerToken: config.requireSecret("bearerToken"),
  name: "orders-access",
  policyGroupName: config.require("policyGroupName"),
  appTags: ["vpc-a"],
  users: config.requireObject<string[]>("users"),
  action: "allow",
  enabled: true,
});

export const appId = app.appId;
export const matchedApps = assignment.matchedApps;
export const policyId = policy.policyId;
```

Create `examples/npa-application/Pulumi.yaml`:

```yaml
name: npa-application
runtime: nodejs
description: Register a private app and reconcile it to a placement-labeled publisher pool.
```

Create `examples/npa-application/package.json`:

```json
{
  "name": "npa-application",
  "private": true,
  "type": "module",
  "dependencies": {
    "@johninnl/pulumi-netskope-publisher": "file:../..",
    "@pulumi/pulumi": "^3.0.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0"
  }
}
```

Create `examples/npa-application/README.md`:

```md
# NPA Application Example

This example deploys a placement-labeled publisher pool, registers a private
application, assigns all apps tagged `vpc-a` to publishers labeled `vpc-a`, and
creates an NPA realtime protection policy for the app tag.

Required config:

```bash
pulumi config set tenantUrl https://example.goskope.com
pulumi config set --secret bearerToken <token>
pulumi config set subnetId subnet-123
pulumi config set --path 'securityGroupIds[0]' sg-123
pulumi config set amiId ami-123
pulumi config set policyGroupName default
pulumi config set --path 'users[0]' user@example.com
```
```

- [ ] **Step 4: Update README scope and quick example**

Modify `README.md` current scope list to add:

```md
- Private application registration: `PrivateApp`
- NPA realtime protection policy rule: `RealtimeProtectionPolicy`
- App-tag to publisher-pool reconciliation: `TagPublisherAssignment`
- Publisher placement labels for Pulumi-side pool selection
```

Add a short section after Quick start:

```md
## Private application access path

Publisher components accept `placementLabels` so Pulumi can group deployed
publishers by logical network or placement. `PrivateApp` registers applications
with Netskope app tags, and `TagPublisherAssignment` reconciles apps with a
matching tag to publishers with the matching placement label.
```

- [ ] **Step 5: Update site component index**

Modify `site/source/admin/component/index.md` to include the three new resources in the component list:

```md
- `PrivateApp` registers a Netskope Private Access private application.
- `TagPublisherAssignment` reconciles private app tags to placement-labeled publisher pools.
- `RealtimeProtectionPolicy` manages one NPA realtime protection rule for app access.
```

- [ ] **Step 6: Run validation commands**

Run:

```bash
npm run typecheck
npm run go:test
npm run registry:check
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add schema.json sdk README.md site/source/admin/component/index.md examples/npa-application
git commit -m "docs: add npa application resource examples"
```

---

### Task 7: Final Integration Check

**Files:**
- Review all changed files

- [ ] **Step 1: Run complete supported checks**

Run:

```bash
npm run typecheck
npm test
npm run go:test
npm run sdk:gen
npm run registry:check
```

Expected: PASS. If `npm test` or `registry:check` rewrites generated docs or schema, inspect and include legitimate generated output.

- [ ] **Step 2: Inspect working tree**

Run:

```bash
git status --short
```

Expected: only intentional files from this plan are modified. Preserve unrelated user changes such as pre-existing `site/package.json` edits unless the user explicitly asks to include them.

- [ ] **Step 3: Review schema for public names**

Run:

```bash
rg -n '"PrivateApp"|"TagPublisherAssignment"|"RealtimeProtectionPolicy"|"placementLabels"' schema.json
```

Expected: all new resources and `placementLabels` appear in generated schema.

- [ ] **Step 4: Commit final generated drift if needed**

If Step 1 or Step 3 produced intentional generated changes not committed by Task 6, commit them:

```bash
git add schema.json sdk docs site README.md examples
git commit -m "chore: refresh generated npa resource artifacts"
```

If there are no generated changes, do not create an empty commit.

---

## Self-Review

Spec coverage:

- Broaden provider into focused NPA deployment provider: Tasks 2, 4, 5, and 6.
- Private app CRUD and explicit adoption: Task 2.
- Inline private app tags: Task 2.
- Realtime protection policy with policy group reference/lookup: Task 5.
- User-defined publisher placement labels: Task 3.
- Tag-based publisher assignment reconciler: Task 4.
- Shared Netskope API client and auth reuse: Task 1.
- Tests with `httptest`: Tasks 1, 2, 4, and 5.
- Docs/examples/schema/SDK updates: Task 6.

Out of scope for this implementation plan:

- Separate `PrivateAppTagAttachment` resource.
- Full policy group lifecycle.
- Alerts, discovery settings, upgrade profiles, SCIM, reporting, and broad steering administration.

Type consistency:

- Public Pulumi names use camelCase: `tenantUrl`, `bearerToken`, `placementLabels`, `appTags`, `publisherPlacementLabels`, `policyGroupName`.
- Go resource names match provider tokens: `PrivateApp`, `TagPublisherAssignment`, `RealtimeProtectionPolicy`.
- Assignment publisher input matches publisher output shape: `publisherId` plus `placementLabels`.
