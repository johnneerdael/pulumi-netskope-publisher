# Provider Architecture Audit Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix provider architecture issues found in the audit by making catalog metadata accurate, validating Go and TypeScript providers consistently, and checking raw child provider tokens/properties against upstream Pulumi Registry schemas.

**Architecture:** Keep the existing catalog-driven model, but make the catalog own external provider package metadata and upstream schema expectations. Add a shared Go validation layer that mirrors TypeScript validation behavior, then call it from raw bootstrap constructors before child resources are created. Extend catalog checks to fetch upstream registry schemas so raw resource tokens, user-data properties, and documented package names cannot drift silently.

**Tech Stack:** TypeScript, Node test runner, Go `testing`, Pulumi Go provider integration mocks, Pulumi Registry schema JSON, GitHub Pages generated docs.

---

## File Structure

- Modify: `src/providerCatalog.ts`
  - Correct external provider package names for UpCloud, Stackit, Equinix, OpenTelekomCloud, Outscale, and Yandex.
  - Add per-provider `registrySchemaUrl` metadata so checks can verify upstream schemas.
- Modify: `test/providerCatalog.test.ts`
  - Add package metadata assertions for providers with audited mismatches.
  - Add schema URL assertions for every catalog-driven provider.
- Create: `test/providerRegistrySchema.test.ts`
  - Unit-test registry schema validation logic with in-memory schemas.
- Create: `scripts/provider-registry-schema-check.mjs`
  - Export pure helpers for validating provider catalog entries against upstream schemas.
  - Run as a CLI to fetch upstream Pulumi Registry schema JSON.
- Modify: `scripts/check-provider-catalog.mjs`
  - Invoke upstream schema validation after local catalog parity checks.
- Modify: `package.json`
  - Ensure `scripts/provider-registry-schema-check.mjs` is packaged.
- Modify: `internal/provider/catalog.go`
  - Add `RequiredOneOf`, `MutuallyExclusive`, and `ExperimentalOptInField` metadata.
- Create: `internal/provider/catalog_validation.go`
  - Implement Go-side catalog argument validation.
- Modify: `internal/provider/provider_test.go`
  - Add failing tests that prove Go rejects invalid Vultr and Hyper-V inputs through construct calls.
- Modify: `internal/provider/components.go`
  - Call Go catalog validation in raw provider constructors and Hyper-V constructor.
- Modify: `README.md`
  - Document the upstream schema audit guard and when to run it.

## Task 1: Correct Catalog Package Metadata

**Files:**
- Modify: `src/providerCatalog.ts`
- Modify: `test/providerCatalog.test.ts`

- [ ] **Step 1: Write failing package metadata tests**

Append this test to `test/providerCatalog.test.ts`:

```ts
test("provider catalog uses installable upstream package metadata", () => {
  const expectedPackages: Record<string, string> = {
    UpcloudPublisher: "@upcloud/pulumi-upcloud",
    StackitPublisher: "@stackitcloud/pulumi-stackit",
    EquinixPublisher: "@equinix-labs/pulumi-equinix",
    OpentelekomcloudPublisher: "terraform-provider:opentelekomcloud/opentelekomcloud",
    OutscalePublisher: "terraform-provider:outscale/outscale",
    TencentcloudPublisher: "terraform-provider:tencentcloudstack/tencentcloud",
    YandexPublisher: "pulumi/yandex",
  };

  for (const [componentName, packageName] of Object.entries(expectedPackages)) {
    assert.equal(providerCatalog[componentName].providerPackage, packageName, `${componentName} providerPackage mismatch`);
  }
});
```

- [ ] **Step 2: Run catalog tests to verify they fail**

Run:

```bash
npm run build && node --test dist/test/providerCatalog.test.js
```

Expected: FAIL with at least `UpcloudPublisher providerPackage mismatch`.

- [ ] **Step 3: Correct package metadata**

In `src/providerCatalog.ts`, replace the affected provider definitions with these exact definitions:

```ts
  provider({ displayName: "UpCloud", componentName: "UpcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "upcloud", required: ["zone"], resourceToken: "upcloud:index/server:Server", providerPackage: "@upcloud/pulumi-upcloud" }),
  provider({ displayName: "Stackit", componentName: "StackitPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "stackit", required: ["projectId", "machineType", "imageId"], resourceToken: "stackit:index/server:Server", providerPackage: "@stackitcloud/pulumi-stackit" }),
  provider({ displayName: "Equinix Metal", componentName: "EquinixPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "equinix", required: ["projectId", "metro", "plan"], resourceToken: "equinix:metal/device:Device", providerPackage: "@equinix-labs/pulumi-equinix" }),
  provider({ displayName: "Outscale", componentName: "OutscalePublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "outscale", required: ["imageId"], resourceToken: "outscale:index/vm:Vm", providerPackage: "terraform-provider:outscale/outscale" }),
  provider({ displayName: "OpenTelekomCloud", componentName: "OpentelekomcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "opentelekomcloud", required: ["networks"], resourceToken: "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", providerPackage: "terraform-provider:opentelekomcloud/opentelekomcloud" }),
  provider({ displayName: "TencentCloud", componentName: "TencentcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "raw", slug: "tencentcloud", required: ["availabilityZone", "imageId"], resourceToken: "tencentcloud:index/instance:Instance", providerPackage: "terraform-provider:tencentcloudstack/tencentcloud" }),
  provider({ displayName: "Yandex Cloud", componentName: "YandexPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "metadata", slug: "yandex", required: ["imageId", "subnetId"], resourceToken: "yandex:index/computeInstance:ComputeInstance", providerPackage: "pulumi/yandex" }),
```

- [ ] **Step 4: Run catalog tests**

Run:

```bash
npm run build && node --test dist/test/providerCatalog.test.js
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add src/providerCatalog.ts test/providerCatalog.test.ts
git commit -m "fix: correct provider catalog package metadata"
```

## Task 2: Add Upstream Registry Schema Metadata

**Files:**
- Modify: `src/providerCatalog.ts`
- Modify: `test/providerCatalog.test.ts`

- [ ] **Step 1: Write failing registry schema metadata tests**

In `test/providerCatalog.test.ts`, inside the existing `"catalog-driven providers declare resource token, adapter, docs, and yaml example"` test, add this assertion after the `resourceToken` assertion:

```ts
    assert.ok(provider.registrySchemaUrl, `${provider.componentName} missing registrySchemaUrl`);
```

Also append this test to `test/providerCatalog.test.ts`:

```ts
test("catalog-driven providers declare upstream registry schema URLs", () => {
  const expectedUrls: Record<string, string> = {
    HcloudPublisher: "https://www.pulumi.com/registry/packages/hcloud/schema.json",
    NutanixPublisher: "https://www.pulumi.com/registry/packages/nutanix/schema.json",
    OpenstackPublisher: "https://www.pulumi.com/registry/packages/openstack/schema.json",
    OvhPublisher: "https://www.pulumi.com/registry/packages/ovh/schema.json",
    ScalewayPublisher: "https://www.pulumi.com/registry/packages/scaleway/schema.json",
    OciPublisher: "https://www.pulumi.com/registry/packages/oci/schema.json",
    AlicloudPublisher: "https://www.pulumi.com/registry/packages/alicloud/schema.json",
    ProxmoxvePublisher: "https://www.pulumi.com/registry/packages/proxmoxve/schema.json",
    DigitaloceanPublisher: "https://www.pulumi.com/registry/packages/digitalocean/schema.json",
    VultrPublisher: "https://www.pulumi.com/registry/packages/vultr/schema.json",
    ExoscalePublisher: "https://www.pulumi.com/registry/packages/exoscale/schema.json",
    UpcloudPublisher: "https://www.pulumi.com/registry/packages/upcloud/schema.json",
    StackitPublisher: "https://www.pulumi.com/registry/packages/stackit/schema.json",
    EquinixPublisher: "https://www.pulumi.com/registry/packages/equinix/schema.json",
    OutscalePublisher: "https://www.pulumi.com/registry/packages/outscale/schema.json",
    OpentelekomcloudPublisher: "https://www.pulumi.com/registry/packages/opentelekomcloud/schema.json",
    TencentcloudPublisher: "https://www.pulumi.com/registry/packages/tencentcloud/schema.json",
    YandexPublisher: "https://www.pulumi.com/registry/packages/yandex/schema.json",
  };

  for (const [componentName, registrySchemaUrl] of Object.entries(expectedUrls)) {
    assert.equal(providerCatalog[componentName].registrySchemaUrl, registrySchemaUrl, `${componentName} registrySchemaUrl mismatch`);
  }
});
```

