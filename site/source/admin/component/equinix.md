---
title: Equinix Metal Component
toc: true
---

# Equinix Metal Component

`EquinixPublisher` creates one Ubuntu 22.04 bootstrap-mode Netskope Publisher per resolved publisher name.

Required platform inputs: `projectId`, `metro`, and `plan`. The default operating system is `ubuntu_22_04`.

## Inputs

Required Netskope inputs: `tenantUrl` and `bearerToken`, unless `registrations` is provided. OAuth2 enrollment is available with `authMode: "oauth2"` and `oauth2` client credentials.

## Bootstrap behavior

The component renders the shared Netskope Publisher cloud-init payload, installs the generic publisher software, and enrolls the VM with a deployment-time registration token.

## Pulumi YAML

```yaml
name: netskope-publisher-equinix
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:EquinixPublisher
    properties:
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      projectId: <project-id>
      metro: am
      plan: c3.small.x86
      operatingSystem: ubuntu_22_04
      bootstrap: true
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { EquinixPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new EquinixPublisher("publisher", {
  namePrefix: "pub",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  projectId: "project-id",
  metro: "AM",
  plan: "c3.small.x86",
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import EquinixPublisher

netskope = pulumi.Config("netskope")
publisher = EquinixPublisher(
    "publisher",
    name_prefix="pub",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    project_id="project-id",
    metro="AM",
    plan="c3.small.x86",
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
    var publisher = new EquinixPublisher("publisher", new EquinixPublisherArgs
    {
        NamePrefix = "pub",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        ProjectId = "project-id",
        Metro = "AM",
        Plan = "c3.small.x86",
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewEquinixPublisher(ctx, "publisher", &netskopepublisher.EquinixPublisherArgs{
	NamePrefix: pulumi.String("pub"),
	Replicas: pulumi.Int(2),
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	BearerToken: netskope.RequireSecret("bearerToken"),
	ProjectId: pulumi.String("project-id"),
	Metro: pulumi.String("AM"),
	Plan: pulumi.String("c3.small.x86"),
})
_ = publisher
```

## Java

```java
var publisher = new EquinixPublisher("publisher", EquinixPublisherArgs.builder()
    .namePrefix("pub")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .projectId("project-id")
    .metro("AM")
    .plan("c3.small.x86")
    .build());
```

## Rust

```rust
let publisher = netskope::equinix_publisher::create(
    ctx,
    "publisher",
    netskope::equinix_publisher::EquinixPublisherArgs::builder()
        .name_prefix("pub")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .project_id("project-id")
        .metro("AM")
        .plan("c3.small.x86")
        .build_struct(),
);
```
