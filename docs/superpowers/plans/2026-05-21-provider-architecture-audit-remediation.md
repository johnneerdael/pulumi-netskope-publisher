# Provider Architecture Audit Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the provider architecture gaps found by the audit so catalog metadata, generated schema, registry checks, runtime outputs, and user-data adapters stay aligned.

**Architecture:** Make the catalog the enforceable contract for source and generated schema, strengthen registry validation from "resource exists" to "resource inputs we use exist", and add tests around output fields the docs promise. Keep the implementation incremental: first lock required-input parity, then registry property checks, then output mapping coverage, then remove the split user-data mode abstraction.

**Tech Stack:** TypeScript, Node test runner, Pulumi dynamic/component resources, Go provider with `pulumi-go-provider`, generated Pulumi schema, Hexo docs.

---

## File Structure

- Modify `src/providerCatalog.ts`: fix `NetskopeRegistration` required metadata and add upstream property placement metadata for catalog-driven providers.
- Modify `src/providerRegistrySchema.ts`: validate every declared upstream property path, including nested `$ref` paths, not only the user-data property.
- Modify `scripts/check-provider-catalog.mjs`: compare catalog `validation.required` against generated `schema.json` `requiredInputs` for every catalog component.
- Modify `test/providerCatalog.test.ts`: assert `NetskopeRegistration` required metadata matches implementation and schema expectations.
- Modify `test/providerValidation.test.ts`: prove `NetskopeRegistration` validation rejects missing `tenantUrl`.
- Modify `test/providerRegistrySchema.test.ts`: add positive/negative tests for declared upstream property paths.
- Modify `internal/provider/catalog.go`: add `NetskopeRegistration` to the Go catalog so parity tests include it.
- Modify `internal/provider/catalog_test.go`: assert the Go catalog includes `NetskopeRegistration` and its required fields.
- Modify `internal/provider/provider_test.go`: add output contract tests for Go providers that currently return empty IP outputs despite available resource fields.
- Modify `internal/provider/components.go`: map available IP output fields for Go providers.
- Modify `src/userDataAdapters.ts`: align adapter mode names with `src/providerCatalog.ts` and expose one adapter map used by factories.
- Modify `src/catalogVmFactory.ts`: use the shared user-data adapter map instead of local mode branching.
- Modify `test/userDataAdapters.test.ts`: assert adapter modes cover every catalog user-data mode that can be rendered through the raw VM factory.

---

### Task 1: Enforce Catalog Required-Input Parity

**Files:**
- Modify: `src/providerCatalog.ts`
- Modify: `scripts/check-provider-catalog.mjs`
- Modify: `test/providerCatalog.test.ts`
- Modify: `test/providerValidation.test.ts`
- Modify: `internal/provider/catalog.go`
- Modify: `internal/provider/catalog_test.go`

- [ ] **Step 1: Add failing TypeScript catalog tests for `NetskopeRegistration`**

Append this to `test/providerCatalog.test.ts`:

```ts
test("NetskopeRegistration catalog required inputs match resource args", () => {
  assert.deepEqual(providerCatalog.NetskopeRegistration.validation.required, ["publisherNames", "tenantUrl"]);
});
```

Append this to `test/providerValidation.test.ts`:

```ts
test("validateProviderArgs rejects NetskopeRegistration without tenantUrl", () => {
  assert.throws(
    () => validateProviderArgs("NetskopeRegistration", { publisherNames: ["pub-1"] }),
    /NetskopeRegistration requires input tenantUrl/,
  );
});
```

- [ ] **Step 2: Run TypeScript tests to verify they fail**

Run:

```bash
npm run build
node --test dist/test/providerCatalog.test.js dist/test/providerValidation.test.js
```

Expected: FAIL with the new `NetskopeRegistration` catalog required-input assertion.

- [ ] **Step 3: Fix `NetskopeRegistration` catalog metadata**

In `src/providerCatalog.ts`, replace the `NetskopeRegistration` provider definition:

