# Provider Audit Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the confirmed provider audit issues: Go Azure private IP parity, Proxmox VE multi-replica VM ID safety and docs, and OpenTelekomCloud image/flavor selector conflicts.

**Architecture:** Keep the current provider framework intact. Add focused validations at the component boundary, keep TypeScript and Go behavior aligned, then regenerate schema/SDK/docs only after source behavior is fixed and covered by tests.

**Tech Stack:** Pulumi TypeScript components, Pulumi Go executable provider, Node test runner, Go tests, generated Pulumi schema/SDKs, Hexo GitHub Pages site.

---

## File Structure

- Modify: `test/azurePublisher.test.ts`
  - Add a TypeScript assertion that Azure publisher outputs include the NIC private IP.
- Modify: `internal/provider/provider_test.go`
  - Add Go executable-provider assertions for Azure private IP output, Proxmox VE duplicate `vmId` validation, and OpenTelekomCloud selector behavior.
- Modify: `internal/provider/components.go`
  - Read Azure private IP from NIC `ipConfigurations`.
  - Reject Proxmox VE `vmId` when more than one publisher is requested.
  - Avoid sending OpenTelekomCloud default `imageName`/`flavorName` when ID selectors are supplied.
- Modify: `src/proxmoxvePublisher.ts`
  - Reject `vmId` when more than one publisher is requested.
- Modify: `src/opentelekomcloudPublisher.ts`
  - Avoid sending default `imageName`/`flavorName` when `imageId`/`flavorId` are supplied.
- Modify: `src/providerCatalog.ts`
  - Add OpenTelekomCloud mutually exclusive selector metadata for `imageName`/`imageId` and `flavorName`/`flavorId`.
- Modify: `site/source/admin/component/proxmoxve.md`
  - Remove unsupported `vmIdStart` and `snippetsDatastoreId` example inputs.
  - Document `vmId` single-publisher limitation.
- Generated after source changes:
  - `schema.json`
  - `sdk/**`
  - `docs/_index.md`
  - `docs/installation-configuration.md`
  - `site/source/_generated/**`

---

### Task 1: Restore Go Azure Private IP Output Parity

**Files:**
- Modify: `test/azurePublisher.test.ts`
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/components.go`

- [ ] **Step 1: Add a TypeScript private IP assertion**

In `test/azurePublisher.test.ts`, inside `test("AzurePublisher creates outputs keyed by publisher name", async () => { ... })`, after:

```ts
  assert.equal(publishers["pub-1"].vmId, "publisher-pub-1-id");
```

add:

```ts
  assert.equal(publishers["pub-1"].privateIp, "10.1.0.10");
```

- [ ] **Step 2: Run the TypeScript Azure test**

Run:

```bash
npm run build && node --test dist/test/azurePublisher.test.js
```

Expected: PASS. This documents that TypeScript already exposes Azure `privateIp`.

- [ ] **Step 3: Add Go mock state for Azure NIC private IP**

In `internal/provider/provider_test.go`, inside `constructAndCollectPublisherOutput`, locate the `NewResourceF` switch. Add this case before the existing cloud VM cases:

```go
				case "azure-native:network:NetworkInterface":
					state = state.Set("ipConfigurations", property.New([]property.Value{property.New(map[string]property.Value{
						"privateIPAddress": property.New("10.1.0.10"),
					})}))
```

- [ ] **Step 4: Add a failing Go Azure private IP case**

In `internal/provider/provider_test.go`, inside `TestProviderOutputsExposeAvailableIPAddresses`, add this case after the opening `cases := []struct { ... }{` line:

```go
		{
			name:  "Azure private IP",
			token: "netskope-publisher:index:AzurePublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":             property.New([]property.Value{property.New("pub-1")}),
				"registrations":     registrationMap("pub-1"),
				"resourceGroupName": property.New("rg"),
				"location":          property.New("westeurope"),
				"subnetId":          property.New("/subscriptions/000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/default"),
				"adminSshPublicKey": property.New("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtest"),
				"imageId":           property.New("/subscriptions/000/resourceGroups/rg/providers/Microsoft.Compute/images/publisher"),
			}),
			outputField: "privateIp",
			expected:    "10.1.0.10",
		},
