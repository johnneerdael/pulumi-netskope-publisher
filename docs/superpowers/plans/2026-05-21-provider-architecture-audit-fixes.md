# Provider Architecture Audit Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Fix the provider bugs found in the architecture audit by aligning NPA resources with the documented Netskope API contracts and Pulumi lifecycle expectations.

**Architecture:** Treat `/Users/jneerdael/Scripts/privateaccess-mcp/swagger.json` as the source of truth for request and response shapes. Repair the low-level Netskope client first, then update resources to use the corrected client operations, and finally add refresh/delete behavior so Pulumi state reflects remote reality.

**Tech Stack:** Go provider code using `pulumi-go-provider`, `net/http` test servers, TypeScript SDK generated from provider schema, npm scripts for test orchestration.

---

## File Structure

- Modify `internal/provider/netskope_client.go`: keep all Netskope HTTP request/response contract structs and methods here. Add API-shaped payloads for NPA rules, app-publisher assignments by ID, and remote read methods.
- Modify `internal/provider/realtime_protection_policy.go`: keep Pulumi lifecycle behavior for `RealtimeProtectionPolicy`. Move dry-run before live lookups and map Pulumi args into swagger-compatible NPA rule payloads.
- Modify `internal/provider/tag_publisher_assignment.go`: keep tag matching, publisher selection, assignment reconciliation, validation, and delete behavior for `TagPublisherAssignment`.
- Modify `internal/provider/private_app.go`: keep Pulumi lifecycle behavior for `PrivateApp`. Add remote read behavior and use the documented update method after the client changes.
- Modify `internal/provider/provider_test.go`: add contract-first unit tests around request bodies, response IDs, dry-run behavior, validation failures, delete behavior, and remote reads.
- Regenerate generated surfaces only if schema output changes: `schema.json`, `src/*.ts`, `bin/*`, and generated docs. These fixes should mostly preserve the public shape, so generation may produce no diff.

## Task 1: Repair RealtimeProtectionPolicy API Contract

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/netskope_client.go`
- Modify: `internal/provider/realtime_protection_policy.go`

- [x] **Step 1: Replace the realtime policy create test with a swagger-shaped contract test**

In `internal/provider/provider_test.go`, replace `TestRealtimeProtectionPolicyCreatesRuleWithPolicyGroupReference` with:

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
	if created["group_id"] != float64(12) {
		t.Fatalf("expected group_id 12 in payload, got %#v", created)
	}
	if created["enabled"] != "1" {
		t.Fatalf("expected enabled string 1 in payload, got %#v", created)
	}
	ruleData := created["rule_data"].(map[string]any)
	if got := ruleData["privateApps"].([]any)[0]; got != float64(44) {
		t.Fatalf("expected privateApps [44], got %#v", ruleData["privateApps"])
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
```

- [x] **Step 2: Run the realtime policy contract test and verify it fails**

Run:

```bash
npm run go:test -- -run TestRealtimeProtectionPolicyCreatesRuleWithPolicyGroupReference -count=1
```

Expected: FAIL because the provider currently sends `name` instead of `rule_name`, `policy_group_id` instead of `group_id`, and expects `data.id` instead of `rule_id`.

- [x] **Step 3: Replace realtime policy client structs with swagger-shaped structs**

In `internal/provider/netskope_client.go`, replace `realtimePolicyPayload` and `realtimePolicyRecord` with:

```go
type realtimePolicyAction struct {
	ActionName string `json:"action_name"`
}

type realtimePolicyRuleData struct {
	PrivateApps         []int                  `json:"privateApps,omitempty"`
	PrivateAppTags      []string               `json:"privateAppTags,omitempty"`
	Users               []string               `json:"users,omitempty"`
	UserGroups          []string               `json:"userGroups,omitempty"`
	MatchCriteriaAction realtimePolicyAction   `json:"match_criteria_action"`
}

type realtimePolicyPayload struct {
	RuleName  string                 `json:"rule_name"`
	GroupID   int                    `json:"group_id,omitempty"`
	GroupName string                 `json:"group_name,omitempty"`
	RuleData  realtimePolicyRuleData `json:"rule_data"`
	Enabled   string                 `json:"enabled"`
}

type realtimePolicyRecord struct {
	RuleID   int    `json:"rule_id"`
	RuleName string `json:"rule_name"`
}
```

Then replace `createRealtimePolicy` and `updateRealtimePolicy` with:

```go
func (client *netskopeClient) createRealtimePolicy(ctx context.Context, payload realtimePolicyPayload) (realtimePolicyRecord, error) {
	var response realtimePolicyRecord
	if err := client.request(ctx, "Create realtime protection policy "+payload.RuleName, http.MethodPost, "/api/v2/policy/npa/rules", payload, &response); err != nil {
		return realtimePolicyRecord{}, err
	}
	return response, nil
}

func (client *netskopeClient) updateRealtimePolicy(ctx context.Context, id int, payload realtimePolicyPayload) (realtimePolicyRecord, error) {
	var response realtimePolicyRecord
	path := fmt.Sprintf("/api/v2/policy/npa/rules/%d", id)
	if err := client.request(ctx, "Update realtime protection policy "+payload.RuleName, http.MethodPatch, path, payload, &response); err != nil {
		return realtimePolicyRecord{}, err
	}
	return response, nil
}
```

- [x] **Step 4: Map Pulumi args to the corrected realtime payload and response ID**

In `internal/provider/realtime_protection_policy.go`, update `Create`, `Update`, and `realtimePolicyPayloadFromArgs`:

```go
func (*RealtimeProtectionPolicy) Create(ctx context.Context, req infer.CreateRequest[RealtimeProtectionPolicyArgs]) (infer.CreateResponse[RealtimeProtectionPolicyOutputs], error) {
	output := RealtimeProtectionPolicyOutputs{RealtimeProtectionPolicyArgs: req.Inputs}
	if req.DryRun {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{ID: req.Inputs.Name, Output: output}, nil
	}

	output, payload, err := realtimePolicyPayloadFromArgs(ctx, req.Inputs)
	if err != nil {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}

	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	created, err := client.createRealtimePolicy(ctx, payload)
	if err != nil {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	output.PolicyID = created.RuleID
	return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{ID: strconv.Itoa(created.RuleID), Output: output}, nil
}

func (*RealtimeProtectionPolicy) Update(ctx context.Context, req infer.UpdateRequest[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]) (infer.UpdateResponse[RealtimeProtectionPolicyOutputs], error) {
	output := RealtimeProtectionPolicyOutputs{RealtimeProtectionPolicyArgs: req.Inputs}
	policyID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, fmt.Errorf("invalid realtime policy ID %q: %w", req.ID, err)
	}
	output.PolicyID = policyID
	if req.DryRun {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{Output: output}, nil
	}

	output, payload, err := realtimePolicyPayloadFromArgs(ctx, req.Inputs)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	output.PolicyID = policyID

	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	updated, err := client.updateRealtimePolicy(ctx, policyID, payload)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	output.PolicyID = updated.RuleID
	return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{Output: output}, nil
}

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
		RuleName: args.Name,
		GroupID:  groupID,
		RuleData: realtimePolicyRuleData{
			PrivateApps:    args.AppIDs,
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

func enabledString(enabled bool) string {
	if enabled {
		return "1"
	}
	return "0"
}
```

- [x] **Step 5: Run the realtime policy contract test and verify it passes**

Run:

```bash
npm run go:test -- -run TestRealtimeProtectionPolicyCreatesRuleWithPolicyGroupReference -count=1
```

Expected: PASS.

- [x] **Step 6: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/netskope_client.go internal/provider/realtime_protection_policy.go
git commit -m "fix: align realtime policy payload with npa api"
```

## Task 2: Make RealtimeProtectionPolicy Preview Offline-Safe

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/realtime_protection_policy.go`

- [x] **Step 1: Add a dry-run test that proves preview does not call Netskope**

Append this test to `internal/provider/provider_test.go`:

```go
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
```

If `stringPtr` does not already exist in the test file, add this helper near the other helpers:

```go
func stringPtr(value string) *string {
	return &value
}
```

- [x] **Step 2: Run the dry-run test and verify it fails before the Task 1 implementation, or passes after Task 1**

Run:

```bash
npm run go:test -- -run TestRealtimeProtectionPolicyDryRunDoesNotResolvePolicyGroupName -count=1
```

Expected after Task 1: PASS because Task 1 moved the dry-run check before `realtimePolicyPayloadFromArgs`.

- [x] **Step 3: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/realtime_protection_policy.go
git commit -m "test: keep realtime policy preview offline"
```

## Task 3: Send Private App IDs for Publisher Assignment

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/netskope_client.go`
- Modify: `internal/provider/tag_publisher_assignment.go`

- [x] **Step 1: Strengthen the assignment test to assert API field names and app IDs**

