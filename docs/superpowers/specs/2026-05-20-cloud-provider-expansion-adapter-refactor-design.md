# Cloud Provider Expansion And Bootstrap Adapter Refactor Design

## Decision

Implement the provider expansion using an adapter refactor plus new providers.

This pass adds the audited strong and good suitability providers, migrates the
existing bootstrap-style providers onto the same shared bootstrap contract, and
keeps the fuller provider capability framework as a direct follow-up.

## Goals

- Add support for the strong and good suitability providers:
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
- Migrate existing bootstrap-style providers to the same adapter model:
  - `HcloudPublisher`
  - `NutanixPublisher`
  - `OpenstackPublisher`
  - `OvhPublisher`
  - `ScalewayPublisher`
  - `OciPublisher`
  - `AlicloudPublisher`
  - `ProxmoxvePublisher`
- Keep official Netskope-image providers on a narrower enrollment and network
  customization path.
- Regenerate schema and SDKs so TypeScript, Python, C#, Go, Java, and Rust
  expose the new components.
- Update README, GitHub Pages docs, installation docs, and examples.

## Non-Goals

- Do not add IBM, IONOS Cloud, Libvirt, Linode, Genesis Cloud, or Snowflake in
  this pass.
- Do not build the full provider capability metadata and documentation generator
  framework in this pass.
- Do not change existing provider argument names unless required for correctness.
- Do not make Pulumi Registry docs claim Rust support, because the Pulumi
  Registry does not officially support Rust.

## Follow-Up

After this pass, build a fuller provider framework with provider capability
metadata, stricter validation, documentation/example generation hooks, and
adapter-driven resource factory registration.

## Architecture

Split the implementation into four layers.

### Bootstrap Payload Renderer

The shared bootstrap renderer creates the Ubuntu 22.04 cloud-init payload used
by all bootstrap-based providers. It owns Publisher installation and enrollment:

- hostname
- default install user
- optional password or SSH keys
- optional default user deletion
- optional static network config
- optional `.nonat`
- bootstrap script URL
- Netskope registration wizard/token execution

The renderer must exist in both TypeScript and Go because the repository ships
both TypeScript component implementations and a Go executable provider.

### Netskope Image Enrollment Renderer

Official Netskope-image providers keep a smaller renderer that supports
enrollment and safe network customization only. It may support password or
private-key customization for the default `ubuntu` user where the platform
already allows it.

This prevents official-image providers such as AWS, Azure, GCP, vSphere, ESXi,
and Hyper-V from accidentally inheriting full vanilla Ubuntu bootstrap
customization.

### User-Data Adapter Contract

Each provider declares how a rendered payload is placed onto its VM resource.

```ts
type UserDataMode =
  | "plain"
  | "base64"
  | "metadata-user-data"
  | "custom-data"
  | "cloud-init-disk"
  | "guestinfo"
  | "startup-script";

interface PublisherUserDataAdapter {
  mode: UserDataMode;
  maxBytes?: number;
  render(payload: pulumi.Output<string>): pulumi.Input<string> | Record<string, pulumi.Input<string>>;
}
```

The Go implementation should use the same conceptual contract with Go-native
types.

### Provider Components

Each component remains responsible for its cloud-specific infrastructure:

- image or template selection
- region or zone
- instance size or type
- networking
- SSH key references
- firewall or security group references
- tags or labels
- provider-specific outputs

Components should not build bootstrap scripts directly. They pass common
publisher arguments into the renderer, ask their adapter for the correctly
encoded placement, and create provider resources.

## Provider Adapter Map

| Provider | Component | Bootstrap placement | Status |
| --- | --- | --- | --- |
| HCloud | `HcloudPublisher` | plain `userData` | Existing, migrate |
| OVH | `OvhPublisher` | plain `userData` | Existing, migrate |
| OpenStack | `OpenstackPublisher` | plain `userData` | Existing, migrate |
| Scaleway | `ScalewayPublisher` | `cloudInit` and `userData["cloud-init"]` | Existing, migrate |
| OCI | `OciPublisher` | base64 metadata `userData` | Existing, migrate |
| Alicloud | `AlicloudPublisher` | base64 `userData` | Existing, migrate |
| Nutanix | `NutanixPublisher` | base64 guest customization field | Existing, migrate |
| Proxmox VE | `ProxmoxvePublisher` | cloud-init snippet/file | Existing, migrate |
| DigitalOcean | `DigitaloceanPublisher` | plain `userData` | New |
| Vultr | `VultrPublisher` | plain `userData` | New |
| Exoscale | `ExoscalePublisher` | plain `userData` | New |
| UpCloud | `UpcloudPublisher` | plain `userData` | New |
| Stackit | `StackitPublisher` | plain `userData` | New |
| Equinix Metal | `EquinixPublisher` | plain `userData` | New |
| Outscale | `OutscalePublisher` | plain `userData` | New |
| OpenTelekomCloud | `OpentelekomcloudPublisher` | plain `userData` | New |
| TencentCloud | `TencentcloudPublisher` | likely `userDataRaw` | New, validate before coding |
| Yandex | `YandexPublisher` | metadata key, likely `user-data` | New, validate before coding |

## Validation Requirements

For each new and migrated provider, verify against the Pulumi registry schema
before implementation:

- resource token exists
- resource has the expected image, OS, or template field
- resource has the expected user-data, cloud-init, metadata, custom-data, or
  cloud-init disk field
- plain vs base64 expectations are understood
- required networking fields are represented in component arguments
- Ubuntu 22.04 can be selected by image ID, image slug, template ID, or
  marketplace reference

The two providers with extra confirmation requirements are:

- TencentCloud: confirm whether `userDataRaw` or `userData` is the correct field
  for plain cloud-init payloads.
- Yandex: confirm the metadata key convention for cloud-init user data.

## Testing

Add or update tests for:

- bootstrap renderer payload content
- adapter output for plain, base64, metadata, custom-data, cloud-init disk, and
  GuestInfo modes
- component mocks proving each provider receives the expected payload placement
- schema generation exposing all new components
- SDK generation for TypeScript, Python, C#, Go, Java, and Rust
- docs and GitHub Pages build

## Documentation

Update:

- `README.md`
- GitHub Pages admin and provider docs
- installation and configuration docs
- provider comparison tables
- examples for Pulumi CLI, TypeScript, Python, C#, Go, Java, and Rust where
  supported

Pulumi Registry docs should include only officially supported Registry languages
and should not mention Rust as an official Registry SDK.
