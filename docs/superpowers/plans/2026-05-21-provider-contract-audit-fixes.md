# Provider Contract Audit Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Fix the provider bugs found in the latest architecture audit so NPA resources match the documented Netskope API contracts and Pulumi lifecycle behavior.

**Architecture:** Keep the low-level Netskope API contract in `internal/provider/netskope_client.go`, then map Pulumi resource inputs into that contract in the resource files. Each bug gets a failing contract test first, followed by the smallest implementation change and a focused verification command.

**Tech Stack:** Go provider code using `pulumi-go-provider`, `net/http/httptest`, local Swagger reference at `/Users/jneerdael/Scripts/privateaccess-mcp/swagger.json`, TypeScript wrapper code, generated Pulumi schema/SDKs, npm test orchestration.

---

## File Structure

- Modify `internal/provider/netskope_client.go`: owns Netskope request and response structs. Fix private app protocol/publisher request types, policy group response parsing, and 404 handling helpers.
- Modify `internal/provider/private_app.go`: owns `PrivateApp` lifecycle and maps Pulumi args into private app API payloads. Keep public `ports` input but emit API `port`.
- Modify `internal/provider/realtime_protection_policy.go`: owns `RealtimeProtectionPolicy` lifecycle. Keep policy group name lookup behavior but consume fixed client records.
- Modify `internal/provider/tag_publisher_assignment.go`: owns tag-to-publisher selection. Change placement label selection from any-label to all-label matching.
- Modify `internal/provider/provider_test.go`: add red tests for each audited bug before implementation.
- Regenerate only after schema-affecting changes. These tasks should not intentionally change the Pulumi schema, so `schema.json` and SDK outputs should remain unchanged after `go run ./cmd/pulumi-resource-netskope-publisher --schema > schema.json`.

## Task 1: Fix Realtime Policy Group Lookup Contract

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/netskope_client.go`

- [x] **Step 1: Replace the policy group lookup fixture with the Swagger response shape**

In `internal/provider/provider_test.go`, update `TestRealtimeProtectionPolicyCreatesRuleWithPolicyGroupReference` so the `GET /api/v2/policy/npa/policygroups` branch returns a bare array with `group_id` and `group_name`:

```go
case r.Method == http.MethodGet && r.URL.Path == "/api/v2/policy/npa/policygroups":
	writeJSON(t, w, []map[string]any{{
		"group_id":   12,
		"group_name": "default",
	}})
```

- [x] **Step 2: Run the focused test and verify it fails**

Run:

```bash
npm run go:test -- -run TestRealtimeProtectionPolicyCreatesRuleWithPolicyGroupReference -count=1
```

Expected: FAIL with an error containing `policy group "default" not found`.

- [x] **Step 3: Change the policy group response model to Swagger field names**

In `internal/provider/netskope_client.go`, replace `policyGroupRecord` and `findPolicyGroupByName` with:

```go
type policyGroupRecord struct {
	ID   int    `json:"group_id"`
	Name string `json:"group_name"`
}

func (client *netskopeClient) findPolicyGroupByName(ctx context.Context, name string) (*policyGroupRecord, error) {
	var response []policyGroupRecord
	if err := client.request(ctx, "List policy groups", http.MethodGet, "/api/v2/policy/npa/policygroups", nil, &response); err != nil {
		return nil, err
	}
	for _, group := range response {
		if group.Name == name {
			return &group, nil
		}
	}
	return nil, nil
}
```

- [x] **Step 4: Run the focused test and verify it passes**

Run:

```bash
npm run go:test -- -run TestRealtimeProtectionPolicyCreatesRuleWithPolicyGroupReference -count=1
```

Expected: PASS.

- [x] **Step 5: Commit the policy group contract fix**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/netskope_client.go
git commit -m "fix: parse realtime policy groups from swagger shape"
```

