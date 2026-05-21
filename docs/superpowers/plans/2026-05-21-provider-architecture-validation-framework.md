# Provider Architecture Validation Framework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the provider architecture audit findings by making catalog validation apply to every publisher component, making Proxmox VE multi-resource bootstrap schema checks explicit, and replacing fragile Go reflection name mapping with Pulumi tag-based validation.

**Architecture:** Keep the existing provider catalog as the source of provider capability metadata, but make every component constructor participate in that metadata instead of only the raw-resource factory. Extend registry schema checks from one resource/property pair to explicit resource/property-path checks so composite providers like Proxmox VE can validate both the snippet resource and VM reference. In Go, use `pulumi` struct tags to resolve catalog input names to fields, then call validation consistently across all VM publisher constructors.

**Tech Stack:** TypeScript, Node test runner, Pulumi Node mocks, Go `testing`, `pulumi-go-provider` integration mocks, Pulumi Registry schema JSON, npm verification scripts.

---

## File Structure

- Modify: `src/providerValidation.ts`
  - Keep catalog validation rules here.
  - Add `validateComponentArgs(componentName, args)` as the constructor-friendly API.
- Modify: all TypeScript publisher components under `src/*Publisher.ts`
  - Add one constructor validation call for every component present in `providerCatalog`.
  - Remove the duplicate raw-factory-only validation once every constructor validates itself.
- Modify: `src/catalogVmFactory.ts`
  - Stop being the only validation entry point for raw-resource providers.
- Modify: `test/nutanixPublisher.test.ts`
  - Prove a direct `createVmPublishers` component rejects missing catalog-required inputs.
- Modify: `test/proxmoxvePublisher.test.ts`
  - Prove Proxmox VE rejects missing catalog-required inputs before child resources are built.
- Modify: `src/providerCatalog.ts`
  - Add explicit `registrySchemaChecks` metadata.
  - Add Proxmox VE checks for `FileLegacy.sourceRaw.data` and `VmLegacy.initialization.userDataFileId`.
- Modify: `src/providerRegistrySchema.ts`
  - Validate nested registry schema paths by resolving local `#/types/...` references.
  - Use explicit `registrySchemaChecks` when present, and keep fallback behavior for simple providers.
- Modify: `test/providerRegistrySchema.test.ts`
  - Add nested `$ref` schema validation tests.
- Modify: `scripts/generate-provider-docs.mjs`
  - Render explicit schema paths for composite user-data placement docs.
- Modify: `internal/provider/catalog_validation.go`
  - Replace `goFieldName` with `pulumi` struct tag lookup.
- Modify: `internal/provider/catalog_test.go`
  - Add validation metadata coverage for a direct-provider component such as Proxmox VE.
- Modify: `internal/provider/provider_test.go`
  - Add construct-time rejection tests for Proxmox VE and OCI.
- Modify: `internal/provider/components.go`
  - Call Go catalog validation in all VM publisher constructors with catalog-required fields, not only the later raw-bootstrap block.
- Modify: `README.md` and generated docs only if `npm run docs:gen` produces meaningful diffs.

## Task 1: Centralize TypeScript Constructor Validation

**Files:**
- Modify: `src/providerValidation.ts`
- Modify: `src/catalogVmFactory.ts`
- Modify: `src/nutanixPublisher.ts`
- Modify: `src/proxmoxvePublisher.ts`
- Modify: `src/openstackPublisher.ts`
- Modify: `src/ovhPublisher.ts`
- Modify: `src/ociPublisher.ts`
- Modify: `src/alicloudPublisher.ts`
- Modify: `src/hcloudPublisher.ts`
- Modify: `src/scalewayPublisher.ts`
- Modify: `src/awsPublisher.ts`
- Modify: `src/azurePublisher.ts`
- Modify: `src/gcpPublisher.ts`
- Modify: `src/vspherePublisher.ts`
- Modify: `src/esxiPublisher.ts`
- Modify: `src/hypervPublisher.ts`
- Modify: `src/kubernetesPublisher.ts`
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
- Modify: `test/nutanixPublisher.test.ts`
- Modify: `test/proxmoxvePublisher.test.ts`

