# Provider Framework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a catalog-driven provider framework for bootstrap VM publishers with capability metadata, validation, docs/example checks, and adapter-driven factories while preserving public component names and Pulumi tokens.

**Architecture:** Add a TypeScript provider catalog as the first source of truth, then layer validation and parity checks on top before migrating resource construction. Keep AWS, Azure, GCP, Kubernetes, vSphere, ESXi, Hyper-V, and NetskopeRegistration bespoke, but include them in metadata for docs and parity checks. Use small factories for catalog-fit raw VM providers instead of a universal provider DSL.

**Tech Stack:** TypeScript Pulumi components, Node test runner, Pulumi mocks, Go executable provider, Pulumi schema generation, Hexo GitHub Pages docs.

---

## File Structure

- Create: `src/providerCatalog.ts`
  - Own provider capability metadata, implementation mode, user-data placement, required/optional input metadata, YAML example values, and docs metadata.
- Create: `test/providerCatalog.test.ts`
  - Validate catalog uniqueness, required fields, provider category correctness, and example metadata shape.
- Create: `src/providerValidation.ts`
  - Validate provider args against catalog rules before resource creation.
- Create: `test/providerValidation.test.ts`
  - Unit-test missing required fields, required-one-of rules, mutually exclusive inputs, and experimental opt-in rules.
- Create: `src/catalogVmFactory.ts`
  - Provide reusable raw VM component factory helpers for catalog-fit providers.
- Modify: `src/digitaloceanPublisher.ts`, `src/vultrPublisher.ts`, `src/exoscalePublisher.ts`, `src/upcloudPublisher.ts`, `src/stackitPublisher.ts`, `src/equinixPublisher.ts`, `src/outscalePublisher.ts`, `src/opentelekomcloudPublisher.ts`, `src/tencentcloudPublisher.ts`, `src/yandexPublisher.ts`
  - Migrate simple raw VM providers to catalog-driven factory calls.
- Create: `scripts/check-provider-catalog.mjs`
  - Check catalog parity against `schema.json`, `src/index.ts`, Go registration, and GitHub Pages docs.
- Modify: `package.json`
  - Add `catalog:check` script and include it in `registry:check` or `release:check`.
- Create: `scripts/generate-provider-docs.mjs`
  - Generate catalog-owned snippets for provider matrix, component index, shared cloud-init table, and Pulumi YAML examples.
- Create: `site/source/_generated/provider-matrix.md`, `site/source/_generated/component-links.md`, `site/source/_generated/shared-cloud-init-table.md`, `site/source/_generated/component-yaml/*.md`
  - Generated docs snippets committed to the repo.
- Modify: `site/source/reference/provider-matrix.md`, `site/source/admin/component/index.md`, `site/source/admin/concepts/shared-cloud-init.md`, `site/source/admin/component/*.md`
  - Replace repeated provider facts with generated snippets or generated YAML blocks.
- Create: `internal/provider/catalog.go`
  - Mirror catalog identity, implementation mode, user-data mode, and validation metadata needed by Go.
- Create: `internal/provider/catalog_test.go`
  - Validate Go catalog tokens and registration parity.
- Modify: `internal/provider/components.go`
  - Add validation calls for catalog providers and refactor repetitive raw bootstrap provider construction only where the helper is clearer than custom code.
- Modify: `scripts/check-registry-readiness.mjs`
  - Consume catalog parity results or invoke `scripts/check-provider-catalog.mjs`.

## Task 1: Add TypeScript Provider Catalog

**Files:**
- Create: `src/providerCatalog.ts`
- Create: `test/providerCatalog.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Write failing catalog tests**

Create `test/providerCatalog.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import { providerCatalog, catalogProviders, catalogDrivenProviders, bespokeProviders } from "../src/providerCatalog";

test("provider catalog has unique names and tokens", () => {
  const names = new Set<string>();
  const tokens = new Set<string>();

  for (const provider of catalogProviders) {
    assert.equal(names.has(provider.componentName), false, `duplicate component ${provider.componentName}`);
    assert.equal(tokens.has(provider.token), false, `duplicate token ${provider.token}`);
    names.add(provider.componentName);
    tokens.add(provider.token);
  }
});

test("provider catalog includes current public components", () => {
  for (const componentName of [
    "AwsPublisher",
    "AzurePublisher",
    "GcpPublisher",
    "KubernetesPublisher",
    "VspherePublisher",
    "EsxiPublisher",
    "HcloudPublisher",
    "NutanixPublisher",
    "OpenstackPublisher",
    "OvhPublisher",
    "ScalewayPublisher",
    "OciPublisher",
    "AlicloudPublisher",
    "ProxmoxvePublisher",
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
    "HypervPublisher",
    "NetskopeRegistration",
  ]) {
    assert.ok(providerCatalog[componentName], `${componentName} missing from provider catalog`);
    assert.equal(providerCatalog[componentName].token, `netskope-publisher:index:${componentName}`);
  }
});