```

- [ ] **Step 5: Run the focused Go test and verify it fails**

Run:

```bash
gofmt -w internal/provider/provider_test.go
go test ./internal/provider -run TestProviderOutputsExposeAvailableIPAddresses -count=1
```

Expected: FAIL with `Azure private IP` reporting `got ""`, because `NewAzurePublisher` currently returns an empty private IP.

- [ ] **Step 6: Add a helper to read the first Azure NIC private IP**

In `internal/provider/components.go`, add this helper near the existing output helpers such as `firstMapFieldOutput`:

```go
func firstAzurePrivateIP(configs azurenetwork.NetworkInterfaceIPConfigurationArrayOutput) pulumi.StringOutput {
	return configs.ApplyT(func(items []azurenetwork.NetworkInterfaceIPConfiguration) string {
		if len(items) == 0 {
			return ""
		}
		if items[0].PrivateIPAddress == nil {
			return ""
		}
		return *items[0].PrivateIPAddress
	}).(pulumi.StringOutput)
}
```

- [ ] **Step 7: Use the Azure private IP helper in Go outputs**

In `internal/provider/components.go`, replace this line in `NewAzurePublisher`:

```go
		outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), publicIPOutput, args.PlacementLabels)
```

with:

```go
		outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), firstAzurePrivateIP(nic.IpConfigurations), publicIPOutput, args.PlacementLabels)
```

- [ ] **Step 8: Run focused Azure parity tests**

Run:

```bash
gofmt -w internal/provider/components.go internal/provider/provider_test.go
go test ./internal/provider -run TestProviderOutputsExposeAvailableIPAddresses -count=1
npm run build && node --test dist/test/azurePublisher.test.js
```

Expected: PASS.

- [ ] **Step 9: Commit Task 1**

Run:

```bash
git add test/azurePublisher.test.ts internal/provider/provider_test.go internal/provider/components.go
git commit -m "fix: expose azure private ip in go provider"
```

---

### Task 2: Make Proxmox VE `vmId` Safe for Multiple Publishers

**Files:**
- Modify: `test/proxmoxvePublisher.test.ts`
- Modify: `internal/provider/provider_test.go`
- Modify: `src/proxmoxvePublisher.ts`
- Modify: `internal/provider/components.go`

- [ ] **Step 1: Add a failing TypeScript test for duplicate Proxmox VE `vmId`**

Append this test to `test/proxmoxvePublisher.test.ts`, after `ProxmoxvePublisher rejects missing catalog-required templateVmId`:

```ts
test("ProxmoxvePublisher rejects vmId with multiple publishers", () => {
  assert.throws(
    () => new ProxmoxvePublisher("duplicate-vmid", {
      names: ["pub-1", "pub-2"],
      tenantUrl: "https://tenant.goskope.com",
      apiToken: pulumi.secret("api-token"),
      nodeName: "pve-1",
      datastoreId: "local",
      templateVmId: 9000,
      vmId: 101,
    }),
    /ProxmoxvePublisher vmId can only be used with exactly one publisher/,
  );
});
```

- [ ] **Step 2: Run the TypeScript test and verify it fails**

Run:

```bash
npm run build && node --test dist/test/proxmoxvePublisher.test.js
```

Expected: FAIL with `Missing expected exception`.

- [ ] **Step 3: Reject multi-publisher `vmId` in TypeScript**

In `src/proxmoxvePublisher.ts`, replace the import block:

```ts
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";
```

with:

```ts
import { resolvePublisherNames } from "./componentCore";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";
```

Then, after:

```ts
    validateComponentArgs("ProxmoxvePublisher", args);
```

add:

```ts
    if (args.vmId !== undefined && resolvePublisherNames(args).length !== 1) {
      throw new Error("ProxmoxvePublisher vmId can only be used with exactly one publisher; omit vmId for provider-assigned IDs or deploy one publisher at a time.");
    }
