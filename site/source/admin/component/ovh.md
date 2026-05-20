---
title: OVH Component
toc: true
---

# OVH Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`OvhPublisher` creates one OVH Public Cloud instance per publisher name.

## Inputs

Required: `serviceName`, `region`, `imageId`, `flavorId`, and
`tenantUrl` plus `bearerToken` unless `registrations` is provided.

Optional platform inputs: `sshKeyName` and `networkId`.

## Image and bootstrap behavior

Use an Ubuntu 22.04 image ID from the OVH Public Cloud project. The
component uses bootstrap mode and passes cloud-init to the instance
`userData` field.

## Outputs

`publisherNames` and secret `publishers`, keyed by publisher name.

## Pulumi CLI

```bash
pulumi config set ovh:endpoint ovh-eu
pulumi config set ovh:applicationKey --secret
pulumi config set ovh:applicationSecret --secret
pulumi config set ovh:consumerKey --secret
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:bearerToken --secret
pulumi up
```

## TypeScript

```ts
const publisher = new OvhPublisher("publisher", {
  namePrefix: "pub-ovh",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  serviceName: config.require("serviceName"),
  region: "GRA11",
  imageId: config.require("imageId"),
  flavorId: config.require("flavorId"),
  sshKeyName: config.get("sshKeyName"),
});
```

## Python

```python
publisher = OvhPublisher(
    "publisher",
    name_prefix="pub-ovh",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    service_name=config.require("serviceName"),
    region="GRA11",
    image_id=config.require("imageId"),
    flavor_id=config.require("flavorId"),
    ssh_key_name=config.get("sshKeyName"),
)
```

## C#

```csharp
var publisher = new OvhPublisher("publisher", new OvhPublisherArgs
{
    NamePrefix = "pub-ovh",
    Replicas = 2,
    TenantUrl = netskope.Require("tenantUrl"),
    BearerToken = netskope.RequireSecret("bearerToken"),
    ServiceName = config.Require("serviceName"),
    Region = "GRA11",
    ImageId = config.Require("imageId"),
    FlavorId = config.Require("flavorId"),
    SshKeyName = config.Get("sshKeyName"),
});
```

## Go

```go
publisher, err := netskopepublisher.NewOvhPublisher(ctx, "publisher", &netskopepublisher.OvhPublisherArgs{
	NamePrefix: pulumi.String("pub-ovh"),
	Replicas:   pulumi.Int(2),
	TenantUrl:  pulumi.String(netskope.Require("tenantUrl")),
	BearerToken:   netskope.RequireSecret("bearerToken"),
	ServiceName: pulumi.String(cfg.Require("serviceName")),
	Region:     pulumi.String("GRA11"),
	ImageId:    pulumi.String(cfg.Require("imageId")),
	FlavorId:   pulumi.String(cfg.Require("flavorId")),
})
```

## Java

```java
var publisher = new OvhPublisher("publisher", OvhPublisherArgs.builder()
    .namePrefix("pub-ovh")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .serviceName(config.require("serviceName"))
    .region("GRA11")
    .imageId(config.require("imageId"))
    .flavorId(config.require("flavorId"))
    .build());
```

## Rust

```rust
let publisher = netskope::ovh_publisher::create(
    ctx,
    "publisher",
    netskope::ovh_publisher::OvhPublisherArgs::builder()
        .name_prefix("pub-ovh")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .service_name("project-id")
        .region("GRA11")
        .image_id("ubuntu-image-id")
        .flavor_id("flavor-id")
        .build_struct(),
);
```
