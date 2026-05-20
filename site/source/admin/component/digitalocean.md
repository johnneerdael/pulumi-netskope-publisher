---
title: DigitalOcean Component
toc: true
---

# DigitalOcean Component

`DigitaloceanPublisher` creates one Ubuntu 22.04 bootstrap-mode Netskope Publisher per resolved publisher name.

Required platform input: `region`. Optional inputs include `size`, `image`, `sshKeys`, `vpcUuid`, `monitoring`, and `ipv6`. The default image is `ubuntu-22-04-x64`.

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
import { DigitaloceanPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new DigitaloceanPublisher("publisher", {
  namePrefix: "pub",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  region: "ams3",
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import DigitaloceanPublisher

netskope = pulumi.Config("netskope")
publisher = DigitaloceanPublisher(
    "publisher",
    name_prefix="pub",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    region="ams3",
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
    var publisher = new DigitaloceanPublisher("publisher", new DigitaloceanPublisherArgs
    {
        NamePrefix = "pub",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        Region = "ams3",
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewDigitaloceanPublisher(ctx, "publisher", &netskopepublisher.DigitaloceanPublisherArgs{
	NamePrefix: pulumi.String("pub"),
	Replicas: pulumi.Int(2),
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	BearerToken: netskope.RequireSecret("bearerToken"),
	Region: pulumi.String("ams3"),
})
_ = publisher
```

## Java

```java
var publisher = new DigitaloceanPublisher("publisher", DigitaloceanPublisherArgs.builder()
    .namePrefix("pub")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .region("ams3")
    .build());
```

## Rust

```rust
let publisher = netskope::digitalocean_publisher::create(
    ctx,
    "publisher",
    netskope::digitalocean_publisher::DigitaloceanPublisherArgs::builder()
        .name_prefix("pub")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .region("ams3")
        .build_struct(),
);
```
