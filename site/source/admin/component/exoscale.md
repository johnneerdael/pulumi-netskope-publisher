---
title: Exoscale Component
toc: true
---

# Exoscale Component

`ExoscalePublisher` creates one Ubuntu 22.04 bootstrap-mode Netskope Publisher per resolved publisher name.

Required platform inputs: `zone`, `type`, `templateId`, and `diskSize`. Use an Ubuntu 22.04 template ID.

## Inputs

Required Netskope inputs: `tenantUrl` and `bearerToken`, unless `registrations` is provided. OAuth2 enrollment is available with `authMode: "oauth2"` and `oauth2` client credentials.

## Bootstrap behavior

The component renders the shared Netskope Publisher cloud-init payload, installs the generic publisher software, and enrolls the VM with a deployment-time registration token.

## Pulumi YAML

```yaml
name: netskope-publisher-exoscale
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:ExoscalePublisher
    properties:
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      zone: ch-gva-2
      templateId: <ubuntu-2204-template-id>
      instanceType: standard.medium
      bootstrap: true
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { ExoscalePublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new ExoscalePublisher("publisher", {
  namePrefix: "pub",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  zone: "ch-gva-2",
  type: "standard.medium",
  templateId: "ubuntu-22-template",
  diskSize: 50,
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import ExoscalePublisher

netskope = pulumi.Config("netskope")
publisher = ExoscalePublisher(
    "publisher",
    name_prefix="pub",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    zone="ch-gva-2",
    type="standard.medium",
    template_id="ubuntu-22-template",
    disk_size=50,
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
    var publisher = new ExoscalePublisher("publisher", new ExoscalePublisherArgs
    {
        NamePrefix = "pub",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        Zone = "ch-gva-2",
        Type = "standard.medium",
        TemplateId = "ubuntu-22-template",
        DiskSize = 50,
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewExoscalePublisher(ctx, "publisher", &netskopepublisher.ExoscalePublisherArgs{
	NamePrefix: pulumi.String("pub"),
	Replicas: pulumi.Int(2),
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	BearerToken: netskope.RequireSecret("bearerToken"),
	Zone: pulumi.String("ch-gva-2"),
	Type: pulumi.String("standard.medium"),
	TemplateId: pulumi.String("ubuntu-22-template"),
	DiskSize: pulumi.Int(50),
})
_ = publisher
```

## Java

```java
var publisher = new ExoscalePublisher("publisher", ExoscalePublisherArgs.builder()
    .namePrefix("pub")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .zone("ch-gva-2")
    .type("standard.medium")
    .templateId("ubuntu-22-template")
    .diskSize(50)
    .build());
```

## Rust

```rust
let publisher = netskope::exoscale_publisher::create(
    ctx,
    "publisher",
    netskope::exoscale_publisher::ExoscalePublisherArgs::builder()
        .name_prefix("pub")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .zone("ch-gva-2")
        .type_("standard.medium")
        .template_id("ubuntu-22-template")
        .disk_size(50)
        .build_struct(),
);
```
