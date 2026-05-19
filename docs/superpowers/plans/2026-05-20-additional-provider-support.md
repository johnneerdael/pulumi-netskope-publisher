# Additional Provider Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add ESXi Native plus bootstrap-only Hcloud, Nutanix, OpenStack, OVH, Scaleway, OCI, and Alicloud publisher components with tests, schema/SDK generation, and GitHub Pages documentation.

**Architecture:** Keep one public component per provider. Extract the repeated VM-backed publisher flow into shared TypeScript and Go helpers so provider files only translate component inputs into provider-native VM resources and output fields. Keep `VspherePublisher` unchanged and add `EsxiPublisher` as the direct-host alternative.

**Tech Stack:** TypeScript Pulumi components, Pulumi Go Provider executable components, provider SDKs from `@pulumiverse/esxi-native`, `@pulumi/hcloud`, `@pierskarsenbarg/nutanix`, `@pulumi/openstack`, `@ovhcloud/pulumi-ovh`, `@pulumiverse/scaleway`, `@pulumi/oci`, `@pulumi/alicloud`, Node test mocks, generated Pulumi SDKs including Java and Rust.

---

## File Structure

- Modify `package.json`: add the eight TypeScript provider SDK dependencies.
- Modify `src/types.ts`: add input interfaces for each new component.
- Modify `src/index.ts`: export each new component.
- Create `src/vmPublisherCore.ts`: shared TypeScript helper for name resolution, registration, bootstrap cloud-init rendering, and secret output map creation.
- Create `src/esxiPublisher.ts`, `src/hcloudPublisher.ts`, `src/nutanixPublisher.ts`, `src/openstackPublisher.ts`, `src/ovhPublisher.ts`, `src/scalewayPublisher.ts`, `src/ociPublisher.ts`, `src/alicloudPublisher.ts`.
- Create `test/*Publisher.test.ts` for each new component.
- Modify `internal/provider/types.go`, `internal/provider/components.go`, `internal/provider/provider.go`: add Go executable provider components and schema tokens.
- Modify `scripts/check-registry-readiness.mjs`: include the new component tokens and files.
- Modify docs under `site/source/admin/component/`, `site/source/reference/provider-matrix.md`, `site/source/admin/index.md`, `site/source/admin/component/index.md`, `README.md`, and `docs/_index.md`.
- Regenerate `schema.json` and SDKs after Go provider changes.

Provider component input contracts:

```ts
export interface EsxiPublisherArgs extends CommonPublisherArgs {
  diskStore: pulumi.Input<string>;
  virtualNetwork: pulumi.Input<string>;
  os?: pulumi.Input<string>;
  memory?: pulumi.Input<number>;
  numVCpus?: pulumi.Input<number>;
  diskSize?: pulumi.Input<number>;
}

export interface HcloudPublisherArgs extends CommonPublisherArgs {
  serverType?: pulumi.Input<string>;
  image?: pulumi.Input<string>;
  location?: pulumi.Input<string>;
  datacenter?: pulumi.Input<string>;
  sshKeys?: pulumi.Input<pulumi.Input<string>[]>;
  firewallIds?: pulumi.Input<pulumi.Input<number>[]>;
  networkId?: pulumi.Input<number>;
  assignPublicIp?: pulumi.Input<boolean>;
}

export interface NutanixPublisherArgs extends CommonPublisherArgs {
  imageName: pulumi.Input<string>;
  subnetName: pulumi.Input<string>;
  clusterName?: pulumi.Input<string>;
  numVCpus?: pulumi.Input<number>;
  numCoresPerVcpu?: pulumi.Input<number>;
  memorySizeMib?: pulumi.Input<number>;
}

export interface OpenstackPublisherArgs extends CommonPublisherArgs {
  imageName: pulumi.Input<string>;
  flavorName: pulumi.Input<string>;
  networkName: pulumi.Input<string>;
  keyPair?: pulumi.Input<string>;
  securityGroups?: pulumi.Input<pulumi.Input<string>[]>;
  availabilityZone?: pulumi.Input<string>;
  assignFloatingIp?: pulumi.Input<boolean>;
  floatingIpPool?: pulumi.Input<string>;
}

export interface OvhPublisherArgs extends CommonPublisherArgs {
  serviceName: pulumi.Input<string>;
  region: pulumi.Input<string>;
  imageName: pulumi.Input<string>;
  flavorName: pulumi.Input<string>;
  sshKeyName?: pulumi.Input<string>;
  networkId?: pulumi.Input<string>;
}

export interface ScalewayPublisherArgs extends CommonPublisherArgs {
  type?: pulumi.Input<string>;
  image?: pulumi.Input<string>;
  zone?: pulumi.Input<string>;
  securityGroupId?: pulumi.Input<string>;
  enableDynamicIp?: pulumi.Input<boolean>;
}

export interface OciPublisherArgs extends CommonPublisherArgs {
  compartmentId: pulumi.Input<string>;
  availabilityDomain: pulumi.Input<string>;
  shape?: pulumi.Input<string>;
  subnetId: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  sshPublicKey?: pulumi.Input<string>;
  assignPublicIp?: pulumi.Input<boolean>;
}

export interface AlicloudPublisherArgs extends CommonPublisherArgs {
  instanceType?: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  vswitchId: pulumi.Input<string>;
  securityGroupIds: pulumi.Input<pulumi.Input<string>[]>;
  keyName?: pulumi.Input<string>;
  allocatePublicIp?: pulumi.Input<boolean>;
}
```