In `TestTagPublisherAssignmentAddsAndRemovesOnlySelectedPublishers`, after the existing `len(putBodies)` assertion, add:

```go
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
```

- [x] **Step 2: Run the assignment contract test and verify it fails**

Run:

```bash
npm run go:test -- -run TestTagPublisherAssignmentAddsAndRemovesOnlySelectedPublishers -count=1
```

Expected: FAIL because the current body contains `private_app_names`.

- [x] **Step 3: Replace assignment client method to accept private app IDs**

In `internal/provider/netskope_client.go`, replace `replacePrivateAppPublishers` with:

```go
func (client *netskopeClient) replacePrivateAppPublishers(ctx context.Context, appIDs []int, publisherIDs []int) error {
	privateAppIDs := make([]string, 0, len(appIDs))
	for _, id := range appIDs {
		privateAppIDs = append(privateAppIDs, strconv.Itoa(id))
	}
	publisherIDValues := make([]string, 0, len(publisherIDs))
	for _, id := range publisherIDs {
		publisherIDValues = append(publisherIDValues, strconv.Itoa(id))
	}
	body := map[string]any{
		"private_app_ids": privateAppIDs,
		"publisher_ids":   publisherIDValues,
	}
	return client.request(ctx, "Replace private app publishers", http.MethodPut, "/api/v2/steering/apps/private/publishers", body, nil)
}
```

- [x] **Step 4: Pass app resource IDs from reconciliation**

In `internal/provider/tag_publisher_assignment.go`, replace:

```go
if err := client.replacePrivateAppPublishers(ctx, []string{app.AppName}, next); err != nil {
	return output, err
}
```

with:

```go
if err := client.replacePrivateAppPublishers(ctx, []int{app.resourceID()}, next); err != nil {
	return output, err
}
```

In `internal/provider/netskope_client.go`, add this method below `privateAppRecordWithPublishers`:

```go
func (app privateAppRecordWithPublishers) resourceID() int {
	if app.AppID != 0 {
		return app.AppID
	}
	return app.ID
}
```

- [x] **Step 5: Run the assignment contract test and verify it passes**

Run:

```bash
npm run go:test -- -run TestTagPublisherAssignmentAddsAndRemovesOnlySelectedPublishers -count=1
```

Expected: PASS.

- [x] **Step 6: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/netskope_client.go internal/provider/tag_publisher_assignment.go
git commit -m "fix: assign publishers by private app id"
```

## Task 4: Validate Empty Publisher Selection

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/tag_publisher_assignment.go`

- [x] **Step 1: Add a validation test for unmatched placement labels**

Append this test to `internal/provider/provider_test.go`:

```go
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
```

- [x] **Step 2: Run the validation test and verify it fails**

Run:

```bash
npm run go:test -- -run TestTagPublisherAssignmentFailsWhenPlacementLabelsSelectNoPublishers -count=1
```

Expected: FAIL because the resource currently accepts an empty selected publisher list.

- [x] **Step 3: Add validation after publisher selection**

In `internal/provider/tag_publisher_assignment.go`, add `fmt` to imports:

```go
import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"
)
```

Then add this check in `reconcileTagPublisherAssignment` immediately after `selected := selectPublishersByPlacement(...)`:

```go
if len(selected) == 0 {
	return TagPublisherAssignmentOutputs{}, fmt.Errorf("publisherPlacementLabels %v did not match any managed publishers", args.PublisherPlacementLabels)
}
```

- [x] **Step 4: Run the validation test and verify it passes**

Run:

```bash
npm run go:test -- -run TestTagPublisherAssignmentFailsWhenPlacementLabelsSelectNoPublishers -count=1
```

Expected: PASS.

- [x] **Step 5: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/tag_publisher_assignment.go
git commit -m "fix: validate publisher placement labels"
```

## Task 5: Delete TagPublisherAssignment Remote Associations

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/netskope_client.go`
- Modify: `internal/provider/tag_publisher_assignment.go`

- [x] **Step 1: Add a delete test that unassigns selected publishers from matched apps**

Append this test to `internal/provider/provider_test.go`:

```go
func TestTagPublisherAssignmentDeleteRemovesSelectedPublishersFromMatchedApps(t *testing.T) {
	var deleteBodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/steering/apps/private":
			writeJSON(t, w, map[string]any{
				"status": "success",
				"data": []map[string]any{{
					"app_id":   10,
					"app_name": "orders",
					"tags":     []map[string]any{{"tag_name": "vpc-a"}},
					"service_publisher_assignments": []map[string]any{
						{"publisher_id": 99},
						{"publisher_id": 101},
					},
				}},
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
```

