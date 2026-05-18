# Pulumi Netskope Publisher Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an AWS-first TypeScript Pulumi component package for Netskope Private Access Publishers, with examples, tests, CI, and GitHub Pages docs.

**Architecture:** The package exposes `AwsPublisher` from `src/index.ts`. `AwsPublisher` composes focused helpers for name derivation, Netskope registration, cloud-init rendering, and AWS EC2 infrastructure. The repository also includes a real `examples/aws-single` Pulumi program and a Hexo/Cactus documentation site adapted from the Terraform module.

**Tech Stack:** TypeScript, Node.js, Pulumi `@pulumi/pulumi`, Pulumi AWS provider `@pulumi/aws`, Node test runner, Hexo/Cactus docs, GitHub Actions.

---

## References

- Design spec: `docs/superpowers/specs/2026-05-18-pulumi-netskope-publisher-design.md`
- Terraform source repo: `/Users/jneerdael/Scripts/terraform-netskope-publisher`
- Pulumi source-based packages: https://www.pulumi.com/docs/iac/guides/building-extending/packages/source-based-plugin/
- `PulumiPlugin.yaml` reference: https://www.pulumi.com/docs/reference/pulumiplugin-yaml/
- Pulumi dynamic providers: https://www.pulumi.com/docs/iac/concepts/providers/dynamic-providers/

## File Structure

Create or modify these files:

- `package.json`: Node package metadata, scripts, dependencies.
- `package-lock.json`: npm lockfile.
- `tsconfig.json`: strict TypeScript config for `src`, `test`, and examples.
- `PulumiPlugin.yaml`: source-based component package manifest with `runtime: nodejs`.
- `README.md`: package overview, quick start, current AWS scope, docs link.
- `src/index.ts`: public exports.
- `src/types.ts`: shared argument and output interfaces.
- `src/names.ts`: pure publisher name derivation.
- `src/cloudInit.ts`: pure cloud-init rendering helpers.
- `src/netskopeClient.ts`: HTTP client wrapper for the Netskope API.
- `src/netskopeRegistration.ts`: Pulumi dynamic resource/provider and BYO registration normalization.
- `src/awsPublisher.ts`: public `AwsPublisher` component.
- `test/names.test.ts`: name derivation unit tests.
- `test/cloudInit.test.ts`: cloud-init unit tests.
- `test/netskopeClient.test.ts`: mocked API client tests.
- `test/netskopeRegistration.test.ts`: registration resource helper tests.
- `test/awsPublisher.test.ts`: Pulumi runtime mocks for component shape and secret outputs.
- `examples/aws-single/Pulumi.yaml`: example Pulumi project.
- `examples/aws-single/package.json`: example dependencies.
- `examples/aws-single/index.ts`: example program.
- `examples/aws-single/README.md`: example walkthrough.
- `site/package.json`, `site/_config.yml`, `site/source/**/*.md`: docs site.
- `.github/workflows/ci.yml`: install, typecheck, tests, docs build.
- `.github/workflows/pages.yml`: GitHub Pages build/deploy.

## Task 1: Package Scaffold

**Files:**
- Create: `package.json`
- Create: `tsconfig.json`
- Create: `PulumiPlugin.yaml`
- Modify: `README.md`

- [ ] **Step 1: Create package metadata**

Write `package.json`:

```json
{
  "name": "@johnneerdael/pulumi-netskope-publisher",
  "version": "0.1.0",
  "description": "Pulumi components for Netskope Private Access Publishers.",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "files": [
    "dist",
    "PulumiPlugin.yaml",
    "README.md",
    "LICENSE"
  ],
  "scripts": {
    "build": "tsc -p tsconfig.json",
    "typecheck": "tsc -p tsconfig.json --noEmit",
    "test": "npm run build && node --test dist/test/**/*.test.js",
    "clean": "rm -rf dist"
  },
  "keywords": [
    "pulumi",
    "netskope",
    "npa",
    "publisher",
    "aws"
  ],
  "author": "John Neerdael",
  "license": "Apache-2.0",
  "dependencies": {
    "@pulumi/aws": "^6.0.0",
    "@pulumi/pulumi": "^3.0.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0"
  },
  "engines": {
    "node": ">=20"
  }
}
```

- [ ] **Step 2: Create TypeScript config**

Write `tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "commonjs",
    "moduleResolution": "node",
    "declaration": true,
    "outDir": "dist",
    "rootDir": ".",
    "strict": true,
    "esModuleInterop": true,
    "forceConsistentCasingInFileNames": true,
    "skipLibCheck": true
  },
  "include": [
    "src/**/*.ts",
    "test/**/*.ts",
    "examples/**/*.ts"
  ]
}
```

- [ ] **Step 3: Create Pulumi source package manifest**

Write `PulumiPlugin.yaml`:

```yaml
runtime: nodejs
```

- [ ] **Step 4: Update README with the initial contract**

Replace `README.md` with:

```markdown
# pulumi-netskope-publisher

Pulumi components for provisioning Netskope Private Access Publishers.

The first version is AWS-first. It mirrors the Terraform AWS module from
`terraform-netskope-publisher`: register or reuse Netskope publisher
records, generate per-publisher cloud-init, and create EC2 instances.

## Current scope

- AWS publisher component: `AwsPublisher`
- Netskope publisher registration by name
- Bring-your-own registration data escape hatch
- GitHub Pages documentation

Azure, GCP, vSphere, and Hyper-V are planned follow-up providers.

## Development

```bash
npm install
npm run typecheck
npm test
```

## Example

See `examples/aws-single` for a Pulumi program that deploys one or more
AWS publishers.
```