test("catalog-driven providers declare resource token, adapter, docs, and yaml example", () => {
  for (const provider of catalogDrivenProviders) {
    assert.ok(provider.resourceToken, `${provider.componentName} missing resourceToken`);
    assert.notEqual(provider.userData.mode, "none", `${provider.componentName} missing user-data mode`);
    assert.ok(provider.docs.summary, `${provider.componentName} missing docs summary`);
    assert.ok(provider.yamlExample.name, `${provider.componentName} missing yaml example name`);
    assert.ok(provider.yamlExample.properties.length > 0, `${provider.componentName} missing yaml properties`);
  }
});

test("bespoke providers are metadata-only", () => {
  for (const provider of bespokeProviders) {
    assert.equal(provider.implementation, "bespoke");
    assert.ok(provider.docs.summary, `${provider.componentName} missing docs summary`);
  }
});
```

- [ ] **Step 2: Run catalog tests to verify they fail**

Run: `npm run build && node --test dist/test/providerCatalog.test.js`

Expected: FAIL because `src/providerCatalog.ts` does not exist.

- [ ] **Step 3: Implement `src/providerCatalog.ts`**

Create `src/providerCatalog.ts`:

```ts
export type ProviderImplementationMode = "catalogRawVm" | "catalogTypedVm" | "catalogSpecializedVm" | "bespoke";
export type ProviderSupportStatus = "supported" | "experimental";
export type BootstrapModel = "bootstrapOnly" | "prebakedSupported" | "marketplaceSupported" | "helm" | "registrationOnly" | "experimental";
export type UserDataMode =
  | "none"
  | "plain"
  | "base64"
  | "metadata"
  | "customData"
  | "ociMetadata"
  | "scalewayDual"
  | "guestInfo"
  | "raw"
  | "proxmoxSnippet";

export interface ProviderInputMetadata {
  name: string;
  required?: boolean;
  secret?: boolean;
  summary: string;
  example?: unknown;
}

export interface ProviderValidationMetadata {
  required?: string[];
  requiredOneOf?: string[][];
  mutuallyExclusive?: string[][];
  experimentalOptInField?: string;
}

export interface ProviderYamlExample {
  name: string;
  properties: Array<[string, unknown]>;
}

export interface ProviderCatalogEntry {
  displayName: string;
  componentName: string;
  token: string;
  support: ProviderSupportStatus;
  providerPackage?: string;
  resourceToken?: string;
  implementation: ProviderImplementationMode;
  bootstrapModel: BootstrapModel;
  userData: {
    mode: UserDataMode;
    property?: string;
    metadataKey?: string;
  };
  inputs: {
    required: ProviderInputMetadata[];
    optional: ProviderInputMetadata[];
  };
  validation: ProviderValidationMetadata;
  docs: {
    slug: string;
    summary: string;
    bootstrapNotes?: string;
  };
  yamlExample: ProviderYamlExample;
}

function token(componentName: string): string {
  return `netskope-publisher:index:${componentName}`;
}

const entries = [
  {
    displayName: "AWS",
    componentName: "AwsPublisher",
    token: token("AwsPublisher"),
    support: "supported",
    providerPackage: "@pulumi/aws",
    implementation: "bespoke",
    bootstrapModel: "prebakedSupported",
    userData: { mode: "base64", property: "userDataBase64" },
    inputs: {
      required: [
        { name: "subnetId", required: true, summary: "EC2 subnet ID.", example: "subnet-0123456789abcdef0" },
        { name: "securityGroupIds", required: true, summary: "EC2 security group IDs.", example: ["sg-0123456789abcdef0"] },
      ],
      optional: [
        { name: "amiId", summary: "Publisher or Ubuntu 22.04 AMI ID." },
        { name: "instanceType", summary: "EC2 instance type.", example: "t3.medium" },
      ],
    },
    validation: { required: ["subnetId", "securityGroupIds"] },
    docs: { slug: "aws", summary: "EC2 instances with optional AMI lookup." },
    yamlExample: {
      name: "netskope-publisher-aws",
      properties: [
        ["namePrefix", "pub-eu"],
        ["replicas", 2],
        ["subnetId", "subnet-0123456789abcdef0"],
        ["securityGroupIds", ["sg-0123456789abcdef0"]],
        ["instanceType", "t3.medium"],
        ["bootstrap", true],
      ],
    },
  },
  {
    displayName: "DigitalOcean",
    componentName: "DigitaloceanPublisher",
    token: token("DigitaloceanPublisher"),
    support: "supported",
    providerPackage: "@pulumi/digitalocean",
    resourceToken: "digitalocean:index/droplet:Droplet",
    implementation: "catalogRawVm",
    bootstrapModel: "bootstrapOnly",
    userData: { mode: "plain", property: "userData" },
    inputs: {
      required: [{ name: "region", required: true, summary: "DigitalOcean region slug.", example: "ams3" }],
      optional: [
        { name: "size", summary: "Droplet size slug.", example: "s-2vcpu-4gb" },
        { name: "image", summary: "Ubuntu 22.04 image slug.", example: "ubuntu-22-04-x64" },
      ],
    },
    validation: { required: ["region"] },
    docs: { slug: "digitalocean", summary: "DigitalOcean Droplets with plain user data." },
    yamlExample: {
      name: "netskope-publisher-digitalocean",
      properties: [
        ["namePrefix", "pub"],
        ["replicas", 2],
        ["region", "ams3"],
        ["size", "s-2vcpu-4gb"],
        ["image", "ubuntu-22-04-x64"],
        ["bootstrap", true],
      ],
    },
  },
] satisfies ProviderCatalogEntry[];