- [x] **Step 2: Run the delete test and verify it fails**

Run:

```bash
npm run go:test -- -run TestTagPublisherAssignmentDeleteRemovesSelectedPublishersFromMatchedApps -count=1
```

Expected: FAIL because `Delete` is currently a no-op.

- [x] **Step 3: Add a delete association client method**

In `internal/provider/netskope_client.go`, add:

```go
func (client *netskopeClient) deletePrivateAppPublishers(ctx context.Context, appIDs []int, publisherIDs []int) error {
	privateAppIDs := make([]string, 0, len(appIDs))
	for _, id := range appIDs {
		privateAppIDs = append(privateAppIDs, strconv.Itoa(id))
	}
	publisherIDValues := make([]string, 0, len(publisherIDs))
	for _, id := range publisherIDs {
		publisherIDValues = append(publisherIDValues, strconv.Itoa(id))
	}
	body := map[string]any{
		"private_app_ids": privateAppIDs,
		"publisher_ids":   publisherIDValues,
	}
	return client.request(ctx, "Delete private app publishers", http.MethodDelete, "/api/v2/steering/apps/private/publishers", body, nil)
}
```

- [x] **Step 4: Implement delete reconciliation**

In `internal/provider/tag_publisher_assignment.go`, replace `Delete` with:

```go
func (*TagPublisherAssignment) Delete(ctx context.Context, req infer.DeleteRequest[TagPublisherAssignmentOutputs]) (infer.DeleteResponse, error) {
	selected := selectPublishersByPlacement(req.State.Publishers, req.State.PublisherPlacementLabels)
	if len(selected) == 0 {
		return infer.DeleteResponse{}, nil
	}

	client := newResourceClient(req.State.TenantURL, req.State.APIToken, req.State.BearerToken, req.State.AuthMode, req.State.OAuth2, http.DefaultClient)
	apps, err := client.listPrivateAppsWithPublishers(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	selectedSet := intSet(selected)
	for _, app := range apps {
		if !appMatchesTags(app.Tags, req.State.AppTags, defaultString(req.State.MatchMode, "any")) {
			continue
		}
		current := currentPublisherIDs(app.ServicePublisherAssignments)
		toRemove := intersectPublisherIDs(current, selectedSet)
		if len(toRemove) == 0 {
			continue
		}
		if err := client.deletePrivateAppPublishers(ctx, []int{app.resourceID()}, toRemove); err != nil {
			return infer.DeleteResponse{}, err
		}
	}
	return infer.DeleteResponse{}, nil
}
```

Add this helper below `reconcilePublisherIDs`:

```go
func intersectPublisherIDs(current []int, selected map[int]bool) []int {
	var values []int
	for _, id := range current {
		if selected[id] {
			values = append(values, id)
		}
	}
	sort.Ints(values)
	return values
}
```

- [x] **Step 5: Run the delete test and verify it passes**

Run:

```bash
npm run go:test -- -run TestTagPublisherAssignmentDeleteRemovesSelectedPublishersFromMatchedApps -count=1
```

Expected: PASS.

- [x] **Step 6: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/netskope_client.go internal/provider/tag_publisher_assignment.go
git commit -m "fix: delete tag publisher assignments"
```

## Task 6: Add Remote Read for PrivateApp and RealtimeProtectionPolicy

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/netskope_client.go`
- Modify: `internal/provider/private_app.go`
- Modify: `internal/provider/realtime_protection_policy.go`

- [x] **Step 1: Add private app and realtime policy read tests**

Append these tests to `internal/provider/provider_test.go`:

```go
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
```

- [x] **Step 2: Run the read tests and verify they fail**

Run:

```bash
npm run go:test -- -run 'Test(PrivateApp|RealtimeProtectionPolicy)ReadDropsResourceWhenRemote.*Missing' -count=1
```

Expected: FAIL because `Read` currently returns prior state without remote calls.

- [x] **Step 3: Add client read methods and 404 detection**

In `internal/provider/netskope_client.go`, add this sentinel near the imports or below the client type:

```go
var errNetskopeNotFound = fmt.Errorf("netskope resource not found")
```

In `request`, before the generic non-2xx error block, add:

```go
if response.StatusCode == http.StatusNotFound {
	return errNetskopeNotFound
}
```