- [ ] **Step 1: Write failing tests for direct-provider validation gaps**

Append this test to `test/nutanixPublisher.test.ts`:

```ts
test("NutanixPublisher rejects missing catalog-required clusterUuid", () => {
  assert.throws(
    () => new NutanixPublisher("missing-cluster", {
      names: ["pub-1"],
      tenantUrl: "https://tenant.goskope.com",
      apiToken: pulumi.secret("api-token"),
    } as any),
    /NutanixPublisher requires input clusterUuid/,
  );
});
```

Append this test to `test/proxmoxvePublisher.test.ts`:

```ts
test("ProxmoxvePublisher rejects missing catalog-required templateVmId", () => {
  assert.throws(
    () => new ProxmoxvePublisher("missing-template", {
      names: ["pub-1"],
      tenantUrl: "https://tenant.goskope.com",
      apiToken: pulumi.secret("api-token"),
      nodeName: "pve-1",
      datastoreId: "local",
    } as any),
    /ProxmoxvePublisher requires input templateVmId/,
  );
});
```

- [ ] **Step 2: Run the focused TypeScript tests and verify they fail**

Run:

```bash
npm run build && node --test dist/test/nutanixPublisher.test.js dist/test/proxmoxvePublisher.test.js
```

Expected: FAIL because both constructors currently use `createVmPublishers` directly and never call `validateProviderArgs`.

- [ ] **Step 3: Add a constructor-friendly validation wrapper**

In `src/providerValidation.ts`, add this exported function above `validateProviderArgs`:

```ts
export function validateComponentArgs(componentName: string, args: unknown): void {
  validateProviderArgs(componentName, args as Record<string, unknown>);
}
```

Keep the existing `validateProviderArgs` implementation unchanged.

- [ ] **Step 4: Remove raw-factory-only validation**

In `src/catalogVmFactory.ts`, remove this import:

```ts
import { validateProviderArgs } from "./providerValidation";
```

Then remove this line from `createCatalogRawVmPublishers`:

```ts
  validateProviderArgs(options.provider.componentName, options.args as Record<string, unknown>);
```

The constructor-level calls added in the next step will own validation for both raw and typed providers.

- [ ] **Step 5: Add constructor validation imports and calls**

For every file listed below, add this import:

```ts
import { validateComponentArgs } from "./providerValidation";
```

Then add the listed validation call immediately after the `super(...)` call in the constructor.