```

- [ ] **Step 4: Run the TypeScript Proxmox VE test**

Run:

```bash
npm run build && node --test dist/test/proxmoxvePublisher.test.js
```

Expected: PASS.

- [ ] **Step 5: Add a failing Go test for duplicate Proxmox VE `vmId`**

In `internal/provider/provider_test.go`, after `TestProxmoxveConstructRejectsMissingTemplateVMID`, add:

```go
func TestProxmoxveConstructRejectsVMIDWithMultiplePublishers(t *testing.T) {
	err := constructPublisherResourceError(t, "netskope-publisher:index:ProxmoxvePublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1"), property.New("pub-2")}),
		"registrations": registrationMap("pub-1", "pub-2"),
		"nodeName":      property.New("pve-1"),
		"datastoreId":   property.New("local"),
		"templateVmId":  property.New(9000.0),
		"vmId":          property.New(101.0),
	}))
	if err == nil || !strings.Contains(err.Error(), "ProxmoxvePublisher vmId can only be used with exactly one publisher") {
		t.Fatalf("expected duplicate vmId guard error, got %v", err)
	}
}
```

- [ ] **Step 6: Run the Go test and verify it fails**

Run:

```bash
gofmt -w internal/provider/provider_test.go
go test ./internal/provider -run TestProxmoxveConstructRejectsVMIDWithMultiplePublishers -count=1
```

Expected: FAIL because the Go constructor currently accepts one `vmId` with multiple publishers.

- [ ] **Step 7: Reject multi-publisher `vmId` in Go**

In `internal/provider/components.go`, inside `NewProxmoxvePublisher`, after:

```go
	if err := validateProviderCatalogArgs("ProxmoxvePublisher", args); err != nil {
		return nil, err
	}
```

add:

```go
	if args.VMID != nil {
		names, err := derivePublisherNames(args.common())
		if err != nil {
			return nil, err
		}
		if len(names) != 1 {
			return nil, fmt.Errorf("ProxmoxvePublisher vmId can only be used with exactly one publisher; omit vmId for provider-assigned IDs or deploy one publisher at a time")
		}
	}
```

- [ ] **Step 8: Run focused Proxmox VE validation tests**

Run:

```bash
gofmt -w internal/provider/components.go internal/provider/provider_test.go
go test ./internal/provider -run 'TestProxmoxveConstructRejectsVMIDWithMultiplePublishers|TestProxmoxveConstructCreatesSnippetBackedVmClone' -count=1
npm run build && node --test dist/test/proxmoxvePublisher.test.js
```

Expected: PASS.

- [ ] **Step 9: Commit Task 2**

Run:

```bash
git add test/proxmoxvePublisher.test.ts src/proxmoxvePublisher.ts internal/provider/provider_test.go internal/provider/components.go
git commit -m "fix: guard proxmox vmid for multiple publishers"
```

---

### Task 3: Fix OpenTelekomCloud Image and Flavor Selector Conflicts

**Files:**
- Modify: `test/additionalCloudPublishers.test.ts`
- Modify: `internal/provider/provider_test.go`
- Modify: `src/providerCatalog.ts`
- Modify: `src/opentelekomcloudPublisher.ts`
- Modify: `internal/provider/catalog.go`
- Modify: `internal/provider/components.go`

- [ ] **Step 1: Add TypeScript tests for OpenTelekomCloud selector behavior**

In `test/additionalCloudPublishers.test.ts`, after `OpentelekomcloudPublisher creates compute instance with plain userData`, add:

```ts
test("OpentelekomcloudPublisher uses imageId and flavorId without default name selectors", async () => {
  const component = new OpentelekomcloudPublisher("otc-id-selectors", baseArgs({
    networks: [{ name: "private" }],
    imageId: "image-id",
    flavorId: "flavor-id",
  }));

  await outputValue(component.publishers);
  const instance = createdResources["opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2"]["otc-id-selectors-pub-1"];

  assert.equal(instance.imageId, "image-id");
  assert.equal(instance.flavorId, "flavor-id");
  assert.equal(instance.imageName, undefined);
  assert.equal(instance.flavorName, undefined);
});