```ts
provider({ displayName: "Netskope Registration", componentName: "NetskopeRegistration", implementation: "bespoke", bootstrapModel: "registrationOnly", userDataMode: "none", slug: "registration", required: ["publisherNames"] }),
```

with:

```ts
provider({ displayName: "Netskope Registration", componentName: "NetskopeRegistration", implementation: "bespoke", bootstrapModel: "registrationOnly", userDataMode: "none", slug: "registration", required: ["publisherNames", "tenantUrl"] }),
```

- [ ] **Step 4: Add generated schema required-input parity to catalog checker**

In `scripts/check-provider-catalog.mjs`, inside the `for (const provider of catalogProviders)` loop, immediately after the existing schema resource check:

```js
  const schemaResource = schema.resources?.[provider.token];
  if (!schemaResource) {
    errors.push(`schema.json missing catalog token ${provider.token}`);
  }
```

replace the current token check block:

```js
  if (!schema.resources?.[provider.token]) {
    errors.push(`schema.json missing catalog token ${provider.token}`);
  }
```

with:

```js
  const schemaResource = schema.resources?.[provider.token];
  if (!schemaResource) {
    errors.push(`schema.json missing catalog token ${provider.token}`);
  } else {
    const catalogRequired = [...(provider.validation.required ?? [])].sort();
    const schemaRequired = [...(schemaResource.requiredInputs ?? [])].sort();
    if (JSON.stringify(catalogRequired) !== JSON.stringify(schemaRequired)) {
      errors.push(`${provider.componentName} catalog required inputs ${catalogRequired.join(",")} do not match schema requiredInputs ${schemaRequired.join(",")}`);
    }
  }
```

- [ ] **Step 5: Add failing Go catalog test for `NetskopeRegistration`**

In `internal/provider/catalog_test.go`, add `NetskopeRegistration` to the `required := []string{...}` list in `TestProviderCatalogIncludesCurrentComponents`:

```go
"NetskopeRegistration",
```

Append this test:

```go
func TestProviderCatalogIncludesRegistrationMetadata(t *testing.T) {
	registration := providerCatalog["NetskopeRegistration"]
	if registration.Implementation != "resource" {
		t.Fatalf("NetskopeRegistration implementation mismatch: %s", registration.Implementation)
	}
	if !containsString(registration.RequiredInputs, "publisherNames") {
		t.Fatalf("NetskopeRegistration missing publisherNames validation metadata: %#v", registration.RequiredInputs)
	}
	if !containsString(registration.RequiredInputs, "tenantUrl") {
		t.Fatalf("NetskopeRegistration missing tenantUrl validation metadata: %#v", registration.RequiredInputs)
	}
}
```

- [ ] **Step 6: Run Go catalog test to verify it fails**

Run:

```bash
go test ./internal/provider -run 'TestProviderCatalogIncludesRegistrationMetadata|TestProviderCatalogIncludesCurrentComponents' -count=1
```

Expected: FAIL because the Go catalog does not include `NetskopeRegistration`.

- [ ] **Step 7: Add `NetskopeRegistration` to Go catalog**

In `internal/provider/catalog.go`, add this entry to `providerCatalog`:

```go
"NetskopeRegistration": providerEntry("Netskope Registration", "NetskopeRegistration", "resource", "none", "publisherNames", "tenantUrl"),
```

- [ ] **Step 8: Run focused parity checks**

Run:

```bash
npm run build
node --test dist/test/providerCatalog.test.js dist/test/providerValidation.test.js
go test ./internal/provider -run 'TestProviderCatalogIncludesRegistrationMetadata|TestProviderCatalogIncludesCurrentComponents' -count=1
npm run catalog:check
```

Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add src/providerCatalog.ts scripts/check-provider-catalog.mjs test/providerCatalog.test.ts test/providerValidation.test.ts internal/provider/catalog.go internal/provider/catalog_test.go
git commit -m "fix: enforce provider catalog required input parity"
```

---

### Task 2: Validate All Declared Upstream Provider Input Paths

**Files:**
- Modify: `src/providerCatalog.ts`
- Modify: `src/providerRegistrySchema.ts`
- Modify: `test/providerRegistrySchema.test.ts`

- [ ] **Step 1: Add registry property-check types and failing tests**

In `test/providerRegistrySchema.test.ts`, append:

```ts
test("validateProviderAgainstRegistrySchema accepts declared upstream property paths", () => {
  const errors = validateProviderAgainstRegistrySchema({
    componentName: "NestedPublisher",
    resourceToken: "example:index/server:Server",
    providerPackage: "@example/provider",
    upstreamPropertyChecks: [{
      resourceToken: "example:index/server:Server",
      propertyPath: ["network", "subnetId"],
      description: "server subnet placement",
    }],
    userData: {
      mode: "plain",
      property: "userData",
    },
  }, {
    name: "example",
    language: { nodejs: { packageName: "@example/provider" } },
    resources: {
      "example:index/server:Server": {
        inputProperties: {
          userData: { type: "string" },
          network: { "$ref": "#/types/example:index/ServerNetwork:ServerNetwork" },
        },
      },
    },
    types: {
      "example:index/ServerNetwork:ServerNetwork": {
        properties: {
          subnetId: { type: "string" },
        },
      },
    },
  });

  assert.deepEqual(errors, []);
});

test("validateProviderAgainstRegistrySchema rejects missing declared upstream property paths", () => {
  const errors = validateProviderAgainstRegistrySchema({
    componentName: "NestedPublisher",
    resourceToken: "example:index/server:Server",
    upstreamPropertyChecks: [{
      resourceToken: "example:index/server:Server",
      propertyPath: ["network", "subnetId"],
      description: "server subnet placement",
    }],
    userData: {
      mode: "plain",
      property: "userData",
    },
  }, {
    name: "example",
    resources: {
      "example:index/server:Server": {
        inputProperties: {
          userData: { type: "string" },
          network: { "$ref": "#/types/example:index/ServerNetwork:ServerNetwork" },
        },
      },
    },
    types: {
      "example:index/ServerNetwork:ServerNetwork": {
        properties: {
          networkId: { type: "string" },
        },
      },
    },
  });

  assert.match(errors.join("\n"), /NestedPublisher upstream resource example:index\/server:Server missing server subnet placement path network\.subnetId/);
});
```

- [ ] **Step 2: Run registry tests to verify they fail**

Run:

```bash
npm run build
node --test dist/test/providerRegistrySchema.test.js
```

Expected: FAIL because `upstreamPropertyChecks` is not implemented.

- [ ] **Step 3: Extend registry schema check interfaces**

In `src/providerCatalog.ts`, rename `ProviderRegistrySchemaCheck` to a generic path check and keep a compatibility alias:

```ts
export interface ProviderRegistrySchemaCheck {
  resourceToken: string;
  propertyPath: string[];
  description: string;
}

export type ProviderUpstreamPropertyCheck = ProviderRegistrySchemaCheck;
```

Add this optional property to `ProviderCatalogEntry`:

```ts
upstreamPropertyChecks?: ProviderUpstreamPropertyCheck[];
```

Add it to `ProviderDefinition`:

```ts
upstreamPropertyChecks?: ProviderUpstreamPropertyCheck[];
```

Add it to the object returned by `provider(definition)`:

```ts
upstreamPropertyChecks: definition.upstreamPropertyChecks,
```

- [ ] **Step 4: Implement upstream property path validation**

In `src/providerRegistrySchema.ts`, add `upstreamPropertyChecks` to `RegistryProviderEntry`:

```ts
  upstreamPropertyChecks?: Array<{
    resourceToken: string;
    propertyPath: string[];
    description: string;
  }>;