- [ ] **Step 2: Run catalog tests to verify they fail**

Run:

```bash
npm run build && node --test dist/test/providerCatalog.test.js
```

Expected: FAIL with `missing registrySchemaUrl`.

- [ ] **Step 3: Add `registrySchemaUrl` to catalog types and definitions**

In `src/providerCatalog.ts`, add this optional field to `ProviderCatalogEntry`:

```ts
  registrySchemaUrl?: string;
```

Add this optional field to `ProviderDefinition`:

```ts
  registrySchemaUrl?: string;
```

In `provider(definition: ProviderDefinition)`, add:

```ts
    registrySchemaUrl: definition.registrySchemaUrl,
```

Then add `registrySchemaUrl` to every catalog-driven provider definition. Example for DigitalOcean:

```ts
  provider({ displayName: "DigitalOcean", componentName: "DigitaloceanPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "digitalocean", required: ["region"], resourceToken: "digitalocean:index/droplet:Droplet", providerPackage: "@pulumi/digitalocean", registrySchemaUrl: "https://www.pulumi.com/registry/packages/digitalocean/schema.json", yamlProperties: [["namePrefix", "pub"], ["replicas", 2], ["region", "ams3"], ["size", "s-2vcpu-4gb"], ["image", "ubuntu-22-04-x64"], ["bootstrap", true]] }),
```

Use the URL map from Step 1 for all other catalog-driven providers.

- [ ] **Step 4: Export the new type field without broad re-export changes**

No `src/index.ts` changes are required because `ProviderCatalogEntry` is already exported as a type.

- [ ] **Step 5: Run catalog tests**

Run:

```bash
npm run build && node --test dist/test/providerCatalog.test.js
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add src/providerCatalog.ts test/providerCatalog.test.ts
git commit -m "feat: track upstream provider schemas"
```

## Task 3: Validate Upstream Registry Schemas

**Files:**
- Create: `scripts/provider-registry-schema-check.mjs`
- Create: `test/providerRegistrySchema.test.ts`
- Modify: `scripts/check-provider-catalog.mjs`
- Modify: `package.json`

- [ ] **Step 1: Write failing schema validation tests**

Create `test/providerRegistrySchema.test.ts`:

```ts
import assert from "node:assert/strict";
import test from "node:test";
import { validateProviderAgainstRegistrySchema } from "../scripts/provider-registry-schema-check.mjs";

const provider = {
  componentName: "ExamplePublisher",
  resourceToken: "example:index/server:Server",
  providerPackage: "@example/provider",
  userData: {
    mode: "plain",
    property: "userData",
  },
};

test("validateProviderAgainstRegistrySchema accepts matching token, package, and user-data property", () => {
  const errors = validateProviderAgainstRegistrySchema(provider, {
    name: "example",
    language: { nodejs: { packageName: "@example/provider" } },
    resources: {
      "example:index/server:Server": {
        inputProperties: {
          userData: { type: "string" },
        },
      },
    },
  });

  assert.deepEqual(errors, []);
});

test("validateProviderAgainstRegistrySchema rejects missing resource token", () => {
  const errors = validateProviderAgainstRegistrySchema(provider, {
    name: "example",
    language: { nodejs: { packageName: "@example/provider" } },
    resources: {},
  });

  assert.match(errors.join("\n"), /ExamplePublisher upstream schema missing resource token example:index\/server:Server/);
});

test("validateProviderAgainstRegistrySchema rejects missing user-data property", () => {
  const errors = validateProviderAgainstRegistrySchema(provider, {
    name: "example",
    language: { nodejs: { packageName: "@example/provider" } },
    resources: {
      "example:index/server:Server": {
        inputProperties: {
          metadata: { type: "object" },
        },
      },
    },
  });

  assert.match(errors.join("\n"), /ExamplePublisher upstream resource example:index\/server:Server missing user-data property userData/);
});

test("validateProviderAgainstRegistrySchema accepts terraform-provider package markers without node package comparison", () => {
  const errors = validateProviderAgainstRegistrySchema({
    ...provider,
    providerPackage: "terraform-provider:example/example",
  }, {
    name: "example",
    language: {},
    resources: {
      "example:index/server:Server": {
        inputProperties: {
          userData: { type: "string" },
        },
      },
    },
  });

  assert.deepEqual(errors, []);
});
```