---

### Task 1: Install Provider SDK Dependencies

**Files:**
- Modify: `package.json`
- Modify: `package-lock.json`

- [ ] **Step 1: Add dependencies**

Run:

```bash
npm install \
  @pulumiverse/esxi-native \
  @pulumi/hcloud \
  @pierskarsenbarg/nutanix \
  @pulumi/openstack \
  @ovhcloud/pulumi-ovh \
  @pulumiverse/scaleway \
  @pulumi/oci \
  @pulumi/alicloud
```

Expected: `package.json` and `package-lock.json` include the eight new packages.

- [ ] **Step 2: Verify package install did not alter unrelated files**

Run:

```bash
git diff -- package.json package-lock.json
```

Expected: only dependency additions and lockfile dependency graph changes.

- [ ] **Step 3: Commit**

```bash
git add package.json package-lock.json
git commit -m "build: add additional provider SDK dependencies"
```

---

### Task 2: Add Shared TypeScript VM Publisher Helper

**Files:**
- Create: `src/vmPublisherCore.ts`
- Test indirectly through Task 3 and Task 4 component tests.

- [ ] **Step 1: Create helper**

Create `src/vmPublisherCore.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import { renderUserDataBase64 } from "./cloudInit";
import { createPublisherOutput, createRegistrations, resolvePublisherNames } from "./componentCore";
import { CommonPublisherArgs, PublisherOutput } from "./types";

export interface VmPublisherRuntime {
  parent: pulumi.ComponentResource;
  componentName: string;
  args: CommonPublisherArgs;
  forceBootstrap?: boolean;
  defaultNonat?: boolean;
}

export interface VmPublisherBuildInput {
  publisherName: string;
  registration: pulumi.Output<{
    publisherId: number;
    registrationToken: string;
    existedBefore?: boolean;
  }>;
  userDataBase64: pulumi.Output<string>;
}

export interface VmPublisherBuildResult {
  vmId: pulumi.Input<string>;
  privateIp: pulumi.Input<string>;
  publicIp?: pulumi.Input<string>;
}

export function createVmPublishers(
  runtime: VmPublisherRuntime,
  build: (input: VmPublisherBuildInput) => VmPublisherBuildResult,
): {
  publisherNames: pulumi.Output<string[]>;
  publishers: pulumi.Output<Record<string, PublisherOutput>>;
} {
  const parentOpts = { parent: runtime.parent };
  const publisherNames = resolvePublisherNames(runtime.args);
  const registrations = createRegistrations(runtime.componentName, publisherNames, runtime.args, parentOpts);
  const publisherOutputs: Record<string, pulumi.Output<PublisherOutput>> = {};

  for (const publisherName of publisherNames) {
    const registration = registrations.apply((allRegistrations) => allRegistrations[publisherName]);
    const userDataBase64 = pulumi.all({
      registration,
      wizardPath: runtime.args.wizardPath,
      bootstrap: runtime.forceBootstrap ? true : runtime.args.bootstrap,
      bootstrapUrl: runtime.args.bootstrapUrl,
      nonat: runtime.args.nonat ?? runtime.defaultNonat ?? false,
      installUser: runtime.args.installUser,
      installUserPassword: runtime.args.installUserPassword,
      installUserPasswordIsHash: runtime.args.installUserPasswordIsHash,
      installUserSshAuthorizedKeys: runtime.args.installUserSshAuthorizedKeys,
      deleteDefaultUser: runtime.args.deleteDefaultUser,
      guestNetworkInterface: runtime.args.guestNetworkInterface,
    }).apply((options: any) =>
      renderUserDataBase64({
        publisherName,
        registrationToken: options.registration.registrationToken,
        wizardPath: options.wizardPath,
        bootstrap: options.bootstrap,
        bootstrapUrl: options.bootstrapUrl,
        nonat: options.nonat,
        installUser: options.installUser,
        installUserPassword: options.installUserPassword,
        installUserPasswordIsHash: options.installUserPasswordIsHash,
        installUserSshAuthorizedKeys: options.installUserSshAuthorizedKeys,
        deleteDefaultUser: options.deleteDefaultUser,
        guestNetworkInterface: options.guestNetworkInterface,
      }),
    );

    const result = build({ publisherName, registration, userDataBase64 });
    publisherOutputs[publisherName] = createPublisherOutput({
      registration,
      vmId: result.vmId,
      privateIp: result.privateIp,
      publicIp: result.publicIp,
    });
  }

  return {
    publisherNames: pulumi.output(publisherNames),
    publishers: pulumi.secret(pulumi.all(publisherOutputs)),
  };
}
```

