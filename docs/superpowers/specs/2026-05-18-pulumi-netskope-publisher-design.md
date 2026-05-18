# Pulumi Netskope Publisher Design

## Context

The existing Terraform repository provisions Netskope Private Access
Publishers on AWS, Azure, GCP, and vSphere. It contains per-platform
Terraform modules, a shared Netskope registration module, a shared
cloud-init module, example configurations, tests, and a Hexo/Cactus
GitHub Pages documentation site.

The Pulumi repository currently contains only an initial README. The
first Pulumi implementation should create the reusable package shape and
ship one complete provider implementation instead of attempting a full
multi-cloud port in a single step.

## Scope

Build the first version as a TypeScript Pulumi component package focused
on AWS. Include shared internals for Netskope registration and cloud-init
so Azure, GCP, vSphere, and Hyper-V can be added later without changing
the public model.

In scope:

- TypeScript package structure for a reusable Pulumi component library.
- Public `AwsPublisher` component.
- Shared Netskope registration resource/provider.
- Shared cloud-init rendering.
- AWS EC2 publisher infrastructure.
- `examples/aws-single` Pulumi program.
- Hexo/Cactus GitHub Pages documentation adapted from the Terraform site.
- CI for typecheck, tests, and docs build.
- GitHub Pages publishing workflow.
- Repository metadata for GitHub/source-based consumption.

Out of scope for the first implementation:

- Azure, GCP, vSphere, and Hyper-V implementations.
- Public Pulumi Registry publishing.
- A docs redesign or migration away from Hexo/Cactus.
- Manual release automation beyond preparing metadata for future use.

## Architecture

The package exposes a small public API and keeps provider-specific work
behind focused internal modules.

```text
src/
  index.ts
  awsPublisher.ts
  netskopeRegistration.ts
  cloudInit.ts
examples/
  aws-single/
site/
  ...
.github/workflows/
  ci.yml
  pages.yml
PulumiPlugin.yaml
package.json
tsconfig.json
```

`AwsPublisher` composes three responsibilities:

1. Derive publisher names from explicit `names` or from `namePrefix` and
   `replicas`.
2. Register or reuse Netskope publisher records and generate
   registration tokens.
3. Render per-publisher cloud-init and create one AWS EC2 instance per
   publisher.

This structure keeps the AWS implementation complete while allowing later
platform components to reuse name derivation, registration, cloud-init,
and output conventions.

## Component API

`AwsPublisher` should mirror the Terraform AWS module where practical.

Core inputs:

- `namePrefix`
- `names`
- `replicas`
- `tenantUrl`
- `apiToken`
- `wizardPath`
- `tags`

AWS inputs:

- `subnetId`
- `securityGroupIds`
- `keyName`
- `instanceType`
- `amiId`
- `associatePublicIpAddress`
- `iamInstanceProfile`
- `ebsOptimized`
- `monitoring`
- `metadataOptions`

Defaults should match the Terraform module unless the Pulumi AWS provider
already supplies the same behavior.

The component should also expose a bring-your-own registration path for
environments where the Netskope API calls are managed outside Pulumi. In
that mode, callers provide pre-created publisher IDs and registration
tokens, and the component only creates cloud infrastructure.

Outputs:

- `publisherNames`
- `publishers`

`publishers` is keyed by publisher name. Each value includes publisher
ID, instance ID, private IP, public IP, and registration token. The
registration token must be a Pulumi secret.

## Netskope Registration

Registration should be represented as Pulumi-managed state, not as
opaque HTTP calls hidden inside every deployment. Use a Pulumi dynamic
provider or equivalent component-internal resource to:

1. List existing publishers from the Netskope tenant.
2. Create any missing publishers by name.
3. Store the resulting publisher IDs in Pulumi state.
4. Generate registration tokens for the current deployment.

The behavior should match Terraform's current flow: existing publishers
are reused by name, missing publishers are created, and one registration
token is generated per publisher.

Error messages should identify the operation, publisher name when
available, and HTTP status. Sensitive values such as API tokens and
registration tokens must not appear in errors or logs.

## Cloud-Init

Cloud-init rendering should be shared across platforms. The AWS
implementation uses user data equivalent to the Terraform cloud-init
module and passes the per-publisher registration token to the Netskope
publisher wizard path.

The renderer should be unit tested independently from AWS resources so
future providers can reuse it without requiring cloud provider tests.

## AWS Infrastructure

The AWS component should create one EC2 instance per publisher name.

Behavior should match the Terraform AWS module:

- Use a caller-supplied `amiId` when provided.
- Otherwise discover the latest Netskope Private Access Publisher AMI
  owned by Netskope.
- Pass cloud-init as base64 user data when required by the Pulumi AWS
  provider.
- Apply tags merged with `Name = <publisher name>`.
- Preserve instance options for key pair, IAM instance profile, EBS
  optimization, monitoring, public IP association, and IMDS metadata
  options.

## Examples

Create `examples/aws-single` as a real Pulumi program that consumes the
local package during development. It should demonstrate:

- Netskope tenant URL and API token via Pulumi config, with the token
  stored as a secret.
- AWS subnet, security groups, key pair, and optional AMI override.
- `pulumi preview`, `pulumi up`, and `pulumi destroy` workflow.

The root README should show local development usage first and describe
GitHub/tag-based consumption once releases are available.

## Documentation

Port the Terraform repo's Hexo/Cactus documentation structure into
`site/` and adapt the content for Pulumi.

The docs should cover:

- What the first version builds.
- Pulumi installation and project setup.
- Stack config and secret handling.
- AWS account and networking prerequisites.
- Netskope tenant/API token preparation.
- First publisher deployment.
- Verification and teardown.
- Component inputs and outputs.
- Operations guidance for state, secrets, upgrades, and troubleshooting.
- Roadmap for Azure, GCP, vSphere, and Hyper-V.

The docs must state clearly that the first Pulumi version is AWS-first
and does not yet have Terraform provider parity.

## Testing And CI

Unit tests should cover:

- Name derivation from `names`, `namePrefix`, and `replicas`.
- Cloud-init rendering for one or more publishers.
- Registration provider behavior with mocked Netskope HTTP responses.
- Secret propagation for registration tokens.

CI should run:

- Dependency installation.
- TypeScript typecheck.
- Unit tests.
- Documentation build.

GitHub Pages should publish the generated Hexo site from a dedicated
workflow.

## Publishing

The first implementation should prepare metadata for source-based GitHub
consumption and future registry work, including `PulumiPlugin.yaml` and
package metadata. Public Pulumi Registry publishing is deferred until the
component API has been exercised and stabilized.
