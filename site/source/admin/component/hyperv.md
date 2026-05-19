---
title: Hyper-V Component
toc: true
---

# Hyper-V Component

`HypervPublisher` is experimental. The upstream Pulumi Hyper-V provider
exists in `pulumi/pulumi-hyperv`, but `@pulumi/hyperv` is not published
to npm, so this component currently acts as an explicit gate.

It requires `enableExperimentalHyperv: true` and then fails with a clear
dependency message until a stable package source exists.

Use the official Netskope VHDX when preparing Hyper-V images:

```text
https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.vhdx
```

## Planned inputs

The current gate does not create VMs. The intended Hyper-V shape mirrors
the Terraform module: virtual switch, VHDX source, CPU and memory sizing,
publisher naming, tenant registration, and NoCloud seed data.

## Pulumi CLI

```bash
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:apiToken --secret
pulumi config set enableExperimentalHyperv true
```

## TypeScript

```ts
import { HypervPublisher } from "@johninnl/pulumi-netskope-publisher";

new HypervPublisher("publisher", {
  enableExperimentalHyperv: true,
});
```

## Python

The generated Python SDK includes `HypervPublisher`, but the component
has the same experimental limitation as TypeScript:

```python
from pulumi_netskope_publisher import HypervPublisher

HypervPublisher("publisher", enable_experimental_hyperv=True)
```

## C#

```csharp
new HypervPublisher("publisher", new HypervPublisherArgs
{
    EnableExperimentalHyperv = true,
});
```

## Go

```go
_, err := netskopepublisher.NewHypervPublisher(ctx, "publisher", &netskopepublisher.HypervPublisherArgs{
	EnableExperimentalHyperv: pulumi.Bool(true),
})
```