| File | Validation call |
|---|---|
| `src/awsPublisher.ts` | `validateComponentArgs("AwsPublisher", args);` |
| `src/azurePublisher.ts` | `validateComponentArgs("AzurePublisher", args);` |
| `src/gcpPublisher.ts` | `validateComponentArgs("GcpPublisher", args);` |
| `src/kubernetesPublisher.ts` | `validateComponentArgs("KubernetesPublisher", args);` |
| `src/vspherePublisher.ts` | `validateComponentArgs("VspherePublisher", args);` |
| `src/esxiPublisher.ts` | `validateComponentArgs("EsxiPublisher", args);` |
| `src/hcloudPublisher.ts` | `validateComponentArgs("HcloudPublisher", args);` |
| `src/nutanixPublisher.ts` | `validateComponentArgs("NutanixPublisher", args);` |
| `src/openstackPublisher.ts` | `validateComponentArgs("OpenstackPublisher", args);` |
| `src/ovhPublisher.ts` | `validateComponentArgs("OvhPublisher", args);` |
| `src/scalewayPublisher.ts` | `validateComponentArgs("ScalewayPublisher", args);` |
| `src/ociPublisher.ts` | `validateComponentArgs("OciPublisher", args);` |
| `src/alicloudPublisher.ts` | `validateComponentArgs("AlicloudPublisher", args);` |
| `src/proxmoxvePublisher.ts` | `validateComponentArgs("ProxmoxvePublisher", args);` |
| `src/digitaloceanPublisher.ts` | `validateComponentArgs("DigitaloceanPublisher", args);` |
| `src/vultrPublisher.ts` | `validateComponentArgs("VultrPublisher", args);` |
| `src/exoscalePublisher.ts` | `validateComponentArgs("ExoscalePublisher", args);` |
| `src/upcloudPublisher.ts` | `validateComponentArgs("UpcloudPublisher", args);` |
| `src/stackitPublisher.ts` | `validateComponentArgs("StackitPublisher", args);` |
| `src/equinixPublisher.ts` | `validateComponentArgs("EquinixPublisher", args);` |
| `src/outscalePublisher.ts` | `validateComponentArgs("OutscalePublisher", args);` |
| `src/opentelekomcloudPublisher.ts` | `validateComponentArgs("OpentelekomcloudPublisher", args);` |
| `src/tencentcloudPublisher.ts` | `validateComponentArgs("TencentcloudPublisher", args);` |
| `src/yandexPublisher.ts` | `validateComponentArgs("YandexPublisher", args);` |
| `src/hypervPublisher.ts` | `validateComponentArgs("HypervPublisher", args);` |

For `src/hypervPublisher.ts`, replace the existing explicit opt-in check:

```ts
    if (args.enableExperimentalHyperv !== true) {
      throw new Error("Hyper-V support is experimental and requires enableExperimentalHyperv: true");
    }
```

with:

```ts
    validateComponentArgs("HypervPublisher", args);
```

- [ ] **Step 6: Run focused TypeScript validation tests**

Run:

```bash
npm run build && node --test dist/test/providerValidation.test.js dist/test/nutanixPublisher.test.js dist/test/proxmoxvePublisher.test.js
```

Expected: PASS.

- [ ] **Step 7: Run all Node tests**

Run:

```bash
npm test
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add src test/nutanixPublisher.test.ts test/proxmoxvePublisher.test.ts
git commit -m "fix: validate all typescript publisher constructors"
```

## Task 2: Model Proxmox VE Multi-Resource Registry Schema Checks

**Files:**
- Modify: `src/providerCatalog.ts`
- Modify: `src/providerRegistrySchema.ts`
- Modify: `test/providerRegistrySchema.test.ts`
- Modify: `test/providerCatalog.test.ts`
- Modify: `scripts/generate-provider-docs.mjs`
- Verify generated docs under `site/source/_generated/`

- [ ] **Step 1: Write failing registry schema tests for nested checks**

Append this test to `test/providerRegistrySchema.test.ts`:

```ts
test("validateProviderAgainstRegistrySchema accepts explicit nested schema checks", () => {
  const errors = validateProviderAgainstRegistrySchema({
    componentName: "CompositePublisher",
    providerPackage: "@example/provider",
    resourceToken: "example:index/vm:Vm",
    userData: {
      mode: "proxmoxSnippet",
    },
    registrySchemaChecks: [{
      resourceToken: "example:index/file:File",
      propertyPath: ["sourceRaw", "data"],
      description: "cloud-init snippet content",
    }, {
      resourceToken: "example:index/vm:Vm",
      propertyPath: ["initialization", "userDataFileId"],
      description: "VM cloud-init user-data file reference",
    }],
  }, {
    name: "example",
    language: { nodejs: { packageName: "@example/provider" } },
    resources: {
      "example:index/file:File": {
        inputProperties: {
          sourceRaw: { "$ref": "#/types/example:index/FileSourceRaw:FileSourceRaw" },
        },
      },
      "example:index/vm:Vm": {
        inputProperties: {
          initialization: { "$ref": "#/types/example:index/VmInitialization:VmInitialization" },
        },
      },
    },
    types: {
      "example:index/FileSourceRaw:FileSourceRaw": {
        properties: {
          data: { type: "string" },
        },
      },
      "example:index/VmInitialization:VmInitialization": {
        properties: {
          userDataFileId: { type: "string" },
        },
      },
    },
  });

  assert.deepEqual(errors, []);
});

test("validateProviderAgainstRegistrySchema rejects missing nested schema check path", () => {
  const errors = validateProviderAgainstRegistrySchema({
    componentName: "CompositePublisher",
    resourceToken: "example:index/vm:Vm",
    userData: {
      mode: "proxmoxSnippet",
    },
    registrySchemaChecks: [{
      resourceToken: "example:index/file:File",
      propertyPath: ["sourceRaw", "data"],
      description: "cloud-init snippet content",
    }],
  }, {
    name: "example",
    resources: {
      "example:index/file:File": {
        inputProperties: {
          sourceRaw: { "$ref": "#/types/example:index/FileSourceRaw:FileSourceRaw" },
        },
      },
    },
    types: {
      "example:index/FileSourceRaw:FileSourceRaw": {
        properties: {
          fileName: { type: "string" },
        },
      },
    },
  });

  assert.match(errors.join("\n"), /CompositePublisher upstream resource example:index\/file:File missing cloud-init snippet content path sourceRaw\.data/);
});
```

- [ ] **Step 2: Write failing catalog metadata test for Proxmox VE checks**

Append this test to `test/providerCatalog.test.ts`:

```ts
test("Proxmox VE declares both cloud-init snippet schema checks", () => {
  assert.deepEqual(providerCatalog.ProxmoxvePublisher.registrySchemaChecks, [{
    resourceToken: "proxmoxve:index/fileLegacy:FileLegacy",
    propertyPath: ["sourceRaw", "data"],
    description: "cloud-init snippet content",
  }, {
    resourceToken: "proxmoxve:index/vmLegacy:VmLegacy",
    propertyPath: ["initialization", "userDataFileId"],
    description: "VM cloud-init user-data file reference",
  }]);
});
```

- [ ] **Step 3: Run focused schema tests and verify they fail**

Run:

```bash
npm run build && node --test dist/test/providerRegistrySchema.test.js dist/test/providerCatalog.test.js
```

Expected: FAIL because `registrySchemaChecks` does not exist and nested paths are not validated.

- [ ] **Step 4: Add schema check types and Proxmox VE metadata**

In `src/providerCatalog.ts`, add this interface above `ProviderCatalogEntry`:

```ts
export interface ProviderRegistrySchemaCheck {
  resourceToken: string;
  propertyPath: string[];
  description: string;
}
```

Add this optional field to both `ProviderCatalogEntry` and `ProviderDefinition`:

```ts
  registrySchemaChecks?: ProviderRegistrySchemaCheck[];
```

In `provider(definition: ProviderDefinition)`, add:

```ts
    registrySchemaChecks: definition.registrySchemaChecks,
```

Replace the Proxmox VE provider definition with this exact entry:

```ts
  provider({
    displayName: "Proxmox VE",
    componentName: "ProxmoxvePublisher",
    implementation: "catalogSpecializedVm",
    bootstrapModel: "bootstrapOnly",
    userDataMode: "proxmoxSnippet",
    slug: "proxmoxve",
    required: ["nodeName", "datastoreId", "templateVmId"],
    resourceToken: "proxmoxve:index/vmLegacy:VmLegacy",
    providerPackage: "@muhlba91/pulumi-proxmoxve",
    registrySchemaChecks: [{
      resourceToken: "proxmoxve:index/fileLegacy:FileLegacy",
      propertyPath: ["sourceRaw", "data"],
      description: "cloud-init snippet content",
    }, {
      resourceToken: "proxmoxve:index/vmLegacy:VmLegacy",
      propertyPath: ["initialization", "userDataFileId"],
      description: "VM cloud-init user-data file reference",
    }],
  }),
```

