# Multi-Provider And Registry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the Pulumi Netskope Publisher package from AWS-only to Azure, GCP, vSphere, experimental Hyper-V readiness, and registry publishing readiness.

**Architecture:** Add one `ComponentResource` per stable platform while reusing existing name derivation, Netskope registration, cloud-init, and output conventions. Extract the common registration/output preparation from `AwsPublisher` into a shared component helper before adding `AzurePublisher`, `GcpPublisher`, and `VspherePublisher`. Treat Hyper-V as an experimental gated component because `@pulumi/hyperv` exists in the upstream GitHub SDK but is not published to npm.

**Tech Stack:** TypeScript, Node.js, Pulumi `@pulumi/pulumi`, `@pulumi/aws`, `@pulumi/azure-native`, `@pulumi/gcp`, `@pulumi/vsphere`, optional GitHub-sourced `@pulumi/hyperv`, Node test runner, Hexo/Cactus docs, GitHub Actions.

---

## Scope Check

This plan covers several independent subsystems. Keep each provider in its own task group and commit after each group:

- Task 1 prepares shared component helpers and dependencies.
- Tasks 2-4 add stable provider components: Azure, GCP, vSphere.
- Task 5 adds Hyper-V as an explicit experimental gate, not a default dependency.
- Tasks 6-7 update examples and docs for the provider matrix.
- Task 8 adds publishing readiness. Public Pulumi Registry publication is a decision gate because Pulumi documentation states public Registry components should use executable-based packages with generated SDKs; the current package is TypeScript source-based.
- Task 9 runs final verification.

## References

- Existing package plan: `docs/superpowers/plans/2026-05-18-pulumi-netskope-publisher.md`
- Existing design spec: `docs/superpowers/specs/2026-05-18-pulumi-netskope-publisher-design.md`
- Terraform source repo: `/Users/jneerdael/Scripts/terraform-netskope-publisher`
- Terraform modules to mirror: `modules/azure`, `modules/gcp`, `modules/vsphere`
- Pulumi packaging components: https://www.pulumi.com/docs/iac/guides/building-extending/components/packaging-components/
- Pulumi vSphere `VirtualMachine`: https://www.pulumi.com/registry/packages/vsphere/api-docs/virtualmachine/
- Pulumi GCP `compute.Instance`: https://www.pulumi.com/registry/packages/gcp/api-docs/compute/instance/
- Pulumi Azure Native `compute.VirtualMachine`: https://www.pulumi.com/registry/packages/azure-native/api-docs/compute/virtualmachine/
- Pulumi Hyper-V upstream SDK: https://github.com/pulumi/pulumi-hyperv

## File Structure

Create or modify these files:

- `package.json`: add stable provider dependencies and release scripts.
- `package-lock.json`: update lockfile.
- `src/types.ts`: add common args plus Azure/GCP/vSphere/Hyper-V args.
- `src/componentCore.ts`: shared publisher name, registration, tag, and output helper functions.
- `src/awsPublisher.ts`: refactor to shared helper without changing public API.
- `src/azurePublisher.ts`: new Azure component.
- `src/gcpPublisher.ts`: new GCP component.
- `src/vspherePublisher.ts`: new vSphere component.
- `src/hypervPublisher.ts`: experimental Hyper-V component behind optional dependency contract.
- `src/index.ts`: export new components and types.
- `test/componentCore.test.ts`: shared helper tests.
- `test/awsPublisher.test.ts`: regression test after refactor.
- `test/azurePublisher.test.ts`: Azure component test with Pulumi mocks.
- `test/gcpPublisher.test.ts`: GCP component test with Pulumi mocks.
- `test/vspherePublisher.test.ts`: vSphere component test with Pulumi mocks.
- `test/hypervPublisher.test.ts`: Hyper-V gate tests.
- `examples/azure-single/*`, `examples/gcp-single/*`, `examples/vsphere-single/*`: platform examples.
- `site/source/admin/component/{azure,gcp,vsphere,hyperv}.md`: component reference docs.
- `site/source/reference/provider-matrix.md`: provider support matrix.
- `site/source/reference/registry-publishing.md`: registry publishing strategy and limits.
- `.github/workflows/ci.yml`: include provider matrix docs and package checks.
- `.github/workflows/release.yml`: npm/GitHub release workflow, with public Registry gate documented.

## Task 1: Shared Component Core And Dependencies

**Files:**
- Modify: `package.json`
- Modify: `package-lock.json`
- Modify: `src/types.ts`
- Create: `src/componentCore.ts`
- Create: `test/componentCore.test.ts`
- Modify: `src/awsPublisher.ts`
- Modify: `test/awsPublisher.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Add stable provider dependencies**

Modify `package.json` dependencies to include:

```json
"@pulumi/azure-native": "^3.18.0",
"@pulumi/gcp": "^9.23.0",
"@pulumi/vsphere": "^4.16.5"
```

Keep `@pulumi/hyperv` out of `dependencies` because it is not published to npm. Hyper-V gets a separate experimental gate in Task 5.

- [ ] **Step 2: Install dependencies**

Run:

```bash
npm install
```

Expected: `package-lock.json` updates and npm exits successfully.

- [ ] **Step 3: Write shared helper tests**

Create `test/componentCore.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import {
  buildNameTag,
  createPublisherOutput,
  normalizeByoRegistrations,
  requireManagedRegistrationInputs,
} from "../src/componentCore";

