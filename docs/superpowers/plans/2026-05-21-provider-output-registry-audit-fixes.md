# Provider Output Registry Audit Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the provider architecture audit findings around registry-backed output mappings, raw-provider IP outputs, registry readiness coverage, and upstream input validation coverage.

**Architecture:** Keep the current catalog-driven provider model, but make the contract more explicit: provider code must map only fields that exist in upstream Pulumi schemas, and registry checks must cover every public provider resource. The implementation adds tests that reproduce the current gaps, then fixes output extraction, readiness token coverage, and catalog upstream property declarations without changing public component APIs.

**Tech Stack:** TypeScript, Node test runner, Pulumi TypeScript mocks, Go provider integration tests, Pulumi Registry schema JSON, `pulumi-go-provider`, Git.

---

## File Structure

- Modify `internal/provider/provider_test.go`: make Go output tests use upstream-correct OVH and Scaleway output shapes, then add output assertions for raw bootstrap providers with registry-backed IP fields.
- Modify `internal/provider/components.go`: add missing raw resource output fields and shared array/object output helpers, then wire Go publisher outputs to upstream-correct properties.
- Modify `src/rawResource.ts`: add a typed helper for reading output properties from untyped raw `pulumi.CustomResource` instances.
- Modify `src/digitaloceanPublisher.ts`, `src/vultrPublisher.ts`, `src/exoscalePublisher.ts`, `src/equinixPublisher.ts`, `src/outscalePublisher.ts`, `src/tencentcloudPublisher.ts`, `src/opentelekomcloudPublisher.ts`, and `src/ovhPublisher.ts`: add `mapOutputs` for upstream IP fields.
- Modify `test/additionalCloudPublishers.test.ts`: overlay mock provider output fields and assert TypeScript raw publishers expose IP outputs when the upstream resource exposes them.
- Modify `scripts/check-registry-readiness.mjs`: derive expected provider resource tokens from `schema.json` and export helper constants for testing.
- Create `test/registryReadinessConfig.test.ts`: assert registry readiness coverage does not drift behind schema resources and TypeScript component sources.
- Modify `src/providerCatalog.ts`: add upstream property checks for every catalog-driven provider input path that component implementations send to child resources.
- Modify `test/providerCatalog.test.ts`: assert all catalog-driven providers declare upstream input property checks.

---

### Task 1: Correct Go OVH and Scaleway Output Shape Tests

**Files:**
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/components.go`

- [ ] **Step 1: Update the failing Go output test to use real upstream property names**

In `internal/provider/provider_test.go`, inside `constructAndCollectPublisherOutput`, replace the OVH and Scaleway mock cases:

```go
				case "ovh:CloudProject/instance:Instance":
					state = state.Set("ipAddresses", property.New([]property.Value{property.New("198.51.100.30")}))
				case "scaleway:instance/server:Server":
					state = state.Set("publicIp", property.New("198.51.100.40"))
```

with:

```go
				case "ovh:CloudProject/instance:Instance":
					state = state.Set("addresses", property.New([]property.Value{property.New(map[string]property.Value{
						"ip":      property.New("198.51.100.30"),
						"version": property.New(4.0),
					})}))
				case "scaleway:instance/server:Server":
					state = state.Set("publicIps", property.New([]property.Value{property.New(map[string]property.Value{
						"address": property.New("198.51.100.40"),
					})}))
					state = state.Set("privateIps", property.New([]property.Value{property.New(map[string]property.Value{
						"address": property.New("10.0.0.40"),
					})}))