## Task 2: Emit Swagger-Compatible Private App Publisher IDs

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/netskope_client.go`
- Modify: `internal/provider/private_app.go`

- [x] **Step 1: Change the publisher payload assertion to expect a string ID**

In `TestPrivateAppCreateIncludesInitialPublishersWhenProvided`, replace:

```go
if publisher["publisher_id"] != float64(101) {
	t.Fatalf("expected publisher_id 101, got %#v", created)
}
```

with:

```go
if publisher["publisher_id"] != "101" {
	t.Fatalf("expected publisher_id string 101, got %#v", created)
}
```

- [x] **Step 2: Run the focused test and verify it fails**

Run:

```bash
npm run go:test -- -run TestPrivateAppCreateIncludesInitialPublishersWhenProvided -count=1
```

Expected: FAIL because the body currently contains numeric `publisher_id`.

- [x] **Step 3: Change the private app publisher payload type to string**

In `internal/provider/netskope_client.go`, replace:

```go
type privateAppPublisher struct {
	PublisherID   int    `json:"publisher_id"`
	PublisherName string `json:"publisher_name,omitempty"`
}
```

with:

```go
type privateAppPublisher struct {
	PublisherID   string `json:"publisher_id"`
	PublisherName string `json:"publisher_name,omitempty"`
}
```

- [x] **Step 4: Convert Pulumi publisher IDs to strings in the payload mapper**

In `internal/provider/private_app.go`, update the `publishers` append block inside `privateAppPayloadFromArgs`:

```go
publishers = append(publishers, privateAppPublisher{
	PublisherID:   strconv.Itoa(publisher.PublisherID),
	PublisherName: stringValue(publisher.PublisherName),
})
```

The file already imports `strconv`, so no import change is needed.

- [x] **Step 5: Run the focused test and verify it passes**

Run:

```bash
npm run go:test -- -run TestPrivateAppCreateIncludesInitialPublishersWhenProvided -count=1
```

Expected: PASS.

- [x] **Step 6: Commit the private app publisher ID fix**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/netskope_client.go internal/provider/private_app.go
git commit -m "fix: send private app publisher ids as strings"
```

## Task 3: Stop Sending Response-Only Private App Protocol Fields

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/netskope_client.go`
- Modify: `internal/provider/private_app.go`

- [x] **Step 1: Add a contract assertion that `ports` is not sent**

In `TestPrivateAppCreateIncludesInitialPublishersWhenProvided`, after decoding `created`, add assertions before the publisher assertions:

```go
protocols := created["protocols"].([]any)
protocol := protocols[0].(map[string]any)
if protocol["port"] != "443" {
	t.Fatalf("expected protocol port 443, got %#v", created)
}
if _, ok := protocol["ports"]; ok {
	t.Fatalf("did not expect response-only protocol ports in request payload: %#v", created)
}
```

- [x] **Step 2: Run the focused test and verify it fails**

Run:

```bash
npm run go:test -- -run TestPrivateAppCreateIncludesInitialPublishersWhenProvided -count=1
```

Expected: FAIL because the request body currently includes `protocols[0].ports`.

- [x] **Step 3: Remove the response-only `ports` JSON field from request payloads**

In `internal/provider/netskope_client.go`, replace:

```go
type privateAppProtocol struct {
	Type  string `json:"type"`
	Ports string `json:"ports,omitempty"`
	Port  string `json:"port,omitempty"`
}
```

with:

```go
type privateAppProtocol struct {
	Type string `json:"type"`
	Port string `json:"port,omitempty"`
}
```

- [x] **Step 4: Update the private app payload mapper to populate only `port`**

In `internal/provider/private_app.go`, replace the protocol append block inside `privateAppPayloadFromArgs` with:

```go
protocols = append(protocols, privateAppProtocol{
	Type: protocol.Type,
	Port: protocol.Ports,
})
```

- [x] **Step 5: Run the focused test and verify it passes**

Run:

```bash
npm run go:test -- -run TestPrivateAppCreateIncludesInitialPublishersWhenProvided -count=1
```

Expected: PASS.

- [x] **Step 6: Commit the private app protocol payload fix**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/netskope_client.go internal/provider/private_app.go
git commit -m "fix: send private app protocol port only"
```

