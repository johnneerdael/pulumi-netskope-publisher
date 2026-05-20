---
title: OpenStack Component
toc: true
---

# OpenStack Component

`OpenstackPublisher` creates one OpenStack compute instance per
publisher name.

## Inputs

Required: `imageName`, `flavorName`, `networkName`, and `tenantUrl` plus
`bearerToken` unless `registrations` is provided.

Optional platform inputs: `keyPair`, `securityGroups`,
`availabilityZone`, `assignFloatingIp`, and `floatingIpPool`.

## Image and bootstrap behavior

Use an Ubuntu 22.04 image name. The component uses bootstrap mode and
passes decoded cloud-init to the compute instance `userData` field.

## Outputs

`publisherNames` and secret `publishers`, keyed by publisher name.

## Pulumi CLI

```bash
pulumi config set openstack:authUrl https://openstack.example.com:5000/v3
pulumi config set openstack:userName admin
pulumi config set openstack:password --secret
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:bearerToken --secret
pulumi up
```

## TypeScript

```ts
const publisher = new OpenstackPublisher("publisher", {
  namePrefix: "pub-os",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  imageName: "Ubuntu 22.04",
  flavorName: "m1.medium",
  networkName: "private",
  securityGroups: ["default"],
});
```

## Python

```python
publisher = OpenstackPublisher(
    "publisher",
    name_prefix="pub-os",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    image_name="Ubuntu 22.04",
    flavor_name="m1.medium",
    network_name="private",
    security_groups=["default"],
)
```

## C#

```csharp
var publisher = new OpenstackPublisher("publisher", new OpenstackPublisherArgs
{
    NamePrefix = "pub-os",
    Replicas = 2,
    TenantUrl = netskope.Require("tenantUrl"),
    BearerToken = netskope.RequireSecret("bearerToken"),
    ImageName = "Ubuntu 22.04",
    FlavorName = "m1.medium",
    NetworkName = "private",
    SecurityGroups = { "default" },
});
```

## Go

```go
publisher, err := netskopepublisher.NewOpenstackPublisher(ctx, "publisher", &netskopepublisher.OpenstackPublisherArgs{
	NamePrefix:     pulumi.String("pub-os"),
	Replicas:       pulumi.Int(2),
	TenantUrl:      pulumi.String(netskope.Require("tenantUrl")),
	BearerToken:       netskope.RequireSecret("bearerToken"),
	ImageName:      pulumi.String("Ubuntu 22.04"),
	FlavorName:     pulumi.String("m1.medium"),
	NetworkName:    pulumi.String("private"),
	SecurityGroups: pulumi.StringArray{pulumi.String("default")},
})
```

## Java

```java
var publisher = new OpenstackPublisher("publisher", OpenstackPublisherArgs.builder()
    .namePrefix("pub-os")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .imageName("Ubuntu 22.04")
    .flavorName("m1.medium")
    .networkName("private")
    .securityGroups("default")
    .build());
```

## Rust

```rust
let publisher = netskope::openstack_publisher::create(
    ctx,
    "publisher",
    netskope::openstack_publisher::OpenstackPublisherArgs::builder()
        .name_prefix("pub-os")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .image_name("Ubuntu 22.04")
        .flavor_name("m1.medium")
        .network_name("private")
        .security_groups(vec!["default".to_string()])
        .build_struct(),
);
```
