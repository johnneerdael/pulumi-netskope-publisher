---
title: Bring your own image
toc: true
---

# Bring your own image

## Bootstrap mode

Use bootstrap mode when the VM starts from stock Ubuntu and installs the
publisher during cloud-init.

```ts
new AwsPublisher("publisher", {
  tenantUrl,
  bearerToken,
  subnetId,
  securityGroupIds: [securityGroupId],
  bootstrap: true,
});
```

AWS and Azure resolve Canonical Ubuntu 22.04 Minimal automatically when
bootstrap is true and no image is supplied. GCP expects an Ubuntu image
and defaults to bootstrap behavior.

Override `bootstrapUrl` when the VM must download the script from an
internal mirror.

## Pre-baked image mode

Use pre-baked mode when the image already includes
`npa_publisher_wizard` at `wizardPath`.

| Platform | Input |
|---|---|
| AWS | `amiId` |
| Azure | `imageId` or `marketplace` |
| GCP | `image` |
| vSphere | `templateName` |

```ts
new AwsPublisher("publisher", {
  tenantUrl,
  bearerToken,
  subnetId,
  securityGroupIds: [securityGroupId],
  bootstrap: false,
  amiId: "ami-0123456789abcdef0",
});
```

The image must have cloud-init enabled and must contain the install user
or allow cloud-init to create it.
