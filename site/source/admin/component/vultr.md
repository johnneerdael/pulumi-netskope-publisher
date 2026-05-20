---
title: Vultr Component
toc: true
---

# Vultr Component

`VultrPublisher` creates one Ubuntu 22.04 bootstrap-mode Netskope Publisher per resolved publisher name.

Required platform inputs: `region` and `plan`. Supply either `osId` or `imageId` for an Ubuntu 22.04 image.

## Inputs

Required Netskope inputs: `tenantUrl` and `bearerToken`, unless `registrations` is provided. OAuth2 enrollment is available with `authMode: "oauth2"` and `oauth2` client credentials.

## Bootstrap behavior

The component renders the shared Netskope Publisher cloud-init payload, installs the generic publisher software, and enrolls the VM with a deployment-time registration token.

## Pulumi YAML

```yaml
name: netskope-publisher-vultr
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:VultrPublisher
    properties:
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      region: ams
      plan: vc2-2c-4gb
      osId: 1743
      bootstrap: true
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { VultrPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new VultrPublisher("publisher", {
  namePrefix: "pub",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  region: "ams",
  plan: "vc2-2c-4gb",
  osId: 1743,
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import VultrPublisher

netskope = pulumi.Config("netskope")
publisher = VultrPublisher(
    "publisher",
    name_prefix="pub",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    region="ams",
    plan="vc2-2c-4gb",
    os_id=1743,
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
    var publisher = new VultrPublisher("publisher", new VultrPublisherArgs
    {
        NamePrefix = "pub",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        Region = "ams",
        Plan = "vc2-2c-4gb",
        OsId = 1743,
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewVultrPublisher(ctx, "publisher", &netskopepublisher.VultrPublisherArgs{
	NamePrefix: pulumi.String("pub"),
	Replicas: pulumi.Int(2),
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	BearerToken: netskope.RequireSecret("bearerToken"),
	Region: pulumi.String("ams"),
	Plan: pulumi.String("vc2-2c-4gb"),
	OsId: pulumi.Int(1743),
})
_ = publisher
```

## Java

```java
var publisher = new VultrPublisher("publisher", VultrPublisherArgs.builder()
    .namePrefix("pub")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .region("ams")
    .plan("vc2-2c-4gb")
    .osId(1743)
    .build());
```

## Rust

```rust
let publisher = netskope::vultr_publisher::create(
    ctx,
    "publisher",
    netskope::vultr_publisher::VultrPublisherArgs::builder()
        .name_prefix("pub")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .region("ams")
        .plan("vc2-2c-4gb")
        .os_id(1743)
        .build_struct(),
);
```