- [ ] **Step 2: Run schema validation test to verify it fails**

Run:

```bash
npm run build && node --test dist/test/providerRegistrySchema.test.js
```

Expected: FAIL because `scripts/provider-registry-schema-check.mjs` does not exist or has no exported function.

- [ ] **Step 3: Create upstream schema checker**

Create `scripts/provider-registry-schema-check.mjs`:

```js
import { get } from "node:https";
import { spawnSync } from "node:child_process";

const userDataModesWithoutSingleProperty = new Set(["none", "scalewayDual", "proxmoxSnippet"]);

export function validateProviderAgainstRegistrySchema(provider, schema) {
  const errors = [];

  if (!provider.resourceToken) {
    return errors;
  }

  const resource = schema.resources?.[provider.resourceToken];
  if (!resource) {
    errors.push(`${provider.componentName} upstream schema missing resource token ${provider.resourceToken}`);
    return errors;
  }

  const expectedPackage = provider.providerPackage;
  const nodePackage = schema.language?.nodejs?.packageName;
  if (expectedPackage && expectedPackage.startsWith("@") && nodePackage && nodePackage !== expectedPackage) {
    errors.push(`${provider.componentName} providerPackage ${expectedPackage} does not match upstream node package ${nodePackage}`);
  }

  const userDataProperty = provider.userData?.property;
  if (userDataProperty && !userDataModesWithoutSingleProperty.has(provider.userData.mode) && !resource.inputProperties?.[userDataProperty]) {
    errors.push(`${provider.componentName} upstream resource ${provider.resourceToken} missing user-data property ${userDataProperty}`);
  }

  return errors;
}

export async function validateProvidersAgainstRegistrySchemas(catalogProviders) {
  const errors = [];

  for (const provider of catalogProviders) {
    if (!provider.registrySchemaUrl || !provider.resourceToken) {
      continue;
    }

    const schema = await fetchJson(provider.registrySchemaUrl);
    errors.push(...validateProviderAgainstRegistrySchema(provider, schema));
  }

  return errors;
}

async function fetchJson(url) {
  return new Promise((resolve, reject) => {
    get(url, { headers: { accept: "application/json" } }, (response) => {
      let body = "";
      response.setEncoding("utf8");
      response.on("data", (chunk) => {
        body += chunk;
      });
      response.on("end", () => {
        if (response.statusCode !== 200) {
          reject(new Error(`${url} returned HTTP ${response.statusCode}`));
          return;
        }
        try {
          resolve(JSON.parse(body));
        } catch (error) {
          reject(new Error(`${url} did not return valid JSON: ${error.message}`));
        }
      });
    }).on("error", reject);
  });
}

async function main() {
  const build = spawnSync("npm", ["run", "build"], { stdio: "inherit" });
  if (build.status !== 0) {
    process.exit(build.status ?? 1);
  }

  const { catalogDrivenProviders } = await import("../dist/src/providerCatalog.js");
  const errors = await validateProvidersAgainstRegistrySchemas(catalogDrivenProviders);
  if (errors.length > 0) {
    for (const error of errors) {
      console.error(`- ${error}`);
    }
    process.exit(1);
  }

  console.log("Provider registry schema check passed.");
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch((error) => {
    console.error(error.message);
    process.exit(1);
  });
}
```

- [ ] **Step 4: Add a declaration file for TypeScript tests**

Create `scripts/provider-registry-schema-check.d.ts`:

```ts
export function validateProviderAgainstRegistrySchema(provider: any, schema: any): string[];
export function validateProvidersAgainstRegistrySchemas(catalogProviders: any[]): Promise<string[]>;
```

- [ ] **Step 5: Wire the checker into local catalog check**

In `scripts/check-provider-catalog.mjs`, add this import:

```js
import { validateProvidersAgainstRegistrySchemas } from "./provider-registry-schema-check.mjs";
```

After the local provider loop and before `if (errors.length > 0)`, add:

```js
errors.push(...await validateProvidersAgainstRegistrySchemas(catalogProviders));
```

- [ ] **Step 6: Add checker files to npm package allowlist**

In `package.json`, add these entries under `files` after `scripts/package-sdks.mjs`:

```json
    "scripts/provider-registry-schema-check.mjs",
    "scripts/provider-registry-schema-check.d.ts",
```