- [ ] **Step 2: Run typecheck**

Run:

```bash
npm run typecheck
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add src/vmPublisherCore.ts
git commit -m "refactor: add shared VM publisher helper"
```

---

### Task 3: Add Hcloud and Scaleway TypeScript Components

**Files:**
- Modify: `src/types.ts`
- Modify: `src/index.ts`
- Create: `src/hcloudPublisher.ts`
- Create: `src/scalewayPublisher.ts`
- Create: `test/hcloudPublisher.test.ts`
- Create: `test/scalewayPublisher.test.ts`

- [ ] **Step 1: Add failing Hcloud test**

Create `test/hcloudPublisher.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { HcloudPublisher } from "../src/hcloudPublisher";
import { PublisherOutput } from "../src/types";

const createdServers: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "hcloud:index/server:Server") {
      createdServers[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          ipv4Address: "198.51.100.20",
          privateNet: [{ ip: "10.0.0.20" }],
        },
      };
    }
    if (args.type === "pulumi-nodejs:dynamic:Resource") {
      return {
        id: "pub-1",
        state: {
          ...args.inputs,
          registrations: {
            "pub-1": { publisherId: 101, registrationToken: "token-101", existedBefore: true },
          },
        },
      };
    }
    return { id: `${args.name}-id`, state: args.inputs };
  },
  call(args) {
    return args.inputs;
  },
});

test("HcloudPublisher creates a bootstrap server", async () => {
  const component = new HcloudPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    networkId: 123,
  });

  const publisherNames = await outputValue(component.publisherNames);
  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);

  assert.deepEqual(publisherNames, ["pub-1"]);
  assert.equal(publishers["pub-1"].vmId, "publisher-pub-1-id");
  assert.equal(createdServers["publisher-pub-1"].image, "ubuntu-22.04");
  assert.match(Buffer.from(createdServers["publisher-pub-1"].userData, "base64").toString("utf8"), /bootstrap\.sh/);
});

async function outputValue<T>(output: pulumi.Output<T>): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    output.apply((value) => {
      resolve(value);
      return value;
    });
    setTimeout(() => reject(new Error("Timed out waiting for Pulumi output")), 5000);
  });
}
```

- [ ] **Step 2: Add failing Scaleway test**

Create `test/scalewayPublisher.test.ts` with this complete content:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { ScalewayPublisher } from "../src/scalewayPublisher";
import { PublisherOutput } from "../src/types";

const createdServers: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "scaleway:index/instanceServer:InstanceServer") {
      createdServers[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          publicIp: "198.51.100.30",
          privateIp: "10.0.0.30",
        },
      };
    }
    if (args.type === "pulumi-nodejs:dynamic:Resource") {
      return {
        id: "pub-1",
        state: {
          ...args.inputs,
          registrations: {
            "pub-1": { publisherId: 101, registrationToken: "token-101", existedBefore: true },
          },
        },
      };
    }
    return { id: `${args.name}-id`, state: args.inputs };
  },
  call(args) {
    return args.inputs;
  },
});