test("buildNameTag merges provider tags with Name", async () => {
  const tags = await outputValue(buildNameTag({ Env: "dev" }, "pub-1"));
  assert.deepEqual(tags, { Env: "dev", Name: "pub-1" });
});

test("normalizeByoRegistrations requires every publisher name", () => {
  assert.throws(
    () => normalizeByoRegistrations(["pub-1"], {}),
    /registrations is missing data for publisher pub-1/,
  );
});

test("requireManagedRegistrationInputs rejects missing tenantUrl", () => {
  assert.throws(
    () => requireManagedRegistrationInputs({ apiToken: "token" }),
    /tenantUrl and apiToken are required when registrations are not provided/,
  );
});

test("createPublisherOutput preserves provider IDs and token", async () => {
  const output = await outputValue(createPublisherOutput({
    registration: pulumi.output({
      publisherId: 101,
      registrationToken: "token-101",
      existedBefore: true,
    }),
    vmId: pulumi.output("vm-1"),
    privateIp: pulumi.output("10.0.0.10"),
    publicIp: pulumi.output(undefined),
  }));

  assert.deepEqual(output, {
    publisherId: 101,
    registrationToken: "token-101",
    vmId: "vm-1",
    privateIp: "10.0.0.10",
    publicIp: undefined,
  });
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

- [ ] **Step 4: Run test to verify failure**

Run:

```bash
npm test
```

Expected: FAIL because `src/componentCore.ts` does not exist.

- [ ] **Step 5: Add shared component helpers**

Create `src/componentCore.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import { derivePublisherNames } from "./names";
import { NetskopeRegistration, RegistrationRecord } from "./netskopeRegistration";
import { CommonPublisherArgs, PublisherOutput, PublisherRegistrationInput } from "./types";

export function resolvePublisherNames(args: CommonPublisherArgs): string[] {
  return derivePublisherNames({
    namePrefix: args.namePrefix,
    names: args.names,
    replicas: args.replicas,
  });
}

export function createRegistrations(
  componentName: string,
  publisherNames: string[],
  args: CommonPublisherArgs,
  opts: pulumi.CustomResourceOptions,
): pulumi.Output<Record<string, RegistrationRecord>> {
  if (args.registrations !== undefined) {
    return pulumi.output(args.registrations).apply((registrations) =>
      normalizeByoRegistrations(publisherNames, registrations),
    );
  }

  const required = requireManagedRegistrationInputs(args);
  return new NetskopeRegistration(`${componentName}-registration`, {
    publisherNames,
    tenantUrl: required.tenantUrl,
    apiToken: required.apiToken,
  }, opts).registrations;
}

export function requireManagedRegistrationInputs(args: CommonPublisherArgs): {
  tenantUrl: pulumi.Input<string>;
  apiToken: pulumi.Input<string>;
} {
  if (args.tenantUrl === undefined || args.apiToken === undefined) {
    throw new Error("tenantUrl and apiToken are required when registrations are not provided");
  }

  return {
    tenantUrl: args.tenantUrl,
    apiToken: args.apiToken,
  };
}

export function normalizeByoRegistrations(
  publisherNames: string[],
  registrations: Record<string, PublisherRegistrationInput>,
): Record<string, RegistrationRecord> {
  return Object.fromEntries(publisherNames.map((publisherName) => {
    const registration = registrations[publisherName];
    if (registration === undefined) {
      throw new Error(`registrations is missing data for publisher ${publisherName}`);
    }

    return [publisherName, {
      publisherId: Number(registration.publisherId),
      registrationToken: String(registration.registrationToken),
      existedBefore: true,
    }];
  }));
}

export function buildNameTag(
  tags: pulumi.Input<Record<string, pulumi.Input<string>>> | undefined,
  publisherName: string,
): pulumi.Output<Record<string, pulumi.Input<string>>> {
  return pulumi.output(tags ?? {}).apply((inputTags) => ({
    ...inputTags,
    Name: publisherName,
  }));
}

export function createPublisherOutput(args: {
  registration: pulumi.Output<RegistrationRecord>;
  vmId: pulumi.Input<string>;
  privateIp: pulumi.Input<string>;
  publicIp: pulumi.Input<string | undefined>;
}): pulumi.Output<PublisherOutput> {
  return pulumi.all([
    args.registration,
    args.vmId,
    args.privateIp,
    args.publicIp,
  ]).apply(([registration, vmId, privateIp, publicIp]) => ({
    publisherId: registration.publisherId,
    registrationToken: registration.registrationToken,
    vmId,
    privateIp,
    publicIp,
  }));
}
```

- [ ] **Step 6: Update shared types**

Modify `src/types.ts` so the common and AWS output contracts become:

```ts
export interface CommonPublisherArgs extends NameArgs {
  tenantUrl?: pulumi.Input<string>;
  apiToken?: pulumi.Input<string>;
  wizardPath?: pulumi.Input<string>;
  tags?: pulumi.Input<Record<string, pulumi.Input<string>>>;
  registrations?: pulumi.Input<Record<string, PublisherRegistrationInput>>;
}

export interface PublisherOutput {
  publisherId: number;
  registrationToken: string;
  vmId: string;
  privateIp: string;
  publicIp?: string;
}

export interface AwsPublisherArgs extends CommonPublisherArgs {
  subnetId: pulumi.Input<string>;
  securityGroupIds: pulumi.Input<pulumi.Input<string>[]>;
  keyName?: pulumi.Input<string>;
  instanceType?: pulumi.Input<string>;
  amiId?: pulumi.Input<string>;
  associatePublicIpAddress?: pulumi.Input<boolean>;
  iamInstanceProfile?: pulumi.Input<string>;
  ebsOptimized?: pulumi.Input<boolean>;
  monitoring?: pulumi.Input<boolean>;
  metadataOptions?: pulumi.Input<MetadataOptions>;
}
```

Keep `PublisherRegistrationInput` and `MetadataOptions` unchanged.

- [ ] **Step 7: Refactor AWS to shared helpers**

Modify `src/awsPublisher.ts` to import:

```ts
import {
  buildNameTag,
  createPublisherOutput,
  createRegistrations,
  resolvePublisherNames,
} from "./componentCore";
```

Remove local `createManagedRegistrations` and `normalizeByoRegistrations`. Replace name resolution and registrations with:

```ts
const publisherNames = resolvePublisherNames(args);
this.publisherNames = pulumi.output(publisherNames);
const registrations = createRegistrations(name, publisherNames, args, parentOpts);
```

Replace tag creation with:

```ts
const tags = buildNameTag(args.tags, publisherName);
```

Replace AWS output creation with:

```ts
publisherOutputs[publisherName] = createPublisherOutput({
  registration,
  vmId: instance.id,
  privateIp: instance.privateIp,
  publicIp: instance.publicIp,
});
```

- [ ] **Step 8: Update AWS test assertion**

In `test/awsPublisher.test.ts`, change:

```ts
assert.equal(publishers["pub-1"].instanceId, "publisher-pub-1-id");
```

to:

```ts
assert.equal(publishers["pub-1"].vmId, "publisher-pub-1-id");
```

- [ ] **Step 9: Export helper module**

Modify `src/index.ts`:

```ts
export * from "./componentCore";
```

- [ ] **Step 10: Verify and commit**

Run:

```bash
npm test
```

Expected: PASS.

Run:

```bash
git add package.json package-lock.json src/types.ts src/componentCore.ts src/awsPublisher.ts src/index.ts test/componentCore.test.ts test/awsPublisher.test.ts
git commit -m "refactor: share publisher component core"
```

## Task 2: Azure Publisher Component

**Files:**
- Modify: `src/types.ts`
- Create: `src/azurePublisher.ts`
- Create: `test/azurePublisher.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Add Azure types**

Append to `src/types.ts`:

```ts
export interface AzureMarketplaceImage {
  publisher: pulumi.Input<string>;
  offer: pulumi.Input<string>;
  sku: pulumi.Input<string>;
  version?: pulumi.Input<string>;
}

export interface AzureOsDisk {
  type?: pulumi.Input<string>;
  sizeGb?: pulumi.Input<number>;
}

export interface AzurePublisherArgs extends CommonPublisherArgs {
  resourceGroupName: pulumi.Input<string>;
  location: pulumi.Input<string>;
  subnetId: pulumi.Input<string>;
  vmSize?: pulumi.Input<string>;
  adminUsername?: pulumi.Input<string>;
  adminSshPublicKey: pulumi.Input<string>;
  networkSecurityGroupId?: pulumi.Input<string>;
  assignPublicIp?: pulumi.Input<boolean>;
  osDisk?: pulumi.Input<AzureOsDisk>;
  imageId?: pulumi.Input<string>;
  marketplace?: pulumi.Input<AzureMarketplaceImage>;
  acceptMarketplaceTerms?: pulumi.Input<boolean>;
}
```

- [ ] **Step 2: Write Azure component test**

Create `test/azurePublisher.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { AzurePublisher } from "../src/azurePublisher";
import { PublisherOutput } from "../src/types";

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "azure-native:network:NetworkInterface") {
      return { id: `${args.name}-id`, state: { ...args.inputs, ipConfigurations: [{ privateIPAddress: "10.1.0.10" }] } };
    }

    if (args.type === "azure-native:network:PublicIPAddress") {
      return { id: `${args.name}-id`, state: { ...args.inputs, ipAddress: "203.0.113.10" } };
    }

    if (args.type === "azure-native:compute:VirtualMachine") {
      return { id: `${args.name}-id`, state: args.inputs };
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

test("AzurePublisher creates outputs keyed by publisher name", async () => {
  const component = new AzurePublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    resourceGroupName: "rg",
    location: "westeurope",
    subnetId: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/default",
    adminSshPublicKey: "ssh-rsa AAAA",
    imageId: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/images/npa",
  });

  const publisherNames = await outputValue(component.publisherNames);
  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);

  assert.deepEqual(publisherNames, ["pub-1"]);
  assert.equal(publishers["pub-1"].publisherId, 101);
  assert.equal(publishers["pub-1"].vmId, "publisher-pub-1-id");
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

- [ ] **Step 3: Run test to verify failure**

Run:

```bash
npm test
```

Expected: FAIL because `src/azurePublisher.ts` does not exist.

- [ ] **Step 4: Implement Azure component**

Create `src/azurePublisher.ts` with this resource mapping:

```ts
import * as azure from "@pulumi/azure-native";
import * as pulumi from "@pulumi/pulumi";
import { renderUserDataBase64 } from "./cloudInit";
import {
  buildNameTag,
  createPublisherOutput,
  createRegistrations,
  resolvePublisherNames,
} from "./componentCore";
import { AzurePublisherArgs, PublisherOutput } from "./types";

export class AzurePublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: AzurePublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope:index:AzurePublisher", name, {}, opts);

    const parentOpts = { parent: this };
    const publisherNames = resolvePublisherNames(args);
    this.publisherNames = pulumi.output(publisherNames);
    const registrations = createRegistrations(name, publisherNames, args, parentOpts);
    const publisherOutputs: Record<string, pulumi.Output<PublisherOutput>> = {};

    const adminUsername = args.adminUsername ?? "ubuntu";
    const vmSize = args.vmSize ?? "Standard_D2s_v5";
    const assignPublicIp = args.assignPublicIp ?? false;
    const osDisk = pulumi.output(args.osDisk ?? {}).apply((disk) => ({
      type: disk.type ?? "Premium_LRS",
      sizeGb: disk.sizeGb ?? 64,
    }));

    for (const publisherName of publisherNames) {
      const registration = registrations.apply((allRegistrations) => allRegistrations[publisherName]);
      const customData = pulumi.all([registration, args.wizardPath]).apply(([record, wizardPath]) =>
        renderUserDataBase64({
          publisherName,
          registrationToken: record.registrationToken,
          wizardPath,
        }),
      );

      const publicIp = assignPublicIp
        ? new azure.network.PublicIPAddress(`${name}-${publisherName}-pip`, {
          publicIpAddressName: `${publisherName}-pip`,
          resourceGroupName: args.resourceGroupName,
          location: args.location,
          publicIPAllocationMethod: "Static",
          sku: { name: "Standard" },
          tags: args.tags,
        }, parentOpts)
        : undefined;

      const nic = new azure.network.NetworkInterface(`${name}-${publisherName}-nic`, {
        networkInterfaceName: `${publisherName}-nic`,
        resourceGroupName: args.resourceGroupName,
        location: args.location,
        tags: args.tags,
        ipConfigurations: [{
          name: "primary",
          subnet: { id: args.subnetId },
          privateIPAllocationMethod: "Dynamic",
          publicIPAddress: publicIp ? { id: publicIp.id } : undefined,
        }],
        networkSecurityGroup: args.networkSecurityGroupId ? { id: args.networkSecurityGroupId } : undefined,
      }, parentOpts);

      const vm = new azure.compute.VirtualMachine(`${name}-${publisherName}`, {
        vmName: publisherName,
        resourceGroupName: args.resourceGroupName,
        location: args.location,
        tags: buildNameTag(args.tags, publisherName),
        hardwareProfile: { vmSize },
        networkProfile: {
          networkInterfaces: [{ id: nic.id, primary: true }],
        },
        osProfile: {
          computerName: publisherName,
          adminUsername,
          customData,
          linuxConfiguration: {
            disablePasswordAuthentication: true,
            ssh: {
              publicKeys: [{
                path: pulumi.interpolate`/home/${adminUsername}/.ssh/authorized_keys`,
                keyData: args.adminSshPublicKey,
              }],
            },
          },
        },
        storageProfile: {
          imageReference: pulumi.output(args.marketplace).apply((marketplace) =>
            args.imageId ? { id: args.imageId } : marketplace ? {
              publisher: marketplace.publisher,
              offer: marketplace.offer,
              sku: marketplace.sku,
              version: marketplace.version ?? "latest",
            } : undefined,
          ),
          osDisk: osDisk.apply((disk) => ({
            createOption: "FromImage",
            caching: "ReadWrite",
            managedDisk: { storageAccountType: disk.type },
            diskSizeGB: disk.sizeGb,
          })),
        },
        plan: pulumi.output(args.marketplace).apply((marketplace) =>
          args.imageId ? undefined : marketplace ? {
            publisher: marketplace.publisher,
            product: marketplace.offer,
            name: marketplace.sku,
          } : undefined,
        ),
      }, parentOpts);

      publisherOutputs[publisherName] = createPublisherOutput({
        registration,
        vmId: vm.id,
        privateIp: pulumi.output(nic.ipConfigurations).apply((configs) => configs?.[0]?.privateIPAddress ?? ""),
        publicIp: publicIp?.ipAddress,
      });
    }

    this.publishers = pulumi.secret(pulumi.all(publisherOutputs));
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
```

The implementation must throw `Provide either imageId or marketplace.` before creating resources when both `args.imageId` and `args.marketplace` are absent.

- [ ] **Step 5: Export Azure component**

Modify `src/index.ts`:

```ts
export * from "./azurePublisher";
```

- [ ] **Step 6: Verify and commit**

Run:

```bash
npm test
```

Expected: PASS.

Run:

```bash
git add src/types.ts src/azurePublisher.ts src/index.ts test/azurePublisher.test.ts
git commit -m "feat: add Azure publisher component"
```

## Task 3: GCP Publisher Component

**Files:**
- Modify: `src/types.ts`
- Create: `src/gcpPublisher.ts`
- Create: `test/gcpPublisher.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Add GCP types**

Append to `src/types.ts`:

```ts
export interface GcpServiceAccount {
  email: pulumi.Input<string>;
  scopes?: pulumi.Input<pulumi.Input<string>[]>;
}

export interface GcpPublisherArgs extends CommonPublisherArgs {
  project: pulumi.Input<string>;
  zone: pulumi.Input<string>;
  network: pulumi.Input<string>;
  subnetwork: pulumi.Input<string>;
  machineType?: pulumi.Input<string>;
  image: pulumi.Input<string>;
  assignPublicIp?: pulumi.Input<boolean>;
  networkTags?: pulumi.Input<pulumi.Input<string>[]>;
  serviceAccount?: pulumi.Input<GcpServiceAccount>;
}
```

- [ ] **Step 2: Create GCP test**

Create `test/gcpPublisher.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { GcpPublisher } from "../src/gcpPublisher";
import { PublisherOutput } from "../src/types";

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "gcp:compute/instance:Instance") {
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          instanceId: `${args.name}-numeric-id`,
          networkInterfaces: [{
            networkIp: "10.2.0.10",
            accessConfigs: [{ natIp: "203.0.113.20" }],
          }],
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

test("GcpPublisher creates outputs keyed by publisher name", async () => {
  const component = new GcpPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    project: "project",
    zone: "europe-west4-a",
    network: "default",
    subnetwork: "default",
    image: "projects/example/global/images/npa",
    assignPublicIp: true,
  });

  const publisherNames = await outputValue(component.publisherNames);
  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);

  assert.deepEqual(publisherNames, ["pub-1"]);
  assert.equal(publishers["pub-1"].publisherId, 101);
  assert.equal(publishers["pub-1"].vmId, "publisher-pub-1-numeric-id");
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