- [ ] **Step 5: Extend registry schema validation**

In `src/providerRegistrySchema.ts`, extend `RegistryProviderEntry` with:

```ts
  registrySchemaChecks?: Array<{
    resourceToken: string;
    propertyPath: string[];
    description: string;
  }>;
```

Extend `RegistrySchema` with:

```ts
  types?: Record<string, {
    properties?: Record<string, unknown>;
  }>;
```

Inside `validateProviderAgainstRegistrySchema`, after the package comparison block and before the existing single-property check, add:

```ts
  if (provider.registrySchemaChecks && provider.registrySchemaChecks.length > 0) {
    for (const check of provider.registrySchemaChecks) {
      const checkedResource = schema.resources?.[check.resourceToken];
      if (!checkedResource) {
        errors.push(`${provider.componentName} upstream schema missing resource token ${check.resourceToken}`);
        continue;
      }
      if (!schemaHasPath(schema, checkedResource.inputProperties ?? {}, check.propertyPath)) {
        errors.push(`${provider.componentName} upstream resource ${check.resourceToken} missing ${check.description} path ${check.propertyPath.join(".")}`);
      }
    }
    return errors;
  }
```

Add these helper functions below `validateProviderAgainstRegistrySchema`:

```ts
function schemaHasPath(schema: RegistrySchema, properties: Record<string, unknown>, path: string[]): boolean {
  let currentProperties: Record<string, unknown> | undefined = properties;
  for (const [index, segment] of path.entries()) {
    const property = currentProperties?.[segment] as { $ref?: string; properties?: Record<string, unknown> } | undefined;
    if (!property) {
      return false;
    }
    if (index === path.length - 1) {
      return true;
    }
    currentProperties = property.properties ?? resolveRefProperties(schema, property.$ref);
  }
  return false;
}

function resolveRefProperties(schema: RegistrySchema, ref: string | undefined): Record<string, unknown> | undefined {
  if (!ref?.startsWith("#/types/")) {
    return undefined;
  }
  const typeToken = ref.slice("#/types/".length);
  return schema.types?.[typeToken]?.properties;
}
```

- [ ] **Step 6: Update generated docs placement rendering**

In `scripts/generate-provider-docs.mjs`, find the generated shared cloud-init table row code that currently renders:

```js
provider.userData.property ?? provider.userData.metadataKey ?? ""
```

Replace it with this helper call:

```js
userDataPlacement(provider)
```

Add this helper function near the bottom of the script:

```js
function userDataPlacement(provider) {
  if (provider.registrySchemaChecks?.length > 0) {
    return provider.registrySchemaChecks
      .map((check) => `${check.resourceToken} ${check.propertyPath.join(".")}`)
      .join("<br>");
  }
  return provider.userData.property ?? provider.userData.metadataKey ?? "";
}
```

- [ ] **Step 7: Run schema tests and catalog check**

Run:

```bash
npm run build && node --test dist/test/providerRegistrySchema.test.js dist/test/providerCatalog.test.js
npm run catalog:check
```

Expected: PASS, including `Provider registry schema check passed.`

- [ ] **Step 8: Regenerate docs and inspect Proxmox placement**

Run:

```bash
npm run docs:gen
rg "ProxmoxvePublisher" site/source/_generated/shared-cloud-init-table.md
```

Expected output includes both `proxmoxve:index/fileLegacy:FileLegacy sourceRaw.data` and `proxmoxve:index/vmLegacy:VmLegacy initialization.userDataFileId`.

- [ ] **Step 9: Commit**

```bash
git add src/providerCatalog.ts src/providerRegistrySchema.ts test/providerRegistrySchema.test.ts test/providerCatalog.test.ts scripts/generate-provider-docs.mjs site/source/_generated
git commit -m "fix: validate composite provider schema placements"
```

