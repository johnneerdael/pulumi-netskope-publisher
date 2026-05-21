# Provider Audit Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the provider architecture audit findings for provider token drift, Go/TypeScript output parity, and the unused Azure marketplace terms input.

**Architecture:** Keep the current dual implementation model, but make the executable Go provider and TypeScript SDK obey the same observable provider contracts. The fixes are test-first: assert canonical child tokens, assert IP output parity for Go providers, and make the Azure marketplace terms behavior explicit instead of silently ignoring an input.

**Tech Stack:** TypeScript, Node test runner, Pulumi TypeScript mocks, Go, `pulumi-go-provider` integration tests, Pulumi Registry resource tokens, Hexo documentation.

---

## File Structure

- Modify `src/scalewayPublisher.ts`: instantiate the canonical Scaleway resource export rather than the legacy alias namespace.
- Modify `test/scalewayPublisher.test.ts`: assert the TypeScript Scaleway component creates the canonical registry resource token.
- Modify `internal/provider/components.go`: update the Go Scaleway child token, map GCP/Hcloud/OpenStack private IP outputs, and add helper output extraction for nested GCP network interface data.
- Modify `internal/provider/provider_test.go`: update Scaleway token assertions and add Go provider output parity tests for GCP, Hcloud, and OpenStack private/public IPs.
- Modify `src/azurePublisher.ts`: fail fast when `acceptMarketplaceTerms: true` is supplied, since the component does not accept marketplace terms.
- Modify `src/types.ts`: document `acceptMarketplaceTerms` as currently unsupported in TypeScript.
- Modify `internal/provider/components.go`: fail fast when Go Azure receives `AcceptMarketplaceTerms == true`.
- Modify `test/azurePublisher.test.ts`: assert TypeScript fails explicitly for `acceptMarketplaceTerms: true`.
- Modify `internal/provider/provider_test.go`: assert Go fails explicitly for `acceptMarketplaceTerms: true`.
- Modify `site/source/admin/component/azure.md`: remove the implication that `acceptMarketplaceTerms` is implemented and document external marketplace terms acceptance.

---

### Task 1: Fix Scaleway Resource Token Drift

**Execution correction:** During execution, `@pulumiverse/scaleway` 1.49.0 showed that
`scaleway:index/instanceServer:InstanceServer` is the deprecated alias and
`scaleway:instance/server:Server` is the current resource token. The implemented
fix keeps both TypeScript and Go on the current token and moves the catalog
registry checks away from the deprecated alias.

**Files:**
- Modify: `test/scalewayPublisher.test.ts`
- Modify: `src/scalewayPublisher.ts`
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/components.go`

- [ ] **Step 1: Write the failing TypeScript token test**

In `test/scalewayPublisher.test.ts`, replace the mock branch:

```ts
    if (args.type === "scaleway:instance/server:Server") {
      createdServers[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          privateIps: [{ address: "10.1.0.10" }],
          publicIps: [{ address: "198.51.100.21" }],
        },
      };
    }
```

with:

```ts
    if (args.type === "scaleway:index/instanceServer:InstanceServer") {
      createdServers[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          privateIps: [{ address: "10.1.0.10" }],
          publicIps: [{ address: "198.51.100.21" }],
        },
      };
    }

    if (args.type === "scaleway:instance/server:Server") {
      throw new Error("ScalewayPublisher must use canonical scaleway:index/instanceServer:InstanceServer token");
    }
