---
title: Outscale Component
toc: true
---

# Outscale Component

`OutscalePublisher` creates one Ubuntu 22.04 bootstrap-mode Netskope Publisher per resolved publisher name.

Required platform input: `imageId`. Optional inputs include `vmType`, `subnetId`, `keypairName`, and `securityGroupIds`.

## Inputs

Required Netskope inputs: `tenantUrl` and `bearerToken`, unless `registrations` is provided. OAuth2 enrollment is available with `authMode: "oauth2"` and `oauth2` client credentials.

## Bootstrap behavior

The component renders the shared Netskope Publisher cloud-init payload, installs the generic publisher software, and enrolls the VM with a deployment-time registration token.

## Pulumi YAML

```yaml
name: netskope-publisher-outscale
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:OutscalePublisher
    properties:
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      imageId: ami-ubuntu2204
      vmType: tinav5.c2r4p2
      subnetId: subnet-example
      securityGroupIds:
        - sg-example
      bootstrap: true
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { OutscalePublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new OutscalePublisher("publisher", {
  namePrefix: "pub",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  imageId: "ami-ubuntu22",
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import OutscalePublisher

netskope = pulumi.Config("netskope")
publisher = OutscalePublisher(
    "publisher",
    name_prefix="pub",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    image_id="ami-ubuntu22",
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
    var publisher = new OutscalePublisher("publisher", new OutscalePublisherArgs
    {
        NamePrefix = "pub",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        ImageId = "ami-ubuntu22",
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewOutscalePublisher(ctx, "publisher", &netskopepublisher.OutscalePublisherArgs{
	NamePrefix: pulumi.String("pub"),
	Replicas: pulumi.Int(2),
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	BearerToken: netskope.RequireSecret("bearerToken"),
	ImageId: pulumi.String("ami-ubuntu22"),
})
_ = publisher
```

## Java

```java
var publisher = new OutscalePublisher("publisher", OutscalePublisherArgs.builder()
    .namePrefix("pub")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .imageId("ami-ubuntu22")
    .build());
```

## Rust

```rust
let publisher = netskope::outscale_publisher::create(
    ctx,
    "publisher",
    netskope::outscale_publisher::OutscalePublisherArgs::builder()
        .name_prefix("pub")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .image_id("ami-ubuntu22")
        .build_struct(),
);
```