- [ ] **Step 3: Run test to verify failure**

Run:

```bash
npm test
```

Expected: FAIL because `src/gcpPublisher.ts` does not exist.

- [ ] **Step 4: Implement GCP component**

Create `src/gcpPublisher.ts` with these behaviors:

- Use `new gcp.compute.Instance`.
- Set `machineType` default to `e2-medium`.
- Set boot disk image from `args.image`.
- Set `metadata["user-data"]` to raw `renderUserData(...)`, matching Terraform.
- Include `accessConfigs: [{}]` only when `assignPublicIp` is true.
- Map tags to GCP labels with Pulumi input support, and use `networkTags` for instance network tags.
- Use `serviceAccount.scopes` default `["https://www.googleapis.com/auth/cloud-platform"]`.
- Return `vmId` from `instance.instanceId`, private IP from first network interface, public IP from first access config NAT IP.

The implementation must define `GcpPublisher extends pulumi.ComponentResource` with public `publisherNames` and `publishers` outputs, call `super("netskope:index:GcpPublisher", name, {}, opts)`, derive names synchronously with `resolvePublisherNames(args)`, create registrations with `createRegistrations`, loop over the derived names, create resources outside `apply`, set `this.publishers = pulumi.secret(pulumi.all(publisherOutputs))`, and call `this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers })`.