- [ ] **Step 5: Install dependencies**

Run:

```bash
npm install
```

Expected: `package-lock.json` is created and npm exits successfully.

- [ ] **Step 6: Verify scaffold**

Run:

```bash
npm run typecheck
```

Expected: FAIL because `src` files do not exist yet or no inputs are found. This confirms scripts are wired; Task 2 creates the first source files.

- [ ] **Step 7: Commit**

Run:

```bash
git add package.json package-lock.json tsconfig.json PulumiPlugin.yaml README.md
git commit -m "chore: scaffold Pulumi package"
```

## Task 2: Shared Types And Name Derivation

**Files:**
- Create: `src/types.ts`
- Create: `src/names.ts`
- Create: `src/index.ts`
- Create: `test/names.test.ts`

- [ ] **Step 1: Write failing name tests**

Write `test/names.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import { derivePublisherNames } from "../src/names";

test("derivePublisherNames returns explicit names unchanged", () => {
  assert.deepEqual(
    derivePublisherNames({ namePrefix: "pub", names: ["alpha", "beta"], replicas: 3 }),
    ["alpha", "beta"],
  );
});

test("derivePublisherNames creates Terraform-compatible numbered names", () => {
  assert.deepEqual(
    derivePublisherNames({ namePrefix: "npa-publisher", replicas: 3 }),
    ["npa-publisher-1", "npa-publisher-2", "npa-publisher-3"],
  );
});

test("derivePublisherNames defaults to one npa-publisher", () => {
  assert.deepEqual(
    derivePublisherNames({}),
    ["npa-publisher-1"],
  );
});

test("derivePublisherNames rejects replicas below one", () => {
  assert.throws(
    () => derivePublisherNames({ replicas: 0 }),
    /replicas must be >= 1/,
  );
});

test("derivePublisherNames rejects an empty names list", () => {
  assert.throws(
    () => derivePublisherNames({ names: [] }),
    /names must contain at least one publisher name/,
  );
});
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
npm test
```

Expected: FAIL with TypeScript errors that `src/names` does not exist.

- [ ] **Step 3: Add shared types**

Write `src/types.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";

export interface NameArgs {
  namePrefix?: string;
  names?: string[];
  replicas?: number;
}

export interface MetadataOptions {
  httpEndpoint?: pulumi.Input<string>;
  httpTokens?: pulumi.Input<string>;
}

export interface PublisherRegistrationInput {
  publisherId: pulumi.Input<number>;
  registrationToken: pulumi.Input<string>;
}

export interface AwsPublisherArgs {
  namePrefix?: string;
  names?: string[];
  replicas?: number;
  tenantUrl?: pulumi.Input<string>;
  apiToken?: pulumi.Input<string>;
  wizardPath?: pulumi.Input<string>;
  tags?: pulumi.Input<Record<string, pulumi.Input<string>>>;
  registrations?: pulumi.Input<Record<string, PublisherRegistrationInput>>;
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

export interface PublisherOutput {
  publisherId: number;
  registrationToken: string;
  instanceId: string;
  privateIp: string;
  publicIp?: string;
}
```

- [ ] **Step 4: Add name derivation implementation**

Write `src/names.ts`:

```ts
import { NameArgs } from "./types";

export function derivePublisherNames(args: NameArgs): string[] {
  if (args.names !== undefined) {
    if (args.names.length === 0) {
      throw new Error("names must contain at least one publisher name");
    }

    return args.names;
  }

  const replicas = args.replicas ?? 1;
  if (replicas < 1) {
    throw new Error("replicas must be >= 1");
  }

  const namePrefix = args.namePrefix ?? "npa-publisher";
  return Array.from({ length: replicas }, (_, index) => `${namePrefix}-${index + 1}`);
}
```

- [ ] **Step 5: Add public exports**

Write `src/index.ts`:

```ts
export * from "./types";
export * from "./names";
```

- [ ] **Step 6: Run tests**

Run:

```bash
npm test
```

Expected: PASS for `names.test.ts`.

- [ ] **Step 7: Commit**

Run:

```bash
git add src/types.ts src/names.ts src/index.ts test/names.test.ts
git commit -m "feat: add publisher name derivation"
```

## Task 3: Cloud-Init Rendering

**Files:**
- Create: `src/cloudInit.ts`
- Create: `test/cloudInit.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Write failing cloud-init tests**

Write `test/cloudInit.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import { renderMetadata, renderUserData, renderUserDataBase64 } from "../src/cloudInit";

test("renderUserData matches Terraform cloud-init structure", () => {
  assert.equal(
    renderUserData({
      publisherName: "pub-1",
      registrationToken: "token-123",
      wizardPath: "/home/ubuntu/npa_publisher_wizard",
    }),
    [
      "#cloud-config",
      "hostname: pub-1",
      "preserve_hostname: false",
      "runcmd:",
      '  - [ /home/ubuntu/npa_publisher_wizard, -token, "token-123" ]',
      "",
    ].join("\n"),
  );
});

