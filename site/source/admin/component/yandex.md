---
title: Yandex Cloud Component
toc: true
---

# Yandex Cloud Component

`YandexPublisher` creates one Ubuntu 22.04 bootstrap-mode Netskope Publisher per resolved publisher name.

Required platform inputs: `imageId` and `subnetId`. The provider places cloud-init in metadata key `user-data`.

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
import { YandexPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new YandexPublisher("publisher", {
  namePrefix: "pub",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  imageId: "ubuntu-22-image",
  subnetId: "subnet-id",
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import YandexPublisher

netskope = pulumi.Config("netskope")
publisher = YandexPublisher(
    "publisher",
    name_prefix="pub",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    image_id="ubuntu-22-image",
    subnet_id="subnet-id",
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
    var publisher = new YandexPublisher("publisher", new YandexPublisherArgs
    {
        NamePrefix = "pub",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        ImageId = "ubuntu-22-image",
        SubnetId = "subnet-id",
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewYandexPublisher(ctx, "publisher", &netskopepublisher.YandexPublisherArgs{
	NamePrefix: pulumi.String("pub"),
	Replicas: pulumi.Int(2),
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	BearerToken: netskope.RequireSecret("bearerToken"),
	ImageId: pulumi.String("ubuntu-22-image"),
	SubnetId: pulumi.String("subnet-id"),
})
_ = publisher
```

## Java

```java
var publisher = new YandexPublisher("publisher", YandexPublisherArgs.builder()
    .namePrefix("pub")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .imageId("ubuntu-22-image")
    .subnetId("subnet-id")
    .build());
```

## Rust

```rust
let publisher = netskope::yandex_publisher::create(
    ctx,
    "publisher",
    netskope::yandex_publisher::YandexPublisherArgs::builder()
        .name_prefix("pub")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .image_id("ubuntu-22-image")
        .subnet_id("subnet-id")
        .build_struct(),
);
```