- [ ] **Step 5: Export GCP component**

Modify `src/index.ts`:

```ts
export * from "./gcpPublisher";
```

- [ ] **Step 6: Verify and commit**

Run:

```bash
npm test
```

Expected: PASS.

Run:

```bash
git add src/types.ts src/gcpPublisher.ts src/index.ts test/gcpPublisher.test.ts
git commit -m "feat: add GCP publisher component"
```

## Task 4: vSphere Publisher Component

**Files:**
- Modify: `src/types.ts`
- Create: `src/vspherePublisher.ts`
- Create: `test/vspherePublisher.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Add vSphere types**

Append to `src/types.ts`:

```ts
export interface VspherePublisherArgs extends CommonPublisherArgs {
  datacenter: pulumi.Input<string>;
  cluster?: pulumi.Input<string>;
  host?: pulumi.Input<string>;
  datastore: pulumi.Input<string>;
  networkName: pulumi.Input<string>;
  templateName: pulumi.Input<string>;
  folder?: pulumi.Input<string>;
  numCpus?: pulumi.Input<number>;
  memory?: pulumi.Input<number>;
}
```

- [ ] **Step 2: Create vSphere test**

Create `test/vspherePublisher.test.ts` with Pulumi mocks for these tokens:

- `vsphere:index/getDatacenter:getDatacenter`
- `vsphere:index/getDatastore:getDatastore`
- `vsphere:index/getNetwork:getNetwork`
- `vsphere:index/getVirtualMachine:getVirtualMachine`
- `vsphere:index/getComputeCluster:getComputeCluster`
- `vsphere:index/virtualMachine:VirtualMachine`
- `pulumi-nodejs:dynamic:Resource`

Assert:

```ts
assert.deepEqual(publisherNames, ["pub-1"]);
assert.equal(publishers["pub-1"].publisherId, 101);
assert.equal(publishers["pub-1"].vmId, "publisher-pub-1-id");
```

- [ ] **Step 3: Run test to verify failure**

Run:

```bash
npm test
```

Expected: FAIL because `src/vspherePublisher.ts` does not exist.

- [ ] **Step 4: Implement vSphere component**

Create `src/vspherePublisher.ts` with these behaviors:

- Use data sources for datacenter, datastore, network, template, and either cluster or host.
- Throw `Provide either vsphere.cluster or vsphere.host.` if both are absent.
- Use `new vsphere.VirtualMachine`.
- Set `numCpus` default `2`, `memory` default `4096`.
- Clone from template ID.
- Use first template disk and first template network interface type.
- Set `extraConfig`:

```ts
{
  "guestinfo.userdata": renderUserDataBase64(...),
  "guestinfo.userdata.encoding": "base64",
  "guestinfo.metadata": Buffer.from(renderMetadata(publisherName), "utf8").toString("base64"),
  "guestinfo.metadata.encoding": "base64",
}
```

- Return `vmId` from VM ID, private IP from `defaultIpAddress`, and public IP as `undefined`.

The implementation must not create Pulumi resources inside an `apply`.

- [ ] **Step 5: Export vSphere component**

Modify `src/index.ts`:

```ts
export * from "./vspherePublisher";
```

- [ ] **Step 6: Verify and commit**

Run:

```bash
npm test
```

Expected: PASS.

Run:

```bash
git add src/types.ts src/vspherePublisher.ts src/index.ts test/vspherePublisher.test.ts
git commit -m "feat: add vSphere publisher component"
```

## Task 5: Hyper-V Experimental Gate

**Files:**
- Modify: `src/types.ts`
- Create: `src/hypervPublisher.ts`
- Create: `test/hypervPublisher.test.ts`
- Modify: `src/index.ts`
- Modify: `README.md`
- Create: `docs/hyperv-experimental.md`

- [ ] **Step 1: Add Hyper-V types**

Append to `src/types.ts`:

```ts
export interface HypervHardDrive {
  path: pulumi.Input<string>;
  controllerType?: pulumi.Input<string>;
  controllerNumber?: pulumi.Input<number>;
  controllerLocation?: pulumi.Input<number>;
}