test("ScalewayPublisher creates a bootstrap server", async () => {
  const component = new ScalewayPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
  });

  const publisherNames = await outputValue(component.publisherNames);
  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);

  assert.deepEqual(publisherNames, ["pub-1"]);
  assert.equal(publishers["pub-1"].vmId, "publisher-pub-1-id");
  assert.equal(createdServers["publisher-pub-1"].image, "ubuntu_jammy");
  assert.match(Buffer.from(createdServers["publisher-pub-1"].userData, "base64").toString("utf8"), /bootstrap\.sh/);
});

async function outputValue<T>(output: pulumi.Output<T>): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    output.apply((value) => {
      resolve(value);
      return value;
    });
    setTimeout(() => reject(new Error("Timed out waiting for Pulumi output")), 5000);
  });
}
```

- [ ] **Step 3: Run tests and confirm failure**

Run:

```bash
npm test -- --test-name-pattern 'HcloudPublisher|ScalewayPublisher'
```

Expected: FAIL because `src/hcloudPublisher.ts` and `src/scalewayPublisher.ts` do not exist.

- [ ] **Step 4: Add type definitions and exports**

Add the `HcloudPublisherArgs` and `ScalewayPublisherArgs` interfaces from the File Structure section to `src/types.ts`. Add exports to `src/index.ts`:

```ts
export * from "./hcloudPublisher";
export * from "./scalewayPublisher";
```

- [ ] **Step 5: Implement HcloudPublisher**

Create `src/hcloudPublisher.ts`:

```ts
import * as hcloud from "@pulumi/hcloud";
import * as pulumi from "@pulumi/pulumi";
import { createVmPublishers } from "./vmPublisherCore";
import { HcloudPublisherArgs, PublisherOutput } from "./types";

