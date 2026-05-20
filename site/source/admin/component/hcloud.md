---
title: Hcloud Component
toc: true
---

# Hcloud Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`HcloudPublisher` creates one Hetzner Cloud server per publisher name.

## Inputs

Required: `tenantUrl` and `bearerToken`, unless `registrations` is provided.

Optional platform inputs: `serverType`, `image`, `location`, `datacenter`,
`sshKeys`, `firewallIds`, `networkId`, and `assignPublicIp`.

## Image and bootstrap behavior

The default image is `ubuntu-22.04`. Hcloud uses bootstrap mode and runs
Netskope's generic installer through cloud-init.

## Outputs

`publisherNames` and secret `publishers`, keyed by publisher name.

## Pulumi CLI

```bash
pulumi config set hcloud:token --secret
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:bearerToken --secret
pulumi up
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { HcloudPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new HcloudPublisher("publisher", {
  namePrefix: "pub-fsn",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  location: "fsn1",
  serverType: "cx22",
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import HcloudPublisher

netskope = pulumi.Config("netskope")
publisher = HcloudPublisher(
    "publisher",
    name_prefix="pub-fsn",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    location="fsn1",
    server_type="cx22",
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
    var publisher = new HcloudPublisher("publisher", new HcloudPublisherArgs
    {
        NamePrefix = "pub-fsn",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        Location = "fsn1",
        ServerType = "cx22",
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewHcloudPublisher(ctx, "publisher", &netskopepublisher.HcloudPublisherArgs{
	NamePrefix: pulumi.String("pub-fsn"),
	Replicas:   pulumi.Int(2),
	TenantUrl:  pulumi.String(netskope.Require("tenantUrl")),
	BearerToken:   netskope.RequireSecret("bearerToken"),
	Location:   pulumi.String("fsn1"),
	ServerType: pulumi.String("cx22"),
})
```

## Java

```java
var publisher = new HcloudPublisher("publisher", HcloudPublisherArgs.builder()
    .namePrefix("pub-fsn")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .location("fsn1")
    .serverType("cx22")
    .build());
```

## Rust

```rust
let publisher = netskope::hcloud_publisher::create(
    ctx,
    "publisher",
    netskope::hcloud_publisher::HcloudPublisherArgs::builder()
        .name_prefix("pub-fsn")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .location("fsn1")
        .server_type("cx22")
        .build_struct(),
);
```