export interface HypervPublisherArgs extends CommonPublisherArgs {
  switchName: pulumi.Input<string>;
  hardDrives: pulumi.Input<pulumi.Input<HypervHardDrive>[]>;
  generation?: pulumi.Input<number>;
  processorCount?: pulumi.Input<number>;
  memorySize?: pulumi.Input<number>;
  dynamicMemory?: pulumi.Input<boolean>;
  minimumMemory?: pulumi.Input<number>;
  maximumMemory?: pulumi.Input<number>;
  autoStartAction?: pulumi.Input<string>;
  autoStopAction?: pulumi.Input<string>;
  enableExperimentalHyperv?: boolean;
}
```

- [ ] **Step 2: Create Hyper-V gate test**

Create `test/hypervPublisher.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import { HypervPublisher } from "../src/hypervPublisher";

test("HypervPublisher requires explicit experimental opt-in", () => {
  assert.throws(
    () => new HypervPublisher("publisher", {
      names: ["pub-1"],
      switchName: "Default Switch",
      hardDrives: [{ path: "C:\\\\VMs\\\\pub-1\\\\disk.vhdx" }],
    }),
    /Hyper-V support is experimental and requires enableExperimentalHyperv: true/,
  );
});
```

- [ ] **Step 3: Implement gated Hyper-V component**

Create `src/hypervPublisher.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import { HypervPublisherArgs, PublisherOutput } from "./types";