test("renderMetadata matches Terraform metadata structure", () => {
  assert.equal(
    renderMetadata("pub-1"),
    [
      "instance-id: pub-1",
      "local-hostname: pub-1",
      "",
    ].join("\n"),
  );
});

test("renderUserDataBase64 encodes rendered user data", () => {
  const raw = renderUserData({
    publisherName: "pub-1",
    registrationToken: "token-123",
    wizardPath: "/home/ubuntu/npa_publisher_wizard",
  });

  assert.equal(renderUserDataBase64({
    publisherName: "pub-1",
    registrationToken: "token-123",
    wizardPath: "/home/ubuntu/npa_publisher_wizard",
  }), Buffer.from(raw, "utf8").toString("base64"));
});
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
npm test
```

Expected: FAIL because `src/cloudInit` does not exist.

- [ ] **Step 3: Implement cloud-init rendering**

Write `src/cloudInit.ts`:

```ts
export interface RenderUserDataArgs {
  publisherName: string;
  registrationToken: string;
  wizardPath?: string;
}

export function renderUserData(args: RenderUserDataArgs): string {
  const wizardPath = args.wizardPath ?? "/home/ubuntu/npa_publisher_wizard";

  return [
    "#cloud-config",
    `hostname: ${args.publisherName}`,
    "preserve_hostname: false",
    "runcmd:",
    `  - [ ${wizardPath}, -token, "${escapeDoubleQuoted(args.registrationToken)}" ]`,
    "",
  ].join("\n");
}

export function renderUserDataBase64(args: RenderUserDataArgs): string {
  return Buffer.from(renderUserData(args), "utf8").toString("base64");
}

export function renderMetadata(publisherName: string): string {
  return [
    `instance-id: ${publisherName}`,
    `local-hostname: ${publisherName}`,
    "",
  ].join("\n");
}

