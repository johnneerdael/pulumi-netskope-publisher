---
title: ESXi Component
toc: true
---

# ESXi Component

`EsxiPublisher` creates one ESXi Native virtual machine per publisher
name. It is direct-host ESXi support and does not replace
`VspherePublisher`.

## Inputs

Required: `diskStore`, `virtualNetwork`, and `tenantUrl` plus `apiToken`
unless `registrations` is provided.

Optional platform inputs: `os`, `memory`, `numVCpus`, and `diskSize`.

## Image and bootstrap behavior

The component passes cloud-init through ESXi guestinfo keys. Use an
Ubuntu 22.04 cloud-init capable template or OVF source prepared for the
ESXi Native provider.

## Outputs

`publisherNames` and secret `publishers`, keyed by publisher name.

## Pulumi CLI

```bash
pulumi config set esxi-native:host https://esxi.example.com/sdk
pulumi config set esxi-native:user root
pulumi config set esxi-native:password --secret
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:apiToken --secret
pulumi up
```

## TypeScript

```ts
const publisher = new EsxiPublisher("publisher", {
  names: ["pub-esxi-1", "pub-esxi-2"],
  tenantUrl: netskope.require("tenantUrl"),
  apiToken: netskope.requireSecret("apiToken"),
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
    api_token=netskope.require_secret("apiToken"),
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
    ApiToken = netskope.RequireSecret("apiToken"),
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
	ApiToken:       netskope.RequireSecret("apiToken"),
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
    .apiToken(netskope.requireSecret("apiToken"))
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
        .api_token("secret-token")
        .disk_store("datastore1")
        .virtual_network("VM Network")
        .memory(4096)
        .num_v_cpus(2)
        .build_struct(),
);
```
