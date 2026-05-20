---
title: Provider Matrix
---

# Provider Matrix

| Platform | Component | Status |
|---|---|---|
| AWS | `AwsPublisher` | Supported |
| Azure | `AzurePublisher` | Supported |
| GCP | `GcpPublisher` | Supported |
| Kubernetes | `KubernetesPublisher` | Supported |
| vSphere | `VspherePublisher` | Supported |
| ESXi Native | `EsxiPublisher` | Supported direct-host ESXi |
| Hcloud | `HcloudPublisher` | Supported bootstrap mode |
| Nutanix | `NutanixPublisher` | Supported bootstrap mode |
| OpenStack | `OpenstackPublisher` | Supported bootstrap mode |
| OVH Public Cloud | `OvhPublisher` | Supported bootstrap mode |
| Scaleway | `ScalewayPublisher` | Supported bootstrap mode |
| OCI | `OciPublisher` | Supported bootstrap mode |
| Alicloud | `AlicloudPublisher` | Supported bootstrap mode |
| Proxmox VE | `ProxmoxvePublisher` | Supported bootstrap mode from template clone |
| Hyper-V | `HypervPublisher` | Experimental gate |

All supported providers share name derivation, Netskope registration,
cloud-init generation, and secret output conventions.

ESXi Native is direct-host ESXi support and does not replace the vSphere
component. Hcloud, Nutanix, OpenStack, OVH, Scaleway, OCI, Alicloud, and
Proxmox VE use bootstrap mode on Ubuntu 22.04 images or templates.

Official Netskope image sources:

- Hyper-V VHDX:
  `https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.vhdx`
- vSphere OVA:
  `https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.ova`

GCP has no public Netskope Publisher image. Use a standard Linux image
and let `GcpPublisher` run the generic bootstrap script, or supply a
custom pre-baked image and set `bootstrap: false`.

Kubernetes uses the `kubernetes-netskope-publisher` Helm chart and
supports both token and API enrollment modes.