## Task 3: Replace Go Field Name Mapping with Pulumi Tag Lookup

**Files:**
- Modify: `internal/provider/catalog_validation.go`
- Create: `internal/provider/catalog_validation_test.go`

- [ ] **Step 1: Write failing Go tests for tag-based validation**

Create `internal/provider/catalog_validation_test.go`:

```go
package provider

import (
	"reflect"
	"testing"
)

type catalogValidationTaggedArgs struct {
	TemplateID string `pulumi:"templateId"`
	VpcUUID    string `pulumi:"vpcUuid,optional"`
	OSID       *int   `pulumi:"osId,optional"`
}

func TestIsGoFieldMissingUsesPulumiTags(t *testing.T) {
	value := reflectValue(catalogValidationTaggedArgs{
		TemplateID: "template-1",
		VpcUUID:    "vpc-1",
	})

	if isGoFieldMissing(value, "templateId") {
		t.Fatalf("expected templateId to resolve through pulumi tag")
	}
	if isGoFieldMissing(value, "vpcUuid") {
		t.Fatalf("expected vpcUuid to resolve through pulumi tag")
	}
	if !isGoFieldMissing(value, "osId") {
		t.Fatalf("expected nil osId pointer to be missing")
	}
}

func TestIsGoFieldMissingTreatsUnknownPulumiTagAsMissing(t *testing.T) {
	value := reflectValue(catalogValidationTaggedArgs{
		TemplateID: "template-1",
	})

	if !isGoFieldMissing(value, "doesNotExist") {
		t.Fatalf("expected unknown field to be missing")
	}
}

func reflectValue(value any) reflect.Value {
	result := reflect.ValueOf(value)
	for result.Kind() == reflect.Pointer {
		result = result.Elem()
	}
	return result
}
```

- [ ] **Step 2: Run focused Go validation tests and verify they fail**

Run:

```bash
go test ./internal/provider -run 'TestIsGoFieldMissing' -count=1
```

Expected: FAIL because `goFieldName` has no `vpcUuid` case and does not read Pulumi tags.

- [ ] **Step 3: Implement Pulumi tag lookup**

In `internal/provider/catalog_validation.go`, replace the body of `isGoFieldMissing` with:

```go
func isGoFieldMissing(value reflect.Value, pulumiName string) bool {
	field, ok := fieldByPulumiName(value, pulumiName)
	if !ok {
		return true
	}
	if field.Kind() == reflect.Pointer {
		return field.IsNil() || field.Elem().IsZero()
	}
	return field.IsZero()
}
```

Replace the body of `boolFieldIsTrue` with:

```go
func boolFieldIsTrue(value reflect.Value, pulumiName string) bool {
	field, ok := fieldByPulumiName(value, pulumiName)
	if !ok {
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
```

Delete `goFieldName` and add these helpers:

```go
func fieldByPulumiName(value reflect.Value, pulumiName string) (reflect.Value, bool) {
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		structField := valueType.Field(i)
		if pulumiTagName(structField.Tag.Get("pulumi")) == pulumiName {
			return value.Field(i), true
		}
	}
	return reflect.Value{}, false
}

func pulumiTagName(tag string) string {
	if tag == "" || tag == "-" {
		return ""
	}
	if comma := strings.Index(tag, ","); comma >= 0 {
		return tag[:comma]
	}
	return tag
}
```

Keep the `strings` import because `validateProviderCatalogArgs` still uses `strings.Join` and `pulumiTagName` uses `strings.Index`.

- [ ] **Step 4: Run Go validation tests**

Run:

```bash
go test ./internal/provider -run 'TestIsGoFieldMissing|TestProviderCatalog|TestVultrConstructRejects' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/provider/catalog_validation.go internal/provider/catalog_validation_test.go
git commit -m "fix: validate go catalog inputs by pulumi tags"
```