export class HypervPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: HypervPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope:index:HypervPublisher", name, {}, opts);

    if (args.enableExperimentalHyperv !== true) {
      throw new Error("Hyper-V support is experimental and requires enableExperimentalHyperv: true");
    }

    throw new Error(
      "Hyper-V support requires @pulumi/hyperv from pulumi/pulumi-hyperv because it is not published to npm.",
    );
  }
}
```

This file is intentionally a runtime gate. Do not import `@pulumi/hyperv` until the dependency can be resolved by package consumers.

- [ ] **Step 4: Export Hyper-V gate**

Modify `src/index.ts`:

```ts
export * from "./hypervPublisher";
```

- [ ] **Step 5: Document Hyper-V status**

Create `docs/hyperv-experimental.md`:

```markdown
# Hyper-V Experimental Status

Hyper-V support is not enabled by default.

The upstream Pulumi Hyper-V provider exists at
`https://github.com/pulumi/pulumi-hyperv`, but `@pulumi/hyperv` is not
published to npm. The package exposes a generated Node SDK in
`sdk/nodejs` and marks itself as `1.0.0-alpha.0+dev`.

This repository keeps `HypervPublisher` behind an explicit runtime gate
until the provider can be consumed through a stable package source.
```

- [ ] **Step 6: Verify and commit**

Run:

```bash
npm test
```

Expected: PASS.

Run:

```bash
git add src/types.ts src/hypervPublisher.ts src/index.ts test/hypervPublisher.test.ts docs/hyperv-experimental.md README.md
git commit -m "feat: add experimental Hyper-V gate"
```

## Task 6: Multi-Provider Examples

**Files:**
- Create: `examples/azure-single/Pulumi.yaml`
- Create: `examples/azure-single/package.json`
- Create: `examples/azure-single/index.ts`
- Create: `examples/azure-single/README.md`
- Create: `examples/gcp-single/Pulumi.yaml`
- Create: `examples/gcp-single/package.json`
- Create: `examples/gcp-single/index.ts`
- Create: `examples/gcp-single/README.md`
- Create: `examples/vsphere-single/Pulumi.yaml`
- Create: `examples/vsphere-single/package.json`
- Create: `examples/vsphere-single/index.ts`
- Create: `examples/vsphere-single/README.md`

- [ ] **Step 1: Create Azure example**

Create `examples/azure-single/index.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import { AzurePublisher } from "@johnneerdael/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new AzurePublisher("publisher", {
  namePrefix: config.get("namePrefix") ?? "npa-publisher",
  replicas: config.getNumber("replicas") ?? 1,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  resourceGroupName: config.require("resourceGroupName"),
  location: config.require("location"),
  subnetId: config.require("subnetId"),
  adminSshPublicKey: config.require("adminSshPublicKey"),
  imageId: config.get("imageId") ?? undefined,
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

Create `examples/azure-single/Pulumi.yaml`:

```yaml
name: azure-single
description: Deploy Netskope Private Access Publishers on Azure.
runtime: nodejs
```

Create `examples/azure-single/package.json`:

```json
{
  "name": "pulumi-netskope-publisher-azure-single-example",
  "private": true,
  "scripts": {
    "build": "tsc -p ../../tsconfig.json",
    "preview": "pulumi preview",
    "up": "pulumi up",
    "destroy": "pulumi destroy"
  },
  "dependencies": {
    "@johnneerdael/pulumi-netskope-publisher": "file:../..",
    "@pulumi/azure-native": "^3.18.0",
    "@pulumi/pulumi": "^3.0.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0"
  }
}
```

