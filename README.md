# pulumi-netskope-publisher

Pulumi components for provisioning Netskope Private Access Publishers.

The package mirrors the Terraform modules from
`terraform-netskope-publisher`: register or reuse Netskope publisher
records, generate per-publisher cloud-init, and create platform VMs.

## Current scope

- AWS publisher component: `AwsPublisher`
- Azure publisher component: `AzurePublisher`
- GCP publisher component: `GcpPublisher`
- vSphere publisher component: `VspherePublisher`
- Experimental Hyper-V gate: `HypervPublisher`
- Netskope publisher registration by name
- Bring-your-own registration data escape hatch
- GitHub Pages documentation

Hyper-V depends on the upstream Pulumi Hyper-V provider becoming
consumable from a stable package source. The official Netskope publisher
VHDX and OVA download URLs are exported from `officialImageSources`.

## Development

```bash
npm install
npm run typecheck
npm test
npm run go:test
npm run registry:check
npm run plugin:dist
```

## Quick start

```ts
import * as pulumi from "@pulumi/pulumi";
import { AwsPublisher } from "@johninnl/pulumi-netskope-publisher";

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

## Examples

See `examples/aws-single`, `examples/azure-single`,
`examples/gcp-single`, and `examples/vsphere-single` for Pulumi programs
that deploy one or more publishers.

## Documentation

Full guides are published with GitHub Pages:

https://johnneerdael.github.io/pulumi-netskope-publisher/

## Pulumi Registry

Registry-facing metadata and docs live in `schema.json` and `docs/`.
Run `npm run registry:check` before opening a Registry submission PR.

`schema.json` sets `pluginDownloadURL` to GitHub Releases. The provider
binary is implemented with `pulumi-go-provider` and exposes the package
components as an executable component provider. Tagged releases build
Pulumi plugin archives named
`pulumi-resource-netskope-publisher-v<version>-<os>-<arch>.tar.gz` and
attach them to the release before publication.

The Go provider constructs AWS, Azure, GCP, and vSphere child resources
and includes a stateful `NetskopeRegistration` resource for creating or
reusing Netskope publisher records. Pre-created `registrations` remain
available as an escape hatch.

## Release automation

Pushes to `main` run `.github/workflows/auto-release.yml`. The workflow
bumps the patch version, updates `package.json`, `package-lock.json`,
`schema.json`, and the Go provider schema version, runs the release
checks, commits the version bump, tags `vX.Y.Z`, publishes npm, and
uploads plugin archives to the GitHub release.

Required repository secret:

- `NPM_TOKEN`: npm automation token for publishing

Repository Actions settings must allow workflows to write repository
contents so the release workflow can push the version commit and tag.

Optional Pulumi Registry PR automation:

- `REGISTRY_PR_TOKEN`: GitHub token that can push to your
  `pulumi/registry` fork and open PRs
- `PULUMI_REGISTRY_FORK`: repository variable with the fork slug, for
  example `johnneerdael/registry`