- [ ] **Step 7: Run focused schema validation test**

Run:

```bash
npm run build && node --test dist/test/providerRegistrySchema.test.js
```

Expected: PASS.

- [ ] **Step 8: Run upstream catalog check**

Run:

```bash
npm run catalog:check
```

Expected: PASS and includes `Provider registry schema check passed.` followed by `Provider catalog parity check passed.`

- [ ] **Step 9: Commit**

```bash
git add scripts/provider-registry-schema-check.mjs scripts/provider-registry-schema-check.d.ts scripts/check-provider-catalog.mjs package.json test/providerRegistrySchema.test.ts
git commit -m "test: validate provider catalog against registry schemas"
```

## Task 4: Add Go Catalog Validation Metadata

**Files:**
- Modify: `internal/provider/catalog.go`
- Modify: `internal/provider/catalog_test.go`

- [ ] **Step 1: Write failing Go metadata tests**

Append this test to `internal/provider/catalog_test.go`:

```go
func TestProviderCatalogOneOfValidationMetadata(t *testing.T) {
	vultr := providerCatalog["VultrPublisher"]
	if len(vultr.RequiredOneOf) != 1 || len(vultr.RequiredOneOf[0]) != 2 {
		t.Fatalf("VultrPublisher missing required-one-of metadata: %#v", vultr.RequiredOneOf)
	}
	if vultr.RequiredOneOf[0][0] != "osId" || vultr.RequiredOneOf[0][1] != "imageId" {
		t.Fatalf("VultrPublisher required-one-of mismatch: %#v", vultr.RequiredOneOf)
	}
	if len(vultr.MutuallyExclusive) != 1 || len(vultr.MutuallyExclusive[0]) != 2 {
		t.Fatalf("VultrPublisher missing mutually-exclusive metadata: %#v", vultr.MutuallyExclusive)
	}
	if vultr.MutuallyExclusive[0][0] != "osId" || vultr.MutuallyExclusive[0][1] != "imageId" {
		t.Fatalf("VultrPublisher mutually-exclusive mismatch: %#v", vultr.MutuallyExclusive)
	}
}
```

- [ ] **Step 2: Run Go catalog tests to verify they fail**

Run:

```bash
go test ./internal/provider -run TestProviderCatalog
```

Expected: FAIL because `RequiredOneOf` and `MutuallyExclusive` fields do not exist.

- [ ] **Step 3: Add validation metadata fields**

In `internal/provider/catalog.go`, update `providerCatalogEntry`:

```go
type providerCatalogEntry struct {
	DisplayName            string
	ComponentName          string
	Token                  string
	Implementation         string
	UserDataMode           string
	RequiredInputs         []string
	RequiredOneOf          [][]string
	MutuallyExclusive      [][]string
	ExperimentalOptInField string
}
```

Replace the `VultrPublisher` entry with:

```go
	"VultrPublisher": {
		DisplayName:       "Vultr",
		ComponentName:     "VultrPublisher",
		Token:             "netskope-publisher:index:VultrPublisher",
		Implementation:    "catalogRawVm",
		UserDataMode:      "plain",
		RequiredInputs:    []string{"region", "plan"},
		RequiredOneOf:     [][]string{{"osId", "imageId"}},
		MutuallyExclusive: [][]string{{"osId", "imageId"}},
	},
```

- [ ] **Step 4: Run Go catalog tests**

Run:

```bash
go test ./internal/provider -run TestProviderCatalog
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/provider/catalog.go internal/provider/catalog_test.go
git commit -m "test: mirror catalog validation metadata in go"
```

## Task 5: Enforce Go Catalog Validation

