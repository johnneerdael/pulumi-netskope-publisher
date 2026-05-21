# Provider Contract Follow-Up Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Close the remaining provider contract bugs found in the second architecture audit.

**Architecture:** Keep the provider's public resources small, but make the low-level Netskope client mirror the documented response envelopes and request field types in `/Users/jneerdael/Scripts/privateaccess-mcp/swagger.json`. Start with failing contract tests that use swagger-shaped fixtures, then update the client/resource code, then regenerate SDK and docs only after the provider schema is intentionally changed.

**Tech Stack:** Go provider code using `pulumi-go-provider`, `net/http/httptest`, TypeScript SDK generated from provider schema, Node test runner, npm scripts.

---

## File Structure

- Modify `internal/provider/netskope_client.go`: owns Netskope HTTP payload structs and response envelope parsing. Add list response structs for `data.private_apps`, response envelopes for NPA rule read/update, string conversion for NPA rule request IDs, and optional private app publishers payload support.
- Modify `internal/provider/private_app.go`: owns `PrivateApp` lifecycle and Pulumi input/output types. Remove unsupported `hosts`, add optional initial `publishers`, and keep create/update/read behavior aligned with the client.
- Modify `internal/provider/realtime_protection_policy.go`: owns `RealtimeProtectionPolicy` lifecycle and mapping from Pulumi args to NPA rule payloads. Convert integer Pulumi inputs to string API fields.
- Modify `internal/provider/tag_publisher_assignment.go`: owns tag matching and assignment reconciliation. Add read-time recomputation without remote writes.
- Modify `internal/provider/provider_test.go`: contract tests for swagger-shaped private app list, NPA rule read/update envelopes, string ID request payloads, assignment refresh, and optional initial publishers.
- Create `test/privateAppSchema.test.ts`: asserts the generated TypeScript/JSON schema no longer advertises unsupported `hosts`.
- Regenerate `schema.json`, `src/privateApp.ts`, SDK outputs, and docs after schema-affecting changes.
- Do not edit release version files or root changelog entries.

## Task 1: Parse Private App List Responses From `data.private_apps`

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/netskope_client.go`

- [x] **Step 1: Change private app list fixtures to the documented shape**

In `internal/provider/provider_test.go`, replace the list response in `TestPrivateAppCreateFailsWhenExistingWithoutAdopt` with:

```go
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
```

In `TestPrivateAppAdoptsExistingByName`, replace the `GET /api/v2/steering/apps/private` response with:

```go
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
```

In `TestTagPublisherAssignmentAddsAndRemovesOnlySelectedPublishers`, replace the `GET /api/v2/steering/apps/private` response with:

```go
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
```

In `TestTagPublisherAssignmentDeleteRemovesSelectedPublishersFromMatchedApps`, replace the `GET /api/v2/steering/apps/private` response with:

```go
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
```

- [x] **Step 2: Run private app list-dependent tests and verify they fail**

Run:

```bash
npm run go:test -- -run 'Test(PrivateAppCreateFailsWhenExistingWithoutAdopt|PrivateAppAdoptsExistingByName|TagPublisherAssignmentAddsAndRemovesOnlySelectedPublishers|TagPublisherAssignmentDeleteRemovesSelectedPublishersFromMatchedApps)' -count=1
```

Expected: FAIL because `listPrivateApps` and `listPrivateAppsWithPublishers` currently expect `data` to be an array.

- [x] **Step 3: Add documented list response structs**

In `internal/provider/netskope_client.go`, add these structs above `listPrivateApps`:

```go
type privateAppsListResponse struct {
	Status string `json:"status"`
	Data   struct {
		PrivateApps []privateAppRecord `json:"private_apps"`
	} `json:"data"`
}

