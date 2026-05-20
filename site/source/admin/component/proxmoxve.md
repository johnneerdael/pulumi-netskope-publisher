---
title: Proxmox VE Component
toc: true
---

# Proxmox VE Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`ProxmoxvePublisher` creates one Proxmox VE VM per publisher name. It
uploads per-publisher cloud-init as a `snippets` file and clones an
existing Ubuntu 22.04 template VM with that snippet attached.

## Inputs

Required: `tenantUrl` and `bearerToken`, unless `registrations` is
provided. Platform required inputs are `nodeName`, `datastoreId`, and
`templateVmId`.

Optional platform inputs: `cloneNodeName`, `vmId`, `poolId`,
`cpuCores`, `memory`, `diskSize`, `networkBridge`, `networkModel`,
`vlanId`, `started`, `onBoot`, `fullClone`, `ipAddress`, `gateway`, and
`nameservers`.

## Image and bootstrap behavior

Prepare an Ubuntu 22.04 cloud-init template in Proxmox VE and enable
`snippets` on the datastore used by `datastoreId`. The component uploads
the Netskope bootstrap cloud-init as a snippet, sets
`initialization.userDataFileId`, and defaults networking to DHCP on
`vmbr0`.

## Outputs

`publisherNames` and secret `publishers`, keyed by publisher name.

## Pulumi YAML

```yaml
name: netskope-publisher-proxmoxve
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:ProxmoxvePublisher
    properties:
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      nodeName: pve-1
      vmIdStart: 4200
      templateVmId: 9000
      datastoreId: local-lvm
      snippetsDatastoreId: local
      networkBridge: vmbr0
      bootstrap: true
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { ProxmoxvePublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const publisher = new ProxmoxvePublisher("publisher", {
  namePrefix: "pub-pve",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  nodeName: "pve-1",
  datastoreId: "local",
  templateVmId: 9000,
  networkBridge: "vmbr0",
});
export const publishers = publisher.publishers;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import ProxmoxvePublisher

netskope = pulumi.Config("netskope")
publisher = ProxmoxvePublisher(
    "publisher",
    name_prefix="pub-pve",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    node_name="pve-1",
    datastore_id="local",
    template_vm_id=9000,
    network_bridge="vmbr0",
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
    var publisher = new ProxmoxvePublisher("publisher", new ProxmoxvePublisherArgs
    {
        NamePrefix = "pub-pve",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        BearerToken = netskope.RequireSecret("bearerToken"),
        NodeName = "pve-1",
        DatastoreId = "local",
        TemplateVmId = 9000,
        NetworkBridge = "vmbr0",
    });
});
```

## Go

```go
publisher, err := netskopepublisher.NewProxmoxvePublisher(ctx, "publisher", &netskopepublisher.ProxmoxvePublisherArgs{
	NamePrefix:    pulumi.String("pub-pve"),
	Replicas:      pulumi.Int(2),
	TenantUrl:     pulumi.String(netskope.Require("tenantUrl")),
	BearerToken:   netskope.RequireSecret("bearerToken"),
	NodeName:      "pve-1",
	DatastoreId:   "local",
	TemplateVmId:  9000,
	NetworkBridge: pulumi.String("vmbr0"),
})
```

## Java

```java
var publisher = new ProxmoxvePublisher("publisher", ProxmoxvePublisherArgs.builder()
    .namePrefix("pub-pve")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .nodeName("pve-1")
    .datastoreId("local")
    .templateVmId(9000)
    .networkBridge("vmbr0")
    .build());
```

## Rust

```rust
let publisher = netskope::proxmoxve_publisher::create(
    ctx,
    "publisher",
    netskope::proxmoxve_publisher::ProxmoxvePublisherArgs::builder()
        .name_prefix("pub-pve")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .node_name("pve-1")
        .datastore_id("local")
        .template_vm_id(9000)
        .network_bridge("vmbr0")
        .build_struct(),
);
```
