# Provider Framework Design

## Goal

Create a catalog-driven provider framework for bootstrap VM providers while
keeping custom implementations for platforms with distinct lifecycle behavior.
The framework should reduce duplicated provider logic, make provider
capabilities explicit, validate provider inputs before resource creation, and
use provider metadata to generate or validate GitHub Pages docs and examples.

## Current State

The repository now supports a broad set of providers. TypeScript has one
component file per provider, a shared `vmPublisherCore`, a `RawResource`
helper, and user-data adapter helpers. Go has equivalent behavior in a large
`internal/provider/components.go` file. Docs and examples are still mostly
hand-authored per component page.

This works, but it has several scaling problems:

- Provider capability details are spread across component code, tests, docs,
  and implementation plans.
- Docs can drift from the schema or implementation.
- Input validation is inconsistent across providers.
- Adding another provider requires repeating the same implementation,
  documentation, YAML example, schema, and test updates.
- Go and TypeScript parity depends on manual discipline.

## Chosen Approach

Use a catalog-driven framework for providers that fit the bootstrap VM pattern.
Keep bespoke implementations for providers whose resource lifecycle or image
behavior is materially different.

This is intentionally not a full DSL/code-generation rewrite. The framework
should create a cleaner provider platform without hiding real differences
between cloud providers.

## Provider Categories

### Catalog-Driven Bootstrap Providers

These providers should move toward metadata plus reusable factories:

- `HcloudPublisher`
- `NutanixPublisher`
- `OpenstackPublisher`
- `OvhPublisher`
- `ScalewayPublisher`
- `OciPublisher`
- `AlicloudPublisher`
- `DigitaloceanPublisher`
- `VultrPublisher`
- `ExoscalePublisher`
- `UpcloudPublisher`
- `StackitPublisher`
- `EquinixPublisher`
- `OutscalePublisher`
- `OpentelekomcloudPublisher`
- `TencentcloudPublisher`
- `YandexPublisher`
- `ProxmoxvePublisher`, if its snippet and clone behavior fits cleanly enough

### Bespoke Providers

These providers should remain custom implementations:

- `AwsPublisher`, because it has AMI lookup and Marketplace versus bootstrap
  behavior.
- `AzurePublisher`, because it owns NIC, public IP, VM, Marketplace, and custom
  data behavior.
- `GcpPublisher`, because it has Compute Engine metadata behavior and GCP
  network/public IP handling.
- `KubernetesPublisher`, because it provisions Helm/chart resources rather than
  VMs.
- `VspherePublisher`, because vSphere template clone behavior is different from
  raw VM bootstrap providers.
- `EsxiPublisher`, because direct ESXi and GuestInfo behavior is different from
  cloud VM providers.
- `HypervPublisher`, because it is still an experimental gate.
- `NetskopeRegistration`, because it is a registration resource rather than a
  platform publisher component.

Bespoke providers should still have catalog metadata for docs, capability
tables, and parity checks, but the catalog should not drive their resource
creation.

## Provider Catalog

Introduce a canonical provider catalog. The catalog can start in TypeScript,
with a generated JSON artifact or a mirrored Go catalog added when the Go
refactor begins.

Each provider entry should include:

- Component identity:
  - display name
  - component class name
  - Pulumi token
  - support status
  - provider package name
  - underlying resource token when applicable
- Implementation mode:
  - `catalogRawVm`
  - `catalogTypedVm`
  - `catalogSpecializedVm`
  - `bespoke`
- Bootstrap model:
  - bootstrap-only
  - pre-baked image supported
  - Marketplace image supported
  - experimental or unsupported
- User-data adapter:
  - plain user data
  - base64 user data
  - metadata key
  - custom data
  - OCI metadata
  - Scaleway dual placement
  - GuestInfo
  - raw user data
  - Proxmox snippet
- Required and optional inputs:
  - common inputs inherited from `CommonPublisherArgs`
  - provider-specific required inputs
  - provider-specific optional inputs
  - defaults and example values
- Validation rules:
  - required fields
  - required-one-of fields
  - mutually exclusive fields
  - bootstrap-only constraints
  - image/template constraints
  - unsupported capability messages