## Task 4: Enforce Go Catalog Validation for All VM Providers

**Files:**
- Modify: `internal/provider/catalog_test.go`
- Modify: `internal/provider/provider_test.go`
- Modify: `internal/provider/components.go`

- [ ] **Step 1: Add Go catalog metadata test for direct providers**

Append this test to `internal/provider/catalog_test.go`:

```go
func TestProviderCatalogDirectProviderValidationMetadata(t *testing.T) {
	proxmox := providerCatalog["ProxmoxvePublisher"]
	if len(proxmox.RequiredInputs) != 3 {
		t.Fatalf("ProxmoxvePublisher required inputs mismatch: %#v", proxmox.RequiredInputs)
	}
	if proxmox.RequiredInputs[0] != "nodeName" || proxmox.RequiredInputs[1] != "datastoreId" || proxmox.RequiredInputs[2] != "templateVmId" {
		t.Fatalf("ProxmoxvePublisher required inputs mismatch: %#v", proxmox.RequiredInputs)
	}

	oci := providerCatalog["OciPublisher"]
	if len(oci.RequiredInputs) != 4 {
		t.Fatalf("OciPublisher required inputs mismatch: %#v", oci.RequiredInputs)
	}
	if oci.RequiredInputs[0] != "compartmentId" || oci.RequiredInputs[1] != "availabilityDomain" || oci.RequiredInputs[2] != "subnetId" || oci.RequiredInputs[3] != "imageId" {
		t.Fatalf("OciPublisher required inputs mismatch: %#v", oci.RequiredInputs)
	}
}
```

- [ ] **Step 2: Add failing construct-time validation tests**

Append these tests near `TestVultrConstructRejectsMissingImageChoice` in `internal/provider/provider_test.go`:

```go
func TestProxmoxveConstructRejectsMissingTemplateVMID(t *testing.T) {
	err := constructPublisherResourceError(t, "netskope-publisher:index:ProxmoxvePublisher", property.NewMap(map[string]property.Value{
		"names":         property.New([]property.Value{property.New("pub-1")}),
		"registrations": registrationMap("pub-1"),
		"nodeName":      property.New("pve-1"),
		"datastoreId":   property.New("local"),
	}))
	if err == nil || !strings.Contains(err.Error(), "ProxmoxvePublisher requires input templateVmId") {
		t.Fatalf("expected Proxmox VE missing templateVmId error, got %v", err)
	}
}

func TestOciConstructRejectsMissingImageID(t *testing.T) {
	err := constructPublisherResourceError(t, "netskope-publisher:index:OciPublisher", property.NewMap(map[string]property.Value{
		"names":              property.New([]property.Value{property.New("pub-1")}),
		"registrations":      registrationMap("pub-1"),
		"compartmentId":      property.New("ocid1.compartment.oc1..example"),
		"availabilityDomain": property.New("AD-1"),
		"subnetId":           property.New("ocid1.subnet.oc1..example"),
	}))
	if err == nil || !strings.Contains(err.Error(), "OciPublisher requires input imageId") {
		t.Fatalf("expected OCI missing imageId error, got %v", err)
	}
}
```

- [ ] **Step 3: Run focused construct tests and verify they fail**

Run:

```bash
go test ./internal/provider -run 'TestProxmoxveConstructRejectsMissingTemplateVMID|TestOciConstructRejectsMissingImageID|TestProviderCatalogDirectProviderValidationMetadata' -count=1
```

Expected: FAIL because `NewProxmoxvePublisher` and `NewOciPublisher` do not call `validateProviderCatalogArgs`.

- [ ] **Step 4: Add validation calls to earlier Go VM constructors**

In `internal/provider/components.go`, add the validation block immediately after successful `ctx.RegisterComponentResource(...)` in each listed constructor:

```go
	if err := validateProviderCatalogArgs("<ComponentName>", args); err != nil {
		return nil, err
	}
```