export class HcloudPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: HcloudPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:HcloudPublisher", name, {}, opts);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userDataBase64 }) => {
      const server = new hcloud.Server(`${name}-${publisherName}`, {
        name: publisherName,
        serverType: args.serverType ?? "cx22",
        image: args.image ?? "ubuntu-22.04",
        location: args.location,
        datacenter: args.datacenter,
        sshKeys: args.sshKeys,
        firewallIds: args.firewallIds,
        userData: userDataBase64,
        publicNets: {
          ipv4Enabled: args.assignPublicIp ?? true,
          ipv6Enabled: false,
        },
        networks: args.networkId ? [{ networkId: args.networkId }] : undefined,
        labels: args.tags,
      }, { parent: this });

      return {
        vmId: server.id,
        privateIp: pulumi.output((server as any).privateNet).apply((nets: any[] | undefined) => nets?.[0]?.ip ?? ""),
        publicIp: (server as any).ipv4Address,
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
```

- [ ] **Step 6: Implement ScalewayPublisher**

Create `src/scalewayPublisher.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import * as scaleway from "@pulumiverse/scaleway";
import { ScalewayPublisherArgs, PublisherOutput } from "./types";
import { createVmPublishers } from "./vmPublisherCore";

export class ScalewayPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: ScalewayPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:ScalewayPublisher", name, {}, opts);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userDataBase64 }) => {
      const server = new scaleway.InstanceServer(`${name}-${publisherName}`, {
        name: publisherName,
        type: args.type ?? "DEV1-S",
        image: args.image ?? "ubuntu_jammy",
        zone: args.zone,
        securityGroupId: args.securityGroupId,
        enableDynamicIp: args.enableDynamicIp ?? true,
        userData: {
          "cloud-init": userDataBase64.apply((value) => Buffer.from(value, "base64").toString("utf8")),
        },
        tags: args.tags ? pulumi.output(args.tags).apply((tags) => Object.entries(tags).map(([key, value]) => `${key}:${value}`)) : undefined,
      }, { parent: this });

      return {
        vmId: server.id,
        privateIp: (server as any).privateIp ?? "",
        publicIp: (server as any).publicIp,
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
```

- [ ] **Step 7: Run tests**

Run:

```bash
npm test -- --test-name-pattern 'HcloudPublisher|ScalewayPublisher'
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add src/types.ts src/index.ts src/vmPublisherCore.ts src/hcloudPublisher.ts src/scalewayPublisher.ts test/hcloudPublisher.test.ts test/scalewayPublisher.test.ts
git commit -m "feat: add Hcloud and Scaleway publishers"
```

---

### Task 4: Add OCI and Alicloud TypeScript Components

**Files:**
- Modify: `src/types.ts`
- Modify: `src/index.ts`
- Create: `src/ociPublisher.ts`
- Create: `src/alicloudPublisher.ts`
- Create: `test/ociPublisher.test.ts`
- Create: `test/alicloudPublisher.test.ts`

- [ ] **Step 1: Add component input interfaces**

Add `OciPublisherArgs` and `AlicloudPublisherArgs` from the File Structure section to `src/types.ts`. Add exports to `src/index.ts`:

```ts
export * from "./ociPublisher";
export * from "./alicloudPublisher";
```

- [ ] **Step 2: Write OCI failing test**

Create `test/ociPublisher.test.ts` with a mock for `oci:Core/instance:Instance`, asserting:

```ts
assert.equal(createdInstances["publisher-pub-1"].sourceDetails.sourceId, "ocid1.image.oc1..ubuntu");
assert.match(Buffer.from(createdInstances["publisher-pub-1"].metadata.user_data, "base64").toString("utf8"), /bootstrap\.sh/);
```

The full test file must include the `pulumi-nodejs:dynamic:Resource`
registration mock and `outputValue` helper used in `test/awsPublisher.test.ts`,
copied into `test/ociPublisher.test.ts` so the file runs independently.

- [ ] **Step 3: Implement OciPublisher**

Create `src/ociPublisher.ts`:

```ts
import * as oci from "@pulumi/oci";
import * as pulumi from "@pulumi/pulumi";
import { createVmPublishers } from "./vmPublisherCore";
import { OciPublisherArgs, PublisherOutput } from "./types";

export class OciPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: OciPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:OciPublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userDataBase64 }) => {
      const instance = new oci.core.Instance(`${name}-${publisherName}`, {
        displayName: publisherName,
        compartmentId: args.compartmentId,
        availabilityDomain: args.availabilityDomain,
        shape: args.shape ?? "VM.Standard.E4.Flex",
        createVnicDetails: {
          subnetId: args.subnetId,
          assignPublicIp: args.assignPublicIp ?? false,
          displayName: `${publisherName}-vnic`,
        },
        sourceDetails: {
          sourceType: "image",
          sourceId: args.imageId,
        },
        metadata: {
          user_data: userDataBase64,
          ssh_authorized_keys: args.sshPublicKey,
        },
        freeformTags: args.tags,
      }, { parent: this });

      return {
        vmId: instance.id,
        privateIp: (instance as any).privateIp ?? "",
        publicIp: (instance as any).publicIp,
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
```

- [ ] **Step 4: Write Alicloud failing test**

Create `test/alicloudPublisher.test.ts` with a mock for `alicloud:ecs/instance:Instance`, asserting `imageId`, `userData`, and output map behavior.

- [ ] **Step 5: Implement AlicloudPublisher**

Create `src/alicloudPublisher.ts`:

```ts
import * as alicloud from "@pulumi/alicloud";
import * as pulumi from "@pulumi/pulumi";
import { AlicloudPublisherArgs, PublisherOutput } from "./types";
import { createVmPublishers } from "./vmPublisherCore";

export class AlicloudPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: AlicloudPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:AlicloudPublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userDataBase64 }) => {
      const instance = new alicloud.ecs.Instance(`${name}-${publisherName}`, {
        instanceName: publisherName,
        instanceType: args.instanceType ?? "ecs.t6-c1m2.large",
        imageId: args.imageId,
        vswitchId: args.vswitchId,
        securityGroups: args.securityGroupIds,
        keyName: args.keyName,
        internetMaxBandwidthOut: args.allocatePublicIp === true ? 10 : 0,
        userData: userDataBase64.apply((value) => Buffer.from(value, "base64").toString("utf8")),
        tags: args.tags,
      }, { parent: this });