```

After the existing `registrySchemaChecks` block and before `if (!provider.resourceToken)`, insert:

```ts
  if (provider.upstreamPropertyChecks && provider.upstreamPropertyChecks.length > 0) {
    for (const check of provider.upstreamPropertyChecks) {
      const checkedResource = schema.resources?.[check.resourceToken];
      if (!checkedResource) {
        errors.push(`${provider.componentName} upstream schema missing resource token ${check.resourceToken}`);
        continue;
      }
      if (!schemaHasPath(schema, checkedResource.inputProperties ?? {}, check.propertyPath)) {
        errors.push(`${provider.componentName} upstream resource ${check.resourceToken} missing ${check.description} path ${check.propertyPath.join(".")}`);
      }
    }
  }
```

Do not return early from this new block; keep the existing resource token and user-data checks running afterwards.

- [ ] **Step 5: Declare property checks for provider-specific high-risk paths**

In `src/providerCatalog.ts`, add `upstreamPropertyChecks` to these provider definitions:

For `OciPublisher`:

```ts
upstreamPropertyChecks: [{
  resourceToken: "oci:Core/instance:Instance",
  propertyPath: ["createVnicDetails", "subnetId"],
  description: "primary VNIC subnet",
}, {
  resourceToken: "oci:Core/instance:Instance",
  propertyPath: ["sourceDetails", "sourceId"],
  description: "image source ID",
}, {
  resourceToken: "oci:Core/instance:Instance",
  propertyPath: ["metadata"],
  description: "cloud-init metadata map",
}],
```

For `OpenstackPublisher`:

```ts
upstreamPropertyChecks: [{
  resourceToken: "openstack:compute/instance:Instance",
  propertyPath: ["networks"],
  description: "instance network attachments",
}],
```

For `YandexPublisher`:

```ts
upstreamPropertyChecks: [{
  resourceToken: "yandex:index/computeInstance:ComputeInstance",
  propertyPath: ["bootDisk", "initializeParams", "imageId"],
  description: "boot disk image",
}, {
  resourceToken: "yandex:index/computeInstance:ComputeInstance",
  propertyPath: ["networkInterfaces", "subnetId"],
  description: "network interface subnet",
}, {
  resourceToken: "yandex:index/computeInstance:ComputeInstance",
  propertyPath: ["metadata"],
  description: "cloud-init metadata map",
}],
```

These checks intentionally include nested object and array-item paths so `schemaHasPath` must support both `$ref` and `items.$ref`.

- [ ] **Step 6: Teach `schemaHasPath` to follow array item refs**

In `src/providerRegistrySchema.ts`, update the local `property` type inside `schemaHasPath`:

```ts
const property = currentProperties?.[segment] as { $ref?: string; properties?: Record<string, unknown>; items?: { $ref?: string; properties?: Record<string, unknown> } } | undefined;
```

Replace:

```ts
currentProperties = property.properties ?? resolveRefProperties(schema, property.$ref);
```

with:

```ts
currentProperties = property.properties
  ?? resolveRefProperties(schema, property.$ref)
  ?? property.items?.properties
  ?? resolveRefProperties(schema, property.items?.$ref);
```

- [ ] **Step 7: Run registry validation**

Run:

```bash
npm run build
node --test dist/test/providerRegistrySchema.test.js
npm run registry:check
```

Expected: PASS. Any failure names a concrete upstream path that must be inspected before changing the declared check.

- [ ] **Step 8: Commit**

```bash
git add src/providerCatalog.ts src/providerRegistrySchema.ts test/providerRegistrySchema.test.ts
git commit -m "fix: validate upstream provider input paths"
```

---

### Task 3: Tighten Go Publisher Output Contracts

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/components.go`

- [ ] **Step 1: Add failing Go output tests for IP fields**

Append this helper near the existing `capturedResource` helpers in `internal/provider/provider_test.go`:

```go
func constructAndCollectPublisherOutput(t *testing.T, token string, inputs property.Map) property.Map {
	t.Helper()

	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.3.1"),
		integration.WithProvider(provider),
		integration.WithMocks(&integration.MockResourceMonitor{
			NewResourceF: func(args integration.MockResourceArgs) (string, property.Map, error) {
				state := args.Inputs
				switch string(args.TypeToken) {
				case "hcloud:index/server:Server":
					state = state.Set("ipv4Address", property.New("203.0.113.10"))
				case "nutanix:index/virtualMachine:VirtualMachine":
					state = state.Set("privateIp", property.New("10.0.0.20"))
				case "ovh:CloudProject/instance:Instance":
					state = state.Set("ipAddresses", property.New([]property.Value{property.New("198.51.100.30")}))
				case "scaleway:instance/server:Server":
					state = state.Set("publicIp", property.New("198.51.100.40"))
				}
				return args.Name + "-id", state, nil
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

	publishers := response.State.Get("publishers").AsMap()
	return publishers.Get("pub-1").AsMap()
}
```

Append this test:

```go
func TestProviderOutputsExposeAvailableIPAddresses(t *testing.T) {
	cases := []struct {
		name        string
		token       string
		inputs      property.Map
		outputField string
		expected    string
	}{
		{
			name:  "Hcloud",
			token: "netskope-publisher:index:HcloudPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
			}),
			outputField: "publicIp",
			expected:    "203.0.113.10",
		},
		{
			name:  "Nutanix",
			token: "netskope-publisher:index:NutanixPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"clusterUuid":   property.New("cluster-uuid"),
			}),
			outputField: "privateIp",
			expected:    "10.0.0.20",
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
			outputField: "publicIp",
			expected:    "198.51.100.30",
		},
		{
			name:  "Scaleway",
			token: "netskope-publisher:index:ScalewayPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
			}),
			outputField: "publicIp",
			expected:    "198.51.100.40",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := constructAndCollectPublisherOutput(t, tc.token, tc.inputs)
			if output.Get(tc.outputField).AsString() != tc.expected {
				t.Fatalf("expected %s %s %q, got %#v", tc.name, tc.outputField, tc.expected, output.Get(tc.outputField))
			}
		})
	}
}
```

- [ ] **Step 2: Run Go output test to verify it fails**

Run:

```bash
go test ./internal/provider -run TestProviderOutputsExposeAvailableIPAddresses -count=1
```

Expected: FAIL for providers whose outputs are still hard-coded to empty strings.

- [ ] **Step 3: Add missing raw VM output fields**

In `internal/provider/components.go`, add fields to `rawVMResource`:

```go
IPAddresses pulumi.StringArrayOutput `pulumi:"ipAddresses"`
```

After the change, the struct is:

```go
type rawVMResource struct {
	pulumi.CustomResourceState

	AccessIPV4       pulumi.StringOutput      `pulumi:"accessIpV4"`
	Address          pulumi.StringOutput      `pulumi:"address"`
	IPAddresses      pulumi.StringArrayOutput `pulumi:"ipAddresses"`
	Ipv4Address      pulumi.StringOutput      `pulumi:"ipv4Address"`
	Ipv4Addresses    pulumi.ArrayOutput       `pulumi:"ipv4Addresses"`
	Networks         pulumi.ArrayOutput       `pulumi:"networks"`
	PrimaryIPAddress pulumi.StringOutput      `pulumi:"primaryIpAddress"`
	PrivateIP        pulumi.StringOutput      `pulumi:"privateIp"`
	PublicIP         pulumi.StringOutput      `pulumi:"publicIp"`
}
```

- [ ] **Step 4: Map available IP outputs**

In `NewNutanixPublisher`, replace:

```go
outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput(), args.PlacementLabels)
```

with:

```go
outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), vm.PrivateIP, pulumi.String("").ToStringOutput(), args.PlacementLabels)
```

In `NewScalewayPublisher`, replace:

```go
outputs[publisherName] = publisherOutput(registration, server.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput(), args.PlacementLabels)
```