test("OpentelekomcloudPublisher rejects conflicting image and flavor selectors", () => {
  assert.throws(
    () => new OpentelekomcloudPublisher("otc-conflicting-image", baseArgs({
      networks: [{ name: "private" }],
      imageName: "Ubuntu 22.04",
      imageId: "image-id",
    })),
    /OpentelekomcloudPublisher accepts only one of: imageName, imageId/,
  );

  assert.throws(
    () => new OpentelekomcloudPublisher("otc-conflicting-flavor", baseArgs({
      networks: [{ name: "private" }],
      flavorName: "s3.medium.2",
      flavorId: "flavor-id",
    })),
    /OpentelekomcloudPublisher accepts only one of: flavorName, flavorId/,
  );
});
```

- [ ] **Step 2: Run the TypeScript tests and verify they fail**

Run:

```bash
npm run build && node --test dist/test/additionalCloudPublishers.test.js
```

Expected: FAIL because `imageName` and `flavorName` are still defaulted when ID selectors are provided, and conflicts are not rejected.

- [ ] **Step 3: Add OpenTelekomCloud selector metadata in TypeScript catalog**

In `src/providerCatalog.ts`, replace the `OpentelekomcloudPublisher` provider entry with this version:

```ts
  provider({ displayName: "OpenTelekomCloud", componentName: "OpentelekomcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "opentelekomcloud", required: ["networks"], resourceToken: "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", providerPackage: "terraform-provider:opentelekomcloud/opentelekomcloud", upstreamPropertyChecks: [
    ...upstreamChecks("opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", [
      [["networks"], "network attachments"],
      [["imageName"], "Ubuntu image selection"],
      [["flavorName"], "flavor selection"],
    ]),
    ...upstreamOutputChecks("opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", [
      [["accessIpV4"], "public IPv4 output"],
    ]),
  ], validation: { mutuallyExclusive: [["imageName", "imageId"], ["flavorName", "flavorId"]] } }),
```

- [ ] **Step 4: Stop defaulting OpenTelekomCloud names when IDs are supplied in TypeScript**

In `src/opentelekomcloudPublisher.ts`, replace these map inputs:

```ts
        imageName: currentArgs.imageName ?? "Ubuntu 22.04",
        imageId: currentArgs.imageId,
        flavorName: currentArgs.flavorName ?? "s3.medium.2",
        flavorId: currentArgs.flavorId,
```

with:

```ts
        imageName: currentArgs.imageId === undefined ? currentArgs.imageName ?? "Ubuntu 22.04" : currentArgs.imageName,
        imageId: currentArgs.imageId,
        flavorName: currentArgs.flavorId === undefined ? currentArgs.flavorName ?? "s3.medium.2" : currentArgs.flavorName,
        flavorId: currentArgs.flavorId,
```

- [ ] **Step 5: Run the TypeScript OpenTelekomCloud tests**

Run:

```bash
npm run build && node --test dist/test/additionalCloudPublishers.test.js dist/test/providerValidation.test.js
```

Expected: PASS.

- [ ] **Step 6: Add Go catalog validation metadata**

In `internal/provider/catalog.go`, replace this line:

```go
	"OpentelekomcloudPublisher": providerEntry("OpenTelekomCloud", "OpentelekomcloudPublisher", "catalogRawVm", "plain", "networks"),
```

with:

```go
	"OpentelekomcloudPublisher": {
		DisplayName:       "OpenTelekomCloud",
		ComponentName:     "OpentelekomcloudPublisher",
		Token:             "netskope-publisher:index:OpentelekomcloudPublisher",
		Implementation:    "catalogRawVm",
		UserDataMode:      "plain",
		RequiredInputs:    []string{"networks"},
		MutuallyExclusive: [][]string{{"imageName", "imageId"}, {"flavorName", "flavorId"}},
	},
```

- [ ] **Step 7: Add Go tests for OpenTelekomCloud selector behavior**

In `internal/provider/provider_test.go`, after `TestOciConstructRejectsMissingImageID`, add:

```go
func TestOpentelekomcloudConstructRejectsConflictingSelectors(t *testing.T) {
	imageErr := constructPublisherResourceError(t, "netskope-publisher:index:OpentelekomcloudPublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"networks":      property.New([]property.Value{property.New(map[string]property.Value{"name": property.New("private")})}),
		"imageName":     property.New("Ubuntu 22.04"),
		"imageId":       property.New("image-id"),
	}))
	if imageErr == nil || !strings.Contains(imageErr.Error(), "OpentelekomcloudPublisher accepts only one of: imageName, imageId") {
		t.Fatalf("expected conflicting image selector error, got %v", imageErr)
	}

	flavorErr := constructPublisherResourceError(t, "netskope-publisher:index:OpentelekomcloudPublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"networks":      property.New([]property.Value{property.New(map[string]property.Value{"name": property.New("private")})}),
		"flavorName":    property.New("s3.medium.2"),
		"flavorId":      property.New("flavor-id"),
	}))
	if flavorErr == nil || !strings.Contains(flavorErr.Error(), "OpentelekomcloudPublisher accepts only one of: flavorName, flavorId") {
		t.Fatalf("expected conflicting flavor selector error, got %v", flavorErr)
	}
}