const remainingComponentNames = [
  "AzurePublisher",
  "GcpPublisher",
  "KubernetesPublisher",
  "VspherePublisher",
  "EsxiPublisher",
  "HcloudPublisher",
  "NutanixPublisher",
  "OpenstackPublisher",
  "OvhPublisher",
  "ScalewayPublisher",
  "OciPublisher",
  "AlicloudPublisher",
  "ProxmoxvePublisher",
  "VultrPublisher",
  "ExoscalePublisher",
  "UpcloudPublisher",
  "StackitPublisher",
  "EquinixPublisher",
  "OutscalePublisher",
  "OpentelekomcloudPublisher",
  "TencentcloudPublisher",
  "YandexPublisher",
  "HypervPublisher",
  "NetskopeRegistration",
];

const generatedEntries = remainingComponentNames.map((componentName): ProviderCatalogEntry => ({
  displayName: componentName.replace(/Publisher$|Registration$/g, ""),
  componentName,
  token: token(componentName),
  support: componentName === "HypervPublisher" ? "experimental" : "supported",
  implementation: [
    "AwsPublisher",
    "AzurePublisher",
    "GcpPublisher",
    "KubernetesPublisher",
    "VspherePublisher",
    "EsxiPublisher",
    "HypervPublisher",
    "NetskopeRegistration",
  ].includes(componentName) ? "bespoke" : "catalogRawVm",
  bootstrapModel: componentName === "KubernetesPublisher" ? "helm" : componentName === "NetskopeRegistration" ? "registrationOnly" : "bootstrapOnly",
  userData: { mode: componentName === "NetskopeRegistration" || componentName === "KubernetesPublisher" ? "none" : "plain" },
  inputs: { required: [], optional: [] },
  validation: {},
  docs: { slug: componentName.replace(/Publisher$|Registration$/g, "").toLowerCase(), summary: `${componentName} metadata.` },
  yamlExample: { name: `netskope-${componentName.replace(/[A-Z]/g, (match, index) => `${index === 0 ? "" : "-"}${match.toLowerCase()}`)}`, properties: [["namePrefix", "pub"]] },
}));

export const catalogProviders = [...entries, ...generatedEntries] satisfies ProviderCatalogEntry[];

export const providerCatalog = Object.fromEntries(catalogProviders.map((provider) => [provider.componentName, provider])) as Record<string, ProviderCatalogEntry>;

export const catalogDrivenProviders = catalogProviders.filter((provider) => provider.implementation !== "bespoke");
export const bespokeProviders = catalogProviders.filter((provider) => provider.implementation === "bespoke");
```

After this task, Task 2 replaces the temporary generated coverage entries with complete provider metadata. The tests in this task establish the catalog contract and public component coverage.

- [ ] **Step 4: Export catalog helpers**

Modify `src/index.ts`:

```ts
export * from "./providerCatalog";
```

- [ ] **Step 5: Run catalog tests**

Run: `npm run build && node --test dist/test/providerCatalog.test.js`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add src/providerCatalog.ts src/index.ts test/providerCatalog.test.ts
git commit -m "feat: add provider catalog metadata"
```

## Task 2: Complete Catalog Metadata And Validation

**Files:**
- Modify: `src/providerCatalog.ts`
- Create: `src/providerValidation.ts`
- Create: `test/providerValidation.test.ts`

- [ ] **Step 1: Write failing validation tests**

Create `test/providerValidation.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import { validateProviderArgs } from "../src/providerValidation";

test("validateProviderArgs rejects missing required fields", () => {
  assert.throws(
    () => validateProviderArgs("DigitaloceanPublisher", { replicas: 1 }),
    /DigitaloceanPublisher requires input region/,
  );
});

test("validateProviderArgs accepts present required fields", () => {
  assert.doesNotThrow(() => validateProviderArgs("DigitaloceanPublisher", { region: "ams3" }));
});

test("validateProviderArgs rejects missing required-one-of fields", () => {
  assert.throws(
    () => validateProviderArgs("VultrPublisher", { region: "ams", plan: "vc2-2c-4gb" }),
    /VultrPublisher requires one of: osId, imageId/,
  );
});

test("validateProviderArgs rejects mutually exclusive fields", () => {
  assert.throws(
    () => validateProviderArgs("VultrPublisher", { region: "ams", plan: "vc2-2c-4gb", osId: 1743, imageId: "img-123" }),
    /VultrPublisher accepts only one of: osId, imageId/,
  );
});

test("validateProviderArgs enforces experimental opt-in", () => {
  assert.throws(
    () => validateProviderArgs("HypervPublisher", { switchName: "Default Switch", hardDrives: [] }),
    /HypervPublisher requires enableExperimentalHyperv: true/,
  );
});
```

- [ ] **Step 2: Run validation tests to verify they fail**

Run: `npm run build && node --test dist/test/providerValidation.test.js`

Expected: FAIL because `src/providerValidation.ts` does not exist.

- [ ] **Step 3: Complete catalog validation metadata**

Replace generated entries in `src/providerCatalog.ts` with explicit entries for every current component. Each catalog-driven provider must have exact required fields:

```ts
// Vultr example entry values to include in providerCatalog:
validation: {
  required: ["region", "plan"],
  requiredOneOf: [["osId", "imageId"]],
  mutuallyExclusive: [["osId", "imageId"]],
},
inputs: {
  required: [
    { name: "region", required: true, summary: "Vultr region slug.", example: "ams" },
    { name: "plan", required: true, summary: "Vultr plan slug.", example: "vc2-2c-4gb" },
  ],
  optional: [
    { name: "osId", summary: "Vultr OS ID for Ubuntu 22.04.", example: 1743 },
    { name: "imageId", summary: "Custom image ID." },
  ],
},
```

Minimum required-field metadata for all providers:

```ts
const requiredByProvider = {
  AwsPublisher: ["subnetId", "securityGroupIds"],
  AzurePublisher: ["resourceGroupName", "location", "subnetId", "adminSshPublicKey"],
  GcpPublisher: ["project", "zone", "network", "subnetwork", "image"],
  VspherePublisher: ["datacenter", "datastore", "networkName", "templateName"],
  EsxiPublisher: ["diskStore", "virtualNetwork"],
  HcloudPublisher: [],
  NutanixPublisher: ["clusterUuid"],
  OpenstackPublisher: ["imageName", "flavorName", "networkName"],
  OvhPublisher: ["serviceName", "region", "imageId", "flavorId"],
  ScalewayPublisher: [],
  OciPublisher: ["compartmentId", "availabilityDomain", "subnetId", "imageId"],
  AlicloudPublisher: ["imageId", "vswitchId", "securityGroupIds"],
  ProxmoxvePublisher: ["nodeName", "datastoreId", "templateVmId"],
  DigitaloceanPublisher: ["region"],
  VultrPublisher: ["region", "plan"],
  ExoscalePublisher: ["zone", "type", "templateId", "diskSize"],
  UpcloudPublisher: ["zone"],
  StackitPublisher: ["projectId", "machineType", "imageId"],
  EquinixPublisher: ["projectId", "metro", "plan"],
  OutscalePublisher: ["imageId"],
  OpentelekomcloudPublisher: ["networks"],
  TencentcloudPublisher: ["availabilityZone", "imageId"],
  YandexPublisher: ["imageId", "subnetId"],
  KubernetesPublisher: [],
  HypervPublisher: ["switchName", "hardDrives"],
  NetskopeRegistration: ["publisherNames"],
};
```

- [ ] **Step 4: Implement `src/providerValidation.ts`**

Create `src/providerValidation.ts`:

```ts
import { providerCatalog } from "./providerCatalog";

export function validateProviderArgs(componentName: string, args: Record<string, unknown>): void {
  const provider = providerCatalog[componentName];
  if (!provider) {
    throw new Error(`Unknown provider component ${componentName}`);
  }

  for (const field of provider.validation.required ?? []) {
    if (isMissing(args[field])) {
      throw new Error(`${componentName} requires input ${field}`);
    }
  }

  for (const group of provider.validation.requiredOneOf ?? []) {
    if (!group.some((field) => !isMissing(args[field]))) {
      throw new Error(`${componentName} requires one of: ${group.join(", ")}`);
    }
  }

  for (const group of provider.validation.mutuallyExclusive ?? []) {
    const present = group.filter((field) => !isMissing(args[field]));
    if (present.length > 1) {
      throw new Error(`${componentName} accepts only one of: ${group.join(", ")}`);
    }
  }

  const optInField = provider.validation.experimentalOptInField;
  if (optInField && args[optInField] !== true) {
    throw new Error(`${componentName} requires ${optInField}: true`);
  }
}

function isMissing(value: unknown): boolean {
  return value === undefined || value === null || value === "";
}
```

- [ ] **Step 5: Run validation tests**

Run: `npm run build && node --test dist/test/providerValidation.test.js`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add src/providerCatalog.ts src/providerValidation.ts test/providerValidation.test.ts
git commit -m "feat: validate provider catalog inputs"
```

## Task 3: Add Catalog Parity Check Script

**Files:**
- Create: `scripts/check-provider-catalog.mjs`
- Modify: `scripts/check-registry-readiness.mjs`
- Modify: `package.json`

- [ ] **Step 1: Write the parity script**

Create `scripts/check-provider-catalog.mjs`:

```js
import { existsSync, readFileSync } from "node:fs";
import { spawnSync } from "node:child_process";

const build = spawnSync("npm", ["run", "build"], { stdio: "inherit" });
if (build.status !== 0) {
  process.exit(build.status ?? 1);
}

const { catalogProviders } = await import("../dist/src/providerCatalog.js");
const schema = JSON.parse(readFileSync("schema.json", "utf8"));
const indexSource = readFileSync("src/index.ts", "utf8");
const goProviderSource = readFileSync("internal/provider/provider.go", "utf8");
const errors = [];