Create `examples/azure-single/README.md`:

```markdown
# Azure Single Example

Deploy one or more Netskope Private Access Publishers on Azure.

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set resourceGroupName rg-npa
pulumi config set location westeurope
pulumi config set subnetId /subscriptions/.../subnets/default
pulumi config set adminSshPublicKey "ssh-rsa AAAA..."
pulumi config set imageId /subscriptions/.../providers/Microsoft.Compute/images/npa
npm install
pulumi preview
pulumi up
```
```

- [ ] **Step 2: Create GCP example**

Create `examples/gcp-single/index.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import { GcpPublisher } from "@johnneerdael/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new GcpPublisher("publisher", {
  namePrefix: config.get("namePrefix") ?? "npa-publisher",
  replicas: config.getNumber("replicas") ?? 1,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  project: config.require("project"),
  zone: config.require("zone"),
  network: config.require("network"),
  subnetwork: config.require("subnetwork"),
  image: config.require("image"),
  assignPublicIp: config.getBoolean("assignPublicIp") ?? false,
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

Create `examples/gcp-single/Pulumi.yaml`:

```yaml
name: gcp-single
description: Deploy Netskope Private Access Publishers on GCP.
runtime: nodejs
```

Create `examples/gcp-single/package.json`:

```json
{
  "name": "pulumi-netskope-publisher-gcp-single-example",
  "private": true,
  "scripts": {
    "build": "tsc -p ../../tsconfig.json",
    "preview": "pulumi preview",
    "up": "pulumi up",
    "destroy": "pulumi destroy"
  },
  "dependencies": {
    "@johnneerdael/pulumi-netskope-publisher": "file:../..",
    "@pulumi/gcp": "^9.23.0",
    "@pulumi/pulumi": "^3.0.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0"
  }
}
```

Create `examples/gcp-single/README.md`:

```markdown
# GCP Single Example

Deploy one or more Netskope Private Access Publishers on GCP.

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set project my-project
pulumi config set zone europe-west4-a
pulumi config set network default
pulumi config set subnetwork default
pulumi config set image projects/my-project/global/images/npa
npm install
pulumi preview
pulumi up
```
```

- [ ] **Step 3: Create vSphere example**

Create `examples/vsphere-single/index.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import { VspherePublisher } from "@johnneerdael/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new VspherePublisher("publisher", {
  namePrefix: config.get("namePrefix") ?? "npa-publisher",
  replicas: config.getNumber("replicas") ?? 1,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  datacenter: config.require("datacenter"),
  cluster: config.get("cluster") ?? undefined,
  host: config.get("host") ?? undefined,
  datastore: config.require("datastore"),
  networkName: config.require("networkName"),
  templateName: config.require("templateName"),
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

Create `examples/vsphere-single/Pulumi.yaml`:

```yaml
name: vsphere-single
description: Deploy Netskope Private Access Publishers on vSphere.
runtime: nodejs
```

Create `examples/vsphere-single/package.json`:

```json
{
  "name": "pulumi-netskope-publisher-vsphere-single-example",
  "private": true,
  "scripts": {
    "build": "tsc -p ../../tsconfig.json",
    "preview": "pulumi preview",
    "up": "pulumi up",
    "destroy": "pulumi destroy"
  },
  "dependencies": {
    "@johnneerdael/pulumi-netskope-publisher": "file:../..",
    "@pulumi/pulumi": "^3.0.0",
    "@pulumi/vsphere": "^4.16.5"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0"
  }
}
```

Create `examples/vsphere-single/README.md`:

```markdown
# vSphere Single Example

Deploy one or more Netskope Private Access Publishers on vSphere.

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set datacenter dc-01
pulumi config set cluster cluster-01
pulumi config set datastore datastore-01
pulumi config set networkName VM Network
pulumi config set templateName npa-publisher-template
npm install
pulumi preview
pulumi up
```
```

- [ ] **Step 4: Verify and commit**

Run:

```bash
npm run typecheck
```

Expected: PASS.

Run:

```bash
git add examples/azure-single examples/gcp-single examples/vsphere-single
git commit -m "docs: add multi-provider examples"
```

## Task 7: Documentation Provider Matrix

**Files:**
- Modify: `README.md`
- Modify: `site/source/index.md`
- Modify: `site/source/reference/roadmap.md`
- Create: `site/source/reference/provider-matrix.md`
- Create: `site/source/admin/component/azure.md`
- Create: `site/source/admin/component/gcp.md`
- Create: `site/source/admin/component/vsphere.md`
- Create: `site/source/admin/component/hyperv.md`
- Create: `site/source/reference/registry-publishing.md`

- [ ] **Step 1: Add provider matrix docs**

Create `site/source/reference/provider-matrix.md`:

```markdown
---
title: Provider Matrix
---

# Provider Matrix

| Platform | Component | Status |
|---|---|---|
| AWS | `AwsPublisher` | Supported |
| Azure | `AzurePublisher` | Supported |
| GCP | `GcpPublisher` | Supported |
| vSphere | `VspherePublisher` | Supported |
| Hyper-V | `HypervPublisher` | Experimental gate |

All supported providers share name derivation, Netskope registration,
cloud-init generation, and secret output conventions.
```

- [ ] **Step 2: Add component docs**

Create `site/source/admin/component/azure.md`:

```markdown
---
title: Azure Component
---