function escapeDoubleQuoted(value: string): string {
  return value.replace(/\\/g, "\\\\").replace(/"/g, '\\"');
}
```

- [ ] **Step 4: Export cloud-init helpers**

Modify `src/index.ts`:

```ts
export * from "./types";
export * from "./names";
export * from "./cloudInit";
```

- [ ] **Step 5: Run tests**

Run:

```bash
npm test
```

Expected: PASS for `names.test.ts` and `cloudInit.test.ts`.

- [ ] **Step 6: Commit**

Run:

```bash
git add src/cloudInit.ts src/index.ts test/cloudInit.test.ts
git commit -m "feat: render publisher cloud-init"
```

## Task 4: Netskope API Client

**Files:**
- Create: `src/netskopeClient.ts`
- Create: `test/netskopeClient.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Write failing client tests**

Write `test/netskopeClient.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import { NetskopeClient } from "../src/netskopeClient";

test("listPublishers parses publisher IDs by name", async () => {
  const requests: Array<{ url: string; init: RequestInit }> = [];
  const client = new NetskopeClient({
    tenantUrl: "https://tenant.goskope.com/",
    apiToken: "secret",
    fetchImpl: async (url, init) => {
      requests.push({ url: String(url), init: init ?? {} });
      return response(200, {
        data: {
          publishers: [
            { publisher_name: "pub-a", publisher_id: "101" },
            { publisher_name: "pub-b", publisher_id: 102 },
          ],
        },
      });
    },
  });

  assert.deepEqual(await client.listPublishers(), {
    "pub-a": 101,
    "pub-b": 102,
  });
  assert.equal(requests[0].url, "https://tenant.goskope.com/api/v2/infrastructure/publishers");
  assert.equal((requests[0].init.headers as Record<string, string>)["Netskope-Api-Token"], "secret");
});

test("createPublisher returns created publisher ID", async () => {
  const client = new NetskopeClient({
    tenantUrl: "https://tenant.goskope.com",
    apiToken: "secret",
    fetchImpl: async () => response(201, { data: { id: "123" } }),
  });

  assert.equal(await client.createPublisher("pub-a"), 123);
});

test("generateRegistrationToken returns token", async () => {
  const client = new NetskopeClient({
    tenantUrl: "https://tenant.goskope.com",
    apiToken: "secret",
    fetchImpl: async () => response(200, { data: { token: "registration-token" } }),
  });

  assert.equal(await client.generateRegistrationToken(123), "registration-token");
});

test("client errors include operation and status without token", async () => {
  const client = new NetskopeClient({
    tenantUrl: "https://tenant.goskope.com",
    apiToken: "super-secret-token",
    fetchImpl: async () => response(403, { message: "forbidden" }),
  });

  await assert.rejects(
    () => client.listPublishers(),
    (error: unknown) => {
      assert.match(String(error), /List publishers failed \(status=403\)/);
      assert.doesNotMatch(String(error), /super-secret-token/);
      return true;
    },
  );
});

function response(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
npm test
```

Expected: FAIL because `src/netskopeClient` does not exist.

- [ ] **Step 3: Implement Netskope API client**

Write `src/netskopeClient.ts`:

```ts
export type FetchLike = typeof fetch;

export interface NetskopeClientArgs {
  tenantUrl: string;
  apiToken: string;
  fetchImpl?: FetchLike;
}

export class NetskopeClient {
  private readonly apiBase: string;
  private readonly apiToken: string;
  private readonly fetchImpl: FetchLike;

  constructor(args: NetskopeClientArgs) {
    this.apiBase = `${args.tenantUrl.replace(/\/+$/, "")}/api/v2/infrastructure/publishers`;
    this.apiToken = args.apiToken;
    this.fetchImpl = args.fetchImpl ?? fetch;
  }

  async listPublishers(): Promise<Record<string, number>> {
    const body = await this.request("List publishers", this.apiBase, { method: "GET" });
    const publishers = body?.data?.publishers ?? [];

    return Object.fromEntries(
      publishers.map((publisher: { publisher_name: string; publisher_id: string | number }) => [
        publisher.publisher_name,
        Number(publisher.publisher_id),
      ]),
    );
  }

  async createPublisher(name: string): Promise<number> {
    const body = await this.request(`Create publisher ${name}`, this.apiBase, {
      method: "POST",
      body: JSON.stringify({ name }),
    });

    return Number(body?.data?.id);
  }

  async generateRegistrationToken(publisherId: number): Promise<string> {
    const body = await this.request(
      `Generate registration token for publisher ${publisherId}`,
      `${this.apiBase}/${publisherId}/registration_token`,
      { method: "POST" },
    );

    return String(body?.data?.token);
  }

  private async request(operation: string, url: string, init: RequestInit): Promise<any> {
    const response = await this.fetchImpl(url, {
      ...init,
      headers: {
        "Netskope-Api-Token": this.apiToken,
        "Accept": "application/json",
        "Content-Type": "application/json",
        ...(init.headers ?? {}),
      },
    });

    const text = await response.text();
    const body = text.length > 0 ? JSON.parse(text) : undefined;

    if (response.status < 200 || response.status >= 300) {
      throw new Error(`${operation} failed (status=${response.status})`);
    }

    return body;
  }
}
```

- [ ] **Step 4: Export client**

Modify `src/index.ts`:

```ts
export * from "./types";
export * from "./names";
export * from "./cloudInit";
export * from "./netskopeClient";
```

- [ ] **Step 5: Run tests**

Run:

```bash
npm test
```

Expected: PASS for names, cloud-init, and client tests.

- [ ] **Step 6: Commit**

Run:

```bash
git add src/netskopeClient.ts src/index.ts test/netskopeClient.test.ts
git commit -m "feat: add Netskope API client"
```

## Task 5: Netskope Registration Resource

**Files:**
- Create: `src/netskopeRegistration.ts`
- Create: `test/netskopeRegistration.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Write failing registration helper tests**

Write `test/netskopeRegistration.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import { resolveRegistrations } from "../src/netskopeRegistration";

test("resolveRegistrations reuses existing publisher and creates missing publisher", async () => {
  const created: string[] = [];
  const tokens: number[] = [];

  const result = await resolveRegistrations({
    publisherNames: ["pub-a", "pub-b"],
    client: {
      listPublishers: async () => ({ "pub-a": 101 }),
      createPublisher: async (name: string) => {
        created.push(name);
        return 202;
      },
      generateRegistrationToken: async (publisherId: number) => {
        tokens.push(publisherId);
        return `token-${publisherId}`;
      },
    },
  });

  assert.deepEqual(created, ["pub-b"]);
  assert.deepEqual(tokens, [101, 202]);
  assert.deepEqual(result, {
    "pub-a": { publisherId: 101, registrationToken: "token-101", existedBefore: true },
    "pub-b": { publisherId: 202, registrationToken: "token-202", existedBefore: false },
  });
});
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
npm test
```

Expected: FAIL because `src/netskopeRegistration` does not exist.

- [ ] **Step 3: Implement registration helper and dynamic resource**

Write `src/netskopeRegistration.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import { NetskopeClient } from "./netskopeClient";

export interface RegistrationRecord {
  publisherId: number;
  registrationToken: string;
  existedBefore: boolean;
}

export interface RegistrationClient {
  listPublishers(): Promise<Record<string, number>>;
  createPublisher(name: string): Promise<number>;
  generateRegistrationToken(publisherId: number): Promise<string>;
}

export interface ResolveRegistrationsArgs {
  publisherNames: string[];
  client: RegistrationClient;
}

export async function resolveRegistrations(args: ResolveRegistrationsArgs): Promise<Record<string, RegistrationRecord>> {
  const existingByName = await args.client.listPublishers();
  const result: Record<string, RegistrationRecord> = {};

  for (const publisherName of args.publisherNames) {
    const existingId = existingByName[publisherName];
    const existedBefore = existingId !== undefined;
    const publisherId = existedBefore ? existingId : await args.client.createPublisher(publisherName);
    const registrationToken = await args.client.generateRegistrationToken(publisherId);

    result[publisherName] = {
      publisherId,
      registrationToken,
      existedBefore,
    };
  }

  return result;
}

export interface NetskopeRegistrationArgs {
  publisherNames: pulumi.Input<string[]>;
  tenantUrl: pulumi.Input<string>;
  apiToken: pulumi.Input<string>;
}

interface NetskopeRegistrationProviderInputs {
  publisherNames: string[];
  tenantUrl: string;
  apiToken: string;
}

interface NetskopeRegistrationProviderOutputs extends NetskopeRegistrationProviderInputs {
  registrations: Record<string, RegistrationRecord>;
}

class NetskopeRegistrationProvider implements pulumi.dynamic.ResourceProvider {
  async create(inputs: NetskopeRegistrationProviderInputs): Promise<pulumi.dynamic.CreateResult> {
    const registrations = await resolveRegistrations({
      publisherNames: inputs.publisherNames,
      client: new NetskopeClient({
        tenantUrl: inputs.tenantUrl,
        apiToken: inputs.apiToken,
      }),
    });

    return {
      id: inputs.publisherNames.join(","),
      outs: {
        ...inputs,
        registrations,
      },
    };
  }

  async diff(
    id: string,
    oldOutputs: NetskopeRegistrationProviderOutputs,
    newInputs: NetskopeRegistrationProviderInputs,
  ): Promise<pulumi.dynamic.DiffResult> {
    return {
      changes: JSON.stringify(oldOutputs.publisherNames) !== JSON.stringify(newInputs.publisherNames)
        || oldOutputs.tenantUrl !== newInputs.tenantUrl,
      replaces: ["publisherNames", "tenantUrl"],
    };
  }
}

export class NetskopeRegistration extends pulumi.dynamic.Resource {
  declare readonly registrations: pulumi.Output<Record<string, RegistrationRecord>>;

  constructor(name: string, args: NetskopeRegistrationArgs, opts?: pulumi.CustomResourceOptions) {
    super(new NetskopeRegistrationProvider(), name, {
      registrations: undefined,
      ...args,
    }, opts);
  }
}
```

- [ ] **Step 4: Export registration resource**

Modify `src/index.ts`:

```ts
export * from "./types";
export * from "./names";
export * from "./cloudInit";
export * from "./netskopeClient";
export * from "./netskopeRegistration";
```

- [ ] **Step 5: Run tests**

Run:

```bash
npm test
```

Expected: PASS for registration helper tests and all earlier tests.

- [ ] **Step 6: Commit**

Run:

```bash
git add src/netskopeRegistration.ts src/index.ts test/netskopeRegistration.test.ts
git commit -m "feat: manage Netskope publisher registration"
```

## Task 6: AWS Publisher Component

**Files:**
- Create: `src/awsPublisher.ts`
- Create: `test/awsPublisher.test.ts`
- Modify: `src/index.ts`

- [ ] **Step 1: Write failing component test with Pulumi mocks**

Write `test/awsPublisher.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { AwsPublisher } from "../src/awsPublisher";

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "aws:ec2/ami:Ami") {
      return { id: "ami-123", state: { id: "ami-123" } };
    }

    if (args.type === "aws:ec2/instance:Instance") {
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          privateIp: "10.0.0.10",
          publicIp: "198.51.100.10",
        },
      };
    }

    if (args.type === "pulumi-nodejs:dynamic:Resource") {
      return {
        id: "pub-1",
        state: {
          ...args.inputs,
          registrations: {
            "pub-1": {
              publisherId: 101,
              registrationToken: "token-101",
              existedBefore: true,
            },
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

test("AwsPublisher creates outputs keyed by publisher name", async () => {
  const component = new AwsPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    subnetId: "subnet-123",
    securityGroupIds: ["sg-123"],
  });

  const publisherNames = await outputValue(component.publisherNames);
  const publishers = await outputValue(component.publishers);

  assert.deepEqual(publisherNames, ["pub-1"]);
  assert.equal(publishers["pub-1"].publisherId, 101);
  assert.equal(publishers["pub-1"].instanceId, "publisher-pub-1-id");
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

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
npm test
```

Expected: FAIL because `src/awsPublisher` does not exist.

- [ ] **Step 3: Implement AWS component**

Write `src/awsPublisher.ts`:

```ts
import * as aws from "@pulumi/aws";
import * as pulumi from "@pulumi/pulumi";
import { renderUserDataBase64 } from "./cloudInit";
import { derivePublisherNames } from "./names";
import { NetskopeRegistration, RegistrationRecord } from "./netskopeRegistration";
import { AwsPublisherArgs, PublisherOutput, PublisherRegistrationInput } from "./types";

export class AwsPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: AwsPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope:index:AwsPublisher", name, {}, opts);

    const parentOpts = { parent: this };

    const publisherNames = derivePublisherNames({
      namePrefix: args.namePrefix,
      names: args.names,
      replicas: args.replicas,
    });

    this.publisherNames = pulumi.output(publisherNames);

    const registrations = args.registrations !== undefined
      ? pulumi.output(args.registrations).apply((byoRegistrations) =>
        normalizeByoRegistrations(publisherNames, byoRegistrations),
      )
      : createManagedRegistrations(name, publisherNames, args, parentOpts);

    const ami = args.amiId
      ? pulumi.output(args.amiId)
      : aws.ec2.getAmiOutput({
        mostRecent: true,
        owners: ["679593333241"],
        filters: [{
          name: "name",
          values: ["Netskope Private Access Publisher*"],
        }],
      }, parentOpts).id;

    const publisherOutputs: Record<string, pulumi.Output<PublisherOutput>> = {};

    for (const publisherName of publisherNames) {
      const registration = registrations.apply((allRegistrations) => allRegistrations[publisherName]);
      const userDataBase64 = pulumi.all([registration, args.wizardPath]).apply(([record, wizardPath]) =>
        renderUserDataBase64({
          publisherName,
          registrationToken: record.registrationToken,
          wizardPath,
        }),
      );

      const tags = pulumi.output(args.tags ?? {}).apply((inputTags) => ({
        ...inputTags,
        Name: publisherName,
      }));

      const instance = new aws.ec2.Instance(`${name}-${publisherName}`, {
        ami,
        instanceType: args.instanceType ?? "t3.medium",
        subnetId: args.subnetId,
        vpcSecurityGroupIds: args.securityGroupIds,
        keyName: args.keyName,
        associatePublicIpAddress: args.associatePublicIpAddress ?? false,
        iamInstanceProfile: args.iamInstanceProfile,
        ebsOptimized: args.ebsOptimized ?? true,
        monitoring: args.monitoring ?? true,
        userDataBase64,
        metadataOptions: pulumi.output(args.metadataOptions ?? {}).apply((metadataOptions) => ({
          httpEndpoint: metadataOptions.httpEndpoint ?? "enabled",
          httpTokens: metadataOptions.httpTokens ?? "required",
        })),
        tags,
      }, parentOpts);

      publisherOutputs[publisherName] = pulumi.all({
        registration,
        instanceId: instance.id,
        privateIp: instance.privateIp,
        publicIp: instance.publicIp,
      }).apply(({ registration, instanceId, privateIp, publicIp }) => ({
        publisherId: registration.publisherId,
        registrationToken: registration.registrationToken,
        instanceId,
        privateIp,
        publicIp,
      }));
    }

    this.publishers = pulumi.secret(pulumi.all(publisherOutputs));

    this.registerOutputs({
      publisherNames: this.publisherNames,
      publishers: this.publishers,
    });
  }
}

function createManagedRegistrations(
  name: string,
  publisherNames: string[],
  args: AwsPublisherArgs,
  opts: pulumi.CustomResourceOptions,
): pulumi.Output<Record<string, RegistrationRecord>> {
  if (args.tenantUrl === undefined || args.apiToken === undefined) {
    throw new Error("tenantUrl and apiToken are required when registrations are not provided");
  }

  return new NetskopeRegistration(`${name}-registration`, {
    publisherNames,
    tenantUrl: args.tenantUrl,
    apiToken: args.apiToken,
  }, opts).registrations;
}

function normalizeByoRegistrations(
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
```

- [ ] **Step 4: Export AWS component**

Modify `src/index.ts`:

```ts
export * from "./types";
export * from "./names";
export * from "./cloudInit";
export * from "./netskopeClient";
export * from "./netskopeRegistration";
export * from "./awsPublisher";
```

- [ ] **Step 5: Run tests**

Run:

```bash
npm test
```

Expected: PASS. The test confirms resources are registered synchronously in the component constructor, while dynamic values remain Pulumi outputs.

- [ ] **Step 6: Commit**

Run:

```bash
git add src/awsPublisher.ts src/index.ts test/awsPublisher.test.ts
git commit -m "feat: add AWS publisher component"
```

## Task 7: AWS Example Program

**Files:**
- Create: `examples/aws-single/Pulumi.yaml`
- Create: `examples/aws-single/package.json`
- Create: `examples/aws-single/index.ts`
- Create: `examples/aws-single/README.md`

- [ ] **Step 1: Create example Pulumi project**

Write `examples/aws-single/Pulumi.yaml`:

```yaml
name: aws-single
description: Deploy Netskope Private Access Publishers on AWS.
runtime: nodejs
```

- [ ] **Step 2: Create example package metadata**

Write `examples/aws-single/package.json`:

```json
{
  "name": "pulumi-netskope-publisher-aws-single-example",
  "private": true,
  "scripts": {
    "build": "tsc -p ../../tsconfig.json",
    "preview": "pulumi preview",
    "up": "pulumi up",
    "destroy": "pulumi destroy"
  },
  "dependencies": {
    "@pulumi/aws": "^6.0.0",
    "@pulumi/pulumi": "^3.0.0",
    "@johnneerdael/pulumi-netskope-publisher": "file:../.."
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0"
  }
}
```

- [ ] **Step 3: Create example program**

Write `examples/aws-single/index.ts`:

```ts
import * as pulumi from "@pulumi/pulumi";
import { AwsPublisher } from "@johnneerdael/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new AwsPublisher("publisher", {
  namePrefix: config.get("namePrefix") ?? "npa-publisher",
  replicas: config.getNumber("replicas") ?? 1,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  subnetId: config.require("subnetId"),
  securityGroupIds: config.requireObject<string[]>("securityGroupIds"),
  keyName: config.get("keyName") ?? undefined,
  amiId: config.get("amiId") ?? undefined,
  associatePublicIpAddress: config.getBoolean("associatePublicIpAddress") ?? false,
  tags: {
    Project: pulumi.getProject(),
    Stack: pulumi.getStack(),
  },
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

- [ ] **Step 4: Create example README**

Write `examples/aws-single/README.md`:

```markdown
# AWS Single Example

This example deploys one or more Netskope Private Access Publishers on
AWS using the local Pulumi component package.

## Configure

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set subnetId subnet-1234567890abcdef0
pulumi config set securityGroupIds '["sg-1234567890abcdef0"]'
pulumi config set replicas 1
```

Optional:

```bash
pulumi config set keyName my-key
pulumi config set amiId ami-1234567890abcdef0
pulumi config set associatePublicIpAddress false
```

## Deploy

```bash
npm install
pulumi preview
pulumi up
```

## Destroy

```bash
pulumi destroy
```
```

- [ ] **Step 5: Typecheck**

Run:

```bash
npm run typecheck
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```bash
git add examples/aws-single
git commit -m "docs: add AWS Pulumi example"
```

## Task 8: GitHub Pages Documentation

**Files:**
- Create: `site/package.json`
- Create: `site/_config.yml`
- Create: `site/source/index.md`
- Create: `site/source/starter/index.md`
- Create: `site/source/starter/01-what-youll-build.md`
- Create: `site/source/starter/02-install-tools.md`
- Create: `site/source/starter/03-aws-account-prep.md`
- Create: `site/source/starter/04-netskope-tenant-prep.md`
- Create: `site/source/starter/05-configure-stack.md`
- Create: `site/source/starter/06-first-publisher.md`
- Create: `site/source/starter/07-verify-online.md`
- Create: `site/source/starter/08-tear-down.md`
- Create: `site/source/admin/index.md`
- Create: `site/source/admin/component/index.md`
- Create: `site/source/admin/component/aws.md`
- Create: `site/source/admin/operations/secrets.md`
- Create: `site/source/reference/roadmap.md`

- [ ] **Step 1: Create docs package**

Write `site/package.json`:

```json
{
  "private": true,
  "scripts": {
    "build": "hexo generate",
    "clean": "hexo clean"
  },
  "dependencies": {
    "hexo": "^7.0.0",
    "hexo-generator-index": "^4.0.0",
    "hexo-generator-archive": "^2.0.0",
    "hexo-renderer-marked": "^6.0.0",
    "hexo-theme-cactus": "^2.1.0"
  }
}
```

- [ ] **Step 2: Create docs config**

Write `site/_config.yml`:

```yaml
title:       pulumi-netskope-publisher
subtitle:    Provision Netskope Publishers on AWS with Pulumi
description: Pulumi components for Netskope Private Access Publishers.
author:      John Neerdael
language:    en
timezone:    Europe/Amsterdam

url:         https://johnneerdael.github.io/pulumi-netskope-publisher
root:        /pulumi-netskope-publisher/
permalink:   :title/

source_dir:    source
public_dir:    public

default_layout: page
theme: cactus
```

- [ ] **Step 3: Create landing docs page**

Write `site/source/index.md`:

```markdown
---
title: pulumi-netskope-publisher
---

# pulumi-netskope-publisher

Provision Netskope Private Access Publishers on AWS with Pulumi.

Start with the [starter walkthrough](/pulumi-netskope-publisher/starter/)
or read the [AWS component reference](/pulumi-netskope-publisher/admin/component/aws/).

This first Pulumi version is AWS-first. Azure, GCP, vSphere, and Hyper-V
are tracked on the roadmap.
```

- [ ] **Step 4: Create starter index**

Write `site/source/starter/index.md`:

```markdown
---
title: Starter
---

# Starter

Follow these steps to deploy a first AWS Netskope publisher with Pulumi:

1. What you'll build
2. Install tools
3. Prepare AWS
4. Prepare Netskope
5. Configure the Pulumi stack
6. Deploy the first publisher
7. Verify it is online
8. Tear down
```

- [ ] **Step 5: Create starter pages**

Write each starter page with focused Pulumi-specific content:

`site/source/starter/01-what-youll-build.md`:

```markdown
---
title: What You'll Build
---

# What You'll Build

The AWS component creates one or more EC2 instances running the Netskope
Private Access Publisher image. During deployment it registers or reuses
publisher records in the Netskope tenant, generates registration tokens,
and passes those tokens to the publisher wizard through cloud-init.
```

`site/source/starter/02-install-tools.md`:

```markdown
---
title: Install Tools
---

# Install Tools

Install:

- Node.js 20 or newer
- npm
- Pulumi CLI
- AWS credentials with permission to create EC2 instances

Verify:

```bash
node --version
npm --version
pulumi version
aws sts get-caller-identity
```
```

`site/source/starter/03-aws-account-prep.md`:

```markdown
---
title: AWS Account Prep
---

# AWS Account Prep

Prepare an existing VPC subnet and security group for the publisher EC2
instance. The component expects `subnetId` and `securityGroupIds`.

The default image lookup searches for the latest AMI named
`Netskope Private Access Publisher*` owned by AWS account `679593333241`.
Use `amiId` to override this lookup.
```

`site/source/starter/04-netskope-tenant-prep.md`:

```markdown
---
title: Netskope Tenant Prep
---

# Netskope Tenant Prep

Create or obtain an API token that can list publishers, create
publishers, and generate publisher registration tokens. Store it in
Pulumi config as a secret.
```

`site/source/starter/05-configure-stack.md`:

```markdown
---
title: Configure Stack
---

# Configure Stack

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set subnetId subnet-1234567890abcdef0
pulumi config set securityGroupIds '["sg-1234567890abcdef0"]'
pulumi config set replicas 1
```
```

`site/source/starter/06-first-publisher.md`:

```markdown
---
title: First Publisher
---

# First Publisher

```bash
npm install
pulumi preview
pulumi up
```

Pulumi creates the Netskope publisher registration and the AWS EC2
instance in one deployment.
```

`site/source/starter/07-verify-online.md`:

```markdown
---
title: Verify Online
---

# Verify Online

Check the Netskope tenant for the publisher name emitted by Pulumi:

```bash
pulumi stack output publisherNames
```

The publisher should appear in Netskope after the instance boots and
cloud-init runs the registration wizard.
```

`site/source/starter/08-tear-down.md`:

```markdown
---
title: Tear Down
---

# Tear Down

```bash
pulumi destroy
```

Destroy removes the AWS infrastructure managed by Pulumi. Review
Netskope publisher records separately before deleting any tenant-side
publisher objects.
```

- [ ] **Step 6: Create admin/reference docs**

Write `site/source/admin/index.md`:

```markdown
---
title: Admin
---

# Admin

Admin docs cover component inputs, outputs, secret handling, state, and
operations for the Pulumi package.
```

Write `site/source/admin/component/index.md`:

```markdown
---
title: Components
---

# Components

The first package version exposes `AwsPublisher`.
```

Write `site/source/admin/component/aws.md`:

```markdown
---
title: AWS Component
---

# AWS Component

`AwsPublisher` creates one EC2 instance per publisher name.

## Required inputs

- `subnetId`
- `securityGroupIds`
- `tenantUrl` and `apiToken`, unless `registrations` is provided

## Common optional inputs

- `namePrefix`
- `names`
- `replicas`
- `amiId`
- `instanceType`
- `keyName`
- `tags`

## Outputs

- `publisherNames`
- `publishers`

`publishers` is keyed by publisher name and contains publisher ID,
registration token, EC2 instance ID, private IP, and public IP.
```

Write `site/source/admin/operations/secrets.md`:

```markdown
---
title: Secret Handling
---

# Secret Handling

Set the Netskope API token as a Pulumi secret:

```bash
pulumi config set apiToken --secret
```

Registration tokens are returned as secret outputs.
```

Write `site/source/reference/roadmap.md`:

```markdown
---
title: Roadmap
---

# Roadmap

The Pulumi package starts with AWS. Planned provider components:

- Azure
- GCP
- vSphere
- Hyper-V

These providers should reuse the shared name derivation, registration,
cloud-init, and output conventions.
```

- [ ] **Step 7: Build docs**

Run:

```bash
cd site
npm install
npm run build
```

Expected: `site/public` is generated successfully.

- [ ] **Step 8: Commit docs**

Run:

```bash
git add site
git commit -m "docs: add Pulumi GitHub Pages site"
```

## Task 9: CI And Pages Workflows

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/pages.yml`

- [ ] **Step 1: Add CI workflow**

Write `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
      - run: npm ci
      - run: npm run typecheck
      - run: npm test
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
          cache-dependency-path: site/package-lock.json
      - run: npm ci
        working-directory: site
      - run: npm run build
        working-directory: site
```

- [ ] **Step 2: Add Pages workflow**

Write `.github/workflows/pages.yml`:

```yaml
name: Pages

on:
  push:
    branches: [main]
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: false

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
          cache-dependency-path: site/package-lock.json
      - run: npm ci
        working-directory: site
      - run: npm run build
        working-directory: site
      - uses: actions/configure-pages@v5
      - uses: actions/upload-pages-artifact@v3
        with:
          path: site/public

  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    needs: build
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

- [ ] **Step 3: Run workflow-equivalent checks locally**

Run:

```bash
npm ci
npm run typecheck
npm test
```

Expected: PASS.

Run:

```bash
cd site
npm ci
npm run build
```

Expected: PASS.

- [ ] **Step 4: Commit workflows**

Run:

```bash
git add .github/workflows/ci.yml .github/workflows/pages.yml
git commit -m "ci: test package and publish docs"
```

## Task 10: Final Verification And README Polish

**Files:**
- Modify: `README.md`
- Inspect: all files changed by Tasks 1-9

- [ ] **Step 1: Expand README quick start**

Update `README.md` to include this usage example:

```markdown
## Quick start

```ts
import * as pulumi from "@pulumi/pulumi";
import { AwsPublisher } from "@johnneerdael/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new AwsPublisher("publisher", {
  namePrefix: "pub-eu",
  replicas: 2,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  subnetId: config.require("subnetId"),
  securityGroupIds: config.requireObject<string[]>("securityGroupIds"),
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

## Documentation

Full guides are published with GitHub Pages:

https://johnneerdael.github.io/pulumi-netskope-publisher/
```
```

- [ ] **Step 2: Run final package checks**

Run:

```bash
npm run clean
npm ci
npm run typecheck
npm test
```

Expected: PASS.

- [ ] **Step 3: Run final docs checks**

Run:

```bash
cd site
npm ci
npm run build
```

Expected: PASS and `site/public` exists.

- [ ] **Step 4: Check git diff**

Run:

```bash
git status --short
git diff --check
```

Expected: only intentional README changes are unstaged or staged; `git diff --check` exits with no whitespace errors.

- [ ] **Step 5: Commit final polish**

Run:

```bash
git add README.md
git commit -m "docs: document Pulumi package usage"
```

- [ ] **Step 6: Record final verification**

Before handing off, report:

- The last commit hash.
- Whether `npm test` passed.
- Whether docs build passed.
- Any known gaps, especially that Azure/GCP/vSphere/Hyper-V and public Pulumi Registry publishing are intentionally deferred.
