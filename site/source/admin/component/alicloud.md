---
title: Alicloud Component
toc: true
---

# Alicloud Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`AlicloudPublisher` creates one Alibaba Cloud ECS instance per publisher
name.

## Inputs

Required: `imageId`, `vswitchId`, `securityGroupIds`, and `tenantUrl`
plus `bearerToken` unless `registrations` is provided.

Optional platform inputs: `instanceType`, `keyName`, and
`allocatePublicIp`.

## Image and bootstrap behavior

Alicloud requires an Ubuntu 22.04 ECS image ID. The component uses
bootstrap mode and passes base64 cloud-init through `userData`.

## Outputs

`publisherNames` and secret `publishers`, keyed by publisher name.

## Pulumi CLI

```bash
pulumi config set alicloud:region eu-central-1
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:bearerToken --secret
pulumi up
```

## TypeScript

```ts
const publisher = new AlicloudPublisher("publisher", {
  namePrefix: "pub-ali",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  imageId: config.require("imageId"),
  vswitchId: config.require("vswitchId"),
  securityGroupIds: [config.require("securityGroupId")],
  instanceType: "ecs.t6-c1m2.large",
});
```

## Python

```python
publisher = AlicloudPublisher(
    "publisher",
    name_prefix="pub-ali",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    image_id=config.require("imageId"),
    vswitch_id=config.require("vswitchId"),
    security_group_ids=[config.require("securityGroupId")],
    instance_type="ecs.t6-c1m2.large",
)
```

## C#

```csharp
var publisher = new AlicloudPublisher("publisher", new AlicloudPublisherArgs
{
    NamePrefix = "pub-ali",
    Replicas = 2,
    TenantUrl = netskope.Require("tenantUrl"),
    BearerToken = netskope.RequireSecret("bearerToken"),
    ImageId = config.Require("imageId"),
    VswitchId = config.Require("vswitchId"),
    SecurityGroupIds = { config.Require("securityGroupId") },
    InstanceType = "ecs.t6-c1m2.large",
});
```

## Go

```go
publisher, err := netskopepublisher.NewAlicloudPublisher(ctx, "publisher", &netskopepublisher.AlicloudPublisherArgs{
	NamePrefix:       pulumi.String("pub-ali"),
	Replicas:         pulumi.Int(2),
	TenantUrl:        pulumi.String(netskope.Require("tenantUrl")),
	BearerToken:         netskope.RequireSecret("bearerToken"),
	ImageId:          pulumi.String(cfg.Require("imageId")),
	VswitchId:        pulumi.String(cfg.Require("vswitchId")),
	SecurityGroupIds: pulumi.StringArray{pulumi.String(cfg.Require("securityGroupId"))},
	InstanceType:     pulumi.String("ecs.t6-c1m2.large"),
})
```

## Java

```java
var publisher = new AlicloudPublisher("publisher", AlicloudPublisherArgs.builder()
    .namePrefix("pub-ali")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .imageId(config.require("imageId"))
    .vswitchId(config.require("vswitchId"))
    .securityGroupIds(config.require("securityGroupId"))
    .instanceType("ecs.t6-c1m2.large")
    .build());
```

## Rust

```rust
let publisher = netskope::alicloud_publisher::create(
    ctx,
    "publisher",
    netskope::alicloud_publisher::AlicloudPublisherArgs::builder()
        .name_prefix("pub-ali")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .image_id("ubuntu_22_04_x64_20G_alibase.vhd")
        .vswitch_id("vsw-123")
        .security_group_ids(vec!["sg-123".to_string()])
        .instance_type("ecs.t6-c1m2.large")
        .build_struct(),
);
```