**Files:**
- Create: `internal/provider/catalog_validation.go`
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/components.go`

- [ ] **Step 1: Write failing Go construct validation tests**

Add this helper near the existing `constructPublisherResource` helper in `internal/provider/provider_test.go`:

```go
func constructPublisherResourceError(t *testing.T, token string, inputs property.Map) error {
	t.Helper()

	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.3.0"),
		integration.WithProvider(provider),
		integration.WithMocks(&integration.MockResourceMonitor{
			NewResourceF: func(args integration.MockResourceArgs) (string, property.Map, error) {
				return args.Name + "-id", args.Inputs, nil
			},
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = server.Construct(p.ConstructRequest{
		Urn:    presource.URN("urn:pulumi:stack::project::" + token + "::publisher"),
		Inputs: inputs,
	})
	return err
}
```

Add these tests near `TestExpandedProviderConstructsBootstrapWithRegistryFields`:

```go
func TestVultrConstructRejectsMissingImageChoice(t *testing.T) {
	err := constructPublisherResourceError(t, "netskope-publisher:index:VultrPublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"region":        property.New("ams"),
		"plan":          property.New("vc2-2c-4gb"),
	}))
	if err == nil || !strings.Contains(err.Error(), "VultrPublisher requires one of: osId, imageId") {
		t.Fatalf("expected Vultr missing image choice error, got %v", err)
	}
}

func TestVultrConstructRejectsConflictingImageChoices(t *testing.T) {
	err := constructPublisherResourceError(t, "netskope-publisher:index:VultrPublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"region":        property.New("ams"),
		"plan":          property.New("vc2-2c-4gb"),
		"osId":          property.New(1743.0),
		"imageId":       property.New("img-123"),
	}))
	if err == nil || !strings.Contains(err.Error(), "VultrPublisher accepts only one of: osId, imageId") {
		t.Fatalf("expected Vultr conflicting image choice error, got %v", err)
	}
}
```

- [ ] **Step 2: Run Go validation tests to verify they fail**

Run:

```bash
go test ./internal/provider -run 'TestVultrConstructRejects'
```

Expected: FAIL because `NewVultrPublisher` currently accepts invalid inputs.

- [ ] **Step 3: Implement Go catalog validation**

Create `internal/provider/catalog_validation.go`:

```go
package provider

import (
	"fmt"
	"reflect"
	"strings"
)

func validateProviderCatalogArgs(componentName string, args any) error {
	entry, ok := providerCatalog[componentName]
	if !ok {
		return fmt.Errorf("unknown provider component %s", componentName)
	}

	value := reflect.ValueOf(args)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return fmt.Errorf("%s args must not be nil", componentName)
		}
		value = value.Elem()
	}

	for _, field := range entry.RequiredInputs {
		if isGoFieldMissing(value, field) {
			return fmt.Errorf("%s requires input %s", componentName, field)
		}
	}

	for _, group := range entry.RequiredOneOf {
		present := false
		for _, field := range group {
			if !isGoFieldMissing(value, field) {
				present = true
				break
			}
		}
		if !present {
			return fmt.Errorf("%s requires one of: %s", componentName, strings.Join(group, ", "))
		}
	}

	for _, group := range entry.MutuallyExclusive {
		present := 0
		for _, field := range group {
			if !isGoFieldMissing(value, field) {
				present++
			}
		}
		if present > 1 {
			return fmt.Errorf("%s accepts only one of: %s", componentName, strings.Join(group, ", "))
		}
	}

	if entry.ExperimentalOptInField != "" && !boolFieldIsTrue(value, entry.ExperimentalOptInField) {
		return fmt.Errorf("%s requires %s: true", componentName, entry.ExperimentalOptInField)
	}

	return nil
}

func isGoFieldMissing(value reflect.Value, pulumiName string) bool {
	field := value.FieldByName(goFieldName(pulumiName))
	if !field.IsValid() {
		return true
	}
	if field.Kind() == reflect.Pointer {
		return field.IsNil() || field.Elem().IsZero()
	}
	return field.IsZero()
}

func boolFieldIsTrue(value reflect.Value, pulumiName string) bool {
	field := value.FieldByName(goFieldName(pulumiName))
	if !field.IsValid() {
		return false
	}
	if field.Kind() == reflect.Bool {
		return field.Bool()
	}
	if field.Kind() == reflect.Pointer && !field.IsNil() && field.Elem().Kind() == reflect.Bool {
		return field.Elem().Bool()
	}
	return false
}