for (const provider of catalogProviders) {
  if (!schema.resources?.[provider.token]) {
    errors.push(`schema.json missing catalog token ${provider.token}`);
  }

  if (provider.componentName !== "NetskopeRegistration" && !indexSource.includes(`./${lowerFirst(provider.componentName)}`)) {
    errors.push(`src/index.ts missing export for ${provider.componentName}`);
  }

  if (provider.componentName !== "NetskopeRegistration" && !goProviderSource.includes(`New${provider.componentName}`) && !goProviderSource.includes(provider.componentName)) {
    errors.push(`internal/provider/provider.go missing Go registration for ${provider.componentName}`);
  }

  const docsPath = `site/source/admin/component/${provider.docs.slug}.md`;
  if (!existsSync(docsPath)) {
    errors.push(`Missing component docs page ${docsPath}`);
  } else {
    const docs = readFileSync(docsPath, "utf8");
    if (!docs.includes("## Pulumi YAML")) {
      errors.push(`${docsPath} missing Pulumi YAML example`);
    }
  }
}

if (errors.length > 0) {
  for (const error of errors) {
    console.error(`- ${error}`);
  }
  process.exit(1);
}

console.log("Provider catalog parity check passed.");

function lowerFirst(value) {
  return value.charAt(0).toLowerCase() + value.slice(1);
}
```

- [ ] **Step 2: Add npm script**

Modify `package.json`:

```json
"catalog:check": "node scripts/check-provider-catalog.mjs"
```

Also update `release:check` so `npm run catalog:check` runs after `npm run registry:check`.

- [ ] **Step 3: Wire registry readiness**

At the end of `scripts/check-registry-readiness.mjs`, before printing success, add:

```js
const catalogCheck = spawnSync("node", ["scripts/check-provider-catalog.mjs"], { stdio: "inherit" });
if (catalogCheck.status !== 0) {
  process.exit(catalogCheck.status ?? 1);
}
```

If this creates a recursion because `catalog:check` invokes `build`, keep `registry:check` invoking the catalog check directly and do not make `catalog:check` invoke `registry:check`.

- [ ] **Step 4: Run parity check**

Run: `npm run catalog:check`

Expected: PASS.

- [ ] **Step 5: Run registry check**

Run: `npm run registry:check`

Expected: PASS and includes `Provider catalog parity check passed.`

- [ ] **Step 6: Commit**

```bash
git add package.json scripts/check-provider-catalog.mjs scripts/check-registry-readiness.mjs
git commit -m "test: check provider catalog parity"
```

## Task 4: Generate Provider Docs Snippets And YAML Examples

**Files:**
- Create: `scripts/generate-provider-docs.mjs`
- Create: `site/source/_generated/provider-matrix.md`
- Create: `site/source/_generated/component-links.md`
- Create: `site/source/_generated/shared-cloud-init-table.md`
- Create: `site/source/_generated/component-yaml/*.md`
- Modify: `package.json`

- [ ] **Step 1: Write generator script**

Create `scripts/generate-provider-docs.mjs`:

```js
import { mkdirSync, writeFileSync } from "node:fs";
import { spawnSync } from "node:child_process";

const build = spawnSync("npm", ["run", "build"], { stdio: "inherit" });
if (build.status !== 0) {
  process.exit(build.status ?? 1);
}

const { catalogProviders } = await import("../dist/src/providerCatalog.js");

mkdirSync("site/source/_generated/component-yaml", { recursive: true });

const providerRows = [
  "| Platform | Component | Status | User-data mode |",
  "|---|---|---|---|",
  ...catalogProviders
    .filter((provider) => provider.componentName.endsWith("Publisher"))
    .map((provider) => `| ${provider.displayName} | \`${provider.componentName}\` | ${provider.support} | ${provider.userData.mode} |`),
].join("\\n");
writeFileSync("site/source/_generated/provider-matrix.md", `${providerRows}\\n`);

const componentLinks = catalogProviders
  .filter((provider) => provider.componentName.endsWith("Publisher"))
  .map((provider) => `- [${provider.displayName}](/pulumi-netskope-publisher/admin/component/${provider.docs.slug}/)`)
  .join("\\n");
writeFileSync("site/source/_generated/component-links.md", `${componentLinks}\\n`);

const adapterRows = [
  "| Component | User-data mode | Property |",
  "|---|---|---|",
  ...catalogProviders
    .filter((provider) => provider.userData.mode !== "none")
    .map((provider) => `| \`${provider.componentName}\` | ${provider.userData.mode} | ${provider.userData.property ?? provider.userData.metadataKey ?? ""} |`),
].join("\\n");
writeFileSync("site/source/_generated/shared-cloud-init-table.md", `${adapterRows}\\n`);

for (const provider of catalogProviders) {
  if (!provider.componentName.endsWith("Publisher")) {
    continue;
  }
  writeFileSync(`site/source/_generated/component-yaml/${provider.docs.slug}.md`, renderYaml(provider));
}

function renderYaml(provider) {
  const lines = [
    "## Pulumi YAML",
    "",
    "```yaml",
    `name: ${provider.yamlExample.name}`,
    "runtime: yaml",
    "config:",
    "  tenantUrl:",
    "    type: String",
    "  bearerToken:",
    "    type: String",
    "    secret: true",
    "resources:",
    "  publisher:",
    `    type: ${provider.token}`,
    "    properties:",
    "      tenantUrl: ${tenantUrl}",
    "      bearerToken: ${bearerToken}",
  ];

  for (const [key, value] of provider.yamlExample.properties) {
    appendProperty(lines, key, value, "      ");
  }

  lines.push("outputs:");
  lines.push("  publisherNames: ${publisher.publisherNames}");
  lines.push("  publishers: ${publisher.publishers}");
  lines.push("```");
  return `${lines.join("\\n")}\\n`;
}

function appendProperty(lines, key, value, indent) {
  if (Array.isArray(value)) {
    lines.push(`${indent}${key}:`);
    for (const item of value) {
      lines.push(`${indent}  - ${item}`);
    }
    return;
  }
  if (value && typeof value === "object") {
    lines.push(`${indent}${key}:`);
    for (const [childKey, childValue] of Object.entries(value)) {
      appendProperty(lines, childKey, childValue, `${indent}  `);
    }
    return;
  }
  lines.push(`${indent}${key}: ${value}`);
}
```

- [ ] **Step 2: Add docs generation script**

Modify `package.json`:

```json
"docs:gen": "node scripts/generate-provider-docs.mjs"
```

- [ ] **Step 3: Run generator**

Run: `npm run docs:gen`

Expected: generated files appear under `site/source/_generated`.

- [ ] **Step 4: Validate generated snippets**

Run: `test -s site/source/_generated/provider-matrix.md && test -s site/source/_generated/component-links.md && test -s site/source/_generated/shared-cloud-init-table.md`

Expected: exit code 0.

- [ ] **Step 5: Commit**

```bash
git add package.json scripts/generate-provider-docs.mjs site/source/_generated
git commit -m "docs: generate provider catalog snippets"
```

## Task 5: Migrate Simple Raw VM Providers To Catalog Factory

**Files:**
- Create: `src/catalogVmFactory.ts`
- Modify: `src/digitaloceanPublisher.ts`
- Modify: `src/vultrPublisher.ts`
- Modify: `src/exoscalePublisher.ts`
- Modify: `src/upcloudPublisher.ts`
- Modify: `src/stackitPublisher.ts`
- Modify: `src/equinixPublisher.ts`
- Modify: `src/outscalePublisher.ts`
- Modify: `src/opentelekomcloudPublisher.ts`
- Modify: `src/tencentcloudPublisher.ts`
- Modify: `src/yandexPublisher.ts`
- Modify: `test/additionalCloudPublishers.test.ts`

- [ ] **Step 1: Add failing factory test assertion**

In `test/additionalCloudPublishers.test.ts`, add one assertion to the DigitalOcean test:

```ts
assert.equal(droplet.name, "pub-1");
```

This should keep passing after the migration and confirms the factory preserves provider-specific mapping.

- [ ] **Step 2: Implement `src/catalogVmFactory.ts`**

Create `src/catalogVmFactory.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import { PublisherCatalogEntry } from "./providerCatalog";
import { RawResource } from "./rawResource";
import { validateProviderArgs } from "./providerValidation";
import { base64UserData, metadataUserData, plainUserData } from "./userDataAdapters";
import { CommonPublisherArgs, PublisherOutput } from "./types";
import { createVmPublishers, VmPublisherBuildInput, VmPublisherBuildResult } from "./vmPublisherCore";

export interface CatalogRawVmComponentArgs<TArgs extends CommonPublisherArgs> {
  parent: pulumi.ComponentResource;
  componentName: string;
  provider: PublisherCatalogEntry;
  args: TArgs;
  mapInputs: (input: VmPublisherBuildInput, args: TArgs) => pulumi.Inputs;
  mapOutputs?: (resource: RawResource) => VmPublisherBuildResult;
}

export function createCatalogRawVmPublishers<TArgs extends CommonPublisherArgs>(
  options: CatalogRawVmComponentArgs<TArgs>,
): {
  publisherNames: pulumi.Output<string[]>;
  publishers: pulumi.Output<Record<string, PublisherOutput>>;
} {
  validateProviderArgs(options.provider.componentName, options.args as Record<string, unknown>);

  return createVmPublishers({
    parent: options.parent,
    componentName: options.componentName,
    args: options.args,
    forceBootstrap: options.provider.bootstrapModel === "bootstrapOnly",
  }, (input) => {
    const resource = new RawResource(
      `${options.componentName}-${input.publisherName}`,
      options.provider.resourceToken!,
      options.mapInputs(input, options.args),
      { parent: options.parent },
    );
    return options.mapOutputs?.(resource) ?? { vmId: resource.id, privateIp: pulumi.output(""), publicIp: pulumi.output("") };
  });
}

export function userDataProperty(provider: PublisherCatalogEntry, input: VmPublisherBuildInput): Record<string, pulumi.Input<unknown>> {
  if (provider.userData.mode === "plain") {
    return { [provider.userData.property ?? "userData"]: plainUserData(input.userData) };
  }
  if (provider.userData.mode === "base64") {
    return { [provider.userData.property ?? "userData"]: base64UserData(input.userData) };
  }
  if (provider.userData.mode === "metadata") {
    return { [provider.userData.property ?? "metadata"]: metadataUserData(input.userData, provider.userData.metadataKey ?? "user-data") };
  }
  if (provider.userData.mode === "raw") {
    return { [provider.userData.property ?? "userDataRaw"]: plainUserData(input.userData) };
  }
  throw new Error(`${provider.componentName} cannot use catalog raw VM factory with user-data mode ${provider.userData.mode}`);
}
```

- [ ] **Step 3: Migrate DigitalOcean as reference**

Replace the constructor body in `src/digitaloceanPublisher.ts` after `super(...)`:

```ts
const provider = providerCatalog.DigitaloceanPublisher;
const outputs = createCatalogRawVmPublishers({
  parent: this,
  componentName: name,
  provider,
  args,
  mapInputs: (input, currentArgs) => ({
    name: input.publisherName,
    region: currentArgs.region,
    size: currentArgs.size ?? "s-2vcpu-4gb",
    image: currentArgs.image ?? "ubuntu-22-04-x64",
    sshKeys: currentArgs.sshKeys,
    vpcUuid: currentArgs.vpcUuid,
    monitoring: currentArgs.monitoring,
    ipv6: currentArgs.ipv6,
    ...userDataProperty(provider, input),
    tags: currentArgs.tags === undefined ? undefined : pulumi.output(currentArgs.tags).apply((tags) => Object.entries(tags).map(([key, value]) => `${key}:${value}`)),
  }),
});
```

Update imports:

```ts
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
```

Remove unused imports for `RawResource`, `plainUserData`, and `createVmPublishers`.

- [ ] **Step 4: Run DigitalOcean test**

Run: `npm test -- --test-name-pattern DigitaloceanPublisher`

Expected: PASS.

- [ ] **Step 5: Migrate the remaining simple raw VM providers**

Apply the same pattern to:

- `src/vultrPublisher.ts`
- `src/exoscalePublisher.ts`
- `src/upcloudPublisher.ts`
- `src/stackitPublisher.ts`
- `src/equinixPublisher.ts`
- `src/outscalePublisher.ts`
- `src/opentelekomcloudPublisher.ts`
- `src/tencentcloudPublisher.ts`
- `src/yandexPublisher.ts`

Use each file's existing property map exactly as the source for `mapInputs`. For Yandex, keep the existing `sshKeys` metadata merge inside `mapInputs` after spreading `userDataProperty(provider, input)`.

- [ ] **Step 6: Run raw provider tests**

Run: `npm test -- --test-name-pattern 'DigitaloceanPublisher|VultrPublisher|ExoscalePublisher|UpcloudPublisher|StackitPublisher|EquinixPublisher|OutscalePublisher|OpentelekomcloudPublisher|TencentcloudPublisher|YandexPublisher'`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add src/catalogVmFactory.ts src/digitaloceanPublisher.ts src/vultrPublisher.ts src/exoscalePublisher.ts src/upcloudPublisher.ts src/stackitPublisher.ts src/equinixPublisher.ts src/outscalePublisher.ts src/opentelekomcloudPublisher.ts src/tencentcloudPublisher.ts src/yandexPublisher.ts test/additionalCloudPublishers.test.ts
git commit -m "refactor: use catalog raw vm factory"
```

## Task 6: Add Go Catalog Parity

**Files:**
- Create: `internal/provider/catalog.go`
- Create: `internal/provider/catalog_test.go`
- Modify: `internal/provider/provider.go`

- [ ] **Step 1: Write Go catalog tests**

Create `internal/provider/catalog_test.go`:

```go
package provider

import "testing"

func TestProviderCatalogIncludesCurrentComponents(t *testing.T) {
	required := []string{
		"AwsPublisher",
		"AzurePublisher",
		"GcpPublisher",
		"KubernetesPublisher",
		"VspherePublisher",
		"EsxiPublisher",
		"HcloudPublisher",
		"NutanixPublisher",
		"OpenstackPublisher",
		"OvhPublisher",
		"ScalewayPublisher",
		"OciPublisher",
		"AlicloudPublisher",
		"ProxmoxvePublisher",
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
		"HypervPublisher",
	}

	for _, name := range required {
		entry, ok := providerCatalog[name]
		if !ok {
			t.Fatalf("%s missing from Go provider catalog", name)
		}
		if entry.Token != "netskope-publisher:index:"+name {
			t.Fatalf("%s token mismatch: %s", name, entry.Token)
		}
	}
}

func TestProviderCatalogValidationMetadata(t *testing.T) {
	entry := providerCatalog["DigitaloceanPublisher"]
	if len(entry.RequiredInputs) != 1 || entry.RequiredInputs[0] != "region" {
		t.Fatalf("DigitaloceanPublisher required inputs mismatch: %#v", entry.RequiredInputs)
	}

	hyperv := providerCatalog["HypervPublisher"]
	if hyperv.ExperimentalOptInField != "enableExperimentalHyperv" {
		t.Fatalf("HypervPublisher missing experimental opt-in metadata")
	}
}
```

- [ ] **Step 2: Run Go catalog tests to verify they fail**

Run: `go test ./internal/provider -run TestProviderCatalog`

Expected: FAIL because `providerCatalog` does not exist.

- [ ] **Step 3: Implement Go catalog**

Create `internal/provider/catalog.go`:

```go
package provider

type providerCatalogEntry struct {
	DisplayName            string
	ComponentName          string
	Token                  string
	Implementation         string
	UserDataMode           string
	RequiredInputs         []string
	ExperimentalOptInField string
}

var providerCatalog = map[string]providerCatalogEntry{
	"DigitaloceanPublisher": {DisplayName: "DigitalOcean", ComponentName: "DigitaloceanPublisher", Token: "netskope-publisher:index:DigitaloceanPublisher", Implementation: "catalogRawVm", UserDataMode: "plain", RequiredInputs: []string{"region"}},
	"VultrPublisher": {DisplayName: "Vultr", ComponentName: "VultrPublisher", Token: "netskope-publisher:index:VultrPublisher", Implementation: "catalogRawVm", UserDataMode: "plain", RequiredInputs: []string{"region", "plan"}},
	"HypervPublisher": {DisplayName: "Hyper-V", ComponentName: "HypervPublisher", Token: "netskope-publisher:index:HypervPublisher", Implementation: "bespoke", UserDataMode: "none", RequiredInputs: []string{"switchName", "hardDrives"}, ExperimentalOptInField: "enableExperimentalHyperv"},
}

func init() {
	defaultEntries := []string{
		"AwsPublisher", "AzurePublisher", "GcpPublisher", "KubernetesPublisher", "VspherePublisher", "EsxiPublisher",
		"HcloudPublisher", "NutanixPublisher", "OpenstackPublisher", "OvhPublisher", "ScalewayPublisher", "OciPublisher",
		"AlicloudPublisher", "ProxmoxvePublisher", "ExoscalePublisher", "UpcloudPublisher", "StackitPublisher",
		"EquinixPublisher", "OutscalePublisher", "OpentelekomcloudPublisher", "TencentcloudPublisher", "YandexPublisher",
	}
	for _, name := range defaultEntries {
		if _, ok := providerCatalog[name]; !ok {
			providerCatalog[name] = providerCatalogEntry{
				DisplayName:   name,
				ComponentName: name,
				Token:         "netskope-publisher:index:" + name,
				Implementation: "bespoke",
				UserDataMode:  "plain",
			}
		}
	}
}
```

- [ ] **Step 4: Run Go tests**

Run: `go test ./internal/provider -run TestProviderCatalog`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/provider/catalog.go internal/provider/catalog_test.go
git commit -m "test: add go provider catalog parity"
```

## Task 7: Integrate Docs Generation Into Site Workflow

**Files:**
- Modify: `site/package.json`
- Modify: `.github/workflows/pages.yml`
- Modify: `.github/workflows/ci.yml`
- Modify: `README.md`

- [ ] **Step 1: Add docs generation before Hexo build**

Modify `site/package.json`:

```json
"prebuild": "cd .. && npm run docs:gen"
```

Keep existing `build: "hexo generate"`.

- [ ] **Step 2: Update CI to verify generated docs are current**

Modify `.github/workflows/ci.yml` before the site build step:

```yaml
      - run: npm run docs:gen
```

- [ ] **Step 3: Update Pages workflow**

Modify `.github/workflows/pages.yml` before `npm run build` in the `site` directory:

```yaml
      - run: npm run docs:gen
```

- [ ] **Step 4: Document provider catalog maintenance**

Add to `README.md` near the development section:

```md
### Provider catalog maintenance

Provider capability metadata lives in `src/providerCatalog.ts`. Run
`npm run docs:gen` after changing provider metadata and run
`npm run catalog:check` before opening a release PR. The generated docs snippets
under `site/source/_generated/` are committed so GitHub Pages builds are
reproducible.
```

- [ ] **Step 5: Run docs generation and site build**

Run: `npm run docs:gen && npm run build --prefix site`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add README.md site/package.json .github/workflows/pages.yml .github/workflows/ci.yml site/source/_generated
git commit -m "ci: generate provider docs before site build"
```

## Task 8: Final Verification And SDK Regeneration

**Files:**
- Regenerate if schema changes: `schema.json`, `sdk/python`, `sdk/dotnet`, `sdk/go`, `sdk/java`, `sdk/rust`

- [ ] **Step 1: Run TypeScript checks**

Run: `npm run typecheck`

Expected: PASS.

- [ ] **Step 2: Run Node tests**

Run: `npm test`

Expected: PASS with all current tests plus new catalog tests.

- [ ] **Step 3: Run Go tests**

Run: `npm run go:test`

Expected: PASS.

- [ ] **Step 4: Run registry and catalog checks**

Run: `npm run registry:check && npm run catalog:check`

Expected: PASS.

- [ ] **Step 5: Build GitHub Pages site**

Run: `npm run build --prefix site`

Expected: PASS.

- [ ] **Step 6: Regenerate SDKs only if schema changed**

Run: `git diff --quiet schema.json || npm run sdk:gen`

Expected: If `schema.json` changed, SDKs regenerate successfully. If Gradle is unavailable locally, record that Java packaging remains a local environment limitation and rely on CI for Java build validation.

- [ ] **Step 7: Check worktree**

Run: `git status --short`

Expected: only intentional generated SDK changes, or clean if no schema changed.

- [ ] **Step 8: Commit final generated changes if any**

If Step 7 shows intentional source or generated SDK changes:

```bash
git add schema.json sdk/python sdk/dotnet sdk/go sdk/java sdk/rust
git commit -m "build: regenerate sdks for provider framework"
```

If there are no changes, do not create an empty commit.
