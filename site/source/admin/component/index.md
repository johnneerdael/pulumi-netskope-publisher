---
title: Components
---

# Components

The package exposes provider-specific publisher components. Each page
lists required inputs, common optional inputs, outputs, and examples for
the Pulumi CLI, TypeScript, Python, C#, Go, Java, and Rust.

- [AWS](/pulumi-netskope-publisher/admin/component/aws/)
- [Azure](/pulumi-netskope-publisher/admin/component/azure/)
- [GCP](/pulumi-netskope-publisher/admin/component/gcp/)
- [Kubernetes](/pulumi-netskope-publisher/admin/component/kubernetes/)
- [vSphere](/pulumi-netskope-publisher/admin/component/vsphere/)
- [ESXi Native](/pulumi-netskope-publisher/admin/component/esxi/)
- [Hcloud](/pulumi-netskope-publisher/admin/component/hcloud/)
- [Nutanix](/pulumi-netskope-publisher/admin/component/nutanix/)
- [OpenStack](/pulumi-netskope-publisher/admin/component/openstack/)
- [OVH](/pulumi-netskope-publisher/admin/component/ovh/)
- [Scaleway](/pulumi-netskope-publisher/admin/component/scaleway/)
- [OCI](/pulumi-netskope-publisher/admin/component/oci/)
- [Alicloud](/pulumi-netskope-publisher/admin/component/alicloud/)
- [Hyper-V (experimental)](/pulumi-netskope-publisher/admin/component/hyperv/)
- [Netskope Registration](/pulumi-netskope-publisher/admin/component/registration/)

The TypeScript package and Go executable provider both support managed
Netskope registration or pre-created registration tokens. Kubernetes
deployments additionally support chart API enrollment.

## Shared inputs

| Input | Description |
|---|---|
| `tenantUrl` | Netskope tenant URL. Required unless `registrations` is supplied. |
| `bearerToken` | Secret Netskope API token. Required unless `registrations` is supplied. |
| `registrations` | Pre-created publisher IDs and registration tokens keyed by publisher name. |
| `namePrefix` | Prefix used when deriving names. Defaults to `npa-publisher`. |
| `names` | Explicit publisher names. Overrides `namePrefix` and `replicas`. |
| `replicas` | Number of derived names when `names` is omitted. |
| `tags` | Platform tags or labels where supported. |
| `wizardPath` | Absolute path to `npa_publisher_wizard`. |

ESXi Native is direct-host ESXi support and does not replace the vSphere
component. Hcloud, Nutanix, OpenStack, OVH, Scaleway, OCI, and Alicloud
use bootstrap mode on Ubuntu 22.04 images.

VM-backed cloud providers accept bootstrap and install-user controls:
`bootstrap`, `bootstrapUrl`, `nonat`, `installUser`,
`installUserPassword`, `installUserPasswordIsHash`,
`installUserSshAuthorizedKeys`, `deleteDefaultUser`, and
`guestNetworkInterface`.

## Shared outputs

| Output | Description |
|---|---|
| `publisherNames` | Resolved publisher names. |
| `publishers` | Secret map keyed by publisher name. Includes `publisherId`, `registrationToken`, `vmId`, `privateIp`, and `publicIp` when applicable. |

Kubernetes also returns `helmReleaseNames`.
