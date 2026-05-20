---
title: Shared cloud-init and user-data adapters
---

# Shared cloud-init and user-data adapters

Bootstrap-mode VM components share one Netskope Publisher cloud-init
payload. The shared renderer creates the install and enrollment script
once per publisher, then each platform component adapts that payload to
the provider-specific field expected by the Pulumi registry package.

The shared payload handles:

- Netskope Publisher registration token injection.
- Optional generic publisher software bootstrap.
- `npa_publisher_wizard` execution.
- `nonat` wizard mode.
- Install user, password, SSH key, and default-user cleanup options.
- Optional guest network interface cloud-init configuration.

## Adapter placement

| Adapter mode | Placement | Components |
|---|---|---|
| Plain user data | `userData`, `user_data`, or equivalent plain text field | AWS, Hcloud, OpenStack, OVH, Proxmox VE snippet content, DigitalOcean, Vultr, Exoscale, UpCloud, Stackit, Equinix Metal, Outscale, OpenTelekomCloud |
| Base64 user data | Provider field expects base64 encoded cloud-init | Alicloud, Nutanix |
| Metadata user data | Metadata map key such as `user-data` | GCP, Yandex Cloud |
| Custom data | Azure Linux VM custom data | Azure |
| OCI metadata | Metadata key `userData` with base64 payload | OCI |
| Scaleway dual placement | `cloudInit` and `userData["cloud-init"]` | Scaleway |
| VMware GuestInfo | `guestinfo.userdata` and encoding metadata | vSphere, ESXi Native |
| Raw user data | Provider field expects raw startup payload | TencentCloud |

## Bootstrap images

Bootstrap-mode components expect an Ubuntu 22.04 image or template with
cloud-init enabled. The component does not create provider images for
you; supply the provider-specific image, template, or OS identifier that
represents Ubuntu 22.04 in your account or region.

Use `bootstrap: false` only when your image already contains the
Netskope Publisher software and `npa_publisher_wizard`. In that mode the
shared cloud-init payload skips the generic bootstrap download and runs
only the enrollment step.

## Customization inputs

VM-backed bootstrap components share these customization inputs:

| Input | Purpose |
|---|---|
| `bootstrap` | Enable or disable the generic publisher software bootstrap. |
| `bootstrapUrl` | Override the generic bootstrap script URL. |
| `nonat` | Pass non-NAT mode to the publisher wizard. |
| `installUser` | Cloud-init user to create or update. |
| `installUserPassword` | Secret password or password hash for the install user. |
| `installUserPasswordIsHash` | Treat `installUserPassword` as a pre-hashed password. |
| `installUserSshAuthorizedKeys` | SSH public keys for the install user. |
| `deleteDefaultUser` | Remove the default Ubuntu user after provisioning. |
| `guestNetworkInterface` | Static guest network configuration rendered into cloud-init. |

See the individual component pages for the platform-specific image,
network, size, and security group inputs.