```

In the Scaleway test case in `TestProviderOutputsExposeAvailableIPAddresses`, change the expected public IP case to keep `outputField: "publicIp"` and `expected: "198.51.100.40"`.

- [ ] **Step 2: Run the focused Go test to verify it fails**

Run:

```bash
gofmt -w internal/provider/provider_test.go
go test ./internal/provider -run TestProviderOutputsExposeAvailableIPAddresses -count=1
```

Expected: FAIL for OVH and Scaleway because `components.go` still reads `ipAddresses` and `publicIp`.

- [ ] **Step 3: Add registry-backed array output fields to `rawVMResource`**

In `internal/provider/components.go`, update `rawVMResource` so it includes `Addresses`, `PrivateIPs`, and `PublicIPs`:

```go
type rawVMResource struct {
	pulumi.CustomResourceState

	AccessIPV4       pulumi.StringOutput      `pulumi:"accessIpV4"`
	Address          pulumi.StringOutput      `pulumi:"address"`
	Addresses        pulumi.ArrayOutput       `pulumi:"addresses"`
	IPAddresses      pulumi.StringArrayOutput `pulumi:"ipAddresses"`
	Ipv4Address      pulumi.StringOutput      `pulumi:"ipv4Address"`
	Ipv4Addresses    pulumi.ArrayOutput       `pulumi:"ipv4Addresses"`
	Networks         pulumi.ArrayOutput       `pulumi:"networks"`
	NicListStatuses  pulumi.AnyOutput         `pulumi:"nicListStatuses"`
	PrimaryIPAddress pulumi.StringOutput      `pulumi:"primaryIpAddress"`
	PrivateIP        pulumi.StringOutput      `pulumi:"privateIp"`
	PrivateIPs       pulumi.ArrayOutput       `pulumi:"privateIps"`
	PublicIP         pulumi.StringOutput      `pulumi:"publicIp"`
	PublicIPs        pulumi.ArrayOutput       `pulumi:"publicIps"`
}
```

- [ ] **Step 4: Add a shared map-array field helper**

In `internal/provider/components.go`, immediately after `firstStringOutput`, add:

```go
func firstMapFieldOutput(values pulumi.ArrayOutput, field string) pulumi.StringOutput {
	return values.ApplyT(func(items []interface{}) string {
		if len(items) == 0 {
			return ""
		}
		item, ok := items[0].(map[string]interface{})
		if !ok {
			return ""
		}
		value, ok := item[field]
		if !ok || value == nil {
			return ""
		}
		return fmt.Sprint(value)
	}).(pulumi.StringOutput)
}
```

- [ ] **Step 5: Wire OVH and Scaleway to registry-backed fields**

In `NewOvhPublisher`, replace:

```go
outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), firstStringOutput(instance.IPAddresses), args.PlacementLabels)
```

with:

```go
outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), firstMapFieldOutput(instance.Addresses, "ip"), args.PlacementLabels)
```

In `NewScalewayPublisher`, replace:

```go
outputs[publisherName] = publisherOutput(registration, server.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), server.PublicIP, args.PlacementLabels)
```

with:

```go
outputs[publisherName] = publisherOutput(registration, server.ID().ToStringOutput(), firstMapFieldOutput(server.PrivateIPs, "address"), firstMapFieldOutput(server.PublicIPs, "address"), args.PlacementLabels)
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
git commit -m "fix: use registry-backed ovh and scaleway output fields"
```

---

### Task 2: Expose Raw Provider IP Outputs in TypeScript and Go

**Files:**
- Modify: `src/rawResource.ts`
- Modify: `test/additionalCloudPublishers.test.ts`
- Modify: `src/digitaloceanPublisher.ts`
- Modify: `src/vultrPublisher.ts`
- Modify: `src/exoscalePublisher.ts`
- Modify: `src/equinixPublisher.ts`
- Modify: `src/outscalePublisher.ts`
- Modify: `src/tencentcloudPublisher.ts`
- Modify: `src/opentelekomcloudPublisher.ts`
- Modify: `src/ovhPublisher.ts`
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/components.go`

- [ ] **Step 1: Add TypeScript mock output overlays**

In `test/additionalCloudPublishers.test.ts`, add this constant after `const createdResources`:

```ts
const mockedResourceOutputs: Record<string, Record<string, any>> = {
  "digitalocean:index/droplet:Droplet": {
    ipv4Address: "203.0.113.51",
    ipv4AddressPrivate: "10.0.0.51",
  },
  "vultr:index/instance:Instance": {
    mainIp: "203.0.113.52",
    internalIp: "10.0.0.52",
  },
  "exoscale:index/computeInstance:ComputeInstance": {
    publicIpAddress: "203.0.113.53",
  },
  "equinix:metal/device:Device": {
    accessPublicIpv4: "203.0.113.54",
    accessPrivateIpv4: "10.0.0.54",
  },
  "outscale:index/vm:Vm": {
    publicIp: "203.0.113.55",
    privateIp: "10.0.0.55",
  },
  "tencentcloud:index/instance:Instance": {
    publicIp: "203.0.113.56",
    privateIp: "10.0.0.56",
  },
  "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2": {
    accessIpV4: "203.0.113.57",
  },
  "ovh:CloudProject/instance:Instance": {
    addresses: [{ ip: "203.0.113.58", version: 4 }],
  },
};
```

Then replace the `newResource` return in the Pulumi mocks:

```ts
    return { id: `${args.name}-id`, state: args.inputs };
```

with:

```ts
    return {
      id: `${args.name}-id`,
      state: {
        ...args.inputs,
        ...(mockedResourceOutputs[args.type] ?? {}),
      },
    };
```

- [ ] **Step 2: Add failing TypeScript IP output assertions**

In `test/additionalCloudPublishers.test.ts`, append:

```ts
test("catalog raw VM publishers expose registry-backed IP outputs", async () => {
  const cases: Array<{
    name: string;
    component: pulumi.ComponentResource & { publishers: pulumi.Output<Record<string, PublisherOutput>> };
    expectedPrivateIp: string;
    expectedPublicIp: string | undefined;
  }> = [{
    name: "DigitalOcean",
    component: new DigitaloceanPublisher("digitalocean-outputs", baseArgs({ region: "ams3" })),
    expectedPrivateIp: "10.0.0.51",
    expectedPublicIp: "203.0.113.51",
  }, {
    name: "Vultr",
    component: new VultrPublisher("vultr-outputs", baseArgs({ region: "ams", plan: "vc2-2c-4gb", osId: 1743 })),
    expectedPrivateIp: "10.0.0.52",
    expectedPublicIp: "203.0.113.52",
  }, {
    name: "Exoscale",
    component: new ExoscalePublisher("exoscale-outputs", baseArgs({ zone: "ch-gva-2", type: "standard.medium", templateId: "template-id", diskSize: 50 })),
    expectedPrivateIp: "",
    expectedPublicIp: "203.0.113.53",
  }, {
    name: "Equinix",
    component: new EquinixPublisher("equinix-outputs", baseArgs({ projectId: "project-id", metro: "AM", plan: "c3.small.x86" })),
    expectedPrivateIp: "10.0.0.54",
    expectedPublicIp: "203.0.113.54",
  }, {
    name: "Outscale",
    component: new OutscalePublisher("outscale-outputs", baseArgs({ imageId: "ami-123" })),
    expectedPrivateIp: "10.0.0.55",
    expectedPublicIp: "203.0.113.55",
  }, {
    name: "TencentCloud",
    component: new TencentcloudPublisher("tencent-outputs", baseArgs({ availabilityZone: "ap-guangzhou-6", imageId: "img-123" })),
    expectedPrivateIp: "10.0.0.56",
    expectedPublicIp: "203.0.113.56",
  }, {
    name: "OpenTelekomCloud",
    component: new OpentelekomcloudPublisher("otc-outputs", baseArgs({ networks: [{ name: "private" }] })),
    expectedPrivateIp: "",
    expectedPublicIp: "203.0.113.57",
  }];

  for (const tc of cases) {
    const publishers = await outputValue<Record<string, PublisherOutput>>(tc.component.publishers);
    assert.equal(publishers["pub-1"].privateIp, tc.expectedPrivateIp, `${tc.name} privateIp mismatch`);
    assert.equal(publishers["pub-1"].publicIp, tc.expectedPublicIp, `${tc.name} publicIp mismatch`);
  }
});
```

- [ ] **Step 3: Run TypeScript tests to verify they fail**

Run:

```bash
npm run build
node --test dist/test/additionalCloudPublishers.test.js
```

Expected: FAIL because the catalog raw VM factory returns empty `privateIp` and `publicIp` when `mapOutputs` is omitted.

- [ ] **Step 4: Add a typed raw resource output helper**

Replace `src/rawResource.ts` with:

```ts
import * as pulumi from "@pulumi/pulumi";

export class RawResource extends pulumi.CustomResource {
  constructor(name: string, type: string, args: pulumi.Inputs, opts?: pulumi.CustomResourceOptions) {
    super(type, name, args, opts);
  }

  output<T>(name: string): pulumi.Output<T> {
    return (this as unknown as Record<string, pulumi.Output<T>>)[name] ?? pulumi.output(undefined as T);
  }
}
```

- [ ] **Step 5: Add TypeScript `mapOutputs` to raw providers**

In `src/digitaloceanPublisher.ts`, add this property to the `createCatalogRawVmPublishers` options object after `mapInputs`:

```ts
      mapOutputs: (resource) => ({
        vmId: resource.id,
        privateIp: resource.output<string>("ipv4AddressPrivate"),
        publicIp: resource.output<string>("ipv4Address"),
      }),
```

In `src/vultrPublisher.ts`, add:

```ts
      mapOutputs: (resource) => ({
        vmId: resource.id,
        privateIp: resource.output<string>("internalIp"),
        publicIp: resource.output<string>("mainIp"),
      }),
```

In `src/exoscalePublisher.ts`, add:

```ts
      mapOutputs: (resource) => ({
        vmId: resource.id,
        privateIp: pulumi.output(""),
        publicIp: resource.output<string>("publicIpAddress"),
      }),
```

In `src/equinixPublisher.ts`, add:

```ts
      mapOutputs: (resource) => ({
        vmId: resource.id,
        privateIp: resource.output<string>("accessPrivateIpv4"),
        publicIp: resource.output<string>("accessPublicIpv4"),
      }),
```

In `src/outscalePublisher.ts`, add:

```ts
      mapOutputs: (resource) => ({
        vmId: resource.id,
        privateIp: resource.output<string>("privateIp"),
        publicIp: resource.output<string>("publicIp"),
      }),
```

In `src/tencentcloudPublisher.ts`, add:

```ts
      mapOutputs: (resource) => ({
        vmId: resource.id,
        privateIp: resource.output<string>("privateIp"),
        publicIp: resource.output<string>("publicIp"),
      }),
```

In `src/opentelekomcloudPublisher.ts`, add:

```ts
      mapOutputs: (resource) => ({
        vmId: resource.id,
        privateIp: pulumi.output(""),
        publicIp: resource.output<string>("accessIpV4"),
      }),
```

In `src/ovhPublisher.ts`, replace the current return block:

```ts
      return {
        vmId: instance.id,
        privateIp: pulumi.output(""),
        publicIp: pulumi.output(""),
      };
```

with:

```ts
      return {
        vmId: instance.id,
        privateIp: pulumi.output(""),
        publicIp: instance.addresses.apply((addresses) => addresses?.[0]?.ip),
      };
```

- [ ] **Step 6: Add Go output test cases for raw bootstrap providers**

In `internal/provider/provider_test.go`, inside `constructAndCollectPublisherOutput`, add these mock cases to the `switch string(args.TypeToken)` block:

```go
				case "digitalocean:index/droplet:Droplet":
					state = state.Set("ipv4Address", property.New("203.0.113.51"))
					state = state.Set("ipv4AddressPrivate", property.New("10.0.0.51"))
				case "vultr:index/instance:Instance":
					state = state.Set("mainIp", property.New("203.0.113.52"))
					state = state.Set("internalIp", property.New("10.0.0.52"))
				case "exoscale:index/computeInstance:ComputeInstance":
					state = state.Set("publicIpAddress", property.New("203.0.113.53"))
				case "equinix:metal/device:Device":
					state = state.Set("accessPublicIpv4", property.New("203.0.113.54"))
					state = state.Set("accessPrivateIpv4", property.New("10.0.0.54"))
				case "outscale:index/vm:Vm":
					state = state.Set("publicIp", property.New("203.0.113.55"))
					state = state.Set("privateIp", property.New("10.0.0.55"))
				case "tencentcloud:index/instance:Instance":
					state = state.Set("publicIp", property.New("203.0.113.56"))
					state = state.Set("privateIp", property.New("10.0.0.56"))
				case "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2":
					state = state.Set("accessIpV4", property.New("203.0.113.57"))
```

Append these cases to the `cases` slice in `TestProviderOutputsExposeAvailableIPAddresses`:

```go
		{
			name:  "DigitalOcean",
			token: "netskope-publisher:index:DigitaloceanPublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"region":        property.New("ams3"),
			}),
			outputField: "publicIp",
			expected:    "203.0.113.51",
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
			outputField: "publicIp",
			expected:    "203.0.113.52",
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
			outputField: "publicIp",
			expected:    "203.0.113.53",
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
			outputField: "publicIp",
			expected:    "203.0.113.54",
		},
		{
			name:  "Outscale",
			token: "netskope-publisher:index:OutscalePublisher",
			inputs: property.NewMap(map[string]property.Value{
				"names":         property.New([]property.Value{property.New("pub-1")}),
				"registrations": registrationMap("pub-1"),
				"imageId":       property.New("ami-123"),
			}),
			outputField: "publicIp",
			expected:    "203.0.113.55",
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
			outputField: "publicIp",
			expected:    "203.0.113.56",
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
			outputField: "publicIp",
			expected:    "203.0.113.57",
		},
```

- [ ] **Step 7: Run output tests to verify Go still fails**

Run:

```bash
gofmt -w internal/provider/provider_test.go
go test ./internal/provider -run TestProviderOutputsExposeAvailableIPAddresses -count=1
```

Expected: FAIL for the newly added Go cases whose output mappings still return empty strings.

- [ ] **Step 8: Add Go raw resource fields**

In `internal/provider/components.go`, update `rawVMResource` so these fields are present:

```go
	AccessPrivateIpv4 pulumi.StringOutput `pulumi:"accessPrivateIpv4"`
	AccessPublicIpv4  pulumi.StringOutput `pulumi:"accessPublicIpv4"`
	InternalIP        pulumi.StringOutput `pulumi:"internalIp"`
	Ipv4AddressPrivate pulumi.StringOutput `pulumi:"ipv4AddressPrivate"`
	MainIP            pulumi.StringOutput `pulumi:"mainIp"`
	PublicIPAddress   pulumi.StringOutput `pulumi:"publicIpAddress"`
```

Place them alphabetically with the existing fields in `rawVMResource`.

- [ ] **Step 9: Wire Go output mappings for raw bootstrap providers**

In `NewDigitaloceanPublisher`, replace:

```go
return rawBootstrapBuildResult{droplet.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput()}, err
```

with:

```go
return rawBootstrapBuildResult{droplet.ID().ToStringOutput(), droplet.Ipv4AddressPrivate, droplet.Ipv4Address}, err
```

In `NewVultrPublisher`, replace the return with:

```go
return rawBootstrapBuildResult{instance.ID().ToStringOutput(), instance.InternalIP, instance.MainIP}, err
```

In `NewExoscalePublisher`, replace the return with:

```go
return rawBootstrapBuildResult{instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), instance.PublicIPAddress}, err
```

In `NewEquinixPublisher`, replace the return with:

```go
return rawBootstrapBuildResult{device.ID().ToStringOutput(), device.AccessPrivateIpv4, device.AccessPublicIpv4}, err
```

In `NewOutscalePublisher`, replace the return with:

```go
return rawBootstrapBuildResult{vm.ID().ToStringOutput(), vm.PrivateIP, vm.PublicIP}, err
```

In `NewOpentelekomcloudPublisher`, replace the return with:

```go
return rawBootstrapBuildResult{instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), instance.AccessIPV4}, err
```

In `NewTencentcloudPublisher`, replace the return with:

```go
return rawBootstrapBuildResult{instance.ID().ToStringOutput(), instance.PrivateIP, instance.PublicIP}, err
```

- [ ] **Step 10: Verify TypeScript and Go output tests**

Run:

```bash
npm run build
node --test dist/test/additionalCloudPublishers.test.js
gofmt -w internal/provider/components.go internal/provider/provider_test.go
go test ./internal/provider -run TestProviderOutputsExposeAvailableIPAddresses -count=1
npm run go:test
```

Expected: PASS.

- [ ] **Step 11: Commit**

```bash
git add src/rawResource.ts src/digitaloceanPublisher.ts src/vultrPublisher.ts src/exoscalePublisher.ts src/equinixPublisher.ts src/outscalePublisher.ts src/tencentcloudPublisher.ts src/opentelekomcloudPublisher.ts src/ovhPublisher.ts test/additionalCloudPublishers.test.ts internal/provider/components.go internal/provider/provider_test.go
git commit -m "fix: expose registry-backed raw provider ip outputs"
```

---

### Task 3: Make Registry Readiness Coverage Dynamic

**Files:**
- Modify: `scripts/check-registry-readiness.mjs`
- Create: `test/registryReadinessConfig.test.ts`

- [ ] **Step 1: Add failing readiness coverage tests**

Create `test/registryReadinessConfig.test.ts` with:

```ts
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { catalogProviders } from "../src/providerCatalog";

function objectLiteralValues(source: string, name: string): string[] {
  const match = source.match(new RegExp(`const ${name} = \\{([\\s\\S]*?)\\};`));
  assert.ok(match, `${name} object not found`);
  return Array.from(match[1].matchAll(/"[^"]+":\\s*"([^"]+)"/g)).map((entry) => entry[1]);
}

test("registry readiness derives expectedResourceTokens from schema resources", () => {
  const script = readFileSync("scripts/check-registry-readiness.mjs", "utf8");

  assert.match(script, /const expectedResourceTokens = schema[\s\S]*Object\.keys\(schema\.resources \?\? \{\}\)[\s\S]*startsWith\("netskope-publisher:index:"\)/);
});

test("registry readiness sourceTokens covers every TypeScript component source", () => {
  const script = readFileSync("scripts/check-registry-readiness.mjs", "utf8");
  const sourceTokenValues = new Set(objectLiteralValues(script, "sourceTokens"));
  const missing = catalogProviders
    .filter((provider) => provider.componentName !== "NetskopeRegistration")
    .filter((provider) => provider.componentName !== "PrivateApp")
    .filter((provider) => provider.componentName !== "TagPublisherAssignment")
    .filter((provider) => provider.componentName !== "RealtimeProtectionPolicy")
    .map((provider) => provider.token)
    .filter((token) => !sourceTokenValues.has(token))
    .sort();

  assert.deepEqual(missing, []);
});
```

- [ ] **Step 2: Run the readiness tests to verify they fail**

Run:

```bash
npm run build
node --test dist/test/registryReadinessConfig.test.js
```

Expected: FAIL because `expectedResourceTokens` is still a hard-coded array and `sourceTokens` omits the newer TypeScript component sources.

- [ ] **Step 3: Replace hard-coded readiness resource tokens with schema-derived tokens**

In `scripts/check-registry-readiness.mjs`, delete the full `const expectedResourceTokens = [...]` block.

After `schema` is parsed, add:

```js
const expectedResourceTokens = schema
  ? Object.keys(schema.resources ?? {})
    .filter((token) => token.startsWith("netskope-publisher:index:"))
    .sort()
  : [];
```

Keep the existing `for (const token of expectedResourceTokens)` loops unchanged.

- [ ] **Step 4: Extend `sourceTokens` to cover all TypeScript component sources**

In `scripts/check-registry-readiness.mjs`, replace the `sourceTokens` object with:

```js
const sourceTokens = {
  "src/awsPublisher.ts": "netskope-publisher:index:AwsPublisher",
  "src/azurePublisher.ts": "netskope-publisher:index:AzurePublisher",
  "src/gcpPublisher.ts": "netskope-publisher:index:GcpPublisher",
  "src/kubernetesPublisher.ts": "netskope-publisher:index:KubernetesPublisher",
  "src/vspherePublisher.ts": "netskope-publisher:index:VspherePublisher",
  "src/esxiPublisher.ts": "netskope-publisher:index:EsxiPublisher",
  "src/hcloudPublisher.ts": "netskope-publisher:index:HcloudPublisher",
  "src/nutanixPublisher.ts": "netskope-publisher:index:NutanixPublisher",
  "src/openstackPublisher.ts": "netskope-publisher:index:OpenstackPublisher",
  "src/ovhPublisher.ts": "netskope-publisher:index:OvhPublisher",
  "src/scalewayPublisher.ts": "netskope-publisher:index:ScalewayPublisher",
  "src/ociPublisher.ts": "netskope-publisher:index:OciPublisher",
  "src/alicloudPublisher.ts": "netskope-publisher:index:AlicloudPublisher",
  "src/proxmoxvePublisher.ts": "netskope-publisher:index:ProxmoxvePublisher",
  "src/digitaloceanPublisher.ts": "netskope-publisher:index:DigitaloceanPublisher",
  "src/vultrPublisher.ts": "netskope-publisher:index:VultrPublisher",
  "src/exoscalePublisher.ts": "netskope-publisher:index:ExoscalePublisher",
  "src/upcloudPublisher.ts": "netskope-publisher:index:UpcloudPublisher",
  "src/stackitPublisher.ts": "netskope-publisher:index:StackitPublisher",
  "src/equinixPublisher.ts": "netskope-publisher:index:EquinixPublisher",
  "src/outscalePublisher.ts": "netskope-publisher:index:OutscalePublisher",
  "src/opentelekomcloudPublisher.ts": "netskope-publisher:index:OpentelekomcloudPublisher",
  "src/tencentcloudPublisher.ts": "netskope-publisher:index:TencentcloudPublisher",
  "src/yandexPublisher.ts": "netskope-publisher:index:YandexPublisher",
  "src/hypervPublisher.ts": "netskope-publisher:index:HypervPublisher",
};
```