with:

```go
outputs[publisherName] = publisherOutput(registration, server.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), server.PublicIP, args.PlacementLabels)
```

Keep `HcloudPublisher` as-is if its test passes: it already maps `server.Ipv4Address` to `publicIp`.

- [ ] **Step 5: Add a helper for OVH first IP address output**

Add this helper near `firstNestedString` in `internal/provider/components.go`:

```go
func firstStringOutput(values pulumi.StringArrayOutput) pulumi.StringOutput {
	return values.ApplyT(func(items []string) string {
		if len(items) == 0 {
			return ""
		}
		return items[0]
	}).(pulumi.StringOutput)
}
```

Then in `NewOvhPublisher`, replace:

```go
outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput(), args.PlacementLabels)
```

with:

```go
outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), firstStringOutput(instance.IPAddresses), args.PlacementLabels)
```

- [ ] **Step 6: Run focused Go tests**

Run:

```bash
gofmt -w internal/provider/components.go internal/provider/provider_test.go
go test ./internal/provider -run 'TestProviderOutputsExposeAvailableIPAddresses|TestAdditionalProviderConstructsBootstrapWithRegistryFields' -count=1
```

Expected: PASS.

- [ ] **Step 7: Run full Go provider tests**

Run:

```bash
npm run go:test
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/provider/components.go internal/provider/provider_test.go
git commit -m "fix: expose provider IP outputs when available"
```

---

### Task 4: Consolidate User-Data Adapter Modes

**Files:**
- Modify: `src/userDataAdapters.ts`
- Modify: `src/catalogVmFactory.ts`
- Modify: `test/userDataAdapters.test.ts`

- [ ] **Step 1: Add failing adapter coverage test**

Update the existing `../src/userDataAdapters` import in `test/userDataAdapters.test.ts` to include `userDataAdapters`:

```ts
import {
  base64UserData,
  guestInfoUserData,
  metadataUserData,
  plainUserData,
  scalewayUserData,
  userDataAdapters,
} from "../src/userDataAdapters";
```

Add the `providerCatalog` import near the other top-level imports:

```ts
import { providerCatalog } from "../src/providerCatalog";
```

Append this test body to `test/userDataAdapters.test.ts`:

```ts
test("userDataAdapters cover raw VM factory provider modes", () => {
  for (const componentName of [
    "DigitaloceanPublisher",
    "VultrPublisher",
    "ExoscalePublisher",
    "UpcloudPublisher",
    "StackitPublisher",
    "EquinixPublisher",
    "OutscalePublisher",
    "OpentelekomcloudPublisher",
    "TencentcloudPublisher",
    "YandexPublisher",
  ]) {
    const mode = providerCatalog[componentName].userData.mode;
    assert.ok(userDataAdapters[mode], `${componentName} mode ${mode} has no adapter`);
  }
});
```

- [ ] **Step 2: Run adapter test to verify it fails**

Run:

```bash
npm run build
node --test dist/test/userDataAdapters.test.js
```

Expected: FAIL because `userDataAdapters` does not exist and mode names differ.

- [ ] **Step 3: Align `userDataAdapters.ts` with catalog mode names**

Replace the local `UserDataMode` type in `src/userDataAdapters.ts` with:

```ts
import type { UserDataMode } from "./providerCatalog";

export type { UserDataMode } from "./providerCatalog";
```

Keep the `pulumi` import.

Add this exported adapter map after the existing adapter functions:

```ts
export const userDataAdapters: Partial<Record<UserDataMode, (payload: pulumi.Output<string>, key?: string) => pulumi.Input<string> | Record<string, pulumi.Input<unknown>>>> = {
  plain: (payload) => plainUserData(payload),
  base64: (payload) => base64UserData(payload),
  metadata: (payload, key = "user-data") => metadataUserData(payload, key),
  raw: (payload) => plainUserData(payload),
  customData: (payload) => customData(payload),
  guestInfo: (payload) => guestInfoUserData(payload),
  scalewayDual: (payload) => scalewayUserData(payload),
  ociMetadata: (payload, key = "userData") => base64MetadataUserData(payload, key),
};
```