Add these methods:

```go
func (client *netskopeClient) getPrivateApp(ctx context.Context, id int) (privateAppRecord, error) {
	var response struct {
		Status string           `json:"status"`
		Data   privateAppRecord `json:"data"`
	}
	path := fmt.Sprintf("/api/v2/steering/apps/private/%d", id)
	if err := client.request(ctx, fmt.Sprintf("Get private app %d", id), http.MethodGet, path, nil, &response); err != nil {
		return privateAppRecord{}, err
	}
	return response.Data, nil
}

func (client *netskopeClient) getRealtimePolicy(ctx context.Context, id int) (realtimePolicyRecord, error) {
	var response realtimePolicyRecord
	path := fmt.Sprintf("/api/v2/policy/npa/rules/%d", id)
	if err := client.request(ctx, fmt.Sprintf("Get realtime protection policy %d", id), http.MethodGet, path, nil, &response); err != nil {
		return realtimePolicyRecord{}, err
	}
	return response, nil
}
```

- [x] **Step 4: Implement resource reads**

In `internal/provider/private_app.go`, replace `Read` with:

```go
func (*PrivateApp) Read(ctx context.Context, req infer.ReadRequest[PrivateAppArgs, PrivateAppOutputs]) (infer.ReadResponse[PrivateAppArgs, PrivateAppOutputs], error) {
	appID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.ReadResponse[PrivateAppArgs, PrivateAppOutputs]{}, fmt.Errorf("invalid private app ID %q: %w", req.ID, err)
	}
	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	app, err := client.getPrivateApp(ctx, appID)
	if err != nil {
		if err == errNetskopeNotFound {
			return infer.ReadResponse[PrivateAppArgs, PrivateAppOutputs]{}, nil
		}
		return infer.ReadResponse[PrivateAppArgs, PrivateAppOutputs]{}, err
	}
	state := req.State
	state.AppID = app.resourceID()
	return infer.ReadResponse[PrivateAppArgs, PrivateAppOutputs]{ID: strconv.Itoa(state.AppID), Inputs: req.Inputs, State: state}, nil
}
```

In `internal/provider/realtime_protection_policy.go`, replace `Read` with:

```go
func (*RealtimeProtectionPolicy) Read(ctx context.Context, req infer.ReadRequest[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]) (infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs], error) {
	policyID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{}, fmt.Errorf("invalid realtime policy ID %q: %w", req.ID, err)
	}
	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	policy, err := client.getRealtimePolicy(ctx, policyID)
	if err != nil {
		if err == errNetskopeNotFound {
			return infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{}, nil
		}
		return infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{}, err
	}
	state := req.State
	state.PolicyID = policy.RuleID
	return infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{ID: strconv.Itoa(state.PolicyID), Inputs: req.Inputs, State: state}, nil
}
```

- [x] **Step 5: Run the read tests and verify they pass**

Run:

```bash
npm run go:test -- -run 'Test(PrivateApp|RealtimeProtectionPolicy)ReadDropsResourceWhenRemote.*Missing' -count=1
```

Expected: PASS.

- [x] **Step 6: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/netskope_client.go internal/provider/private_app.go internal/provider/realtime_protection_policy.go
git commit -m "fix: detect removed npa resources during refresh"
```

## Task 7: Use Documented PUT for PrivateApp Update

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/netskope_client.go`

- [x] **Step 1: Update the adoption test to expect PUT**

In `TestPrivateAppAdoptsExistingByName`, replace:

```go
	case r.Method == http.MethodPatch && r.URL.Path == "/api/v2/steering/apps/private/44":
```

with:

```go
	case r.Method == http.MethodPut && r.URL.Path == "/api/v2/steering/apps/private/44":
```

Replace:

```go
if !patched {
	t.Fatalf("expected adopted app to be reconciled with PATCH")
}
```

with:

```go
if !patched {
	t.Fatalf("expected adopted app to be reconciled with PUT")
}
```

- [x] **Step 2: Run the adoption test and verify it fails**

Run:

```bash
npm run go:test -- -run TestPrivateAppAdoptsExistingByName -count=1
```

Expected: FAIL because the client still uses `PATCH`.

- [x] **Step 3: Change the client update method to PUT**

In `internal/provider/netskope_client.go`, replace:

```go
if err := client.request(ctx, "Update private app "+payload.AppName, http.MethodPatch, path, payload, &response); err != nil {
```

with:

```go
if err := client.request(ctx, "Update private app "+payload.AppName, http.MethodPut, path, payload, &response); err != nil {
```

- [x] **Step 4: Run the adoption test and verify it passes**

Run:

```bash
npm run go:test -- -run TestPrivateAppAdoptsExistingByName -count=1
```

Expected: PASS.

- [x] **Step 5: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/netskope_client.go
git commit -m "fix: update private apps with documented method"
```

## Task 8: Add Guardrails for PrivateApp Host Shape

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/private_app.go`

- [x] **Step 1: Add a validation test for `hosts`**

Append this test to `internal/provider/provider_test.go`:

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

- [x] **Step 2: Run the validation test and verify it fails**

Run:

```bash
npm run go:test -- -run TestPrivateAppRejectsHostsUntilApiShapeIsConfirmed -count=1
```

Expected: FAIL because `hosts` is currently accepted and sent as an array.

- [x] **Step 3: Add explicit validation before live API calls**

In `internal/provider/private_app.go`, add this helper:

```go
func validatePrivateAppArgs(args PrivateAppArgs) error {
	if len(args.Hosts) > 0 {
		return fmt.Errorf("hosts is not supported by the documented private app API; use host")
	}
	return nil
}
```

Call it in `Create` after the dry-run block and before creating the client:

```go
if err := validatePrivateAppArgs(req.Inputs); err != nil {
	return infer.CreateResponse[PrivateAppOutputs]{}, err
}
```

Call it in `Update` after parsing the ID and before the dry-run block:

```go
if err := validatePrivateAppArgs(req.Inputs); err != nil {
	return infer.UpdateResponse[PrivateAppOutputs]{}, err
}
```

In `privateAppPayloadFromArgs`, remove the `Hosts` array override and keep:

```go
host := any(args.Host)
```

- [x] **Step 4: Run the validation test and verify it passes**

Run:

```bash
npm run go:test -- -run TestPrivateAppRejectsHostsUntilApiShapeIsConfirmed -count=1
```

Expected: PASS.

- [x] **Step 5: Commit**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/private_app.go
git commit -m "fix: reject unsupported private app hosts array"
```

## Task 9: Run Full Verification and Regenerate if Needed

**Files:**
- Possibly modify: `schema.json`
- Possibly modify: `src/privateApp.ts`
- Possibly modify: `src/realtimeProtectionPolicy.ts`
- Possibly modify: `src/tagPublisherAssignment.ts`
- Possibly modify: generated docs under `docs/`

- [x] **Step 1: Run Go provider tests**

Run:

```bash
npm run go:test
```

Expected: PASS for `./cmd/pulumi-resource-netskope-publisher` and `./internal/provider`.

- [x] **Step 2: Run JS/build tests**

Run:

```bash
npm test
```

Expected: PASS. This runs the build and Node test harness.

- [x] **Step 3: Regenerate SDK and docs only if schema changed**

Run:

```bash
npm run sdk:gen
```

Expected: either no diff, or generated TypeScript/schema updates that reflect changed provider metadata.

Run:

```bash
npm run docs:gen
```

Expected: either no diff, or docs updates that reflect changed provider metadata.

- [x] **Step 4: Inspect the final diff**

Run:

```bash
git diff --stat
```

Expected: diffs are limited to provider code, tests, this plan, and generated files if generation changed output.

Run:

```bash
git diff -- internal/provider/netskope_client.go internal/provider/private_app.go internal/provider/realtime_protection_policy.go internal/provider/tag_publisher_assignment.go internal/provider/provider_test.go
```

Expected: all API request bodies use documented field names; dry-run does not call live policy group lookup; read detects 404; assignment delete removes mappings only.

- [x] **Step 5: Commit verification or generated changes**

If `npm run sdk:gen` or `npm run docs:gen` produced changes, run:

```bash
git add schema.json src docs
git commit -m "chore: regenerate provider surfaces"
```

If generation produced no changes, skip this commit.

## Self-Review

- Spec coverage: The plan covers realtime policy contract repair, preview safety, app ID assignment, placement validation, assignment delete behavior, refresh detection, private app update method alignment, and the ambiguous `hosts` API shape.
- Placeholder scan: The plan contains no deferred implementation markers and no generic edge-case instructions.
- Type consistency: Realtime IDs use `RuleID`; private app IDs use `resourceID()`; assignment requests use `private_app_ids` and `publisher_ids`; `Groups` maps to `userGroups`.