type privateAppsWithPublishersListResponse struct {
	Status string `json:"status"`
	Data   struct {
		PrivateApps []privateAppRecordWithPublishers `json:"private_apps"`
	} `json:"data"`
}
```

- [x] **Step 4: Use the documented list response structs**

In `internal/provider/netskope_client.go`, replace `listPrivateApps` with:

```go
func (client *netskopeClient) listPrivateApps(ctx context.Context) ([]privateAppRecord, error) {
	var response privateAppsListResponse
	if err := client.request(ctx, "List private apps", http.MethodGet, "/api/v2/steering/apps/private", nil, &response); err != nil {
		return nil, err
	}
	return response.Data.PrivateApps, nil
}
```

Replace `listPrivateAppsWithPublishers` with:

```go
func (client *netskopeClient) listPrivateAppsWithPublishers(ctx context.Context) ([]privateAppRecordWithPublishers, error) {
	var response privateAppsWithPublishersListResponse
	if err := client.request(ctx, "List private apps", http.MethodGet, "/api/v2/steering/apps/private", nil, &response); err != nil {
		return nil, err
	}
	return response.Data.PrivateApps, nil
}
```

- [x] **Step 5: Run the private app list-dependent tests and verify they pass**

Run:

```bash
npm run go:test -- -run 'Test(PrivateAppCreateFailsWhenExistingWithoutAdopt|PrivateAppAdoptsExistingByName|TagPublisherAssignmentAddsAndRemovesOnlySelectedPublishers|TagPublisherAssignmentDeleteRemovesSelectedPublishersFromMatchedApps)' -count=1
```

Expected: PASS.

- [x] **Step 6: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/netskope_client.go
git commit -m "fix: parse private app list envelope"
```

## Task 2: Repair Realtime Policy Read/Update Envelopes and String Request IDs

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/netskope_client.go`
- Modify: `internal/provider/realtime_protection_policy.go`

- [x] **Step 1: Update the realtime create test to expect API string IDs**

In `TestRealtimeProtectionPolicyCreatesRuleWithPolicyGroupReference`, replace:

```go
if created["group_id"] != float64(12) {
	t.Fatalf("expected group_id 12 in payload, got %#v", created)
}
```

with:

```go
if created["group_id"] != "12" {
	t.Fatalf("expected group_id string 12 in payload, got %#v", created)
}
```

Replace:

```go
if got := ruleData["privateApps"].([]any)[0]; got != float64(44) {
	t.Fatalf("expected privateApps [44], got %#v", ruleData["privateApps"])
}
```

with:

```go
if got := ruleData["privateApps"].([]any)[0]; got != "44" {
	t.Fatalf("expected privateApps [44] as strings, got %#v", ruleData["privateApps"])
}
```

- [x] **Step 2: Add read/update tests for the documented NPA policy envelopes**

Append these tests to `internal/provider/provider_test.go` near the other realtime policy tests:

```go
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
```

Add this helper near `stringPtr` if it does not exist:

```go
func intPtr(value int) *int {
	return &value
}
```

- [x] **Step 3: Run realtime policy tests and verify they fail**

Run:

```bash
npm run go:test -- -run 'TestRealtimeProtectionPolicy(CreatesRuleWithPolicyGroupReference|ReadParsesPolicyEnvelope|UpdateParsesPolicyEnvelope)' -count=1
```

Expected: FAIL because the provider still sends numeric `group_id`/`privateApps` and parses read/update responses as direct policy records.

- [x] **Step 4: Change realtime policy payload fields to API strings**

In `internal/provider/netskope_client.go`, replace `realtimePolicyRuleData` and `realtimePolicyPayload` with:

```go
type realtimePolicyRuleData struct {
	PrivateApps         []string             `json:"privateApps,omitempty"`
	PrivateAppTags      []string             `json:"privateAppTags,omitempty"`
	Users               []string             `json:"users,omitempty"`
	UserGroups          []string             `json:"userGroups,omitempty"`
	MatchCriteriaAction realtimePolicyAction `json:"match_criteria_action"`
}

