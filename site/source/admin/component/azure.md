---
title: Azure Component
---

# Azure Component

`AzurePublisher` creates one Linux virtual machine per publisher name.

Required inputs: `resourceGroupName`, `location`, `subnetId`,
`adminSshPublicKey`, and one of `imageId`, `marketplace`, or
`bootstrap: true`.

Optional inputs include `vmSize`, `adminUsername`,
`networkSecurityGroupId`, `assignPublicIp`, `osDisk`, `tags`,
`namePrefix`, `names`, `replicas`, `bootstrap`, `bootstrapUrl`,
`nonat`, `installUser`, `installUserPassword`,
`installUserPasswordIsHash`, `installUserSshAuthorizedKeys`,
`deleteDefaultUser`, and `guestNetworkInterface`.

When `bootstrap: true` and no image is supplied, the component uses the
Canonical Ubuntu 22.04 Minimal marketplace image and installs the
publisher through Netskope's generic bootstrap script. `adminUsername`
defaults to `installUser`, so the Azure admin account and cloud-init
install user stay aligned.

Bootstrap mode also lets cloud-init manage the Linux install user,
extra authorized SSH keys, an optional password, and an optional netplan
override for the guest's primary interface.

Outputs: `publisherNames` and secret `publishers`.