- [ ] **Step 5: Verify readiness coverage**

Run:

```bash
npm run build
node --test dist/test/registryReadinessConfig.test.js
npm run registry:check
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add scripts/check-registry-readiness.mjs test/registryReadinessConfig.test.ts
git commit -m "test: keep registry readiness coverage in sync"
```

---

### Task 4: Require Upstream Property Checks for Every Catalog Provider

**Files:**
- Modify: `test/providerCatalog.test.ts`
- Modify: `src/providerCatalog.ts`
- Modify: `src/providerRegistrySchema.ts`
- Modify: `test/providerRegistrySchema.test.ts`

- [ ] **Step 1: Add failing catalog coverage test**

Append this to `test/providerCatalog.test.ts`:

```ts
test("catalog-driven providers declare upstream property checks for child resource inputs", () => {
  const missing = catalogDrivenProviders
    .filter((provider) => (provider.registrySchemaChecks?.length ?? 0) === 0)
    .filter((provider) => (provider.upstreamPropertyChecks?.length ?? 0) === 0)
    .map((provider) => provider.componentName)
    .sort();

  assert.deepEqual(missing, []);
});
```

- [ ] **Step 2: Run catalog tests to verify they fail**

Run:

```bash
npm run build
node --test dist/test/providerCatalog.test.js
```

Expected: FAIL with the providers that do not declare `upstreamPropertyChecks` or `registrySchemaChecks`.

- [ ] **Step 3: Extend registry validation to distinguish input and output paths**

In `src/providerCatalog.ts`, replace `ProviderRegistrySchemaCheck` with:

```ts
export interface ProviderRegistrySchemaCheck {
  resourceToken: string;
  propertyPath: string[];
  description: string;
  propertyKind?: "input" | "output";
}
```

In `src/providerRegistrySchema.ts`, update the `RegistryProviderEntry` check shapes so both `registrySchemaChecks` and `upstreamPropertyChecks` include `propertyKind?: "input" | "output"`:

```ts
  registrySchemaChecks?: Array<{
    resourceToken: string;
    propertyPath: string[];
    description: string;
    propertyKind?: "input" | "output";
  }>;
  upstreamPropertyChecks?: Array<{
    resourceToken: string;
    propertyPath: string[];
    description: string;
    propertyKind?: "input" | "output";
  }>;
```

Update `RegistrySchemaResource` in `src/providerRegistrySchema.ts`:

```ts
export interface RegistrySchemaResource {
  inputProperties?: Record<string, unknown>;
  properties?: Record<string, unknown>;
}
```

Add this helper above `validateProviderAgainstRegistrySchema`:

```ts
function checkRegistryPath(provider: RegistryProviderEntry, schema: RegistrySchema, check: { resourceToken: string; propertyPath: string[]; description: string; propertyKind?: "input" | "output" }): string | undefined {
  const checkedResource = schema.resources?.[check.resourceToken];
  if (!checkedResource) {
    return `${provider.componentName} upstream schema missing resource token ${check.resourceToken}`;
  }

  const propertyKind = check.propertyKind ?? "input";
  const properties = propertyKind === "output"
    ? checkedResource.properties ?? {}
    : checkedResource.inputProperties ?? {};

  if (!schemaHasPath(schema, properties, check.propertyPath)) {
    return `${provider.componentName} upstream resource ${check.resourceToken} missing ${propertyKind} ${check.description} path ${check.propertyPath.join(".")}`;
  }

  return undefined;
}
```

Then replace both duplicated `registrySchemaChecks` and `upstreamPropertyChecks` loops in `validateProviderAgainstRegistrySchema` with:

```ts
  for (const check of provider.registrySchemaChecks ?? []) {
    const error = checkRegistryPath(provider, schema, check);
    if (error) {
      errors.push(error);
    }
  }

  for (const check of provider.upstreamPropertyChecks ?? []) {
    const error = checkRegistryPath(provider, schema, check);
    if (error) {
      errors.push(error);
    }
  }
```

- [ ] **Step 4: Add helpers for compact upstream checks**

In `src/providerCatalog.ts`, after `function userDataProperty`, add:

```ts
function upstreamChecks(resourceToken: string, checks: Array<[string[], string]>): ProviderUpstreamPropertyCheck[] {
  return checks.map(([propertyPath, description]) => ({
    resourceToken,
    propertyPath,
    description,
  }));
}

function upstreamOutputChecks(resourceToken: string, checks: Array<[string[], string]>): ProviderUpstreamPropertyCheck[] {
  return checks.map(([propertyPath, description]) => ({
    resourceToken,
    propertyPath,
    description,
    propertyKind: "output",
  }));
}
```

- [ ] **Step 5: Add upstream property checks for existing bootstrap providers**

In `src/providerCatalog.ts`, add these `upstreamPropertyChecks` properties to the corresponding provider definitions.

For `HcloudPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("hcloud:index/server:Server", [
  [["serverType"], "server type"],
  [["image"], "server image"],
  [["publicNets"], "public network configuration"],
  [["userData"], "cloud-init user data"],
]).concat(upstreamOutputChecks("hcloud:index/server:Server", [
  [["ipv4Address"], "public IPv4 output"],
  [["networks"], "private network output"],
])),
```

For `NutanixPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("nutanix:index/virtualMachine:VirtualMachine", [
  [["clusterUuid"], "cluster UUID"],
  [["guestCustomizationCloudInitUserData"], "cloud-init user data"],
  [["diskLists"], "image disk list"],
  [["nicLists"], "network interface list"],
]).concat(upstreamOutputChecks("nutanix:index/virtualMachine:VirtualMachine", [
  [["nicListStatuses"], "network status outputs"],
])),
```

For `OvhPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("ovh:CloudProject/instance:Instance", [
  [["billingPeriod"], "billing period"],
  [["bootFrom", "imageId"], "boot image ID"],
  [["flavor", "flavorId"], "flavor ID"],
  [["network"], "instance network"],
  [["userData"], "cloud-init user data"],
]).concat(upstreamOutputChecks("ovh:CloudProject/instance:Instance", [
  [["addresses"], "instance addresses output"],
])),
```

For `ScalewayPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("scaleway:index/instanceServer:InstanceServer", [
  [["type"], "server type"],
  [["image"], "server image"],
  [["cloudInit"], "cloud-init user data"],
  [["userData"], "instance user data map"],
]).concat(upstreamOutputChecks("scaleway:index/instanceServer:InstanceServer", [
  [["publicIps"], "public IP outputs"],
  [["privateIps"], "private IP outputs"],
])),
```

For `AlicloudPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("alicloud:ecs/instance:Instance", [
  [["instanceName"], "instance name"],
  [["instanceType"], "instance type"],
  [["imageId"], "image ID"],
  [["vswitchId"], "vSwitch ID"],
  [["securityGroups"], "security groups"],
  [["userData"], "cloud-init user data"],
]).concat(upstreamOutputChecks("alicloud:ecs/instance:Instance", [
  [["primaryIpAddress"], "primary private IP output"],
  [["publicIp"], "public IP output"],
])),
```

- [ ] **Step 6: Add upstream property checks for new raw providers**

In `src/providerCatalog.ts`, add these `upstreamPropertyChecks` properties to the corresponding provider definitions.

For `DigitaloceanPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("digitalocean:index/droplet:Droplet", [
  [["region"], "region"],
  [["size"], "droplet size"],
  [["image"], "image"],
  [["userData"], "cloud-init user data"],
]).concat(upstreamOutputChecks("digitalocean:index/droplet:Droplet", [
  [["ipv4Address"], "public IPv4 output"],
  [["ipv4AddressPrivate"], "private IPv4 output"],
])),
```

For `VultrPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("vultr:index/instance:Instance", [
  [["region"], "region"],
  [["plan"], "plan"],
  [["osId"], "operating system ID"],
  [["imageId"], "image ID"],
  [["userData"], "cloud-init user data"],
]).concat(upstreamOutputChecks("vultr:index/instance:Instance", [
  [["mainIp"], "public IP output"],
  [["internalIp"], "private IP output"],
])),
```

For `ExoscalePublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("exoscale:index/computeInstance:ComputeInstance", [
  [["zone"], "zone"],
  [["type"], "instance type"],
  [["templateId"], "template ID"],
  [["diskSize"], "disk size"],
  [["userData"], "cloud-init user data"],
]).concat(upstreamOutputChecks("exoscale:index/computeInstance:ComputeInstance", [
  [["publicIpAddress"], "public IP output"],
])),
```

For `UpcloudPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("upcloud:index/server:Server", [
  [["hostname"], "hostname"],
  [["zone"], "zone"],
  [["plan"], "plan"],
  [["template"], "template"],
  [["metadata"], "metadata service toggle"],
  [["userData"], "cloud-init user data"],
]),
```

For `StackitPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("stackit:index/server:Server", [
  [["projectId"], "project ID"],
  [["machineType"], "machine type"],
  [["imageId"], "image ID"],
  [["networkInterfaces"], "network interfaces"],
  [["userData"], "cloud-init user data"],
]),
```

For `EquinixPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("equinix:metal/device:Device", [
  [["projectId"], "project ID"],
  [["metro"], "metro"],
  [["plan"], "plan"],
  [["operatingSystem"], "operating system"],
  [["userData"], "cloud-init user data"],
]).concat(upstreamOutputChecks("equinix:metal/device:Device", [
  [["accessPublicIpv4"], "public IPv4 output"],
  [["accessPrivateIpv4"], "private IPv4 output"],
])),
```

