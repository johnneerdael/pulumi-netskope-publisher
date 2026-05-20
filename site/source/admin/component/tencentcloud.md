---
title: TencentCloud Component
toc: true
---

# TencentCloud Component

`TencentcloudPublisher` creates one Ubuntu 22.04 bootstrap-mode Netskope Publisher per resolved publisher name.

Required platform inputs: `availabilityZone` and `imageId`. The provider uses `userDataRaw` for plain cloud-init.

## Inputs

Required Netskope inputs: `tenantUrl` and `bearerToken`, unless `registrations` is provided. OAuth2 enrollment is available with `authMode: "oauth2"` and `oauth2` client credentials.

## Bootstrap behavior

The component renders the shared Netskope Publisher cloud-init payload, installs the generic publisher software, and enrolls the VM with a deployment-time registration token.

## Pulumi CLI

```bash
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:bearerToken --secret
pulumi up
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { TencentcloudPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new TencentcloudPublisher("publisher", {
  namePrefix: "pub",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  availabilityZone: "ap-guangzhou-6",
  imageId: "img-ubuntu22",
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import TencentcloudPublisher

netskope = pulumi.Config("netskope")
publisher = TencentcloudPublisher(
    "publisher",
    name_prefix="pub",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    availability_zone="ap-guangzhou-6",
    image_id="img-ubuntu22",
)
pulumi.export("publishers", publisher.publishers)
```

## C#

```csharp
using Pulumi;
using JohninNL.Pulumi.NetskopePublisher;

return await Deployment.RunAsync(() =>
{
    var netskope = new Config("netskope");
    var publisher = new TencentcloudPublisher("publisher", new TencentcloudPublisherArgs
    {
        NamePrefix = "pub",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        AvailabilityZone = "ap-guangzhou-6",
        ImageId = "img-ubuntu22",
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewTencentcloudPublisher(ctx, "publisher", &netskopepublisher.TencentcloudPublisherArgs{
	NamePrefix: pulumi.String("pub"),
	Replicas: pulumi.Int(2),
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	BearerToken: netskope.RequireSecret("bearerToken"),
	AvailabilityZone: pulumi.String("ap-guangzhou-6"),
	ImageId: pulumi.String("img-ubuntu22"),
})
_ = publisher
```

## Java

```java
var publisher = new TencentcloudPublisher("publisher", TencentcloudPublisherArgs.builder()
    .namePrefix("pub")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .availabilityZone("ap-guangzhou-6")
    .imageId("img-ubuntu22")
    .build());
```

## Rust

```rust
let publisher = netskope::tencentcloud_publisher::create(
    ctx,
    "publisher",
    netskope::tencentcloud_publisher::TencentcloudPublisherArgs::builder()
        .name_prefix("pub")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .availability_zone("ap-guangzhou-6")
        .image_id("img-ubuntu22")
        .build_struct(),
);
```