## Task 4: Require All Publisher Placement Labels

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/tag_publisher_assignment.go`

- [x] **Step 1: Add a failing test for compound placement labels**

In `internal/provider/provider_test.go`, add this test after `TestTagPublisherAssignmentFailsWhenPlacementLabelsSelectNoPublishers`:

```go
func TestTagPublisherAssignmentRequiresAllPlacementLabels(t *testing.T) {
	_, err := createTagPublisherAssignmentResource(t, property.NewMap(map[string]property.Value{
		"tenantUrl":                property.New("https://tenant.example"),
		"bearerToken":              property.New("api-token"),
		"appTags":                  property.New([]property.Value{property.New("vpc-a")}),
		"publisherPlacementLabels": property.New([]property.Value{property.New("aws"), property.New("vpc-a")}),
		"publishers": property.New(map[string]property.Value{
			"pub-wrong-vpc": property.New(map[string]property.Value{
				"publisherId": property.New(202.0),
				"placementLabels": property.New([]property.Value{
					property.New("aws"),
					property.New("vpc-b"),
				}),
			}),
		}),
	}))
	if err == nil {
		t.Fatalf("expected validation error when publisher matches only one placement label")
	}
	if !strings.Contains(err.Error(), `publisherPlacementLabels [aws vpc-a] did not match any managed publishers`) {
		t.Fatalf("expected placement label validation error, got %v", err)
	}
}
```

- [x] **Step 2: Run the focused test and verify it fails**

Run:

```bash
npm run go:test -- -run TestTagPublisherAssignmentRequiresAllPlacementLabels -count=1
```

Expected: FAIL because the current selection matches `aws` and wrongly selects the publisher.

- [x] **Step 3: Replace any-label matching with all-label matching**

In `internal/provider/tag_publisher_assignment.go`, replace `selectPublishersByPlacement` with:

```go
func selectPublishersByPlacement(publishers map[string]PublisherAssignmentInput, labels []string) []int {
	var selected []int
	for _, publisher := range publishers {
		if containsAllStrings(publisher.PlacementLabels, labels) {
			selected = append(selected, publisher.PublisherID)
		}
	}
	sort.Ints(selected)
	return selected
}
```

- [x] **Step 4: Replace `intersectsStringSet` with `containsAllStrings`**

In `internal/provider/tag_publisher_assignment.go`, delete `intersectsStringSet` and add:

```go
func containsAllStrings(values []string, required []string) bool {
	if len(required) == 0 {
		return false
	}
	set := stringSet(values)
	for _, value := range required {
		if !set[value] {
			return false
		}
	}
	return true
}
```

- [x] **Step 5: Run the focused tests and verify they pass**

Run:

```bash
npm run go:test -- -run 'TestTagPublisherAssignment(AddsAndRemovesOnlySelectedPublishers|FailsWhenPlacementLabelsSelectNoPublishers|RequiresAllPlacementLabels|DeleteRemovesSelectedPublishersFromMatchedApps|ReadRecomputesRemoteState)' -count=1
```

Expected: PASS.

- [x] **Step 6: Commit the placement label matching fix**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/tag_publisher_assignment.go
git commit -m "fix: require all publisher placement labels"
```

