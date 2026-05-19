---
title: GCP Component
---

# GCP Component

`GcpPublisher` creates one Compute Engine instance per publisher name.

Required inputs: `project`, `zone`, `network`, `subnetwork`, and `image`.
Use a standard Linux image such as:

```text
projects/ubuntu-os-cloud/global/images/family/ubuntu-2204-lts
```

GCP does not provide a public Netskope Publisher image. By default,
`GcpPublisher` runs Netskope's generic bootstrap script from cloud-init
and then runs the publisher wizard with the generated registration
token.

Optional inputs include `machineType`, `assignPublicIp`, `networkTags`,
`serviceAccount`, `bootstrap`, `bootstrapUrl`, `nonat`, `installUser`,
`installUserPassword`, `installUserPasswordIsHash`,
`installUserSshAuthorizedKeys`, `deleteDefaultUser`,
`guestNetworkInterface`, `tags`, `namePrefix`, `names`, and `replicas`.

Set `bootstrap: false` only when booting a custom image that already has
the publisher software and `npa_publisher_wizard` installed.

Bootstrap mode also lets cloud-init manage the Linux install user,
authorized SSH keys, an optional password, and an optional netplan
override for the guest's primary interface.

Outputs: `publisherNames` and secret `publishers`.
