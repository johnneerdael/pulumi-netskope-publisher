---
title: Netskope Publisher
meta_desc: Pulumi components for provisioning Netskope Private Access Publishers.
layout: package
---

The Netskope Publisher package provides Pulumi component resources for
provisioning Netskope Private Access Publishers on AWS, Azure, Google
Cloud, and vSphere. The package mirrors the Terraform module pattern:
register or reuse Netskope publisher records, generate per-publisher
cloud-init, and create the virtual machines that run the publisher
appliance.

## Components

- `AwsPublisher` creates EC2-backed publishers.
- `AzurePublisher` creates Azure virtual machine backed publishers.
- `GcpPublisher` creates Google Compute Engine backed publishers.
- `VspherePublisher` creates vSphere virtual machine backed publishers.
- `HypervPublisher` is experimental and remains opt-in.

Each component accepts either Netskope tenant credentials for automatic
publisher registration or pre-created registration tokens keyed by
publisher name.

## TypeScript example

```typescript
import * as pulumi from "@pulumi/pulumi";
import { AwsPublisher } from "@johninnl/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new AwsPublisher("publisher", {
  namePrefix: "pub-eu",
  replicas: 2,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  subnetId: config.require("subnetId"),
  securityGroupIds: config.requireObject<string[]>("securityGroupIds"),
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

## Python example

```python
import pulumi
from pulumi_netskope_publisher import AwsPublisher

config = pulumi.Config()

publisher = AwsPublisher(
    "publisher",
    name_prefix="pub-eu",
    replicas=2,
    tenant_url=config.require("tenantUrl"),
    api_token=config.require_secret("apiToken"),
    subnet_id=config.require("subnetId"),
    security_group_ids=config.require_object("securityGroupIds"),
)

pulumi.export("publisherNames", publisher.publisher_names)
pulumi.export("publishers", publisher.publishers)
```

## C# example

```csharp
using System.Collections.Generic;
using Pulumi;
using Pulumi.NetskopePublisher;

return await Deployment.RunAsync(() =>
{
    var config = new Config();

    var publisher = new AwsPublisher("publisher", new AwsPublisherArgs
    {
        NamePrefix = "pub-eu",
        Replicas = 2,
        TenantUrl = config.Require("tenantUrl"),
        ApiToken = config.RequireSecret("apiToken"),
        SubnetId = config.Require("subnetId"),
        SecurityGroupIds = config.RequireObject<string[]>("securityGroupIds"),
    });

    return new Dictionary<string, object?>
    {
        ["publisherNames"] = publisher.PublisherNames,
        ["publishers"] = publisher.Publishers,
    };
});
```

Provider-specific examples are available in the repository under
`examples/`.