## Task 5: Make Remote Deletes Idempotent on 404

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/private_app.go`
- Modify: `internal/provider/realtime_protection_policy.go`

- [x] **Step 1: Add a failing test for private app delete after remote removal**

In `internal/provider/provider_test.go`, add this test after `TestPrivateAppReadDropsResourceWhenRemoteAppIsMissing`:

```go
func TestPrivateAppDeleteIgnoresRemoteNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	resource := PrivateApp{}
	_, err := resource.Delete(t.Context(), infer.DeleteRequest[PrivateAppOutputs]{
		ID: "44",
		State: PrivateAppOutputs{
			PrivateAppArgs: PrivateAppArgs{
				TenantURL:   server.URL,
				BearerToken: stringPtr("api-token"),
			},
			AppID: 44,
		},
	})
	if err != nil {
		t.Fatalf("expected delete to ignore remote 404, got %v", err)
	}
}
```

- [x] **Step 2: Add a failing test for realtime policy delete after remote removal**

In `internal/provider/provider_test.go`, add this test after `TestRealtimeProtectionPolicyReadDropsResourceWhenRemotePolicyIsMissing`:

```go
func TestRealtimeProtectionPolicyDeleteIgnoresRemoteNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	resource := RealtimeProtectionPolicy{}
	_, err := resource.Delete(t.Context(), infer.DeleteRequest[RealtimeProtectionPolicyOutputs]{
		ID: "55",
		State: RealtimeProtectionPolicyOutputs{
			RealtimeProtectionPolicyArgs: RealtimeProtectionPolicyArgs{
				TenantURL:   server.URL,
				BearerToken: stringPtr("api-token"),
			},
			PolicyID: 55,
		},
	})
	if err != nil {
		t.Fatalf("expected delete to ignore remote 404, got %v", err)
	}
}
```

- [x] **Step 3: Run the focused tests and verify they fail**

Run:

```bash
npm run go:test -- -run 'Test(PrivateApp|RealtimeProtectionPolicy)DeleteIgnoresRemoteNotFound' -count=1
```

Expected: FAIL because deletes return `errNetskopeNotFound`.

- [x] **Step 4: Ignore remote not found in private app delete**

In `internal/provider/private_app.go`, replace the return in `Delete` with:

```go
err = client.deletePrivateApp(ctx, appID)
if err == errNetskopeNotFound {
	return infer.DeleteResponse{}, nil
}
return infer.DeleteResponse{}, err
```

- [x] **Step 5: Ignore remote not found in realtime policy delete**

In `internal/provider/realtime_protection_policy.go`, replace the return in `Delete` with:

```go
err = client.deleteRealtimePolicy(ctx, policyID)
if err == errNetskopeNotFound {
	return infer.DeleteResponse{}, nil
}
return infer.DeleteResponse{}, err
```

- [x] **Step 6: Run the focused tests and verify they pass**

Run:

```bash
npm run go:test -- -run 'Test(PrivateApp|RealtimeProtectionPolicy)DeleteIgnoresRemoteNotFound' -count=1
```

Expected: PASS.

- [x] **Step 7: Commit idempotent delete behavior**

Run:

```bash
git add internal/provider/provider_test.go internal/provider/private_app.go internal/provider/realtime_protection_policy.go
git commit -m "fix: ignore remote not found during deletes"
```

## Task 6: Regenerate Schema and Verify No Public Schema Drift

**Files:**
- Check: `schema.json`
- Check: `sdk/rust/schema.json`
- Check generated SDK directories for unexpected changes.

- [x] **Step 1: Format Go files**

Run:

```bash
gofmt -w internal/provider/netskope_client.go internal/provider/private_app.go internal/provider/realtime_protection_policy.go internal/provider/tag_publisher_assignment.go internal/provider/provider_test.go
```

Expected: command exits 0.

- [x] **Step 2: Regenerate provider schema**

Run:

```bash
go run ./cmd/pulumi-resource-netskope-publisher --schema > schema.json
cp schema.json sdk/rust/schema.json
```

Expected: command exits 0.

- [x] **Step 3: Confirm schema changes are absent or only formatting-neutral**

Run:

```bash
git diff -- schema.json sdk/rust/schema.json
```

Expected: no diff. If there is a diff only from pretty-printing, normalize both files:

```bash
jq -c . schema.json > /tmp/schema.json
mv /tmp/schema.json schema.json
jq -c . sdk/rust/schema.json > /tmp/rust-schema.json
mv /tmp/rust-schema.json sdk/rust/schema.json
```

Then run:

```bash
git diff -- schema.json sdk/rust/schema.json
```

Expected: no diff.

- [x] **Step 4: Run full Go verification**

Run:

```bash
npm run go:test
```

Expected: PASS for `./cmd/pulumi-resource-netskope-publisher` and `./internal/provider`.

- [x] **Step 5: Run full TypeScript and Node verification**

Run:

```bash
npm test
```

Expected: PASS with all Node tests passing.

- [x] **Step 6: Check final working tree**

Run:

```bash
git status --short
```

Expected: only intentional source/test changes from this plan are present, or clean if every task commit has already been made.

- [x] **Step 7: Commit any final formatting-only changes**

If `gofmt` changed files after the previous task commits, run:

```bash
git add internal/provider/netskope_client.go internal/provider/private_app.go internal/provider/realtime_protection_policy.go internal/provider/tag_publisher_assignment.go internal/provider/provider_test.go
git commit -m "chore: format provider contract fixes"
```

If there are no changes, do not create an empty commit.

## Self-Review

- Spec coverage: all five audit findings are covered by Tasks 1-5, and full verification/schema checks are covered by Task 6.
- Placeholder scan: no `TBD`, `TODO`, or unspecified implementation steps remain.
- Type consistency: publisher IDs remain public Pulumi `int` inputs and become API `string` request fields; private app public `ports` remains the Pulumi input while API payload emits singular `port`; policy group lookup uses `group_id/group_name` consistently.

