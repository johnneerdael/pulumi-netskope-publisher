# Cloud Provider Expansion Adapter Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the strong and good suitability cloud providers while migrating bootstrap-based providers to a shared cloud-init adapter model.

**Architecture:** Keep cloud-init rendering centralized, add a small adapter layer for payload placement, and keep provider components responsible only for provider-specific infrastructure inputs. Use typed provider SDKs where the repo already depends on them, and raw Pulumi resource tokens for new providers whose registry schemas do not expose a stable Node SDK package.

**Tech Stack:** TypeScript Pulumi components, Go executable Pulumi provider, Pulumi package schema generation, generated Python/.NET/Go/Java/Rust SDKs, Hexo GitHub Pages docs.

---

## File Structure

- Create: `src/userDataAdapters.ts`
  - TypeScript adapter helpers for plain, base64, metadata, custom-data, cloud-init disk/snippet, GuestInfo, and Scaleway dual placement.
- Create: `src/rawResource.ts`
  - Minimal TypeScript `pulumi.CustomResource` wrapper for providers without usable Node SDK packages.
- Modify: `src/vmPublisherCore.ts`
  - Return rendered payloads through adapter helpers while preserving `userData` and `userDataBase64` for existing component call sites during migration.
- Modify: `src/types.ts`
  - Add args for the 10 new providers.
- Create: `src/digitaloceanPublisher.ts`, `src/vultrPublisher.ts`, `src/exoscalePublisher.ts`, `src/upcloudPublisher.ts`, `src/stackitPublisher.ts`, `src/equinixPublisher.ts`, `src/outscalePublisher.ts`, `src/opentelekomcloudPublisher.ts`, `src/tencentcloudPublisher.ts`, `src/yandexPublisher.ts`
  - New TypeScript component resources.
- Modify: `src/index.ts`
  - Export new components and adapter helpers.
- Create tests under `test/*Publisher.test.ts` for all new TypeScript components.
- Create: `internal/provider/userdata.go`
  - Go adapter helpers mirroring the TypeScript adapter modes.
- Modify: `internal/provider/components.go`
  - Migrate existing bootstrap providers and add the 10 new components.
- Modify: `internal/provider/provider.go`
  - Register new component constructors and update description.
- Modify: `internal/provider/provider_test.go`
  - Add adapter tests and constructor tests for every new provider.
- Modify: `package.json`, `package-lock.json`
  - Add only new provider SDK dependencies that have stable Node packages and are worth typing. Use raw resources for Terraform-bridge packages without Node packages.
- Modify: `README.md`, `docs/_index.md`, `docs/installation-configuration.md`, `site/source/**/*.md`
  - Document the expanded provider set and examples.
- Regenerate: `schema.json`, `sdk/python`, `sdk/dotnet`, `sdk/go`, `sdk/java`, `sdk/rust`, and `site/public`.

## Provider Tokens And Placement

| Component | Resource token | User-data placement | Default Ubuntu selector |
| --- | --- | --- | --- |
| `DigitaloceanPublisher` | `digitalocean:index/droplet:Droplet` | plain `userData` | `image: "ubuntu-22-04-x64"` |
| `VultrPublisher` | `vultr:index/instance:Instance` | plain `userData` | `osId` or `imageId`, user supplied |
| `ExoscalePublisher` | `exoscale:index/computeInstance:ComputeInstance` | plain `userData` | `templateId`, user supplied |
| `UpcloudPublisher` | `upcloud:index/server:Server` | plain `userData` | `template: "01000000-0000-4000-8000-000030220200"` unless overridden |
| `StackitPublisher` | `stackit:index/server:Server` | plain `userData` | `imageId`, user supplied |
| `EquinixPublisher` | `equinix:metal/device:Device` | plain `userData` | `operatingSystem: "ubuntu_22_04"` |
| `OutscalePublisher` | `outscale:index/vm:Vm` | plain `userData` | `imageId`, user supplied |
| `OpentelekomcloudPublisher` | `opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2` | plain `userData` | `imageName: "Ubuntu 22.04"` unless overridden |
| `TencentcloudPublisher` | `tencentcloud:index/instance:Instance` | plain `userDataRaw` | `imageId`, user supplied |
| `YandexPublisher` | `yandex:index/computeInstance:ComputeInstance` | metadata key `user-data` | `bootDisk.imageId`, user supplied |

## Task 1: Add User-Data Adapter Tests And TypeScript Helpers

**Files:**
- Create: `src/userDataAdapters.ts`
- Test: `test/userDataAdapters.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Write failing adapter tests**

Create `test/userDataAdapters.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import {
  base64UserData,
  guestInfoUserData,
  metadataUserData,
  plainUserData,
  scalewayUserData,
} from "../src/userDataAdapters";