# Azure Component

`AzurePublisher` creates one Linux virtual machine per publisher name.

Required inputs: `resourceGroupName`, `location`, `subnetId`,
`adminSshPublicKey`, and either `imageId` or `marketplace`.

Optional inputs include `vmSize`, `adminUsername`,
`networkSecurityGroupId`, `assignPublicIp`, `osDisk`, `tags`,
`namePrefix`, `names`, and `replicas`.

Outputs: `publisherNames` and secret `publishers`.
```

Create `site/source/admin/component/gcp.md`:

```markdown
---
title: GCP Component
---

# GCP Component

`GcpPublisher` creates one Compute Engine instance per publisher name.

Required inputs: `project`, `zone`, `network`, `subnetwork`, and `image`.

Optional inputs include `machineType`, `assignPublicIp`, `networkTags`,
`serviceAccount`, `tags`, `namePrefix`, `names`, and `replicas`.

Outputs: `publisherNames` and secret `publishers`.
```

Create `site/source/admin/component/vsphere.md`:

```markdown
---
title: vSphere Component
---

# vSphere Component

`VspherePublisher` clones one VM per publisher name from an existing
template.

Required inputs: `datacenter`, `datastore`, `networkName`,
`templateName`, and either `cluster` or `host`.

Optional inputs include `folder`, `numCpus`, `memory`, `tags`,
`namePrefix`, `names`, and `replicas`.

Outputs: `publisherNames` and secret `publishers`.
```

Create `site/source/admin/component/hyperv.md`:

```markdown
---
title: Hyper-V Component
---

# Hyper-V Component

`HypervPublisher` is an experimental gate. The upstream Pulumi Hyper-V
provider exists in `pulumi/pulumi-hyperv`, but `@pulumi/hyperv` is not
published to npm.

The component requires `enableExperimentalHyperv: true` and then fails
with a clear dependency message until a stable package source exists.
```

- [ ] **Step 3: Add registry publishing docs**

Create `site/source/reference/registry-publishing.md`:

```markdown
---
title: Registry Publishing
---

# Registry Publishing

This package is currently a TypeScript source-based component package.
It can be consumed from Git references and published to npm for
TypeScript users.

Pulumi's public Registry path for broadly consumable components expects
an executable-based package with generated SDKs. Moving this package to
public Registry publication requires a separate provider packaging track.

Immediate supported release path:

1. Publish the TypeScript package to npm.
2. Tag GitHub releases.
3. Use Git references or npm for consumption.
4. Revisit executable-based packaging before public Pulumi Registry
   submission.
```

- [ ] **Step 4: Update README and site landing page**

Update `README.md` and `site/source/index.md` to list AWS, Azure, GCP, and vSphere as supported, and Hyper-V as experimental.

- [ ] **Step 5: Build docs and commit**

Run:

```bash
cd site
npm run build
```

Expected: PASS.

Run:

```bash
git add README.md site/source
git commit -m "docs: document provider matrix and registry path"
```

## Task 8: Release And Registry Readiness

**Files:**
- Modify: `package.json`
- Create: `.github/workflows/release.yml`
- Create: `docs/registry-publication-checklist.md`

- [ ] **Step 1: Add release scripts**

Modify `package.json` scripts:

```json
"prepack": "npm run clean && npm run build",
"release:check": "npm ci && npm run typecheck && npm test"
```

- [ ] **Step 2: Add npm/GitHub release workflow**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write
  id-token: write

jobs:
  npm:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
          registry-url: https://registry.npmjs.org
      - run: npm ci
      - run: npm run typecheck
      - run: npm test
      - run: npm publish --provenance --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
      - uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true
```

- [ ] **Step 3: Add public Registry checklist**

Create `docs/registry-publication-checklist.md`:

```markdown
# Pulumi Registry Publication Checklist

The current package is TypeScript source-based. Public Pulumi Registry
publication requires an executable-based package track with generated
SDKs.

Before requesting public Registry listing:

- Decide whether to rewrite the package as a Go executable provider.
- Generate schema and SDKs for Node.js, Python, Go, .NET, and Java.
- Publish SDKs to public language package feeds.
- Publish provider binaries for supported platforms.
- Add Registry metadata, examples, and API docs.
- Keep the TypeScript source-based package as either the canonical
  package or a compatibility package, but do not publish two packages
  with conflicting APIs.
```

- [ ] **Step 4: Verify and commit**

Run:

```bash
npm run typecheck
npm test
```

Expected: PASS.

Run:

```bash
git add package.json .github/workflows/release.yml docs/registry-publication-checklist.md
git commit -m "ci: prepare package release workflow"
```

## Task 9: Final Verification

**Files:**
- Inspect all changed files.

- [ ] **Step 1: Clean package build**

Run:

```bash
npm run clean
npm ci
npm run typecheck
npm test
```

Expected: PASS.

- [ ] **Step 2: Clean docs build**

Run:

```bash
cd site
npm ci
npm run build
```

Expected: PASS.

- [ ] **Step 3: Check git hygiene**

Run:

```bash
git status --short
git diff --check
```

Expected: clean except for intentional uncommitted work before final commit, and no whitespace errors.

- [ ] **Step 4: Final report**

Report:

- Latest commit hash.
- Provider components added.
- Whether `npm test` passed.
- Whether docs build passed.
- Hyper-V status as experimental gated.
- Public Registry status as release-ready documentation plus npm release workflow, with executable-based Registry publication still requiring a separate provider-packaging track.
