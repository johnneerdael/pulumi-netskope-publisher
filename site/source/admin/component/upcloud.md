---
title: UpCloud Component
toc: true
---

# UpCloud Component

`UpcloudPublisher` creates one Ubuntu 22.04 bootstrap-mode Netskope Publisher per resolved publisher name.

Required platform input: `zone`. Optional inputs include `plan`, `template`, and `networkInterfaces`. The default template targets Ubuntu 22.04.

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
import { UpcloudPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new UpcloudPublisher("publisher", {
  namePrefix: "pub",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  zone: "nl-ams1",
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import UpcloudPublisher

netskope = pulumi.Config("netskope")
publisher = UpcloudPublisher(
    "publisher",
    name_prefix="pub",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    zone="nl-ams1",
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
    var publisher = new UpcloudPublisher("publisher", new UpcloudPublisherArgs
    {
        NamePrefix = "pub",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        Zone = "nl-ams1",
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewUpcloudPublisher(ctx, "publisher", &netskopepublisher.UpcloudPublisherArgs{
	NamePrefix: pulumi.String("pub"),
	Replicas: pulumi.Int(2),
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	BearerToken: netskope.RequireSecret("bearerToken"),
	Zone: pulumi.String("nl-ams1"),
})
_ = publisher
```

## Java

```java
var publisher = new UpcloudPublisher("publisher", UpcloudPublisherArgs.builder()
    .namePrefix("pub")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .zone("nl-ams1")
    .build());
```

## Rust

```rust
let publisher = netskope::upcloud_publisher::create(
    ctx,
    "publisher",
    netskope::upcloud_publisher::UpcloudPublisherArgs::builder()
        .name_prefix("pub")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .zone("nl-ams1")
        .build_struct(),
);
```
