---
title: ESXi Component
toc: true
---

# ESXi Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`EsxiPublisher` creates one ESXi Native virtual machine per publisher
name. It is direct-host ESXi support and does not replace
`VspherePublisher`.

## Inputs

Required: `diskStore`, `virtualNetwork`, and `tenantUrl` plus `bearerToken`
unless `registrations` is provided.

Optional platform inputs: `os`, `memory`, `numVCpus`, and `diskSize`.

## Image and bootstrap behavior

The component passes cloud-init through ESXi guestinfo keys. Use an
Ubuntu 22.04 cloud-init capable template or OVF source prepared for the
ESXi Native provider.

## Outputs

`publisherNames` and secret `publishers`, keyed by publisher name.

## Pulumi YAML

```yaml
name: netskope-publisher-esxi
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:EsxiPublisher
    properties:
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      datastore: datastore1
      networkName: VM Network
      ovfSource: /images/ubuntu-22.04.ova
      diskStore: datastore1
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```

## TypeScript

```ts
const publisher = new EsxiPublisher("publisher", {
  names: ["pub-esxi-1", "pub-esxi-2"],
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  diskStore: "datastore1",
  virtualNetwork: "VM Network",
  memory: 4096,
  numVCpus: 2,
});
```

## Python

```python
publisher = EsxiPublisher(
    "publisher",
    names=["pub-esxi-1", "pub-esxi-2"],
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    disk_store="datastore1",
    virtual_network="VM Network",
    memory=4096,
    num_v_cpus=2,
)
```

## C#

```csharp
var publisher = new EsxiPublisher("publisher", new EsxiPublisherArgs
{
    Names = { "pub-esxi-1", "pub-esxi-2" },
    TenantUrl = netskope.Require("tenantUrl"),
    BearerToken = netskope.RequireSecret("bearerToken"),
    DiskStore = "datastore1",
    VirtualNetwork = "VM Network",
    Memory = 4096,
    NumVCpus = 2,
});
```

## Go

```go
publisher, err := netskopepublisher.NewEsxiPublisher(ctx, "publisher", &netskopepublisher.EsxiPublisherArgs{
	Names:          pulumi.StringArray{pulumi.String("pub-esxi-1"), pulumi.String("pub-esxi-2")},
	TenantUrl:      pulumi.String(netskope.Require("tenantUrl")),
	BearerToken:       netskope.RequireSecret("bearerToken"),
	DiskStore:      pulumi.String("datastore1"),
	VirtualNetwork: pulumi.String("VM Network"),
	Memory:         pulumi.Int(4096),
	NumVCpus:       pulumi.Int(2),
})
```

## Java

```java
var publisher = new EsxiPublisher("publisher", EsxiPublisherArgs.builder()
    .names("pub-esxi-1", "pub-esxi-2")
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .diskStore("datastore1")
    .virtualNetwork("VM Network")
    .memory(4096)
    .numVCpus(2)
    .build());
```

## Rust

```rust
let publisher = netskope::esxi_publisher::create(
    ctx,
    "publisher",
    netskope::esxi_publisher::EsxiPublisherArgs::builder()
        .names(vec!["pub-esxi-1".to_string(), "pub-esxi-2".to_string()])
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .disk_store("datastore1")
        .virtual_network("VM Network")
        .memory(4096)
        .num_v_cpus(2)
        .build_struct(),
);
```