- [ ] **Step 4: Update raw VM factory to use adapter map**

In `src/catalogVmFactory.ts`, replace imports:

```ts
import { base64UserData, metadataUserData, plainUserData } from "./userDataAdapters";
```

with:

```ts
import { userDataAdapters } from "./userDataAdapters";
```

Replace the full `userDataProperty` function with:

```ts
export function userDataProperty(provider: ProviderCatalogEntry, input: VmPublisherBuildInput): Record<string, pulumi.Input<unknown>> {
  if (provider.userData.mode === "scalewayDual" || provider.userData.mode === "guestInfo" || provider.userData.mode === "ociMetadata" || provider.userData.mode === "customData") {
    throw new Error(`${provider.componentName} cannot use catalog raw VM factory with user-data mode ${provider.userData.mode}`);
  }

  const adapter = userDataAdapters[provider.userData.mode];
  if (!adapter) {
    throw new Error(`${provider.componentName} cannot use catalog raw VM factory with user-data mode ${provider.userData.mode}`);
  }

  const property = provider.userData.property;
  const rendered = adapter(input.userData, provider.userData.metadataKey);

  if (provider.userData.mode === "metadata") {
    return { [property ?? "metadata"]: rendered };
  }

  return { [property ?? "userData"]: rendered };
}
```

- [ ] **Step 5: Run adapter and raw provider tests**

Run:

```bash
npm run build
node --test dist/test/userDataAdapters.test.js dist/test/additionalCloudPublishers.test.js dist/test/ociPublisher.test.js dist/test/scalewayPublisher.test.js
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add src/userDataAdapters.ts src/catalogVmFactory.ts test/userDataAdapters.test.ts
git commit -m "refactor: consolidate user data adapter modes"
```

---

### Task 5: Final Verification and Documentation Regeneration

**Files:**
- Modify if generated: `site/source/_generated/*`
- Modify if generated: `schema.json`
- Modify if generated: `sdk/**`

- [ ] **Step 1: Regenerate docs**

Run:

```bash
npm run docs:gen
```

Expected: PASS.

- [ ] **Step 2: Run full verification**

Run:

```bash
npm run typecheck
npm test
npm run go:test
npm run registry:check
npm run catalog:check
npm run build --prefix site
```

Expected: all commands exit 0.

- [ ] **Step 3: Regenerate SDKs only if schema changed**

Run:

```bash
git diff --quiet schema.json || npm run sdk:gen
```

Expected: no output if `schema.json` did not change; otherwise SDK generation succeeds.

- [ ] **Step 4: Inspect final diff**

Run:

```bash
git status --short --branch
git diff --stat
```

Expected: only intended files from this plan are modified.

- [ ] **Step 5: Commit generated outputs after docs or SDK generation**

Run:

```bash
if ! git diff --quiet -- site/source/_generated schema.json sdk; then
  git add site/source/_generated schema.json sdk
  git commit -m "chore: refresh generated provider artifacts"
fi
```

Expected: creates a commit only when generated artifacts changed.

- [ ] **Step 6: Final status**

Run:

```bash
git status --short --branch
```

Expected: clean worktree or only unrelated pre-existing files that are explicitly not part of this plan.

---

## Self-Review

**Spec coverage:** This plan covers all four audit findings: `NetskopeRegistration` catalog drift, shallow registry validation, Go output contract gaps, and split/unused user-data adapter modes.

**Placeholder scan:** No placeholders, TBDs, or generic "add tests" steps remain. Each code-changing step includes exact file paths and code.

**Type consistency:** The plan consistently uses `upstreamPropertyChecks`, `ProviderUpstreamPropertyCheck`, `userDataAdapters`, `NetskopeRegistration`, and existing provider/user-data mode names.