```

- [ ] **Step 2: Run the focused TypeScript test to verify it fails**

Run:

```bash
npm run build
node --test dist/test/scalewayPublisher.test.js
```

Expected: FAIL with `ScalewayPublisher must use canonical scaleway:index/instanceServer:InstanceServer token`.

- [ ] **Step 3: Switch TypeScript Scaleway to the canonical resource class**

In `src/scalewayPublisher.ts`, replace:

```ts
      const server = new scaleway.instance.Server(`${name}-${publisherName}`, {
```

with:

```ts
      const server = new scaleway.InstanceServer(`${name}-${publisherName}`, {
```

Keep the existing inputs unchanged:

```ts
        name: publisherName,
        type: args.type ?? "DEV1-M",
        image: args.image ?? "ubuntu_jammy",
        zone: args.zone,
        securityGroupId: args.securityGroupId,
        enableDynamicIp: args.enableDynamicIp ?? true,
        cloudInit: userDataPlacement.cloudInit as pulumi.Input<string>,
        userData: userDataPlacement.userData as pulumi.Input<Record<string, pulumi.Input<string>>>,
        tags: pulumi.output(args.tags ?? {}).apply((tags) =>
          Object.entries(tags).map(([key, value]) => `${key}=${value}`),
        ),
```

- [ ] **Step 4: Update the Go token tests to expect the canonical token**

In `internal/provider/provider_test.go`, replace every expected Scaleway child token:

```go
"scaleway:instance/server:Server"
```

with:

```go
"scaleway:index/instanceServer:InstanceServer"
```

Specifically update the cases in `TestAdditionalProviderConstructsExpectedChildResources`, `TestAdditionalProviderConstructsBootstrapWithRegistryFields`, and the mock branch inside `constructAndCollectPublisherOutput`.

- [ ] **Step 5: Run the focused Go tests to verify they fail**

Run:

```bash
gofmt -w internal/provider/provider_test.go
go test ./internal/provider -run 'TestAdditionalProviderConstructsExpectedChildResources|TestAdditionalProviderConstructsBootstrapWithRegistryFields|TestProviderOutputsExposeAvailableIPAddresses' -count=1
```

Expected: FAIL for Scaleway because `components.go` still registers `scaleway:instance/server:Server`.

- [ ] **Step 6: Switch the Go Scaleway registration token**

In `internal/provider/components.go`, inside `NewScalewayPublisher`, replace:

```go
		err := ctx.RegisterResource("scaleway:instance/server:Server", name+"-"+publisherName, pulumi.Map{
```

with:

```go
		err := ctx.RegisterResource("scaleway:index/instanceServer:InstanceServer", name+"-"+publisherName, pulumi.Map{
```

- [ ] **Step 7: Verify Scaleway token fixes**

Run:

```bash
npm run build
node --test dist/test/scalewayPublisher.test.js
gofmt -w internal/provider/components.go internal/provider/provider_test.go
go test ./internal/provider -run 'TestAdditionalProviderConstructsExpectedChildResources|TestAdditionalProviderConstructsBootstrapWithRegistryFields|TestProviderOutputsExposeAvailableIPAddresses' -count=1
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add src/scalewayPublisher.ts test/scalewayPublisher.test.ts internal/provider/components.go internal/provider/provider_test.go
git commit -m "fix: use canonical scaleway registry resource token"
```

---

### Task 2: Restore Go Provider IP Output Parity

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/components.go`

- [ ] **Step 1: Add failing Go output assertions for GCP, Hcloud, and OpenStack**

In `internal/provider/provider_test.go`, inside `constructAndCollectPublisherOutput`, add or update these mock cases:

```go
				case "gcp:compute/instance:Instance":
					state = state.Set("instanceId", property.New("gcp-instance-id"))
					state = state.Set("networkInterfaces", property.New([]property.Value{property.New(map[string]property.Value{
						"networkIp": property.New("10.2.0.10"),
						"accessConfigs": property.New([]property.Value{property.New(map[string]property.Value{
							"natIp": property.New("203.0.113.20"),
						})}),
					})}))
				case "hcloud:index/server:Server":
					state = state.Set("ipv4Address", property.New("203.0.113.10"))
					state = state.Set("networks", property.New([]property.Value{property.New(map[string]property.Value{
						"ip": property.New("10.0.0.10"),
					})}))
				case "openstack:compute/instance:Instance":
					state = state.Set("accessIpV4", property.New("198.51.100.25"))
					state = state.Set("networks", property.New([]property.Value{property.New(map[string]property.Value{
						"fixedIpV4": property.New("10.5.0.10"),
						"port":      property.New("port-123"),
					})}))
```

In `TestProviderOutputsExposeAvailableIPAddresses`, add these cases:

```go
		{
			name:  "GCP private IP",
			token: "netskope-publisher:index:GcpPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"project":       property.New("project"),
				"zone":          property.New("europe-west4-a"),
				"network":       property.New("default"),
				"subnetwork":    property.New("default"),
				"image":         property.New("projects/example/global/images/npa"),
				"assignPublicIp": property.New(true),
			}),
			outputField: "privateIp",
			expected:    "10.2.0.10",
		},
		{
			name:  "GCP public IP",
			token: "netskope-publisher:index:GcpPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"project":       property.New("project"),
				"zone":          property.New("europe-west4-a"),
				"network":       property.New("default"),
				"subnetwork":    property.New("default"),
				"image":         property.New("projects/example/global/images/npa"),
				"assignPublicIp": property.New(true),
			}),
			outputField: "publicIp",
			expected:    "203.0.113.20",
		},
		{
			name:  "Hcloud private IP",
			token: "netskope-publisher:index:HcloudPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"networkId":     property.New(123.0),
			}),
			outputField: "privateIp",
			expected:    "10.0.0.10",
		},
		{
			name:  "OpenStack private IP",
			token: "netskope-publisher:index:OpenstackPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"imageName":     property.New("Ubuntu 22.04"),
				"flavorName":    property.New("m1.medium"),
				"networkName":   property.New("private"),
			}),
			outputField: "privateIp",
			expected:    "10.5.0.10",
		},
```

- [ ] **Step 2: Run the focused Go output test to verify it fails**

Run:

```bash
gofmt -w internal/provider/provider_test.go
go test ./internal/provider -run TestProviderOutputsExposeAvailableIPAddresses -count=1
```

Expected: FAIL for GCP, Hcloud, and OpenStack private/public IP outputs because `components.go` still returns empty strings for those fields.

- [ ] **Step 3: Add `NetworkInterfaces` to the raw VM resource**

In `internal/provider/components.go`, update `rawVMResource` by adding:

```go
	NetworkInterfaces  pulumi.ArrayOutput       `pulumi:"networkInterfaces"`
```

The relevant section should include:

```go
	Networks           pulumi.ArrayOutput       `pulumi:"networks"`
	NetworkInterfaces  pulumi.ArrayOutput       `pulumi:"networkInterfaces"`
	NicListStatuses    pulumi.AnyOutput         `pulumi:"nicListStatuses"`
```

- [ ] **Step 4: Add a helper for GCP nested NAT output extraction**

In `internal/provider/components.go`, immediately after `firstMapFieldOutput`, add:

```go
func firstNestedMapFieldOutput(values pulumi.ArrayOutput, arrayField string, field string) pulumi.StringOutput {
	return values.ApplyT(func(items []interface{}) string {
		if len(items) == 0 {
			return ""
		}
		item, ok := items[0].(map[string]interface{})
		if !ok {
			return ""
		}
		nestedValues, ok := item[arrayField].([]interface{})
		if !ok || len(nestedValues) == 0 {
			return ""
		}
		nested, ok := nestedValues[0].(map[string]interface{})
		if !ok {
			return ""
		}
		value, ok := nested[field]
		if !ok || value == nil {
			return ""
		}
		return fmt.Sprint(value)
	}).(pulumi.StringOutput)
}
```

- [ ] **Step 5: Wire GCP, Hcloud, and OpenStack output mappings**

In `NewGcpPublisher`, replace:

```go
		outputs[publisherName] = publisherOutput(registration, instance.InstanceId, pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput(), args.PlacementLabels)
```

with:

```go
		outputs[publisherName] = publisherOutput(
			registration,
			instance.InstanceId,
			firstMapFieldOutput(instance.NetworkInterfaces, "networkIp"),
			firstNestedMapFieldOutput(instance.NetworkInterfaces, "accessConfigs", "natIp"),
			args.PlacementLabels,
		)
```

In `NewHcloudPublisher`, replace:

```go
		outputs[publisherName] = publisherOutput(registration, server.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), server.Ipv4Address, args.PlacementLabels)
```

with:

```go
		outputs[publisherName] = publisherOutput(registration, server.ID().ToStringOutput(), firstMapFieldOutput(server.Networks, "ip"), server.Ipv4Address, args.PlacementLabels)
```

In `NewOpenstackPublisher`, replace:

```go
		outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), publicIP, args.PlacementLabels)
```

with:

```go
		outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), firstMapFieldOutput(instance.Networks, "fixedIpV4"), publicIP, args.PlacementLabels)
```

- [ ] **Step 6: Verify focused and full Go tests**

Run:

```bash
gofmt -w internal/provider/components.go internal/provider/provider_test.go
go test ./internal/provider -run TestProviderOutputsExposeAvailableIPAddresses -count=1
npm run go:test
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/provider/components.go internal/provider/provider_test.go
git commit -m "fix: align go provider ip outputs with typescript"
```

---

### Task 3: Make Azure Marketplace Terms Input Explicit

**Files:**
- Modify: `test/azurePublisher.test.ts`
- Modify: `src/azurePublisher.ts`
- Modify: `src/types.ts`
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/components.go`
- Modify: `site/source/admin/component/azure.md`

- [ ] **Step 1: Add a failing TypeScript test for unsupported terms acceptance**

In `test/azurePublisher.test.ts`, add this test after the existing Azure publisher creation test:

```ts
test("AzurePublisher rejects acceptMarketplaceTerms because terms acceptance is external", () => {
  assert.throws(() => new AzurePublisher("publisher-terms", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    resourceGroupName: "rg",
    location: "westeurope",
    subnetId: "/subscriptions/000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/default",
    adminSshPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtest",
    marketplace: {
      publisher: "netskope",
      offer: "private-access-publisher",
      sku: "publisher",
      version: "latest",
    },
    acceptMarketplaceTerms: true,
  }), /acceptMarketplaceTerms is not implemented/);
});
```

- [ ] **Step 2: Run the focused TypeScript test to verify it fails**

Run:

```bash
npm run build
node --test dist/test/azurePublisher.test.js
```

Expected: FAIL because `acceptMarketplaceTerms` is currently ignored.

- [ ] **Step 3: Reject unsupported TypeScript marketplace terms acceptance**

In `src/azurePublisher.ts`, immediately after:

```ts
    const bootstrap = args.bootstrap ?? false;
```

add:

```ts
    if (args.acceptMarketplaceTerms === true) {
      throw new Error("acceptMarketplaceTerms is not implemented; accept Azure marketplace terms outside this component before deploying marketplace images.");
    }
```

- [ ] **Step 4: Document the TypeScript argument as unsupported**

In `src/types.ts`, replace:

```ts
  acceptMarketplaceTerms?: pulumi.Input<boolean>;
```

with:

```ts
  /**
   * Azure marketplace terms acceptance is not implemented by this component.
   * Accept marketplace image terms outside this component before deployment.
   */
  acceptMarketplaceTerms?: pulumi.Input<boolean>;
```

- [ ] **Step 5: Add a failing Go test for unsupported terms acceptance**

In `internal/provider/provider_test.go`, add this test near the other construct validation tests:

```go
func TestAzurePublisherRejectsUnsupportedMarketplaceTermsAcceptance(t *testing.T) {
	err := constructPublisherResourceError(t, "netskope-publisher:index:AzurePublisher", property.NewMap(map[string]property.Value{
		"names":                 property.New([]property.Value{property.New("pub-1")}),
		"registrations":         registrationMap("pub-1"),
		"resourceGroupName":     property.New("rg"),
		"location":              property.New("westeurope"),
		"subnetId":              property.New("/subscriptions/000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/default"),
		"adminSshPublicKey":     property.New("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtest"),
		"acceptMarketplaceTerms": property.New(true),
		"marketplace": property.New(map[string]property.Value{
			"publisher": property.New("netskope"),
			"offer":     property.New("private-access-publisher"),
			"sku":       property.New("publisher"),
			"version":   property.New("latest"),
		}),
	}))
	if err == nil || !strings.Contains(err.Error(), "acceptMarketplaceTerms is not implemented") {
		t.Fatalf("expected unsupported acceptMarketplaceTerms error, got %v", err)
	}
}
```

- [ ] **Step 6: Run the focused Go test to verify it fails**

Run:

```bash
gofmt -w internal/provider/provider_test.go
go test ./internal/provider -run TestAzurePublisherRejectsUnsupportedMarketplaceTermsAcceptance -count=1
```

Expected: FAIL because Go Azure currently ignores `AcceptMarketplaceTerms`.

- [ ] **Step 7: Reject unsupported Go marketplace terms acceptance**

In `internal/provider/components.go`, inside `NewAzurePublisher`, immediately after the image/marketplace/bootstrap guard:

```go
	if args.ImageID == nil && args.Marketplace == nil && !defaultBool(args.Bootstrap, false) {
		return nil, fmt.Errorf("provide imageId, marketplace, or set bootstrap: true")
	}
```

add:

```go
	if defaultBool(args.AcceptMarketplaceTerms, false) {
		return nil, fmt.Errorf("acceptMarketplaceTerms is not implemented; accept Azure marketplace terms outside this component before deploying marketplace images")
	}
```

- [ ] **Step 8: Update Azure component docs**

In `site/source/admin/component/azure.md`, replace:

```md
- `vmSize`, `adminUsername`, `networkSecurityGroupId`,
  `assignPublicIp`, `osDisk`, `acceptMarketplaceTerms`
```

with:

```md
- `vmSize`, `adminUsername`, `networkSecurityGroupId`,
  `assignPublicIp`, `osDisk`
```

Then add this paragraph immediately below the optional input list:

```md
Marketplace terms acceptance is not performed by this component. If you
use a third-party marketplace image, accept the Azure marketplace terms
outside this component before running `pulumi up`. With `bootstrap: true`,
the component uses the Canonical Ubuntu image path and does not require
Netskope marketplace terms.
```

- [ ] **Step 9: Verify focused Azure tests and docs generation**

Run:

```bash
npm run build
node --test dist/test/azurePublisher.test.js
go test ./internal/provider -run TestAzurePublisherRejectsUnsupportedMarketplaceTermsAcceptance -count=1
npm run build --prefix site
```

Expected: PASS.

- [ ] **Step 10: Commit**

```bash
git add src/azurePublisher.ts src/types.ts test/azurePublisher.test.ts internal/provider/components.go internal/provider/provider_test.go site/source/admin/component/azure.md
git commit -m "fix: reject unsupported azure marketplace terms input"
```

---

### Task 4: Full Verification and Generated Artifact Check

**Files:**
- May modify generated artifacts only if `schema.json`, `sdk/`, or generated site content changes.

- [ ] **Step 1: Regenerate provider docs**

Run:

```bash
npm run docs:gen
```

Expected: PASS. If generated docs change, inspect them before committing.

- [ ] **Step 2: Run full validation**

Run:

```bash
npm run typecheck
npm test
npm run go:test
npm run registry:check
npm run catalog:check
npm run build --prefix site
```

Expected: all commands PASS.

- [ ] **Step 3: Regenerate SDKs only if schema changed**

Run:

```bash
git diff --quiet schema.json || npm run sdk:gen
```

Expected: no output if `schema.json` did not change. If SDKs regenerate, inspect `git diff -- sdk schema.json`.

- [ ] **Step 4: Commit generated artifacts if needed**

Run:

```bash
if ! git diff --quiet -- site/source/_generated schema.json sdk; then
  git add site/source/_generated schema.json sdk
  git commit -m "chore: refresh generated provider artifacts"
fi
```

Expected: either creates a generated artifact commit or does nothing.

- [ ] **Step 5: Confirm clean status**

Run:

```bash
git status --short --branch
```

Expected: no uncommitted changes. The branch may still be ahead of `origin/main` if GitHub credentials have not been fixed.

---

## Self-Review

**Spec coverage:** The plan covers all three audit findings: Scaleway token drift, Go/TypeScript IP output parity drift, and unused Azure `acceptMarketplaceTerms`.

**Placeholder scan:** No task uses TBD/TODO/fill-in language. Each code-changing step includes exact snippets and exact commands.

**Type consistency:** The plan consistently uses `scaleway:index/instanceServer:InstanceServer`, `acceptMarketplaceTerms`, `NetworkInterfaces`, `firstMapFieldOutput`, and `firstNestedMapFieldOutput`.