type realtimePolicyPayload struct {
	RuleName  string                 `json:"rule_name"`
	GroupID   string                 `json:"group_id,omitempty"`
	GroupName string                 `json:"group_name,omitempty"`
	RuleData  realtimePolicyRuleData `json:"rule_data"`
	Enabled   string                 `json:"enabled"`
}
```

- [x] **Step 5: Parse read/update policy envelopes**

In `internal/provider/netskope_client.go`, add:

```go
type realtimePolicyEnvelope struct {
	Status string               `json:"status"`
	Data   realtimePolicyRecord `json:"data"`
}
```

Replace `updateRealtimePolicy` with:

```go
func (client *netskopeClient) updateRealtimePolicy(ctx context.Context, id int, payload realtimePolicyPayload) (realtimePolicyRecord, error) {
	var response realtimePolicyEnvelope
	path := fmt.Sprintf("/api/v2/policy/npa/rules/%d", id)
	if err := client.request(ctx, "Update realtime protection policy "+payload.RuleName, http.MethodPatch, path, payload, &response); err != nil {
		return realtimePolicyRecord{}, err
	}
	return response.Data, nil
}
```

Replace `getRealtimePolicy` with:

```go
func (client *netskopeClient) getRealtimePolicy(ctx context.Context, id int) (realtimePolicyRecord, error) {
	var response realtimePolicyEnvelope
	path := fmt.Sprintf("/api/v2/policy/npa/rules/%d", id)
	if err := client.request(ctx, fmt.Sprintf("Get realtime protection policy %d", id), http.MethodGet, path, nil, &response); err != nil {
		return realtimePolicyRecord{}, err
	}
	return response.Data, nil
}
```

- [x] **Step 6: Convert Pulumi integer args to API string fields**

In `internal/provider/realtime_protection_policy.go`, replace the tail of `realtimePolicyPayloadFromArgs` with:

```go
	groupIDString := ""
	if groupID != 0 {
		groupIDString = strconv.Itoa(groupID)
	}

	output.ResolvedPolicyGroupID = groupID
	return output, realtimePolicyPayload{
		RuleName: args.Name,
		GroupID:  groupIDString,
		RuleData: realtimePolicyRuleData{
			PrivateApps:    intStrings(args.AppIDs),
			PrivateAppTags: args.AppTags,
			Users:          args.Users,
			UserGroups:     args.Groups,
			MatchCriteriaAction: realtimePolicyAction{
				ActionName: args.Action,
			},
		},
		Enabled: enabledString(args.Enabled),
	}, nil
}

func intStrings(values []int) []string {
	strings := make([]string, 0, len(values))
	for _, value := range values {
		strings = append(strings, strconv.Itoa(value))
	}
	return strings
}
```

- [x] **Step 7: Run realtime policy tests and verify they pass**

Run:

```bash
npm run go:test -- -run 'TestRealtimeProtectionPolicy(CreatesRuleWithPolicyGroupReference|ReadParsesPolicyEnvelope|UpdateParsesPolicyEnvelope)' -count=1
```

Expected: PASS.

- [x] **Step 8: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/netskope_client.go internal/provider/realtime_protection_policy.go
git commit -m "fix: align realtime policy read and update contracts"
```