async function outputValue<T>(value: pulumi.Output<T>): Promise<T> {
  return await new Promise<T>((resolve) => value.apply((resolved) => {
    resolve(resolved);
    return resolved;
  }));
}

test("plainUserData returns the payload unchanged", async () => {
  const result = plainUserData(pulumi.output("#cloud-config"));
  assert.equal(await outputValue(result), "#cloud-config");
});

test("base64UserData encodes the payload", async () => {
  const result = base64UserData(pulumi.output("#cloud-config"));
  assert.equal(await outputValue(result), Buffer.from("#cloud-config", "utf8").toString("base64"));
});

test("metadataUserData places payload under the requested key", async () => {
  const result = metadataUserData(pulumi.output("#cloud-config"), "user-data");
  assert.equal(await outputValue(result["user-data"] as pulumi.Output<string>), "#cloud-config");
});

test("guestInfoUserData emits base64 guestinfo keys", async () => {
  const result = guestInfoUserData(pulumi.output("#cloud-config"));
  assert.equal(await outputValue(result["guestinfo.userdata"] as pulumi.Output<string>), Buffer.from("#cloud-config", "utf8").toString("base64"));
  assert.equal(result["guestinfo.userdata.encoding"], "base64");
});

test("scalewayUserData emits both cloudInit and userData cloud-init keys", async () => {
  const result = scalewayUserData(pulumi.output("#cloud-config"));
  assert.equal(await outputValue(result.cloudInit as pulumi.Output<string>), "#cloud-config");
  const map = result.userData as Record<string, pulumi.Input<string>>;
  assert.equal(await outputValue(map["cloud-init"] as pulumi.Output<string>), "#cloud-config");
});
```

- [ ] **Step 2: Run the failing test**

Run: `npm test -- --test-name-pattern userDataAdapters`

Expected: FAIL because `src/userDataAdapters.ts` does not exist.

- [ ] **Step 3: Implement TypeScript adapter helpers**

Create `src/userDataAdapters.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";

export type UserDataMode =
  | "plain"
  | "base64"
  | "metadata-user-data"
  | "custom-data"
  | "cloud-init-disk"
  | "guestinfo"
  | "startup-script";

export interface PublisherUserDataAdapter {
  mode: UserDataMode;
  maxBytes?: number;
  render(payload: pulumi.Output<string>): pulumi.Input<string> | Record<string, pulumi.Input<string>>;
}

export function plainUserData(payload: pulumi.Output<string>): pulumi.Output<string> {
  return payload;
}

export function base64UserData(payload: pulumi.Output<string>): pulumi.Output<string> {
  return payload.apply((value) => Buffer.from(value, "utf8").toString("base64"));
}

export function metadataUserData(payload: pulumi.Output<string>, key = "user-data"): Record<string, pulumi.Input<string>> {
  return { [key]: payload };
}

export function base64MetadataUserData(payload: pulumi.Output<string>, key = "userData"): Record<string, pulumi.Input<string>> {
  return { [key]: base64UserData(payload) };
}

export function customData(payload: pulumi.Output<string>): pulumi.Output<string> {
  return base64UserData(payload);
}

export function guestInfoUserData(payload: pulumi.Output<string>): Record<string, pulumi.Input<string>> {
  return {
    "guestinfo.userdata": base64UserData(payload),
    "guestinfo.userdata.encoding": "base64",
  };
}

export function scalewayUserData(payload: pulumi.Output<string>): Record<string, pulumi.Input<unknown>> {
  return {
    cloudInit: payload,
    userData: {
      "cloud-init": payload,
    },
  };
}
```

Modify `src/index.ts`:

```ts
export * from "./userDataAdapters";
```

- [ ] **Step 4: Run adapter tests**

Run: `npm test -- --test-name-pattern userDataAdapters`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add src/userDataAdapters.ts src/index.ts test/userDataAdapters.test.ts
git commit -m "refactor: add publisher user-data adapters"
```

## Task 2: Add Raw Resource Helper For New TypeScript Providers

**Files:**
- Create: `src/rawResource.ts`
- Test: `test/rawResource.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Write failing raw resource test**

Create `test/rawResource.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "../src/rawResource";

test("RawResource registers the requested token and inputs", async () => {
  const seen: Array<{ type: string; inputs: pulumi.Inputs }> = [];
  pulumi.runtime.setMocks({
    newResource(args) {
      seen.push({ type: args.type, inputs: args.inputs });
      return { id: `${args.name}-id`, state: args.inputs };
    },
    call(args) {
      return args.inputs;
    },
  });

  new RawResource("example", "example:index/server:Server", { userData: "#cloud-config" });

  assert.equal(seen[0].type, "example:index/server:Server");
  assert.equal(seen[0].inputs.userData, "#cloud-config");
});
```

- [ ] **Step 2: Run the failing test**

Run: `npm test -- --test-name-pattern RawResource`

Expected: FAIL because `src/rawResource.ts` does not exist.

- [ ] **Step 3: Implement RawResource**

Create `src/rawResource.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";

