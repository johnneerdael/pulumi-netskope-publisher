---
title: vSphere Component
toc: true
---

# vSphere Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`VspherePublisher` clones one VM per publisher name from an existing
template and passes registration data through guestinfo cloud-init.

## Inputs

Required:

- `datacenter`
- `datastore`
- `networkName`
- `templateName`
- either `cluster` or `host`
- `tenantUrl` and `bearerToken`, unless `registrations` is provided

Optional inputs include `folder`, `numCpus`, `memory`, `tags`,
`namePrefix`, `names`, `replicas`, and `wizardPath`.

Prepare the template from the official Netskope OVA:

```text
https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.ova
```

The template must have VMware guestinfo cloud-init support enabled.

## Outputs

- `publisherNames`
- secret `publishers`

## Pulumi CLI

```bash
pulumi new typescript
pulumi config set vsphere:user administrator@vsphere.local
pulumi config set vsphere:password --secret
pulumi config set vsphere:vsphereServer vcsa.lab.local
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:bearerToken --secret
pulumi config set datacenter Lab
pulumi config set datastore datastore1
pulumi config set networkName VM Network
pulumi config set templateName npa-publisher-template
pulumi config set cluster Cluster
pulumi up
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { VspherePublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const config = new pulumi.Config();

const publisher = new VspherePublisher("publisher", {
  namePrefix: "pub-vsphere",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  datacenter: config.require("datacenter"),
  datastore: config.require("datastore"),
  networkName: config.require("networkName"),
  templateName: config.require("templateName"),
  cluster: config.require("cluster"),
  folder: config.get("folder"),
  numCpus: 2,
  memory: 4096,
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import VspherePublisher

netskope = pulumi.Config("netskope")
config = pulumi.Config()

publisher = VspherePublisher(
    "publisher",
    name_prefix="pub-vsphere",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    datacenter=config.require("datacenter"),
    datastore=config.require("datastore"),
    network_name=config.require("networkName"),
    template_name=config.require("templateName"),
    cluster=config.require("cluster"),
    folder=config.get("folder"),
    num_cpus=2,
    memory=4096,
)

pulumi.export("publisherNames", publisher.publisher_names)
pulumi.export("publishers", pulumi.Output.secret(publisher.publishers))
```

## C#

```csharp
var publisher = new VspherePublisher("publisher", new VspherePublisherArgs
{
    NamePrefix = "pub-vsphere",
    Replicas = 2,
    TenantUrl = netskope.Require("tenantUrl"),
    BearerToken = netskope.RequireSecret("bearerToken"),
    Datacenter = config.Require("datacenter"),
    Datastore = config.Require("datastore"),
    NetworkName = config.Require("networkName"),
    TemplateName = config.Require("templateName"),
    Cluster = config.Require("cluster"),
    Folder = config.Get("folder"),
    NumCpus = 2,
    Memory = 4096,
});
```

## Go

```go
publisher, err := netskopepublisher.NewVspherePublisher(ctx, "publisher", &netskopepublisher.VspherePublisherArgs{
	NamePrefix:   pulumi.String("pub-vsphere"),
	Replicas:     pulumi.Int(2),
	TenantUrl:    pulumi.String(netskope.Require("tenantUrl")),
	BearerToken:     netskope.RequireSecret("bearerToken"),
	Datacenter:   pulumi.String(cfg.Require("datacenter")),
	Datastore:    pulumi.String(cfg.Require("datastore")),
	NetworkName:  pulumi.String(cfg.Require("networkName")),
	TemplateName: pulumi.String(cfg.Require("templateName")),
	Cluster:      pulumi.String(cfg.Require("cluster")),
	Folder:       pulumi.StringPtr(cfg.Get("folder")),
	NumCpus:      pulumi.Int(2),
	Memory:       pulumi.Int(4096),
})
if err != nil {
	return err
}
ctx.Export("publisherNames", publisher.PublisherNames)
ctx.Export("publishers", pulumi.ToSecret(publisher.Publishers))
```

## Java

```java
var publisher = new VspherePublisher("publisher", VspherePublisherArgs.builder()
    .namePrefix("pub-vsphere")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .datacenter(config.require("datacenter"))
    .datastore(config.require("datastore"))
    .networkName(config.require("networkName"))
    .templateName(config.require("templateName"))
    .cluster(config.require("cluster"))
    .folder(config.get("folder").orElse(null))
    .numCpus(2)
    .memory(4096)
    .build());

ctx.export("publisherNames", publisher.publisherNames());
ctx.export("publishers", Output.secret(publisher.publishers()));
```

## Rust

```rust
let publisher = netskope::vsphere_publisher::create(
    ctx,
    "publisher",
    netskope::vsphere_publisher::VspherePublisherArgs::builder()
        .name_prefix("pub-vsphere")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .datacenter("Lab")
        .datastore("datastore1")
        .network_name("VM Network")
        .template_name("npa-publisher-template")
        .cluster("Cluster")
        .num_cpus(2)
        .memory(4096)
        .build_struct(),
);

add_export("publisherNames", &publisher.publisher_names);
add_export("publishers", &publisher.publishers);
```