## Task 3: Recompute TagPublisherAssignment During Refresh

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/tag_publisher_assignment.go`

- [x] **Step 1: Add a read test that recomputes matched apps and selected publishers**

Append this test near the other `TagPublisherAssignment` tests in `internal/provider/provider_test.go`:

```go
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
```

- [x] **Step 2: Run the read test and verify it fails**

Run:

```bash
npm run go:test -- -run TestTagPublisherAssignmentReadRecomputesRemoteState -count=1
```

Expected: FAIL because `Read` currently returns stale state without remote recomputation.

- [x] **Step 3: Add a read-only summary function**

In `internal/provider/tag_publisher_assignment.go`, add this function below `reconcileTagPublisherAssignment`:

```go
func summarizeTagPublisherAssignment(ctx context.Context, args TagPublisherAssignmentArgs) (TagPublisherAssignmentOutputs, error) {
	selected := selectPublishersByPlacement(args.Publishers, args.PublisherPlacementLabels)
	if len(selected) == 0 {
		return TagPublisherAssignmentOutputs{}, fmt.Errorf("publisherPlacementLabels %v did not match any managed publishers", args.PublisherPlacementLabels)
	}

	output := TagPublisherAssignmentOutputs{
		TagPublisherAssignmentArgs: args,
		SelectedPublishers:         selected,
	}

	client := newResourceClient(args.TenantURL, args.APIToken, args.BearerToken, args.AuthMode, args.OAuth2, http.DefaultClient)
	apps, err := client.listPrivateAppsWithPublishers(ctx)
	if err != nil {
		return output, err
	}

	for _, app := range apps {
		if appMatchesTags(app.Tags, args.AppTags, defaultString(args.MatchMode, "any")) {
			output.MatchedApps = append(output.MatchedApps, app.AppName)
		}
	}
	sort.Strings(output.MatchedApps)
	return output, nil
}
```

- [x] **Step 4: Use the read-only summary from `Read`**

In `internal/provider/tag_publisher_assignment.go`, replace `Read` with:

```go
func (*TagPublisherAssignment) Read(ctx context.Context, req infer.ReadRequest[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs]) (infer.ReadResponse[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs], error) {
	state, err := summarizeTagPublisherAssignment(ctx, req.Inputs)
	if err != nil {
		return infer.ReadResponse[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs]{}, err
	}
	return infer.ReadResponse[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs]{ID: req.ID, Inputs: req.Inputs, State: state}, nil
}
```

- [x] **Step 5: Run the read test and verify it passes**

Run:

```bash
npm run go:test -- -run TestTagPublisherAssignmentReadRecomputesRemoteState -count=1
```

Expected: PASS.

- [x] **Step 6: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/tag_publisher_assignment.go
git commit -m "fix: refresh tag publisher assignments"
```

## Task 4: Remove Unsupported `hosts` From the Public PrivateApp Schema

**Files:**
- Modify: `internal/provider/private_app.go`
- Modify: `internal/provider/provider_test.go`
- Create: `test/privateAppSchema.test.ts`
- Regenerate: `schema.json`
- Regenerate: `src/privateApp.ts`

- [x] **Step 1: Delete the runtime rejection test**

In `internal/provider/provider_test.go`, delete the entire `TestPrivateAppRejectsHostsUntilApiShapeIsConfirmed` function:

```go
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
```

- [x] **Step 2: Add a schema test that `hosts` is not advertised**

Create `test/privateAppSchema.test.ts`:

```ts
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const schema = JSON.parse(readFileSync("schema.json", "utf8"));

test("PrivateApp schema does not expose unsupported hosts input", () => {
  const privateApp = schema.resources["netskope-publisher:index:PrivateApp"];
  assert.ok(privateApp, "PrivateApp resource must exist in schema");
  assert.equal(privateApp.inputProperties.hosts, undefined);
  assert.equal(privateApp.properties.hosts, undefined);
});
```

- [x] **Step 3: Run the schema test and verify it fails**

Run:

```bash
npm run build
```

Expected: PASS.

Run:

```bash
node --test dist/test/privateAppSchema.test.js
```

Expected: FAIL because `schema.json` still contains the `hosts` input.

- [x] **Step 4: Remove `hosts` from the provider type and validation path**

In `internal/provider/private_app.go`, remove this field from `PrivateAppArgs`:

```go
Hosts []string `pulumi:"hosts,optional"`
```

In `Create`, remove:

```go
if err := validatePrivateAppArgs(req.Inputs); err != nil {
	return infer.CreateResponse[PrivateAppOutputs]{}, err
}
```

In `Update`, remove:

```go
if err := validatePrivateAppArgs(req.Inputs); err != nil {
	return infer.UpdateResponse[PrivateAppOutputs]{}, err
}
```

Delete the helper:

```go
func validatePrivateAppArgs(args PrivateAppArgs) error {
	if len(args.Hosts) > 0 {
		return fmt.Errorf("hosts is not supported by the documented private app API; use host")
	}
	return nil
}
```

- [x] **Step 5: Regenerate schema and SDK surfaces**

Run:

```bash
npm run sdk:gen
```

Expected: generated schema and SDK no longer include `hosts` for `PrivateApp`.

Run:

```bash
npm run docs:gen
```

Expected: generated docs no longer include `hosts` for `PrivateApp`.

- [x] **Step 6: Run the schema test and verify it passes**

Run:

```bash
npm run build
```

Expected: PASS.

Run:

```bash
node --test dist/test/privateAppSchema.test.js
```

Expected: PASS.

- [x] **Step 7: Restore generated build cache churn if SDK generation removes tracked cache artifacts**

Run:

```bash
git status --short
```

If tracked generated cache artifacts under `sdk/java/.gradle`, `sdk/java/build`, `sdk/rust/target`, or `sdk/rust/Cargo.lock` appear as unrelated deletions, restore only those paths:

```bash
git restore sdk/java/.gradle sdk/java/build sdk/rust/target sdk/rust/Cargo.lock
```

Expected: remaining generated diffs are source/schema/docs changes, not build cache deletions.

- [x] **Step 8: Commit**

Run:

```bash
git add internal/provider/private_app.go internal/provider/provider_test.go test/privateAppSchema.test.ts schema.json src docs sdk
git commit -m "fix: remove unsupported private app hosts input"
```

## Task 5: Support Initial Publishers on PrivateApp Create and Update

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/private_app.go`
- Modify: `internal/provider/netskope_client.go`
- Regenerate: `schema.json`
- Regenerate: `src/privateApp.ts`

- [x] **Step 1: Add a private app create test for initial publishers**

Append this test near the other `PrivateApp` tests in `internal/provider/provider_test.go`:

```go
func TestPrivateAppCreateIncludesInitialPublishersWhenProvided(t *testing.T) {
	var created map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/steering/apps/private":
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data":   map[string]any{"private_apps": []map[string]any{}},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/steering/apps/private":
			if err := json.NewDecoder(r.Body).Decode(&created); err != nil {
				t.Fatal(err)
			}
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
		"publishers": property.New([]property.Value{property.New(map[string]property.Value{
			"publisherId":   property.New(101.0),
			"publisherName": property.New("pub-a"),
		})}),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "44" {
		t.Fatalf("expected created ID 44, got %q", response.ID)
	}
	publishers := created["publishers"].([]any)
	publisher := publishers[0].(map[string]any)
	if publisher["publisher_id"] != float64(101) {
		t.Fatalf("expected publisher_id 101, got %#v", created)
	}
	if publisher["publisher_name"] != "pub-a" {
		t.Fatalf("expected publisher_name pub-a, got %#v", created)
	}
}
```

- [x] **Step 2: Run the private app publisher test and verify it fails**

Run:

```bash
npm run go:test -- -run TestPrivateAppCreateIncludesInitialPublishersWhenProvided -count=1
```

Expected: FAIL because `PrivateAppArgs` and `privateAppPayload` do not include `publishers`.

- [x] **Step 3: Add publisher input and payload types**

In `internal/provider/private_app.go`, add this type below `PrivateAppProtocol`:

```go
type PrivateAppPublisher struct {
	PublisherID   int     `pulumi:"publisherId"`
	PublisherName *string `pulumi:"publisherName,optional"`
}
```

Add this field to `PrivateAppArgs`:

```go
Publishers []PrivateAppPublisher `pulumi:"publishers,optional"`
```

In `internal/provider/netskope_client.go`, add this type below `privateAppTag`:

```go
type privateAppPublisher struct {
	PublisherID   int    `json:"publisher_id"`
	PublisherName string `json:"publisher_name,omitempty"`
}
```

Add this field to `privateAppPayload`:

```go
Publishers []privateAppPublisher `json:"publishers,omitempty"`
```

- [x] **Step 4: Map publisher args into the private app payload**

In `internal/provider/private_app.go`, add this block inside `privateAppPayloadFromArgs` after tags are built:

```go
	publishers := make([]privateAppPublisher, 0, len(args.Publishers))
	for _, publisher := range args.Publishers {
		publishers = append(publishers, privateAppPublisher{
			PublisherID:   publisher.PublisherID,
			PublisherName: stringValue(publisher.PublisherName),
		})
	}