export class RawResource extends pulumi.CustomResource {
  constructor(name: string, type: string, args: pulumi.Inputs, opts?: pulumi.CustomResourceOptions) {
    super(type, name, args, opts);
  }
}
```

Modify `src/index.ts`:

```ts
export * from "./rawResource";
```

- [ ] **Step 4: Run raw resource tests**

Run: `npm test -- --test-name-pattern RawResource`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add src/rawResource.ts src/index.ts test/rawResource.test.ts
git commit -m "feat: add raw provider resource helper"
```

## Task 3: Migrate Existing TypeScript Bootstrap Providers To Adapters

**Files:**
- Modify: `src/hcloudPublisher.ts`, `src/nutanixPublisher.ts`, `src/openstackPublisher.ts`, `src/ovhPublisher.ts`, `src/scalewayPublisher.ts`, `src/ociPublisher.ts`, `src/alicloudPublisher.ts`, `src/proxmoxvePublisher.ts`
- Test: existing `test/*Publisher.test.ts`

- [ ] **Step 1: Run existing provider tests before refactor**

Run: `npm test -- --test-name-pattern 'HcloudPublisher|NutanixPublisher|OpenstackPublisher|OvhPublisher|ScalewayPublisher|OciPublisher|AlicloudPublisher|ProxmoxvePublisher'`

Expected: PASS before the refactor starts.

- [ ] **Step 2: Replace direct base64/plain conversions with adapter helpers**

Apply these exact mappings:

| File | Replace placement with |
| --- | --- |
| `src/hcloudPublisher.ts` | `plainUserData(userData)` |
| `src/openstackPublisher.ts` | `plainUserData(userData)` |
| `src/ovhPublisher.ts` | `plainUserData(userData)` |
| `src/scalewayPublisher.ts` | `scalewayUserData(userData)` |
| `src/ociPublisher.ts` | `base64MetadataUserData(userData, "userData")` |
| `src/alicloudPublisher.ts` | `base64UserData(userData)` |
| `src/nutanixPublisher.ts` | `base64UserData(userData)` |
| `src/proxmoxvePublisher.ts` | `plainUserData(userData)` for snippet content |

Use imports from `./userDataAdapters`. Do not change public args or output names.

- [ ] **Step 3: Run migrated provider tests**

Run: `npm test -- --test-name-pattern 'HcloudPublisher|NutanixPublisher|OpenstackPublisher|OvhPublisher|ScalewayPublisher|OciPublisher|AlicloudPublisher|ProxmoxvePublisher'`

Expected: PASS with no schema changes.

- [ ] **Step 4: Commit**

```bash
git add src/hcloudPublisher.ts src/nutanixPublisher.ts src/openstackPublisher.ts src/ovhPublisher.ts src/scalewayPublisher.ts src/ociPublisher.ts src/alicloudPublisher.ts src/proxmoxvePublisher.ts
git commit -m "refactor: migrate bootstrap providers to user-data adapters"
```

## Task 4: Add TypeScript Args And Components For New Providers

**Files:**
- Modify: `src/types.ts`, `src/index.ts`
- Create: `src/digitaloceanPublisher.ts`, `src/vultrPublisher.ts`, `src/exoscalePublisher.ts`, `src/upcloudPublisher.ts`, `src/stackitPublisher.ts`, `src/equinixPublisher.ts`, `src/outscalePublisher.ts`, `src/opentelekomcloudPublisher.ts`, `src/tencentcloudPublisher.ts`, `src/yandexPublisher.ts`
- Test: create matching `test/*Publisher.test.ts`

- [ ] **Step 1: Add args interfaces to `src/types.ts`**

Add these interfaces after `ProxmoxvePublisherArgs`:

```ts
export interface DigitaloceanPublisherArgs extends CommonPublisherArgs {
  region: pulumi.Input<string>;
  size?: pulumi.Input<string>;
  image?: pulumi.Input<string>;
  sshKeys?: pulumi.Input<pulumi.Input<string>[]>;
  vpcUuid?: pulumi.Input<string>;
  monitoring?: pulumi.Input<boolean>;
  ipv6?: pulumi.Input<boolean>;
}

export interface VultrPublisherArgs extends CommonPublisherArgs {
  region: pulumi.Input<string>;
  plan: pulumi.Input<string>;
  osId?: pulumi.Input<number>;
  imageId?: pulumi.Input<string>;
  sshKeyIds?: pulumi.Input<pulumi.Input<string>[]>;
  vpc2Ids?: pulumi.Input<pulumi.Input<string>[]>;
  enableIpv6?: pulumi.Input<boolean>;
}

export interface ExoscalePublisherArgs extends CommonPublisherArgs {
  zone: pulumi.Input<string>;
  type: pulumi.Input<string>;
  templateId: pulumi.Input<string>;
  diskSize: pulumi.Input<number>;
  sshKeys?: pulumi.Input<pulumi.Input<string>[]>;
  securityGroupIds?: pulumi.Input<pulumi.Input<string>[]>;
  networkIds?: pulumi.Input<pulumi.Input<string>[]>;
}

export interface UpcloudPublisherArgs extends CommonPublisherArgs {
  zone: pulumi.Input<string>;
  plan?: pulumi.Input<string>;
  hostname?: pulumi.Input<string>;
  template?: pulumi.Input<string>;
  sshKeys?: pulumi.Input<pulumi.Input<string>[]>;
  networkInterfaces?: pulumi.Input<pulumi.Input<Record<string, pulumi.Input<unknown>>>[]>;
}

export interface StackitPublisherArgs extends CommonPublisherArgs {
  projectId: pulumi.Input<string>;
  machineType: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  availabilityZone?: pulumi.Input<string>;
  keypairName?: pulumi.Input<string>;
  networkInterfaces?: pulumi.Input<pulumi.Input<Record<string, pulumi.Input<unknown>>>[]>;
}

export interface EquinixPublisherArgs extends CommonPublisherArgs {
  projectId: pulumi.Input<string>;
  metro: pulumi.Input<string>;
  plan: pulumi.Input<string>;
  operatingSystem?: pulumi.Input<string>;
  billingCycle?: pulumi.Input<string>;
  projectSshKeyIds?: pulumi.Input<pulumi.Input<string>[]>;
  userSshKeyIds?: pulumi.Input<pulumi.Input<string>[]>;
}

export interface OutscalePublisherArgs extends CommonPublisherArgs {
  imageId: pulumi.Input<string>;
  vmType?: pulumi.Input<string>;
  subnetId?: pulumi.Input<string>;
  keypairName?: pulumi.Input<string>;
  securityGroupIds?: pulumi.Input<pulumi.Input<string>[]>;
  placementSubregionName?: pulumi.Input<string>;
}

export interface OpentelekomcloudPublisherArgs extends CommonPublisherArgs {
  imageName?: pulumi.Input<string>;
  imageId?: pulumi.Input<string>;
  flavorName?: pulumi.Input<string>;
  flavorId?: pulumi.Input<string>;
  networks: pulumi.Input<pulumi.Input<Record<string, pulumi.Input<unknown>>>[]>;
  keyPair?: pulumi.Input<string>;
  availabilityZone?: pulumi.Input<string>;
  securityGroups?: pulumi.Input<pulumi.Input<string>[]>;
}

export interface TencentcloudPublisherArgs extends CommonPublisherArgs {
  availabilityZone: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  instanceType?: pulumi.Input<string>;
  subnetId?: pulumi.Input<string>;
  vpcId?: pulumi.Input<string>;
  keyName?: pulumi.Input<string>;
  securityGroups?: pulumi.Input<pulumi.Input<string>[]>;
  systemDiskType?: pulumi.Input<string>;
  systemDiskSize?: pulumi.Input<number>;
}

export interface YandexPublisherArgs extends CommonPublisherArgs {
  zone?: pulumi.Input<string>;
  platformId?: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  subnetId: pulumi.Input<string>;
  cores?: pulumi.Input<number>;
  memory?: pulumi.Input<number>;
  coreFraction?: pulumi.Input<number>;
  nat?: pulumi.Input<boolean>;
  sshKeys?: pulumi.Input<pulumi.Input<string>[]>;
}
```

- [ ] **Step 2: Add component implementation pattern**

For each new component, use `createVmPublishers`, `RawResource`, and the adapter from the provider map. The DigitalOcean component is the reference implementation:

```ts
import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { createVmPublishers } from "./vmPublisherCore";
import { plainUserData } from "./userDataAdapters";
import { DigitaloceanPublisherArgs, PublisherOutput } from "./types";

export class DigitaloceanPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: DigitaloceanPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:DigitaloceanPublisher", name, {}, opts);

    const result = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userData }) => {
      const droplet = new RawResource(`${name}-${publisherName}`, "digitalocean:index/droplet:Droplet", {
        name: publisherName,
        region: args.region,
        size: args.size ?? "s-2vcpu-4gb",
        image: args.image ?? "ubuntu-22-04-x64",
        sshKeys: args.sshKeys,
        vpcUuid: args.vpcUuid,
        monitoring: args.monitoring,
        ipv6: args.ipv6,
        userData: plainUserData(userData),
        tags: args.tags ? pulumi.output(args.tags).apply((tags) => Object.entries(tags).map(([key, value]) => `${key}:${value}`)) : undefined,
      }, { parent: this });

      return {
        vmId: droplet.id,
        privateIp: pulumi.output(""),
      };
    });

    this.publisherNames = result.publisherNames;
    this.publishers = result.publishers;
    this.registerOutputs(result);
  }
}
```

Use the same structure for the other nine files with the token and fields from the Provider Tokens table.

- [ ] **Step 3: Export new components**

Modify `src/index.ts`:

```ts
export * from "./digitaloceanPublisher";
export * from "./vultrPublisher";
export * from "./exoscalePublisher";
export * from "./upcloudPublisher";
export * from "./stackitPublisher";
export * from "./equinixPublisher";
export * from "./outscalePublisher";
export * from "./opentelekomcloudPublisher";
export * from "./tencentcloudPublisher";
export * from "./yandexPublisher";
```

- [ ] **Step 4: Add TypeScript mock tests for each new provider**

Create one test file per provider. The DigitalOcean test is the reference:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { DigitaloceanPublisher } from "../src/digitaloceanPublisher";

