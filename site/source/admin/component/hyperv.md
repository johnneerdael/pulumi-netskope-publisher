---
title: Hyper-V Component
toc: true
---

# Hyper-V Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

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

## Pulumi YAML

```yaml
name: netskope-publisher-hyperv
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:HypervPublisher
    properties:
      enableExperimentalHyperv: true
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
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

## Java

```java
new HypervPublisher("publisher", HypervPublisherArgs.builder()
    .enableExperimentalHyperv(true)
    .build());
```

## Rust

```rust
let publisher = netskope::hyperv_publisher::create(
    ctx,
    "publisher",
    netskope::hyperv_publisher::HypervPublisherArgs::builder()
        .enable_experimental_hyperv(true)
        .build_struct(),
);
```
