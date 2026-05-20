---
title: Azure Component
toc: true
---

# Azure Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`AzurePublisher` creates one Linux virtual machine per publisher name.

## Inputs

Required:

- `resourceGroupName`
- `location`
- `subnetId`
- `adminSshPublicKey`
- one of `imageId`, `marketplace`, or `bootstrap: true`
- `tenantUrl` and `bearerToken`, unless `registrations` is provided

Common optional inputs:

- `vmSize`, `adminUsername`, `networkSecurityGroupId`,
  `assignPublicIp`, `osDisk`, `acceptMarketplaceTerms`
- `namePrefix`, `names`, `replicas`, `tags`
- `bootstrap`, `bootstrapUrl`, `nonat`, `installUser`,
  `installUserPassword`, `installUserPasswordIsHash`,
  `installUserSshAuthorizedKeys`, `deleteDefaultUser`,
  `guestNetworkInterface`

When `bootstrap` is true and no image is supplied, the component uses
Canonical Ubuntu 22.04 Minimal and installs the publisher at first boot.
`adminUsername` defaults to `installUser`, keeping Azure SSH and
cloud-init ownership aligned.

## Outputs

- `publisherNames`
- secret `publishers`

## Pulumi YAML

```yaml
name: netskope-publisher-azure
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:AzurePublisher
    properties:
      namePrefix: pub-eu
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      resourceGroupName: rg-npa
      location: westeurope
      subnetId: /subscriptions/.../subnets/npa
      adminSshPublicKey: ssh-rsa AAAA...
      vmSize: Standard_D2s_v5
      assignPublicIp: false
      bootstrap: true
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { AzurePublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const config = new pulumi.Config();

const publisher = new AzurePublisher("publisher", {
  namePrefix: "pub-az",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  resourceGroupName: config.require("resourceGroupName"),
  location: config.require("location"),
  subnetId: config.require("subnetId"),
  adminSshPublicKey: config.require("adminSshPublicKey"),
  vmSize: "Standard_D2s_v5",
  assignPublicIp: false,
  bootstrap: true,
  tags: { service: "npa" },
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import AzurePublisher

netskope = pulumi.Config("netskope")
config = pulumi.Config()

publisher = AzurePublisher(
    "publisher",
    name_prefix="pub-az",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    resource_group_name=config.require("resourceGroupName"),
    location=config.require("location"),
    subnet_id=config.require("subnetId"),
    admin_ssh_public_key=config.require("adminSshPublicKey"),
    vm_size="Standard_D2s_v5",
    assign_public_ip=False,
    bootstrap=True,
    tags={"service": "npa"},
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

    var publisher = new AzurePublisher("publisher", new AzurePublisherArgs
    {
        NamePrefix = "pub-az",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        ResourceGroupName = config.Require("resourceGroupName"),
        Location = config.Require("location"),
        SubnetId = config.Require("subnetId"),
        AdminSshPublicKey = config.Require("adminSshPublicKey"),
        VmSize = "Standard_D2s_v5",
        AssignPublicIp = false,
        Bootstrap = true,
        Tags = { { "service", "npa" } },
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

		publisher, err := netskopepublisher.NewAzurePublisher(ctx, "publisher", &netskopepublisher.AzurePublisherArgs{
			NamePrefix:        pulumi.String("pub-az"),
			Replicas:          pulumi.Int(2),
			TenantUrl:         pulumi.String(netskope.Require("tenantUrl")),
			BearerToken:          netskope.RequireSecret("bearerToken"),
			ResourceGroupName: pulumi.String(cfg.Require("resourceGroupName")),
			Location:          pulumi.String(cfg.Require("location")),
			SubnetId:          pulumi.String(cfg.Require("subnetId")),
			AdminSshPublicKey: pulumi.String(cfg.Require("adminSshPublicKey")),
			VmSize:            pulumi.String("Standard_D2s_v5"),
			AssignPublicIp:    pulumi.Bool(false),
			Bootstrap:         pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"service": pulumi.String("npa"),
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
var publisher = new AzurePublisher("publisher", AzurePublisherArgs.builder()
    .namePrefix("pub-az")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .resourceGroupName(config.require("resourceGroupName"))
    .location(config.require("location"))
    .subnetId(config.require("subnetId"))
    .adminSshPublicKey(config.require("adminSshPublicKey"))
    .vmSize("Standard_D2s_v5")
    .assignPublicIp(false)
    .bootstrap(true)
    .build());

ctx.export("publisherNames", publisher.publisherNames());
ctx.export("publishers", Output.secret(publisher.publishers()));
```

## Rust

```rust
let publisher = netskope::azure_publisher::create(
    ctx,
    "publisher",
    netskope::azure_publisher::AzurePublisherArgs::builder()
        .name_prefix("pub-az")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .resource_group_name("rg-npa")
        .location("westeurope")
        .subnet_id("/subscriptions/.../subnets/npa")
        .admin_ssh_public_key("ssh-rsa AAAA...")
        .vm_size("Standard_D2s_v5")
        .assign_public_ip(false)
        .bootstrap(true)
        .build_struct(),
);

add_export("publisherNames", &publisher.publisher_names);
add_export("publishers", &publisher.publishers);
```
