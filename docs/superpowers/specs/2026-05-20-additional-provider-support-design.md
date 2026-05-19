---
title: Additional Provider Support Design
date: 2026-05-20
---

# Additional Provider Support Design

## Goal

Add Pulumi components for additional infrastructure targets while
preserving the existing registration, naming, cloud-init, and output
contracts used by AWS, Azure, GCP, Kubernetes, and vSphere.

## Providers

Add new components:

- `EsxiPublisher`
- `HcloudPublisher`
- `NutanixPublisher`
- `OpenstackPublisher`
- `OvhPublisher`
- `ScalewayPublisher`
- `OciPublisher`
- `AlicloudPublisher`

Keep `VspherePublisher`. ESXi Native provisions directly against an ESXi
host without vCenter; vSphere remains the vCenter/template path. The two
components solve different deployment shapes and should coexist.

## Common Behavior

Each VM-backed component resolves publisher names from `names`, or from
`namePrefix` and `replicas`, then creates or reuses Netskope publisher
registrations unless `registrations` is supplied. Each publisher gets
provider-specific compute resources and a rendered cloud-init payload
containing its registration token.

All new cloud providers except ESXi use bootstrap mode by default and
should not attempt to use Netskope marketplace images. They provision an
Ubuntu 22.04 VM, run Netskope's generic bootstrap script, and then run
`npa_publisher_wizard`.

Each component returns:

- `publisherNames`
- secret `publishers`, keyed by publisher name

Each `publishers` entry should include `publisherId`,
`registrationToken`, `vmId`, `privateIp`, and `publicIp` when the
provider exposes them.

## Image Strategy

Provider image handling is deliberately conservative:

- Use Ubuntu 22.04 defaults only where the provider has a stable, public,
  documented image identifier.
- Always support an explicit image input.
- Require explicit image input where images are tenant, project, region,
  or catalog specific.

This avoids fragile lookups across OpenStack, OCI, Nutanix, OVH, and
other providers whose image catalogs vary by account or region.

## Implementation Shape

Add shared helpers for the repeated VM component flow:

- Resolve names.
- Create or reuse registrations.
- Render per-publisher cloud-init with bootstrap enabled.
- Build secret publisher output maps.

Provider-specific files should own only provider-specific resource
arguments and output extraction. This keeps each component readable and
prevents a single monolithic multi-provider implementation.

Update:

- TypeScript components in `src/`.
- Public exports in `src/index.ts`.
- Input/output types in `src/types.ts`.
- Pulumi schema and Go provider bindings.
- Unit tests with Pulumi mocks for each provider.
- Admin docs, provider matrix, README, registry docs, and SDK examples.
- Generated SDKs, including Java and Rust.

## Testing

For each provider component:

- Assert publisher names resolve correctly.
- Assert generated user-data includes bootstrap behavior.
- Assert VM resources are created with the expected image, network, and
  cloud-init/user-data attachment fields.
- Assert `publishers` is keyed by publisher name and includes VM/IP
  details where available.

Run the existing verification set after implementation:

- `npm run typecheck`
- `npm test`
- `npm run go:test`
- `npm run sdk:gen`
- `npm run registry:check`
- `npm run build` in `site/`
- `cargo check` in `sdk/rust`

Java SDK compilation is verified in CI where Gradle is installed.

## Documentation

Add one admin component page per new provider with:

- Required inputs.
- Common optional inputs.
- Bootstrap/image behavior.
- Pulumi CLI setup.
- TypeScript, Python, C#, Go, Java, and Rust examples.

Update the admin overview and provider matrix so the new providers are
discoverable from GitHub Pages.