For `OutscalePublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("outscale:index/vm:Vm", [
  [["imageId"], "image ID"],
  [["vmType"], "VM type"],
  [["subnetId"], "subnet ID"],
  [["securityGroupIds"], "security groups"],
  [["userData"], "cloud-init user data"],
]).concat(upstreamOutputChecks("outscale:index/vm:Vm", [
  [["publicIp"], "public IP output"],
  [["privateIp"], "private IP output"],
])),
```

For `OpentelekomcloudPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", [
  [["imageName"], "image name"],
  [["imageId"], "image ID"],
  [["flavorName"], "flavor name"],
  [["flavorId"], "flavor ID"],
  [["networks"], "network attachments"],
  [["userData"], "cloud-init user data"],
]).concat(upstreamOutputChecks("opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", [
  [["accessIpV4"], "public IPv4 output"],
])),
```

For `TencentcloudPublisher`:

```ts
upstreamPropertyChecks: upstreamChecks("tencentcloud:index/instance:Instance", [
  [["availabilityZone"], "availability zone"],
  [["imageId"], "image ID"],
  [["instanceType"], "instance type"],
  [["subnetId"], "subnet ID"],
  [["userDataRaw"], "raw cloud-init user data"],
  [["userDataReplaceOnChange"], "user data replacement toggle"],
]).concat(upstreamOutputChecks("tencentcloud:index/instance:Instance", [
  [["publicIp"], "public IP output"],
  [["privateIp"], "private IP output"],
])),
```

Keep the existing OpenStack, OCI, Proxmox VE, and Yandex checks in place.

- [ ] **Step 7: Add regression tests for output paths and array-item `$ref` paths**

Append these tests to `test/providerRegistrySchema.test.ts`:

```ts
test("validateProviderAgainstRegistrySchema accepts declared upstream output property paths", () => {
  const errors = validateProviderAgainstRegistrySchema({
    componentName: "OutputPublisher",
    resourceToken: "example:index/server:Server",
    upstreamPropertyChecks: [{
      resourceToken: "example:index/server:Server",
      propertyPath: ["publicIps", "address"],
      description: "public IP address",
      propertyKind: "output",
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
        },
        properties: {
          publicIps: {
            type: "array",
            items: { "$ref": "#/types/example:index/ServerPublicIp:ServerPublicIp" },
          },
        },
      },
    },
    types: {
      "example:index/ServerPublicIp:ServerPublicIp": {
        properties: {
          address: { type: "string" },
        },
      },
    },
  });

  assert.deepEqual(errors, []);
});

test("validateProviderAgainstRegistrySchema follows array item refs in declared property paths", () => {
  const errors = validateProviderAgainstRegistrySchema({
    componentName: "ArrayPublisher",
    resourceToken: "example:index/server:Server",
    upstreamPropertyChecks: [{
      resourceToken: "example:index/server:Server",
      propertyPath: ["interfaces", "address"],
      description: "interface address",
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
          interfaces: {
            type: "array",
            items: { "$ref": "#/types/example:index/ServerInterface:ServerInterface" },
          },
        },
      },
    },
    types: {
      "example:index/ServerInterface:ServerInterface": {
        properties: {
          address: { type: "string" },
        },
      },
    },
  });

  assert.deepEqual(errors, []);
});
```

- [ ] **Step 8: Verify catalog and live registry schema checks**

Run:

```bash
npm run build
node --test dist/test/providerCatalog.test.js dist/test/providerRegistrySchema.test.js
npm run catalog:check
npm run registry:check
```

Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add src/providerCatalog.ts src/providerRegistrySchema.ts test/providerCatalog.test.ts test/providerRegistrySchema.test.ts
git commit -m "test: require upstream schema checks for provider inputs"
```

---

### Task 5: Final Verification and Generated Artifacts

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

- [ ] **Step 3: Regenerate SDKs after schema changes**

Run:

```bash
git diff --quiet schema.json || npm run sdk:gen
```

Expected: no output when `schema.json` is unchanged; otherwise SDK generation exits 0.

- [ ] **Step 4: Inspect final diff**

Run:

```bash
git status --short --branch
git diff --stat
```

Expected: only intended files from this plan are modified.

- [ ] **Step 5: Commit generated artifacts**

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

Expected: clean worktree or only unrelated pre-existing files outside this plan.

---

## Self-Review

**Spec coverage:** This plan covers all four audit findings: wrong Go OVH/Scaleway registry output shapes, missing raw-provider IP output mappings, stale registry readiness coverage, and opt-in-only upstream property checks.

**Placeholder scan:** No placeholders, TBDs, or generic "add tests" steps remain. Every code-changing step includes exact paths, code snippets, commands, and expected results.

**Type consistency:** The plan consistently uses existing names: `RawResource.output`, `mapOutputs`, `rawVMResource`, `firstMapFieldOutput`, `upstreamPropertyChecks`, `ProviderUpstreamPropertyCheck`, and current provider component names.