test("DigitaloceanPublisher creates droplet with plain cloud-init userData", async () => {
  const seen: Array<{ type: string; inputs: pulumi.Inputs }> = [];
  pulumi.runtime.setMocks({
    newResource(args) {
      seen.push({ type: args.type, inputs: args.inputs });
      if (args.type === "netskope-publisher:index:NetskopeRegistration") {
        return { id: `${args.name}-id`, state: { registrations: { "pub-1": { publisherId: 123, registrationToken: "token", existedBefore: true } } } };
      }
      return { id: `${args.name}-id`, state: args.inputs };
    },
    call(args) {
      return args.inputs;
    },
  });

  new DigitaloceanPublisher("publisher", {
    names: ["pub-1"],
    registrations: {
      "pub-1": { publisherId: 123, registrationToken: "token" },
    },
    region: "ams3",
  });

  const droplet = seen.find((resource) => resource.type === "digitalocean:index/droplet:Droplet");
  assert.ok(droplet);
  assert.equal(droplet.inputs.image, "ubuntu-22-04-x64");
  assert.match(String(droplet.inputs.userData), /#cloud-config|Output/);
});
```

For providers using special placement, assert the exact field:

- `TencentcloudPublisher`: `userDataRaw`
- `YandexPublisher`: `metadata["user-data"]`
- all other new providers: `userData`

- [ ] **Step 5: Run new TypeScript provider tests**

Run: `npm test -- --test-name-pattern 'DigitaloceanPublisher|VultrPublisher|ExoscalePublisher|UpcloudPublisher|StackitPublisher|EquinixPublisher|OutscalePublisher|OpentelekomcloudPublisher|TencentcloudPublisher|YandexPublisher'`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add src/types.ts src/index.ts src/*Publisher.ts test/*Publisher.test.ts
git commit -m "feat: add bootstrap publishers for additional clouds"
```

## Task 5: Add Go User-Data Adapter Helpers

**Files:**
- Create: `internal/provider/userdata.go`
- Modify: `internal/provider/provider_test.go`

- [ ] **Step 1: Add failing Go adapter tests**

Append tests to `internal/provider/provider_test.go`:

```go
func TestUserDataAdaptersRenderPlacement(t *testing.T) {
	payload := pulumi.String("#cloud-config").ToStringOutput()

	if got := plainUserData(payload); got == nil {
		t.Fatalf("expected plain user data output")
	}

	if got := metadataUserData(payload, "user-data"); got["user-data"] == nil {
		t.Fatalf("expected metadata user-data key")
	}

	if got := guestInfoUserData(payload); got["guestinfo.userdata.encoding"] == nil || got["guestinfo.userdata"] == nil {
		t.Fatalf("expected guestinfo userdata and encoding keys")
	}

	if got := scalewayUserData(payload); got["cloudInit"] == nil || got["userData"] == nil {
		t.Fatalf("expected Scaleway cloudInit and userData keys")
	}
}
```

- [ ] **Step 2: Run failing Go test**

Run: `go test ./internal/provider -run TestUserDataAdaptersRenderPlacement`

Expected: FAIL because helper functions are not defined.

- [ ] **Step 3: Implement Go helpers**

Create `internal/provider/userdata.go`:

```go
package provider

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

func plainUserData(payload pulumi.StringOutput) pulumi.StringOutput {
	return payload
}

func base64UserData(payload pulumi.StringOutput) pulumi.StringOutput {
	return payload.ApplyT(func(value string) string {
		return encodeBase64String(value)
	}).(pulumi.StringOutput)
}

func metadataUserData(payload pulumi.StringOutput, key string) pulumi.Map {
	return pulumi.Map{key: payload}
}

func base64MetadataUserData(payload pulumi.StringOutput, key string) pulumi.Map {
	return pulumi.Map{key: base64UserData(payload)}
}

func guestInfoUserData(payload pulumi.StringOutput) pulumi.Map {
	return pulumi.Map{
		"guestinfo.userdata":          base64UserData(payload),
		"guestinfo.userdata.encoding": pulumi.String("base64"),
	}
}

func scalewayUserData(payload pulumi.StringOutput) pulumi.Map {
	return pulumi.Map{
		"cloudInit": payload,
		"userData": pulumi.Map{"cloud-init": payload},
	}
}
```

If `encodeBase64String` does not already exist, add this function to `internal/provider/userdata.go`:

```go
func encodeBase64String(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}
```

and import `encoding/base64`.

- [ ] **Step 4: Run Go adapter tests**

Run: `go test ./internal/provider -run TestUserDataAdaptersRenderPlacement`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/provider/userdata.go internal/provider/provider_test.go
git commit -m "refactor: add go user-data adapter helpers"
```

## Task 6: Migrate Existing Go Bootstrap Providers To Adapters

**Files:**
- Modify: `internal/provider/components.go`
- Test: `internal/provider/provider_test.go`

- [ ] **Step 1: Run existing Go provider placement tests**

Run: `go test ./internal/provider -run 'TestAdditionalProviderConstructsBootstrapWithRegistryFields|TestProxmoxveConstructCreatesSnippetBackedVmClone'`

Expected: PASS before migration.

- [ ] **Step 2: Replace inline placement logic with adapter helpers**

In `internal/provider/components.go`, use:

- `plainUserData(rendered)` for HCloud, OpenStack, OVH, and Proxmox snippet content
- `base64UserData(rendered)` for Alicloud and Nutanix
- `base64MetadataUserData(rendered, "userData")` for OCI
- `scalewayUserData(rendered)` for Scaleway
- `guestInfoUserData(rendered)` for GuestInfo-style resources that remain bootstrap-based

Use `renderUserDataOutputWithOptions(...)` as the canonical rendered plain payload, then adapt it. Avoid decoding base64 just to re-encode it.

- [ ] **Step 3: Run migrated Go tests**

Run: `go test ./internal/provider -run 'TestAdditionalProviderConstructsBootstrapWithRegistryFields|TestProxmoxveConstructCreatesSnippetBackedVmClone'`

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/provider/components.go
git commit -m "refactor: migrate go bootstrap providers to adapters"
```

## Task 7: Add Go Components For New Providers

**Files:**
- Modify: `internal/provider/components.go`
- Modify: `internal/provider/provider.go`
- Modify: `internal/provider/provider_test.go`

- [ ] **Step 1: Add failing constructor type tests**

Extend `TestAdditionalProviderConstructsCreateProviderChildren` with these cases:

```go
{
	name:  "DigitalOcean",
	token: "netskope-publisher:index:DigitaloceanPublisher",
	inputs: property.NewMap(map[string]property.Value{
		"names": property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"region": property.New("ams3"),
	}),
	expected: "digitalocean:index/droplet:Droplet",
},
{
	name:  "Vultr",
	token: "netskope-publisher:index:VultrPublisher",
	inputs: property.NewMap(map[string]property.Value{
		"names": property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"region": property.New("ams"),
		"plan": property.New("vc2-2c-4gb"),
		"osId": property.New(1743.0),
	}),
	expected: "vultr:index/instance:Instance",
},
{
	name:  "Exoscale",
	token: "netskope-publisher:index:ExoscalePublisher",
	inputs: property.NewMap(map[string]property.Value{
		"names": property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"zone": property.New("ch-gva-2"),
		"type": property.New("standard.medium"),
		"templateId": property.New("template-id"),
		"diskSize": property.New(50.0),
	}),
	expected: "exoscale:index/computeInstance:ComputeInstance",
},
{
	name:  "UpCloud",
	token: "netskope-publisher:index:UpcloudPublisher",
	inputs: property.NewMap(map[string]property.Value{
		"names": property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"zone": property.New("nl-ams1"),
	}),
	expected: "upcloud:index/server:Server",
},
{
	name:  "Stackit",
	token: "netskope-publisher:index:StackitPublisher",
	inputs: property.NewMap(map[string]property.Value{
		"names": property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"projectId": property.New("project-id"),
		"machineType": property.New("g1.2"),
		"imageId": property.New("image-id"),
	}),
	expected: "stackit:index/server:Server",
},
{
	name:  "Equinix",
	token: "netskope-publisher:index:EquinixPublisher",
	inputs: property.NewMap(map[string]property.Value{
		"names": property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"projectId": property.New("project-id"),
		"metro": property.New("AM"),
		"plan": property.New("c3.small.x86"),
	}),
	expected: "equinix:metal/device:Device",
},
{
	name:  "Outscale",
	token: "netskope-publisher:index:OutscalePublisher",
	inputs: property.NewMap(map[string]property.Value{
		"names": property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"imageId": property.New("ami-123"),
	}),
	expected: "outscale:index/vm:Vm",
},
{
	name:  "OpenTelekomCloud",
	token: "netskope-publisher:index:OpentelekomcloudPublisher",
	inputs: property.NewMap(map[string]property.Value{
		"names": property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"networks": property.New([]property.Value{property.New(map[string]property.Value{"name": property.New("private")})}),
	}),
	expected: "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2",
},
{
	name:  "TencentCloud",
	token: "netskope-publisher:index:TencentcloudPublisher",
	inputs: property.NewMap(map[string]property.Value{
		"names": property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"availabilityZone": property.New("ap-guangzhou-6"),
		"imageId": property.New("img-123"),
	}),
	expected: "tencentcloud:index/instance:Instance",
},
{
	name:  "Yandex",
	token: "netskope-publisher:index:YandexPublisher",
	inputs: property.NewMap(map[string]property.Value{
		"names": property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"imageId": property.New("image-id"),
		"subnetId": property.New("subnet-id"),
	}),
	expected: "yandex:index/computeInstance:ComputeInstance",
},
```

- [ ] **Step 2: Run failing constructor tests**

Run: `go test ./internal/provider -run TestAdditionalProviderConstructsCreateProviderChildren`

Expected: FAIL because the new component tokens are not registered.

- [ ] **Step 3: Add Go args structs and constructors**

In `internal/provider/components.go`, add args structs matching the TypeScript interfaces from Task 4. Each struct embeds the existing common publisher fields via the same pattern used by `HcloudPublisherArgs`.

For each constructor:

- resolve names and registrations using the existing component helpers
- render plain payload with `renderUserDataOutputWithOptions(..., cloudInitOptionsFromCommon(args.common(), true))`
- place payload with the adapter from the Provider Tokens table
- create a `rawVMResource` with the exact resource token
- return publisher outputs with `vmId` from the raw resource ID and blank IPs where the provider schema does not expose predictable output names in mocks

- [ ] **Step 4: Register components**

In `internal/provider/provider.go`, add:

```go
infer.ComponentF(NewDigitaloceanPublisher),
infer.ComponentF(NewVultrPublisher),
infer.ComponentF(NewExoscalePublisher),
infer.ComponentF(NewUpcloudPublisher),
infer.ComponentF(NewStackitPublisher),
infer.ComponentF(NewEquinixPublisher),
infer.ComponentF(NewOutscalePublisher),
infer.ComponentF(NewOpentelekomcloudPublisher),
infer.ComponentF(NewTencentcloudPublisher),
infer.ComponentF(NewYandexPublisher),
```

Update the provider description to include the new providers.

- [ ] **Step 5: Add Go bootstrap placement assertions**

Extend `TestAdditionalProviderConstructsBootstrapWithRegistryFields` with one case per new provider. Assert:

- plain providers: `assertBootstrapUserData(t, inputs.Get("userData").AsString())`
- TencentCloud: `assertBootstrapUserData(t, inputs.Get("userDataRaw").AsString())`
- Yandex: `assertBootstrapUserData(t, inputs.Get("metadata").AsMap().Get("user-data").AsString())`

- [ ] **Step 6: Run Go provider tests**

Run: `go test ./internal/provider -run 'TestAdditionalProviderConstructsCreateProviderChildren|TestAdditionalProviderConstructsBootstrapWithRegistryFields'`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/provider/components.go internal/provider/provider.go internal/provider/provider_test.go
git commit -m "feat: add go components for expanded clouds"
```

## Task 8: Regenerate Schema And SDKs

**Files:**
- Regenerate: `schema.json`, `sdk/python`, `sdk/dotnet`, `sdk/go`, `sdk/java`, `sdk/rust`

- [ ] **Step 1: Run schema and SDK generation**

Run: `npm run sdk:gen`

Expected: PASS and generated SDKs include the 10 new component classes/types.

- [ ] **Step 2: Verify schema contains the new components**

Run:

```bash
node - <<'NODE'
const schema = JSON.parse(require('fs').readFileSync('schema.json', 'utf8'));
for (const name of ['DigitaloceanPublisher','VultrPublisher','ExoscalePublisher','UpcloudPublisher','StackitPublisher','EquinixPublisher','OutscalePublisher','OpentelekomcloudPublisher','TencentcloudPublisher','YandexPublisher']) {
  const token = `netskope-publisher:index:${name}`;
  if (!schema.resources[token]) throw new Error(`missing ${token}`);
}
console.log('schema contains expanded cloud providers');
NODE
```

Expected: prints `schema contains expanded cloud providers`.

- [ ] **Step 3: Build generated Java and Rust SDKs**

Run: `cd sdk/java && gradle build`

Expected: PASS.

Run: `cd sdk/rust && cargo check`

Expected: PASS.

- [ ] **Step 4: Commit generated outputs**

```bash
git add schema.json sdk/python sdk/dotnet sdk/go sdk/java sdk/rust
git commit -m "build: regenerate sdks for expanded clouds"
```

## Task 9: Update README And GitHub Pages Docs

**Files:**
- Modify: `README.md`
- Modify: `docs/_index.md`, `docs/installation-configuration.md`
- Modify: `site/source/admin/component/index.md`
- Create: one `site/source/admin/component/*.md` page for each new provider
- Modify: `site/source/reference/provider-matrix.md`, `site/source/reference/roadmap.md`, `site/source/reference/sdk-installation.md`

- [ ] **Step 1: Update provider lists and matrix**

Add the 10 new providers to README, docs, and site provider matrix. Mark:

- bootstrap image model: Ubuntu 22.04
- enrollment: Netskope registration token generated at deployment time
- user-data mode: plain, `userDataRaw`, or metadata `user-data`
- status: supported

- [ ] **Step 2: Add component docs**

Create pages for:

- `site/source/admin/component/digitalocean.md`
- `site/source/admin/component/vultr.md`
- `site/source/admin/component/exoscale.md`
- `site/source/admin/component/upcloud.md`
- `site/source/admin/component/stackit.md`
- `site/source/admin/component/equinix.md`
- `site/source/admin/component/outscale.md`
- `site/source/admin/component/opentelekomcloud.md`
- `site/source/admin/component/tencentcloud.md`
- `site/source/admin/component/yandex.md`

Each page must include:

- purpose
- minimum required inputs
- Ubuntu 22.04 image/template guidance
- networking notes
- token and OAuth2 enrollment examples
- Pulumi CLI config example
- TypeScript, Python, C#, Go, Java, and Rust examples for GitHub Pages docs

- [ ] **Step 3: Keep Pulumi Registry docs language-safe**

In `docs/_index.md` and `docs/installation-configuration.md`, include TypeScript, Python, Go, C#, and Java. Do not list Rust as an official Pulumi Registry language.

- [ ] **Step 4: Build the site**

Run: `cd site && npm install && npm run build`

Expected: PASS and `site/public` is regenerated.

- [ ] **Step 5: Commit docs**

```bash
git add README.md docs site/source site/public
git commit -m "docs: document expanded cloud provider support"
```

## Task 10: Full Verification And Release Readiness

**Files:**
- Potentially modify generated files only if verification exposes deterministic generator drift.

- [ ] **Step 1: Run TypeScript checks**

Run: `npm run typecheck`

Expected: PASS.

Run: `npm test`

Expected: PASS.

- [ ] **Step 2: Run Go tests**

Run: `npm run go:test`

Expected: PASS.

- [ ] **Step 3: Run registry and packaging checks**

Run: `npm run registry:check`

Expected: PASS.

Run: `npm run sdk:pack`

Expected: PASS.

- [ ] **Step 4: Run full release check**

Run: `npm run release:check`

Expected: PASS.

- [ ] **Step 5: Review final diff**

Run: `git status --short`

Expected: clean or only intended generated packaging artifacts.

Run: `git log --oneline -8`

Expected: shows the task commits in order.

- [ ] **Step 6: Final commit if verification changed generated files**

If verification changed generated files, commit them:

```bash
git add schema.json sdk docs site/public package-lock.json
git commit -m "build: refresh generated provider artifacts"
```

If no files changed, do not create an empty commit.

## Self-Review

- Spec coverage: provider expansion, existing migration, SDK regeneration, docs, and follow-up framework are all represented.
- Placeholder scan: no open-ended implementation placeholders remain; provider-specific tokens, fields, test commands, and doc files are listed.
- Type consistency: TypeScript and Go component names match the approved spec and schema token naming style.
