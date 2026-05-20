---
title: AWS Component
toc: true
---

# AWS Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`AwsPublisher` creates one EC2 instance per publisher name and registers
each instance with Netskope.

## Inputs

Required:

- `subnetId`
- `securityGroupIds`
- `tenantUrl` and `bearerToken`, unless `registrations` is provided

Common optional inputs:

- Naming: `namePrefix`, `names`, `replicas`
- EC2: `amiId`, `instanceType`, `keyName`,
  `associatePublicIpAddress`, `iamInstanceProfile`, `ebsOptimized`,
  `monitoring`, `metadataOptions`
- Bootstrap: `bootstrap`, `bootstrapUrl`, `nonat`, `installUser`,
  `installUserPassword`, `installUserPasswordIsHash`,
  `installUserSshAuthorizedKeys`, `deleteDefaultUser`,
  `guestNetworkInterface`
- Metadata: `tags`, `wizardPath`, `registrations`

When `bootstrap` is true and `amiId` is omitted, the component resolves a
Canonical Ubuntu 22.04 Minimal AMI and installs the publisher with
Netskope's generic bootstrap script. Leave `bootstrap` false or omitted
when using a pre-baked Publisher AMI.

## Outputs

- `publisherNames`
- secret `publishers`, keyed by publisher name

Each `publishers` entry includes publisher ID, registration token,
instance ID, private IP, and public IP when assigned.

## Pulumi CLI

```bash
pulumi new typescript
pulumi config set aws:region eu-west-1
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:bearerToken --secret
pulumi config set subnetId subnet-0123456789abcdef0
pulumi config set securityGroupId sg-0123456789abcdef0
pulumi config set keyName npa-admin
pulumi up
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { AwsPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const config = new pulumi.Config();

const publisher = new AwsPublisher("publisher", {
  namePrefix: "pub-eu",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  subnetId: config.require("subnetId"),
  securityGroupIds: [config.require("securityGroupId")],
  keyName: config.get("keyName"),
  instanceType: "t3.medium",
  associatePublicIpAddress: false,
  bootstrap: true,
  tags: {
    service: "npa",
    managedBy: "pulumi",
  },
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import AwsPublisher

netskope = pulumi.Config("netskope")
config = pulumi.Config()

publisher = AwsPublisher(
    "publisher",
    name_prefix="pub-eu",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    subnet_id=config.require("subnetId"),
    security_group_ids=[config.require("securityGroupId")],
    key_name=config.get("keyName"),
    instance_type="t3.medium",
    associate_public_ip_address=False,
    bootstrap=True,
    tags={
        "service": "npa",
        "managedBy": "pulumi",
    },
)

pulumi.export("publisherNames", publisher.publisher_names)
pulumi.export("publishers", pulumi.Output.secret(publisher.publishers))
```

## C#

```csharp
using System.Collections.Generic;
using Pulumi;
using JohninNL.Pulumi.NetskopePublisher;

return await Deployment.RunAsync(() =>
{
    var netskope = new Config("netskope");
    var config = new Config();

    var publisher = new AwsPublisher("publisher", new AwsPublisherArgs
    {
        NamePrefix = "pub-eu",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        SubnetId = config.Require("subnetId"),
        SecurityGroupIds = { config.Require("securityGroupId") },
        KeyName = config.Get("keyName"),
        InstanceType = "t3.medium",
        AssociatePublicIpAddress = false,
        Bootstrap = true,
        Tags =
        {
            { "service", "npa" },
            { "managedBy", "pulumi" },
        },
    });

    return new Dictionary<string, object?>
    {
        ["publisherNames"] = publisher.PublisherNames,
        ["publishers"] = Output.CreateSecret(publisher.Publishers),
    };
});
```

## Go

```go
package main

import (
	"github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		netskope := config.New(ctx, "netskope")
		cfg := config.New(ctx, "")

		publisher, err := netskopepublisher.NewAwsPublisher(ctx, "publisher", &netskopepublisher.AwsPublisherArgs{
			NamePrefix:               pulumi.String("pub-eu"),
			Replicas:                 pulumi.Int(2),
			TenantUrl:                pulumi.String(netskope.Require("tenantUrl")),
			BearerToken:                 netskope.RequireSecret("bearerToken"),
			SubnetId:                 pulumi.String(cfg.Require("subnetId")),
			SecurityGroupIds:         pulumi.StringArray{pulumi.String(cfg.Require("securityGroupId"))},
			KeyName:                  pulumi.StringPtr(cfg.Get("keyName")),
			InstanceType:             pulumi.String("t3.medium"),
			AssociatePublicIpAddress: pulumi.Bool(false),
			Bootstrap:                pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"service":   pulumi.String("npa"),
				"managedBy": pulumi.String("pulumi"),
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("publisherNames", publisher.PublisherNames)
		ctx.Export("publishers", pulumi.ToSecret(publisher.Publishers))
		return nil
	})
}
```

## Java

```java
package myproject;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.netskopepublisher.AwsPublisher;
import com.pulumi.netskopepublisher.AwsPublisherArgs;
import com.pulumi.Config;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var netskope = new Config("netskope");
            var config = new Config();

            var publisher = new AwsPublisher("publisher", AwsPublisherArgs.builder()
                .namePrefix("pub-eu")
                .replicas(2)
                .tenantUrl(netskope.require("tenantUrl"))
                .bearerToken(netskope.requireSecret("bearerToken"))
                .subnetId(config.require("subnetId"))
                .securityGroupIds(config.require("securityGroupId"))
                .keyName(config.get("keyName").orElse(null))
                .instanceType("t3.medium")
                .associatePublicIpAddress(false)
                .bootstrap(true)
                .build());

            ctx.export("publisherNames", publisher.publisherNames());
            ctx.export("publishers", Output.secret(publisher.publishers()));
        });
    }
}
```

## Rust

```rust
mod netskope {
    pulumi_gestalt_rust::include_provider!("netskope-publisher");
}

use anyhow::Result;
use pulumi_gestalt_rust::*;

fn main() {
    run(pulumi_main).unwrap();
}

fn pulumi_main(ctx: &Context) -> Result<()> {
    let publisher = netskope::aws_publisher::create(
        ctx,
        "publisher",
        netskope::aws_publisher::AwsPublisherArgs::builder()
            .name_prefix("pub-eu")
            .replicas(2)
            .tenant_url("https://tenant.goskope.com")
            .bearer_token("secret-token")
            .subnet_id("subnet-0123456789abcdef0")
            .security_group_ids(vec!["sg-0123456789abcdef0".to_string()])
            .instance_type("t3.medium")
            .associate_public_ip_address(false)
            .bootstrap(true)
            .build_struct(),
    );

    add_export("publisherNames", &publisher.publisher_names);
    add_export("publishers", &publisher.publishers);
    Ok(())
}
```
