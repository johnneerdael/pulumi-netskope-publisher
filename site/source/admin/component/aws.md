---
title: AWS Component
---

# AWS Component

`AwsPublisher` creates one EC2 instance per publisher name.

## Required inputs

- `subnetId`
- `securityGroupIds`
- `tenantUrl` and `apiToken`, unless `registrations` is provided

## Common optional inputs

- `namePrefix`
- `names`
- `replicas`
- `amiId`
- `instanceType`
- `keyName`
- `tags`
- `bootstrap`
- `bootstrapUrl`
- `nonat`
- `installUser`
- `installUserPassword`
- `installUserPasswordIsHash`
- `installUserSshAuthorizedKeys`
- `deleteDefaultUser`
- `guestNetworkInterface`

When `bootstrap: true` and `amiId` is omitted, the component resolves a
Canonical Ubuntu 22.04 Minimal AMI and installs the publisher through
Netskope's generic bootstrap script. Leave `bootstrap` unset or false to
use the Netskope Publisher AMI path.

Bootstrap mode also lets cloud-init manage the Linux install user,
authorized SSH keys, an optional password, and an optional netplan
override for the guest's primary interface.

## Outputs

- `publisherNames`
- `publishers`

`publishers` is keyed by publisher name and contains publisher ID,
registration token, EC2 instance ID, private IP, and public IP.