Use these exact component names:

| Constructor | ComponentName |
|---|---|
| `NewAwsPublisher` | `AwsPublisher` |
| `NewAzurePublisher` | `AzurePublisher` |
| `NewGcpPublisher` | `GcpPublisher` |
| `NewVspherePublisher` | `VspherePublisher` |
| `NewEsxiPublisher` | `EsxiPublisher` |
| `NewHcloudPublisher` | `HcloudPublisher` |
| `NewNutanixPublisher` | `NutanixPublisher` |
| `NewOpenstackPublisher` | `OpenstackPublisher` |
| `NewOvhPublisher` | `OvhPublisher` |
| `NewScalewayPublisher` | `ScalewayPublisher` |
| `NewOciPublisher` | `OciPublisher` |
| `NewAlicloudPublisher` | `AlicloudPublisher` |
| `NewProxmoxvePublisher` | `ProxmoxvePublisher` |

Do not add a validation call to `NewKubernetesPublisher` in this task because its catalog required input list is empty.

- [ ] **Step 5: Run Go formatting and focused tests**

Run:

```bash
gofmt -w internal/provider/components.go internal/provider/catalog_test.go internal/provider/provider_test.go
go test ./internal/provider -run 'TestProxmoxveConstructRejectsMissingTemplateVMID|TestOciConstructRejectsMissingImageID|TestProviderCatalogDirectProviderValidationMetadata|TestExpandedProviderConstructs|TestVultrConstructRejects' -count=1
```

Expected: PASS.

- [ ] **Step 6: Run all Go tests**

Run:

```bash
npm run go:test
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/provider/components.go internal/provider/catalog_test.go internal/provider/provider_test.go
git commit -m "fix: validate all go vm publisher constructors"
```

## Task 5: Final Verification and Documentation Refresh

**Files:**
- Modify only if generated output changes: `site/source/_generated/*`
- Verify only: `schema.json`, SDK directories

- [ ] **Step 1: Regenerate docs**

Run:

```bash
npm run docs:gen
```

Expected: PASS.

- [ ] **Step 2: Run TypeScript checks**

Run:

```bash
npm run typecheck
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

Expected: PASS with both `Provider registry schema check passed.` and `Provider catalog parity check passed.`

- [ ] **Step 5: Build GitHub Pages site**

Run:

```bash
npm run build --prefix site
```

Expected: PASS.

- [ ] **Step 6: Check schema and SDK drift**

Run:

```bash
git diff --quiet schema.json || npm run sdk:gen
```

Expected: exits 0. If SDK generation runs, inspect and commit only intentional schema/SDK changes.

- [ ] **Step 7: Commit generated docs or SDK drift if needed**

If `git status --short` shows intentional generated docs or SDK changes from this task, commit them:

```bash
git add site/source/_generated schema.json sdk/python sdk/dotnet sdk/go sdk/java sdk/rust
git commit -m "build: refresh generated provider artifacts"
```

If `git status --short` shows no intentional generated changes, skip this commit.

- [ ] **Step 8: Final status check**

Run:

```bash
git status --short --branch
```

Expected: branch is ahead by the new commits, with only pre-existing unrelated dirty files if those were present before execution.

## Self-Review Notes

- Spec coverage: Task 1 fixes TypeScript validation bypasses. Task 2 fixes Proxmox VE composite schema metadata and checker coverage. Tasks 3 and 4 fix Go validation fragility and coverage. Task 5 verifies the full provider and docs surface.
- Placeholder scan: No task uses TBD, TODO, “similar to”, or open-ended “add tests” instructions. Each code-changing step includes exact snippets or exact file/component mappings.
- Type consistency: `registrySchemaChecks`, `propertyPath`, `resourceToken`, and `description` are used consistently across catalog metadata and registry validation. Go validation uses Pulumi tag names such as `templateId`, `imageId`, and `vpcUuid`, matching catalog input names.
