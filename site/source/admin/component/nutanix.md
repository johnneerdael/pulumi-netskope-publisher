---
title: Nutanix Component
toc: true
---

# Nutanix Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`NutanixPublisher` creates one Nutanix VM per publisher name.

## Inputs

Required: `clusterUuid` and `tenantUrl` plus `bearerToken` unless
`registrations` is provided.

Optional platform inputs: `imageUuid`, `subnetUuid`, `numVCpus`,
`numCoresPerVcpu`, and `memorySizeMib`.

## Image and bootstrap behavior

Use an Ubuntu 22.04 image UUID when setting `imageUuid`. The component
uses Nutanix guest customization and passes base64 cloud-init as user
data.

## Outputs

`publisherNames` and secret `publishers`, keyed by publisher name.

## Pulumi CLI

```bash
pulumi config set nutanix:endpoint prism.example.com
pulumi config set nutanix:username admin
pulumi config set nutanix:password --secret
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:bearerToken --secret
pulumi up
```

## TypeScript

```ts
const publisher = new NutanixPublisher("publisher", {
  namePrefix: "pub-ntnx",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  clusterUuid: config.require("clusterUuid"),
  imageUuid: config.require("imageUuid"),
  subnetUuid: config.require("subnetUuid"),
});
```

## Python

```python
publisher = NutanixPublisher(
    "publisher",
    name_prefix="pub-ntnx",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    cluster_uuid=config.require("clusterUuid"),
    image_uuid=config.require("imageUuid"),
    subnet_uuid=config.require("subnetUuid"),
)
```

## C#

```csharp
var publisher = new NutanixPublisher("publisher", new NutanixPublisherArgs
{
    NamePrefix = "pub-ntnx",
    Replicas = 2,
    TenantUrl = netskope.Require("tenantUrl"),
    BearerToken = netskope.RequireSecret("bearerToken"),
    ClusterUuid = config.Require("clusterUuid"),
    ImageUuid = config.Require("imageUuid"),
    SubnetUuid = config.Require("subnetUuid"),
});
```

## Go

```go
publisher, err := netskopepublisher.NewNutanixPublisher(ctx, "publisher", &netskopepublisher.NutanixPublisherArgs{
	NamePrefix:  pulumi.String("pub-ntnx"),
	Replicas:    pulumi.Int(2),
	TenantUrl:   pulumi.String(netskope.Require("tenantUrl")),
	BearerToken:    netskope.RequireSecret("bearerToken"),
	ClusterUuid: pulumi.String(cfg.Require("clusterUuid")),
	ImageUuid:   pulumi.String(cfg.Require("imageUuid")),
	SubnetUuid:  pulumi.String(cfg.Require("subnetUuid")),
})
```

## Java

```java
var publisher = new NutanixPublisher("publisher", NutanixPublisherArgs.builder()
    .namePrefix("pub-ntnx")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .clusterUuid(config.require("clusterUuid"))
    .imageUuid(config.require("imageUuid"))
    .subnetUuid(config.require("subnetUuid"))
    .build());
```

## Rust

```rust
let publisher = netskope::nutanix_publisher::create(
    ctx,
    "publisher",
    netskope::nutanix_publisher::NutanixPublisherArgs::builder()
        .name_prefix("pub-ntnx")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .cluster_uuid("cluster-uuid")
        .image_uuid("image-uuid")
        .subnet_uuid("subnet-uuid")
        .build_struct(),
);
```
