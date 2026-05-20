---
title: OpenTelekomCloud Component
toc: true
---

# OpenTelekomCloud Component

`OpentelekomcloudPublisher` creates one Ubuntu 22.04 bootstrap-mode Netskope Publisher per resolved publisher name.

Required platform input: `networks`. Defaults use `imageName: "Ubuntu 22.04"` and `flavorName: "s3.medium.2"` unless overridden.

## Inputs

Required Netskope inputs: `tenantUrl` and `bearerToken`, unless `registrations` is provided. OAuth2 enrollment is available with `authMode: "oauth2"` and `oauth2` client credentials.

## Bootstrap behavior

The component renders the shared Netskope Publisher cloud-init payload, installs the generic publisher software, and enrolls the VM with a deployment-time registration token.

## Pulumi YAML

```yaml
name: netskope-publisher-opentelekomcloud
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:OpentelekomcloudPublisher
    properties:
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      availabilityZone: eu-de-01
      imageName: Ubuntu 22.04
      flavorName: s3.large.2
      networkId: <network-id>
      bootstrap: true
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { OpentelekomcloudPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new OpentelekomcloudPublisher("publisher", {
  namePrefix: "pub",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  networks: [{ name: "private" }],
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import OpentelekomcloudPublisher

netskope = pulumi.Config("netskope")
publisher = OpentelekomcloudPublisher(
    "publisher",
    name_prefix="pub",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    networks=[{"name": "private"}],
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
    var publisher = new OpentelekomcloudPublisher("publisher", new OpentelekomcloudPublisherArgs
    {
        NamePrefix = "pub",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        Networks = new[] { new Dictionary<string, object?> { ["name"] = "private" } },
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewOpentelekomcloudPublisher(ctx, "publisher", &netskopepublisher.OpentelekomcloudPublisherArgs{
	NamePrefix: pulumi.String("pub"),
	Replicas: pulumi.Int(2),
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	BearerToken: netskope.RequireSecret("bearerToken"),
	Networks: []map[string]interface{}{{"name": "private"}},
})
_ = publisher
```

## Java

```java
var publisher = new OpentelekomcloudPublisher("publisher", OpentelekomcloudPublisherArgs.builder()
    .namePrefix("pub")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .networks(Map.of("name", "private"))
    .build());
```

## Rust

```rust
let publisher = netskope::opentelekomcloud_publisher::create(
    ctx,
    "publisher",
    netskope::opentelekomcloud_publisher::OpentelekomcloudPublisherArgs::builder()
        .name_prefix("pub")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .networks(vec![pulumi_gestalt_rust::types::to_value(maplit::hashmap!{"name" => "private"})])
        .build_struct(),
);
```
