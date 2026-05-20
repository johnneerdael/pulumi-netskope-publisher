# Unoffical Netskope NPA Publisher — Pulumi Deployment Guide

[![npm](https://img.shields.io/npm/v/%40johninnl%2Fpulumi-netskope-publisher?logo=npm)](https://www.npmjs.com/package/@johninnl/pulumi-netskope-publisher)
[![PyPI](https://img.shields.io/pypi/v/pulumi-netskope-publisher?logo=pypi)](https://pypi.org/project/pulumi-netskope-publisher/)
[![NuGet](https://img.shields.io/nuget/v/JohninNL.Pulumi.NetskopePublisher?logo=nuget)](https://www.nuget.org/packages/JohninNL.Pulumi.NetskopePublisher)
[![Go Reference](https://pkg.go.dev/badge/github.com/johnneerdael/pulumi-netskope-publisher.svg)](https://pkg.go.dev/github.com/johnneerdael/pulumi-netskope-publisher)
[![Java SDK](https://img.shields.io/badge/Java-GitHub%20Packages-blue?logo=apachemaven)](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
[![Crates.io](https://img.shields.io/crates/v/pulumi-netskope-publisher?logo=rust)](https://crates.io/crates/pulumi-netskope-publisher)

Pulumi components for provisioning Netskope Private Access Publishers
with TypeScript, Python, C#, Go, Java, and Rust SDKs.

The package mirrors the Terraform modules from
`terraform-netskope-publisher`: register or reuse Netskope publisher
records, generate per-publisher cloud-init, and create platform VMs.
For Kubernetes, it installs the `kubernetes-netskope-publisher` Helm
chart and supports both token and API enrollment modes.

AWS, Azure, and GCP support the same bootstrap-mode cloud-init controls
as the Terraform modules: `bootstrap`, `bootstrapUrl`, `nonat`,
`installUser`, `installUserPassword`, `installUserPasswordIsHash`,
`installUserSshAuthorizedKeys`, `deleteDefaultUser`, and
`guestNetworkInterface`.

## Current scope

- AWS publisher component: `AwsPublisher`
- Azure publisher component: `AzurePublisher`
- GCP publisher component: `GcpPublisher`
- Kubernetes publisher component: `KubernetesPublisher`
- vSphere publisher component: `VspherePublisher`
- ESXi Native publisher component: `EsxiPublisher`
- Hcloud publisher component: `HcloudPublisher`
- Nutanix publisher component: `NutanixPublisher`
- OpenStack publisher component: `OpenstackPublisher`
- OVH Public Cloud publisher component: `OvhPublisher`
- Scaleway publisher component: `ScalewayPublisher`
- OCI publisher component: `OciPublisher`
- Alicloud publisher component: `AlicloudPublisher`
- Proxmox VE publisher component: `ProxmoxvePublisher`
- Experimental Hyper-V gate: `HypervPublisher`
- Netskope publisher registration by name
- Bring-your-own registration data escape hatch
- GitHub Pages documentation

ESXi Native is direct-host ESXi support and does not replace the vSphere
component. Hcloud, Nutanix, OpenStack, OVH, Scaleway, OCI, Alicloud, and
Proxmox VE use bootstrap mode on Ubuntu 22.04 images.

Hyper-V depends on the upstream Pulumi Hyper-V provider becoming
consumable from a stable package source. The official Netskope publisher
VHDX and OVA download URLs are exported from `officialImageSources`.

## Development

```bash
npm install
npm run typecheck
npm test
npm run go:test
npm run sdk:gen
npm run sdk:pack
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
  bearerToken: config.requireSecret("bearerToken"),
  subnetId: config.require("subnetId"),
  securityGroupIds: config.requireObject<string[]>("securityGroupIds"),
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

## Examples

See `examples/aws-single`, `examples/azure-single`,
`examples/gcp-single`, `examples/kubernetes-kind`, and
`examples/vsphere-single` for Pulumi programs that deploy one or more
publishers.

## Documentation

Full guides are published with GitHub Pages:

https://johnneerdael.github.io/pulumi-netskope-publisher/

## Pulumi Registry

Registry-facing metadata and docs live in `schema.json`, `docs/`, and
the generated SDKs under `sdk/`. Run `npm run registry:check` before
opening a Registry submission PR.

`schema.json` sets `pluginDownloadURL` to GitHub Releases. The provider
binary is implemented with `pulumi-go-provider` and exposes the package
components as an executable component provider. Tagged releases build
Pulumi plugin archives named
`pulumi-resource-netskope-publisher-v<version>-<os>-<arch>.tar.gz` and
attach them to the release before publication.

The release also publishes the TypeScript SDK to npm, the Python SDK to
PyPI, the C# SDK to NuGet, and the Go SDK through the tagged GitHub
module path
`github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher`.
The Java SDK is published as `com.pulumi:netskope-publisher` to GitHub
Packages for this repository by default, and the Rust SDK is published
as [`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
on crates.io.

The Go provider constructs cloud, virtualization, Kubernetes, and
registration child resources and includes a stateful
`NetskopeRegistration` resource for creating or reusing Netskope
publisher records. Pre-created
`registrations` remain available as an escape hatch.

## Release automation

Pushes to `main` run `.github/workflows/auto-release.yml`. The workflow
bumps the patch version, updates `package.json`, `package-lock.json`,
`schema.json`, the generated SDKs, and the Go provider schema version,
runs the release checks, commits the version bump, tags `vX.Y.Z`,
publishes the SDK packages, and uploads plugin archives to the GitHub
release.

Required repository secret:

- `NPM_TOKEN`: npm automation token for publishing
- `PYPI_API_TOKEN`: PyPI token for publishing the Python SDK
- `NUGET_API_KEY`: NuGet API key for publishing the C# SDK
- `CARGO_REGISTRY_TOKEN`: crates.io token for publishing the Rust SDK

Java SDK publishing defaults to GitHub Packages with the workflow
`GITHUB_TOKEN`. To publish to another Maven-compatible repository, set
the `JAVA_MAVEN_REPOSITORY_URL` repository variable. Set
`JAVA_MAVEN_GROUP_ID` when publishing to a registry that requires a
verified namespace, such as Maven Central. Configure
`JAVA_MAVEN_AUTH_BASE64` as a base64-encoded `username:password` secret
for bearer-token publishing endpoints such as Maven Central's OSSRH
staging API, or configure `JAVA_MAVEN_USERNAME` and
`JAVA_MAVEN_PASSWORD` repository secrets for Basic-auth Maven
repositories. Maven Central publishing also requires signed artifacts;
configure `JAVA_SIGNING_KEY` and `JAVA_SIGNING_PASSWORD` with an
ASCII-armored PGP private key and its password before publishing there.

Repository Actions settings must allow workflows to write repository
contents and packages so the release workflow can push the version
commit and tag and publish the Java package.

Optional Pulumi Registry PR automation:

- `REGISTRY_PR_TOKEN`: GitHub token that can push to your
  `pulumi/registry` fork and open PRs
- `PULUMI_REGISTRY_FORK`: repository variable with the fork slug, for
  example `johnneerdael/registry`