```

Add this field to the returned `privateAppPayload`:

```go
Publishers: publishers,
```

- [x] **Step 5: Run the private app publisher test and verify it passes**

Run:

```bash
npm run go:test -- -run TestPrivateAppCreateIncludesInitialPublishersWhenProvided -count=1
```

Expected: PASS.

- [x] **Step 6: Regenerate schema, SDK, and docs**

Run:

```bash
npm run sdk:gen
```

Expected: generated `PrivateApp` inputs include optional `publishers`.

Run:

```bash
npm run docs:gen
```

Expected: generated provider docs include optional `publishers`.

- [x] **Step 7: Restore generated build cache churn if needed**

Run:

```bash
git status --short
```

If tracked generated cache artifacts under `sdk/java/.gradle`, `sdk/java/build`, `sdk/rust/target`, or `sdk/rust/Cargo.lock` appear as unrelated deletions, restore only those paths:

```bash
git restore sdk/java/.gradle sdk/java/build sdk/rust/target sdk/rust/Cargo.lock
```

Expected: remaining generated diffs are source/schema/docs changes, not build cache deletions.

- [x] **Step 8: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/private_app.go internal/provider/netskope_client.go schema.json src docs sdk
git commit -m "feat: support initial private app publishers"
```

## Task 6: Full Verification

**Files:**
- Verify all modified files.

- [x] **Step 1: Format Go files**

Run:

```bash
gofmt -w internal/provider/netskope_client.go internal/provider/private_app.go internal/provider/realtime_protection_policy.go internal/provider/tag_publisher_assignment.go internal/provider/provider_test.go
```

Expected: command exits 0.

- [x] **Step 2: Run Go provider tests**

Run:

```bash
npm run go:test
```

Expected: PASS for `./cmd/pulumi-resource-netskope-publisher` and `./internal/provider`.

- [x] **Step 3: Run TypeScript build and Node tests**

Run:

```bash
npm test
```

Expected: TypeScript build succeeds and all Node tests pass.

- [x] **Step 4: Confirm generated outputs are intentional**

Run:

```bash
git status --short
```

Expected: modified files are limited to provider code, tests, schema/SDK/docs outputs, and plan files. No `.gradle`, Java build output, Rust target output, or release-version files should appear.

Run:

```bash
git diff --stat
```

Expected: diff shows the contract fixes, generated schema/SDK/doc changes for `hosts` removal and optional `publishers`, and no release-version bump.

- [x] **Step 5: Commit final generated cleanup if needed**

If Task 4 or Task 5 generation produced changes that were not committed in their task commits, run:

```bash
git add schema.json src docs sdk test/privateAppSchema.test.ts
git commit -m "chore: refresh generated provider surfaces"
```

If there are no uncommitted generated changes, skip this commit.

## Self-Review

- Spec coverage: Task 1 fixes `data.private_apps`; Task 2 fixes realtime read/update envelopes and string request IDs; Task 3 fixes assignment refresh; Task 4 removes unsupported `hosts`; Task 5 adds optional initial `publishers`; Task 6 verifies the whole branch.
- Placeholder scan: No deferred markers, generic edge-case instructions, or references to undefined functions remain. Helpers `stringPtr`, `intPtr`, and `intStrings` are defined in the tasks that use them.
- Type consistency: API-facing `group_id`, `privateApps`, `private_app_ids`, and `publisher_ids` are strings. Pulumi-facing `policyGroupId`, `appIds`, `publisherId`, and `policyId` remain integers.
