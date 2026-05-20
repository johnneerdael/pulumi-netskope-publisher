---
title: Stackit Component
toc: true
---

# Stackit Component

`StackitPublisher` creates one Ubuntu 22.04 bootstrap-mode Netskope Publisher per resolved publisher name.

Required platform inputs: `projectId`, `machineType`, and `imageId`. Use an Ubuntu 22.04 image ID.

## Inputs

Required Netskope inputs: `tenantUrl` and `bearerToken`, unless `registrations` is provided. OAuth2 enrollment is available with `authMode: "oauth2"` and `oauth2` client credentials.

## Bootstrap behavior

The component renders the shared Netskope Publisher cloud-init payload, installs the generic publisher software, and enrolls the VM with a deployment-time registration token.

## Pulumi YAML

```yaml
name: netskope-publisher-stackit
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:StackitPublisher
    properties:
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      availabilityZone: eu01-1
      imageId: <ubuntu-2204-image-id>
      machineType: g1.2
      bootstrap: true
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { StackitPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new StackitPublisher("publisher", {
  namePrefix: "pub",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  projectId: "project-id",
  machineType: "g1.2",
  imageId: "ubuntu-22-image",
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import StackitPublisher

netskope = pulumi.Config("netskope")
publisher = StackitPublisher(
    "publisher",
    name_prefix="pub",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    project_id="project-id",
    machine_type="g1.2",
    image_id="ubuntu-22-image",
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
    var publisher = new StackitPublisher("publisher", new StackitPublisherArgs
    {
        NamePrefix = "pub",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        ProjectId = "project-id",
        MachineType = "g1.2",
        ImageId = "ubuntu-22-image",
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewStackitPublisher(ctx, "publisher", &netskopepublisher.StackitPublisherArgs{
	NamePrefix: pulumi.String("pub"),
	Replicas: pulumi.Int(2),
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	BearerToken: netskope.RequireSecret("bearerToken"),
	ProjectId: pulumi.String("project-id"),
	MachineType: pulumi.String("g1.2"),
	ImageId: pulumi.String("ubuntu-22-image"),
})
_ = publisher
```

## Java

```java
var publisher = new StackitPublisher("publisher", StackitPublisherArgs.builder()
    .namePrefix("pub")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .projectId("project-id")
    .machineType("g1.2")
    .imageId("ubuntu-22-image")
    .build());
```

## Rust

```rust
let publisher = netskope::stackit_publisher::create(
    ctx,
    "publisher",
    netskope::stackit_publisher::StackitPublisherArgs::builder()
        .name_prefix("pub")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .project_id("project-id")
        .machine_type("g1.2")
        .image_id("ubuntu-22-image")
        .build_struct(),
);
```