- Output mapping:
  - VM ID expression
  - private IP expression
  - optional public IP expression
- Documentation metadata:
  - short description
  - required input summary
  - optional input summary
  - bootstrap notes
  - provider-specific caveats
  - Pulumi YAML example values

## Resource Factories

The framework should provide reusable factories rather than one generic mega
factory.

TypeScript factories:

- Raw VM resource factory for providers backed by `RawResource`.
- Typed VM resource factory for providers using installed typed SDKs.
- Adapter placement helpers for each user-data mode.
- Shared output construction using `createVmPublishers`.
- Shared validation before resource construction.

Go factories:

- Shared registration/name/bootstrap rendering helpers remain the base.
- Add provider metadata and validation helpers that mirror the catalog.
- Refactor repetitive bootstrap components into small provider build functions
  that consume metadata where Go's static schema constraints allow it.
- Keep typed argument structs for schema generation and SDK quality.

The framework must keep public component names and Pulumi tokens stable unless
a specific breaking change is explicitly approved later.

## Validation Model

Validation should run before resources are created. It should produce provider
specific error messages that explain the missing or incompatible input.

Initial validation rules should cover:

- Missing required platform inputs.
- Missing all values from a required-one-of group.
- Mutually exclusive image fields.
- Bootstrap-only providers receiving unsupported pre-baked image settings.
- Providers that require a user-supplied Ubuntu 22.04 image/template.
- Providers that require networking inputs such as subnet, network, security
  group, VPC, or project IDs.
- Experimental providers that require explicit opt-in.

Validation rules must be unit-testable without cloud credentials.

## Docs And Examples Generation

The catalog should become the source of truth for repeated docs content.

Generate or validate:

- Provider matrix rows.
- Component overview links.
- Component required and optional input summaries for catalog-driven providers.
- Shared capability tables.
- Pulumi YAML examples.
- Shared cloud-init adapter table rows.

Per-provider markdown pages can keep narrative text, caveats, and language
examples, but repeated provider facts should be generated or checked against
the catalog. The first implementation should prefer generated snippets or
validation checks over replacing the entire docs system.

## Testing And Parity Checks

Add tests or scripts that enforce catalog parity:

- Every catalog provider token exists in `schema.json`.
- Every catalog provider has a TypeScript export.
- Every catalog provider is registered in the Go provider.
- Every catalog provider has a GitHub Pages component page.
- Every catalog provider has a Pulumi YAML example.
- User-data adapter metadata matches the implemented resource property.
- Validation rules fail for representative missing and invalid inputs.
- Docs generation or docs validation fails when catalog data and markdown drift.

Existing tests for component behavior should remain. New framework tests should
focus on metadata correctness, validation behavior, and resource factory output
shape.

## Migration Plan

1. Add the provider catalog and validation types without changing behavior.
2. Add catalog parity checks for schema tokens, exports, Go registration, and
   docs pages.
3. Generate or validate Pulumi YAML examples and provider matrix rows from the
   catalog.
4. Move the simplest raw-resource bootstrap providers to catalog-driven
   factories.
5. Migrate remaining catalog-fit bootstrap providers where the abstraction
   stays clear.
6. Add Go metadata parity and validation helpers.
7. Refactor Go bootstrap providers toward catalog-driven helper functions while
   keeping typed argument structs.
8. Leave bespoke providers custom, but include them in docs and parity
   metadata.

## Non-Goals

- Do not replace every provider with one universal factory.
- Do not change public component names or Pulumi tokens in this pass.
- Do not remove bespoke provider implementations where they encode real cloud
  lifecycle differences.
- Do not generate all TypeScript and Go provider code from a custom DSL in the
  first implementation.
- Do not add new cloud providers as part of this framework task unless needed
  for a focused validation example.

## Success Criteria

- A new provider can be added by defining catalog metadata, a small factory
  mapping, and focused provider-specific code only where needed.
- Provider capability docs and Pulumi YAML examples are generated or validated
  from the catalog.
- Invalid provider inputs fail with clear messages before resource creation.
- TypeScript and Go provider parity is checked automatically.
- Existing component names, tokens, generated SDKs, and user-facing behavior
  remain stable.