func TestOpentelekomcloudConstructUsesIDSelectorsWithoutDefaultNames(t *testing.T) {
	resources := constructAndCollectResources(t, "netskope-publisher:index:OpentelekomcloudPublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"networks":      property.New([]property.Value{property.New(map[string]property.Value{"name": property.New("private")})}),
		"imageId":       property.New("image-id"),
		"flavorId":      property.New("flavor-id"),
	}))
	instance := findResourceByType(t, resources, "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2")
	if got := instance.Inputs.Get("imageId").AsString(); got != "image-id" {
		t.Fatalf("expected imageId, got %q", got)
	}
	if got := instance.Inputs.Get("flavorId").AsString(); got != "flavor-id" {
		t.Fatalf("expected flavorId, got %q", got)
	}
	if instance.Inputs.Get("imageName").IsString() && instance.Inputs.Get("imageName").AsString() != "" {
		t.Fatalf("did not expect default imageName when imageId is supplied")
	}
	if instance.Inputs.Get("flavorName").IsString() && instance.Inputs.Get("flavorName").AsString() != "" {
		t.Fatalf("did not expect default flavorName when flavorId is supplied")
	}
}
```

- [ ] **Step 8: Run the Go tests and verify ID-selector test fails before implementation**

Run:

```bash
gofmt -w internal/provider/catalog.go internal/provider/provider_test.go
go test ./internal/provider -run 'TestOpentelekomcloudConstructRejectsConflictingSelectors|TestOpentelekomcloudConstructUsesIDSelectorsWithoutDefaultNames' -count=1
```

Expected after Step 6: conflict test PASS, ID-selector test FAIL because Go still defaults `imageName` and `flavorName`.

- [ ] **Step 9: Stop defaulting OpenTelekomCloud names when IDs are supplied in Go**

In `internal/provider/components.go`, inside `NewOpentelekomcloudPublisher`, before `ctx.RegisterResource("opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", ...)`, add:

```go
		imageName := args.ImageName
		if args.ImageID == nil && imageName == nil {
			defaultImageName := "Ubuntu 22.04"
			imageName = &defaultImageName
		}
		flavorName := args.FlavorName
		if args.FlavorID == nil && flavorName == nil {
			defaultFlavorName := "s3.medium.2"
			flavorName = &defaultFlavorName
		}
```

Then replace these inputs:

```go
			"imageName":        pulumi.String(defaultString(args.ImageName, "Ubuntu 22.04")),
			"imageId":          stringPtrInput(args.ImageID),
			"flavorName":       pulumi.String(defaultString(args.FlavorName, "s3.medium.2")),
			"flavorId":         stringPtrInput(args.FlavorID),
```

with:

```go
			"imageName":        stringPtrInput(imageName),
			"imageId":          stringPtrInput(args.ImageID),
			"flavorName":       stringPtrInput(flavorName),
			"flavorId":         stringPtrInput(args.FlavorID),
```

- [ ] **Step 10: Run focused OpenTelekomCloud tests**

Run:

```bash
gofmt -w internal/provider/catalog.go internal/provider/components.go internal/provider/provider_test.go
go test ./internal/provider -run 'TestOpentelekomcloudConstructRejectsConflictingSelectors|TestOpentelekomcloudConstructUsesIDSelectorsWithoutDefaultNames|TestExpandedProviderConstructsBootstrapWithRegistryFields' -count=1
npm run build && node --test dist/test/additionalCloudPublishers.test.js dist/test/providerValidation.test.js
```

Expected: PASS.

- [ ] **Step 11: Commit Task 3**

Run:

```bash
git add test/additionalCloudPublishers.test.ts src/providerCatalog.ts src/opentelekomcloudPublisher.ts internal/provider/catalog.go internal/provider/components.go internal/provider/provider_test.go
git commit -m "fix: validate opentelekomcloud selector inputs"
```

---

### Task 4: Correct Proxmox VE GitHub Pages Documentation

**Files:**
- Modify: `site/source/admin/component/proxmoxve.md`

- [ ] **Step 1: Update the Proxmox VE input docs**

In `site/source/admin/component/proxmoxve.md`, replace:

```md
Optional platform inputs: `cloneNodeName`, `vmId`, `poolId`,
`cpuCores`, `memory`, `diskSize`, `networkBridge`, `networkModel`,
`vlanId`, `started`, `onBoot`, `fullClone`, `ipAddress`, `gateway`, and
`nameservers`.
```

with:

```md
Optional platform inputs: `cloneNodeName`, `vmId`, `poolId`,
`cpuCores`, `memory`, `diskSize`, `networkBridge`, `networkModel`,
`vlanId`, `started`, `onBoot`, `fullClone`, `ipAddress`, `gateway`, and
`nameservers`.