func goFieldName(pulumiName string) string {
	switch pulumiName {
	case "osId":
		return "OSID"
	case "imageId":
		return "ImageID"
	case "projectId":
		return "ProjectID"
	case "subnetId":
		return "SubnetID"
	case "vpcId":
		return "VpcID"
	case "enableExperimentalHyperv":
		return "EnableExperimentalHyperv"
	case "hardDrives":
		return "HardDrives"
	default:
		return strings.ToUpper(pulumiName[:1]) + pulumiName[1:]
	}
}
```

- [ ] **Step 4: Call validation in raw provider constructors**

In `internal/provider/components.go`, add this block immediately after each raw expanded provider component registration succeeds for:

- `NewDigitaloceanPublisher`
- `NewVultrPublisher`
- `NewExoscalePublisher`
- `NewUpcloudPublisher`
- `NewStackitPublisher`
- `NewEquinixPublisher`
- `NewOutscalePublisher`
- `NewOpentelekomcloudPublisher`
- `NewTencentcloudPublisher`
- `NewYandexPublisher`

Example for `NewVultrPublisher`:

```go
	if err := validateProviderCatalogArgs("VultrPublisher", args); err != nil {
		return nil, err
	}
```

Place it after:

```go
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
```

- [ ] **Step 5: Reuse validation in Hyper-V**

In `NewHypervPublisher`, replace the existing experimental opt-in `if` block with:

```go
	if err := validateProviderCatalogArgs("HypervPublisher", args); err != nil {
		return nil, err
	}
```

- [ ] **Step 6: Run Go validation tests**

Run:

```bash
go test ./internal/provider -run 'TestVultrConstructRejects|TestProviderCatalog'
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
git add internal/provider/catalog_validation.go internal/provider/components.go internal/provider/provider_test.go
git commit -m "fix: enforce catalog validation in go provider"
```

## Task 6: Document Provider Audit Guards

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update provider catalog maintenance docs**

In `README.md`, replace the existing `### Provider catalog maintenance` section with:

```md
### Provider catalog maintenance

Provider capability metadata lives in `src/providerCatalog.ts`. Run
`npm run docs:gen` after changing provider metadata and run
`npm run catalog:check` before opening a release PR.

`npm run catalog:check` validates local schema/export/docs parity and also
fetches upstream Pulumi Registry schema JSON for raw child providers. This
guards raw resource tokens, user-data placement properties, and install package
metadata that TypeScript and Go mocks cannot catch.

The generated docs snippets under `site/source/_generated/` are committed so
GitHub Pages builds are reproducible.
```

- [ ] **Step 2: Run docs generation**

Run:

```bash
npm run docs:gen
```

Expected: PASS.

- [ ] **Step 3: Run README-related checks**

Run:

```bash
npm run catalog:check
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add README.md site/source/_generated
git commit -m "docs: document provider catalog audit checks"
```

## Task 7: Final Verification

**Files:**
- Verify only; no planned source edits.

- [ ] **Step 1: Run TypeScript typecheck**

Run:

```bash
npm run typecheck
```

Expected: PASS.

- [ ] **Step 2: Run Node tests**

Run:

```bash
npm test
```

Expected: PASS.

- [ ] **Step 3: Run Go tests**

Run:

```bash
npm run go:test
```

Expected: PASS.

- [ ] **Step 4: Run registry and catalog checks**

Run:

```bash
npm run registry:check && npm run catalog:check
```

Expected: PASS.

- [ ] **Step 5: Build GitHub Pages site**

Run:

```bash
npm run build --prefix site
```

Expected: PASS.

- [ ] **Step 6: Check schema drift**

Run:

```bash
git diff --quiet schema.json || npm run sdk:gen
```

Expected: If `schema.json` changed, SDKs regenerate successfully. If it did not change, command exits 0 with no output.

- [ ] **Step 7: Check worktree**

Run:

```bash
git status --short
```

Expected: clean worktree after commits, or only intentional regenerated SDK changes.

- [ ] **Step 8: Commit regenerated SDKs if needed**

If Step 7 shows intentional schema or SDK changes:

```bash
git add schema.json sdk/python sdk/dotnet sdk/go sdk/java sdk/rust
git commit -m "build: regenerate sdks for provider audit fixes"
```

If Step 7 is clean, do not create an empty commit.

## Self-Review Notes

- Spec coverage: The plan addresses all four audit findings: wrong package metadata, missing upstream schema validation, missing Go validation parity, and documentation for the new guardrail.
- Placeholder scan: No `TBD`, `TODO`, or “similar to” implementation steps remain.
- Type consistency: TypeScript uses `registrySchemaUrl` consistently on `ProviderCatalogEntry`; Go validation uses `providerCatalogEntry.RequiredOneOf` and `MutuallyExclusive`; tests call `validateProviderAgainstRegistrySchema`.
