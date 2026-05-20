---
title: Architecture overview
---

# Architecture overview

`pulumi-netskope-publisher` exposes provider-specific Pulumi components
instead of one switchable root component. Import only the component for
the platform you are provisioning:

```ts
import { AwsPublisher } from "@johninnl/pulumi-netskope-publisher";
```

Each platform component follows the same pattern:

1. Resolve publisher names from `names`, or from `namePrefix` and
   `replicas`.
2. Create or reuse Netskope publisher records through
   `NetskopeRegistration`, unless `registrations` is supplied directly.
3. Render per-publisher cloud-init with the registration token.
4. Provision the VM, pod, or chart resources for the target platform.
5. Return `publisherNames` and secret `publishers` outputs keyed by
   publisher name.

## Two install paths

| Path | Image | Cloud-init behavior |
|---|---|---|
| Bootstrap | Stock Ubuntu image | Runs Netskope's generic `bootstrap.sh`, then `npa_publisher_wizard -token <token>`. |
| Pre-baked | Netskope Publisher image, marketplace image, custom image, OVA, or VHDX | Runs the wizard already present on the image. |

AWS and Azure support both paths. GCP uses bootstrap mode by default
because there is no public Netskope Publisher image. vSphere clones an
existing template, Hyper-V is currently an experimental gate, and
Kubernetes installs the publisher Helm chart instead of booting a VM.
Hcloud, Nutanix, OpenStack, OVH, Scaleway, OCI, Alicloud, and Proxmox
VE use bootstrap mode.

## What each component owns

| Component | Responsibility | Main providers |
|---|---|---|
| `NetskopeRegistration` | List/create publisher records and generate registration tokens | Pulumi dynamic provider |
| `AwsPublisher` | EC2 instances, optional Canonical AMI lookup, EC2 user data | `@pulumi/aws` |
| `AzurePublisher` | NICs, optional public IPs, Linux VMs, custom data | `@pulumi/azure-native` |
| `GcpPublisher` | Compute Engine instances and metadata user-data | `@pulumi/gcp` |
| `VspherePublisher` | VM clones with guestinfo cloud-init | `@pulumi/vsphere` |
| `EsxiPublisher` | Direct ESXi VM creation with guestinfo cloud-init | `@pulumiverse/esxi-native` |
| `HcloudPublisher` | Hetzner Cloud servers and user data | `@pulumi/hcloud` |
| `NutanixPublisher` | Nutanix VMs and guest customization cloud-init | `@pierskarsenbarg/nutanix` |
| `OpenstackPublisher` | OpenStack instances and optional floating IPs | `@pulumi/openstack` |
| `OvhPublisher` | OVH Public Cloud instances and user data | `@ovhcloud/pulumi-ovh` |
| `ScalewayPublisher` | Scaleway instances and cloud-init | `@pulumiverse/scaleway` |
| `OciPublisher` | OCI instances and metadata user data | `@pulumi/oci` |
| `AlicloudPublisher` | Alibaba Cloud ECS instances and user data | `@pulumi/alicloud` |
| `ProxmoxvePublisher` | Proxmox VE cloud-init snippets and VM template clones | `@muhlba91/pulumi-proxmoxve` |
| `KubernetesPublisher` | Helm chart releases and token/API Secrets | `@pulumi/kubernetes` |
| `HypervPublisher` | Experimental placeholder until `@pulumi/hyperv` is published | none |

## Provider isolation

Use the component that matches the provider already configured in the
stack. A stack that only creates `AwsPublisher` resources does not need
Azure, GCP, vSphere, or Kubernetes provider credentials.

See also: [Registration flow](/pulumi-netskope-publisher/admin/concepts/registration-flow/).