`vmId` is only valid when exactly one publisher is created. For HA pairs
or larger replica sets, omit `vmId` and let Proxmox VE assign VM IDs, or
deploy each VM as its own component with a distinct `vmId`.
```

- [ ] **Step 2: Replace the invalid YAML example inputs**

In `site/source/admin/component/proxmoxve.md`, replace the YAML properties block:

```yaml
      nodeName: pve-1
      vmIdStart: 4200
      templateVmId: 9000
      datastoreId: local-lvm
      snippetsDatastoreId: local
      networkBridge: vmbr0
      bootstrap: true
```

with:

```yaml
      nodeName: pve-1
      templateVmId: 9000
      datastoreId: local
      networkBridge: vmbr0
      bootstrap: true
```

- [ ] **Step 3: Verify no unsupported Proxmox VE docs inputs remain**

Run:

```bash
rg -n "vmIdStart|snippetsDatastoreId" site/source/admin/component/proxmoxve.md README.md docs site/source
```

Expected: no output.

- [ ] **Step 4: Build the GitHub Pages site**

Run:

```bash
npm run build --prefix site
```

Expected: PASS.

- [ ] **Step 5: Commit Task 4**

Run:

```bash
git add site/source/admin/component/proxmoxve.md
git commit -m "docs: correct proxmox ve vm id examples"
```

---

### Task 5: Regenerate Schema, SDKs, and Verify Everything

**Files:**
- Generated:
  - `schema.json`
  - `sdk/**`
  - `docs/_index.md`
  - `docs/installation-configuration.md`
  - `site/source/_generated/**`

- [ ] **Step 1: Regenerate provider docs and schema-sensitive artifacts**

Run:

```bash
npm run docs:gen
npm run sdk:gen
```

Expected: both commands complete successfully.

- [ ] **Step 2: Run TypeScript verification**

Run:

```bash
npm run typecheck
npm test
```

Expected: PASS.

- [ ] **Step 3: Run Go verification**

Run:

```bash
npm run go:test
```

Expected: PASS.

- [ ] **Step 4: Run registry and catalog verification**

Run:

```bash
npm run registry:check
npm run catalog:check
```

Expected: PASS.

- [ ] **Step 5: Build GitHub Pages docs**

Run:

```bash
npm run build --prefix site
```

Expected: PASS.

- [ ] **Step 6: Inspect generated diff**

Run:

```bash
git status --short
git diff --stat
```

Expected: only source changes from Tasks 1-4 plus generated schema/SDK/docs changes caused by those source changes.

- [ ] **Step 7: Commit generated artifacts if needed**

If `git status --short` shows generated file changes, run:

```bash
git add schema.json sdk docs/_index.md docs/installation-configuration.md site/source/_generated
git commit -m "chore: refresh generated provider artifacts"
```

If `git status --short` shows no generated file changes, do not create an empty commit.

- [ ] **Step 8: Final branch status**

Run:

```bash
git status --short --branch
git log --oneline -5
```

Expected: clean worktree. Latest commits should include the four fix commits and optional generated-artifacts commit.

---

## Self-Review

**Spec coverage:** The plan covers all four confirmed audit findings:
- Go Azure private IP parity: Task 1.
- Proxmox VE duplicate `vmId` with multiple publishers: Task 2.
- Proxmox VE docs using unsupported inputs: Task 4.
- OpenTelekomCloud selector conflict/default behavior: Task 3.
- Generated schema/SDK/docs and full verification: Task 5.

**Placeholder scan:** No task uses TBD, TODO, "similar to", or open-ended "add tests" instructions. Each code-changing step includes concrete snippets and commands.

**Type consistency:** The plan consistently uses existing names: `AzurePublisher`, `ProxmoxvePublisher`, `OpentelekomcloudPublisher`, `vmId`, `imageName`, `imageId`, `flavorName`, `flavorId`, `privateIp`, `resolvePublisherNames`, `validateProviderCatalogArgs`, and existing test helpers.
