---
title: Hyper-V Component
---

# Hyper-V Component

`HypervPublisher` is an experimental gate. The upstream Pulumi Hyper-V
provider exists in `pulumi/pulumi-hyperv`, but `@pulumi/hyperv` is not
published to npm.

The component requires `enableExperimentalHyperv: true` and then fails
with a clear dependency message until a stable package source exists.

Use the official Netskope VHDX when preparing Hyper-V images:

```text
https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.vhdx
```