      return {
        vmId: instance.id,
        privateIp: (instance as any).privateIp,
        publicIp: (instance as any).publicIp,
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
```

- [ ] **Step 6: Run tests and commit**

Run:

```bash
npm test -- --test-name-pattern 'OciPublisher|AlicloudPublisher'
```

Expected: PASS.

Commit:

```bash
git add src/types.ts src/index.ts src/ociPublisher.ts src/alicloudPublisher.ts test/ociPublisher.test.ts test/alicloudPublisher.test.ts
git commit -m "feat: add OCI and Alicloud publishers"
```

---

### Task 5: Add ESXi, Nutanix, OpenStack, and OVH TypeScript Components

**Files:**
- Modify: `src/types.ts`
- Modify: `src/index.ts`
- Create: `src/esxiPublisher.ts`
- Create: `src/nutanixPublisher.ts`
- Create: `src/openstackPublisher.ts`
- Create: `src/ovhPublisher.ts`
- Create: `test/esxiPublisher.test.ts`
- Create: `test/nutanixPublisher.test.ts`
- Create: `test/openstackPublisher.test.ts`
- Create: `test/ovhPublisher.test.ts`

- [ ] **Step 1: Add interfaces and exports**

Add `EsxiPublisherArgs`, `NutanixPublisherArgs`, `OpenstackPublisherArgs`, and `OvhPublisherArgs` from the File Structure section to `src/types.ts`. Add exports:

```ts
export * from "./esxiPublisher";
export * from "./nutanixPublisher";
export * from "./openstackPublisher";
export * from "./ovhPublisher";
```

- [ ] **Step 2: Implement EsxiPublisher test and component**

Use `@pulumiverse/esxi-native` `VirtualMachine`. First inspect
`node_modules/@pulumiverse/esxi-native/virtualMachine.d.ts`.
If the generated `VirtualMachineArgs` type contains a `userData`,
`guestinfo`, `guestInfo`, or `extraConfig` input, attach the rendered
cloud-init through that field. If none of those fields exists, implement
`EsxiPublisher` as an explicit gate that throws:

```text
EsxiPublisher requires an esxi-native VirtualMachine user-data or guestinfo input; prepare a template with pre-seeded bootstrap data or use VspherePublisher.
```

Test must assert resource type `esxi-native:index/virtualMachine:VirtualMachine`, `diskStore`, `networkInterfaces`, and output map behavior.

- [ ] **Step 3: Implement NutanixPublisher test and component**

Use `@pierskarsenbarg/nutanix` VM resource found in the installed SDK. If the installed SDK does not expose a VM resource that accepts cloud-init/user-data, do not fake it; implement the component as an explicit unsupported gate that throws `NutanixPublisher requires a Nutanix VM resource with guest customization/user-data support`.

Test must assert either VM creation with cloud-init or the explicit gate error.

- [ ] **Step 4: Implement OpenstackPublisher test and component**

Use `@pulumi/openstack` compute instance resource. The component should pass:

```ts
imageName: args.imageName,
flavorName: args.flavorName,
networks: [{ name: args.networkName }],
keyPair: args.keyPair,
securityGroups: args.securityGroups,
availabilityZone: args.availabilityZone,
userData: userDataBase64.apply((value) => Buffer.from(value, "base64").toString("utf8")),
```

If `assignFloatingIp` is true, add a floating IP association resource from the OpenStack SDK and use that address as `publicIp`.

- [ ] **Step 5: Implement OvhPublisher test and component**

Inspect `@ovhcloud/pulumi-ovh` for a public cloud compute instance resource. If the provider exposes only Kubernetes and service-management resources and no instance resource, implement `OvhPublisher` as an explicit unsupported gate with message `OvhPublisher requires an OVH public cloud instance resource; use OpenstackPublisher for OVH Public Cloud VM deployments`.

Test must assert either VM creation with cloud-init or the explicit gate error.

- [ ] **Step 6: Run tests and commit**

Run:

```bash
npm test -- --test-name-pattern 'EsxiPublisher|NutanixPublisher|OpenstackPublisher|OvhPublisher'
```

Expected: PASS.

Commit:

```bash
git add src/types.ts src/index.ts src/esxiPublisher.ts src/nutanixPublisher.ts src/openstackPublisher.ts src/ovhPublisher.ts test/esxiPublisher.test.ts test/nutanixPublisher.test.ts test/openstackPublisher.test.ts test/ovhPublisher.test.ts
git commit -m "feat: add ESXi Nutanix OpenStack and OVH publishers"
```

---

### Task 6: Add Go Provider Schema Components

**Files:**
- Modify: `internal/provider/types.go`
- Modify: `internal/provider/components.go`
- Modify: `internal/provider/provider.go`
- Modify: `scripts/check-registry-readiness.mjs`
- Test: `internal/provider/provider_test.go`

- [ ] **Step 1: Add Go input structs**

Add these Go input structs to `internal/provider/components.go`, following the existing `AwsPublisherArgs` pattern with common fields embedded explicitly:

- `EsxiPublisherArgs`: common fields plus `DiskStore string`, `VirtualNetwork string`, `OS *string`, `Memory *int`, `NumVCpus *int`, `DiskSize *int`.
- `HcloudPublisherArgs`: common fields plus `ServerType *string`, `Image *string`, `Location *string`, `Datacenter *string`, `SSHKeys []string`, `FirewallIds []int`, `NetworkID *int`, `AssignPublicIP *bool`.
- `NutanixPublisherArgs`: common fields plus `ImageName string`, `SubnetName string`, `ClusterName *string`, `NumVCpus *int`, `NumCoresPerVcpu *int`, `MemorySizeMib *int`.
- `OpenstackPublisherArgs`: common fields plus `ImageName string`, `FlavorName string`, `NetworkName string`, `KeyPair *string`, `SecurityGroups []string`, `AvailabilityZone *string`, `AssignFloatingIP *bool`, `FloatingIPPool *string`.
- `OvhPublisherArgs`: common fields plus `ServiceName string`, `Region string`, `ImageName string`, `FlavorName string`, `SSHKeyName *string`, `NetworkID *string`.
- `ScalewayPublisherArgs`: common fields plus `Type *string`, `Image *string`, `Zone *string`, `SecurityGroupID *string`, `EnableDynamicIP *bool`.
- `OciPublisherArgs`: common fields plus `CompartmentID string`, `AvailabilityDomain string`, `Shape *string`, `SubnetID string`, `ImageID string`, `SSHPublicKey *string`, `AssignPublicIP *bool`.
- `AlicloudPublisherArgs`: common fields plus `InstanceType *string`, `ImageID string`, `VswitchID string`, `SecurityGroupIDs []string`, `KeyName *string`, `AllocatePublicIP *bool`.

Use Pulumi field tags that match the TypeScript input names, for example `pulumi:"diskStore"`, `pulumi:"assignPublicIp,optional"`, and `pulumi:"securityGroupIds"`.

Each new component struct must embed `pulumi.ResourceState`, embed its args struct, and expose:

```go
PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
```

- [ ] **Step 2: Add constructors**

For each new component, add `NewHcloudPublisher`, `NewScalewayPublisher`, `NewOciPublisher`, `NewAlicloudPublisher`, `NewEsxiPublisher`, `NewNutanixPublisher`, `NewOpenstackPublisher`, and `NewOvhPublisher`.

For Go provider parity in this pass, constructors may register component resources and return clear unsupported errors for providers whose Go SDK resource implementation is not available yet. Do not emit schema-only resources that silently do nothing.

- [ ] **Step 3: Register components**

Modify `internal/provider/provider.go`:

```go
infer.ComponentF(NewEsxiPublisher),
infer.ComponentF(NewHcloudPublisher),
infer.ComponentF(NewNutanixPublisher),
infer.ComponentF(NewOpenstackPublisher),
infer.ComponentF(NewOvhPublisher),
infer.ComponentF(NewScalewayPublisher),
infer.ComponentF(NewOciPublisher),
infer.ComponentF(NewAlicloudPublisher),
```

Update provider description and keywords to include the new providers.

- [ ] **Step 4: Update registry readiness check**

Add expected tokens:

```js
"netskope-publisher:index:EsxiPublisher",
"netskope-publisher:index:HcloudPublisher",
"netskope-publisher:index:NutanixPublisher",
"netskope-publisher:index:OpenstackPublisher",
"netskope-publisher:index:OvhPublisher",
"netskope-publisher:index:ScalewayPublisher",
"netskope-publisher:index:OciPublisher",
"netskope-publisher:index:AlicloudPublisher",
```

Add matching `sourceTokens` entries for each new `src/*Publisher.ts`.

- [ ] **Step 5: Run Go tests**

Run:

```bash
npm run go:test
npm run registry:check
```

Expected: PASS and schema contains every new token.

- [ ] **Step 6: Commit**

```bash
git add internal/provider/types.go internal/provider/components.go internal/provider/provider.go scripts/check-registry-readiness.mjs
git commit -m "feat: expose additional publishers in provider schema"
```

---

### Task 7: Regenerate Schema and SDKs

**Files:**
- Modify: `schema.json`
- Modify: `sdk/python/**`
- Modify: `sdk/dotnet/**`
- Modify: `sdk/go/**`
- Modify: `sdk/java/**`
- Modify: `sdk/rust/schema.json`

- [ ] **Step 1: Generate schema from Go provider**

Run:

```bash
go run ./cmd/pulumi-resource-netskope-publisher --schema > schema.json
```

Expected: `schema.json` includes all new component resource tokens.

- [ ] **Step 2: Generate SDKs**

Run:

```bash
npm run sdk:gen
```

Expected: Python, .NET, Go, Java, and Rust SDKs include new component types.

- [ ] **Step 3: Run generation checks**

Run:

```bash
npm run registry:check
CARGO_TARGET_DIR=/tmp/pulumi-netskope-rust-target cargo check --manifest-path sdk/rust/Cargo.toml
```

Expected: PASS. Java compilation is verified in CI where Gradle is installed.

- [ ] **Step 4: Commit**

```bash
git add schema.json sdk
git commit -m "chore: regenerate SDKs for additional publishers"
```

---

### Task 8: Update GitHub Pages and Registry Documentation

**Files:**
- Modify: `README.md`
- Modify: `docs/_index.md`
- Modify: `site/source/index.md`
- Modify: `site/source/reference/provider-matrix.md`
- Modify: `site/source/admin/index.md`
- Modify: `site/source/admin/component/index.md`
- Create: `site/source/admin/component/esxi.md`
- Create: `site/source/admin/component/hcloud.md`
- Create: `site/source/admin/component/nutanix.md`
- Create: `site/source/admin/component/openstack.md`
- Create: `site/source/admin/component/ovh.md`
- Create: `site/source/admin/component/scaleway.md`
- Create: `site/source/admin/component/oci.md`
- Create: `site/source/admin/component/alicloud.md`

- [ ] **Step 1: Update overview docs**

Add all new components to README, package docs, site landing page, admin landing page, component overview, and provider matrix. Explicitly state:

```md
ESXi Native is direct-host ESXi support and does not replace the vSphere component.
Hcloud, Nutanix, OpenStack, OVH, Scaleway, OCI, and Alicloud use bootstrap mode on Ubuntu 22.04 images.
```

- [ ] **Step 2: Add component pages**

Each new page must include:

- Required inputs.
- Bootstrap/image behavior.
- Outputs.
- Pulumi CLI setup.
- TypeScript, Python, C#, Go, Java, and Rust examples.

Each page must follow this concrete section order:

```md
---
title: <Provider> Component
toc: true
---

# <Provider> Component

## Inputs

## Image and bootstrap behavior

## Outputs

## Pulumi CLI

## TypeScript

## Python

## C#

## Go

## Java

## Rust
```

Do not add release-version notes or changelog entries.

- [ ] **Step 3: Build site**

Run:

```bash
npm run build
```

from `site/`.

Expected: PASS and generated routes for all new component pages.

- [ ] **Step 4: Commit**

```bash
git add README.md docs/_index.md site/source
git commit -m "docs: add additional provider guides"
```

---

### Task 9: Final Verification

**Files:**
- All changed files.

- [ ] **Step 1: Run local verification**

Run:

```bash
npm run typecheck
npm test
npm run go:test
npm run sdk:gen
npm run registry:check
npm run plugin:dist
npm run build
```

Run site build:

```bash
npm run build
```

from `site/`.

Run Rust SDK check:

```bash
CARGO_TARGET_DIR=/tmp/pulumi-netskope-rust-target cargo check --manifest-path sdk/rust/Cargo.toml
```

Run whitespace check:

```bash
git diff --check
```

Expected: all commands pass. If `gradle` is not installed locally, document that Java SDK compilation is covered by CI and do not claim local Java compilation passed.

- [ ] **Step 2: Inspect git status**

Run:

```bash
git status --short --branch
git log --oneline -8
```

Expected: branch is ahead by the task commits, with no unstaged changes.

- [ ] **Step 3: Report**

Report commit hashes, verification results, and any provider-specific limitations that were implemented as explicit gates.
