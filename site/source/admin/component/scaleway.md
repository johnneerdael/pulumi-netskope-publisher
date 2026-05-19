---
title: Scaleway Component
toc: true
---

# Scaleway Component

`ScalewayPublisher` creates one Scaleway Instance per publisher name.

## Inputs

Required: `tenantUrl` and `apiToken`, unless `registrations` is provided.

Optional platform inputs: `type`, `image`, `zone`, `securityGroupId`, and
`enableDynamicIp`.

## Image and bootstrap behavior

The default image is `ubuntu_jammy`. Scaleway uses bootstrap mode and
passes cloud-init through the instance `cloudInit` and `userData` fields.

## Outputs

`publisherNames` and secret `publishers`, keyed by publisher name.

## Pulumi CLI

```bash
pulumi config set scaleway:access_key --secret
pulumi config set scaleway:secret_key --secret
pulumi config set scaleway:project_id <project-id>
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:apiToken --secret
pulumi up
```

## TypeScript

```ts
const publisher = new ScalewayPublisher("publisher", {
  namePrefix: "pub-fr",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  apiToken: netskope.requireSecret("apiToken"),
  zone: "fr-par-1",
  type: "DEV1-M",
});
```

## Python

```python
publisher = ScalewayPublisher(
    "publisher",
    name_prefix="pub-fr",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    api_token=netskope.require_secret("apiToken"),
    zone="fr-par-1",
    type="DEV1-M",
)
```

## C#

```csharp
var publisher = new ScalewayPublisher("publisher", new ScalewayPublisherArgs
{
    NamePrefix = "pub-fr",
    Replicas = 2,
    TenantUrl = netskope.Require("tenantUrl"),
    ApiToken = netskope.RequireSecret("apiToken"),
    Zone = "fr-par-1",
    Type = "DEV1-M",
});
```

## Go

```go
publisher, err := netskopepublisher.NewScalewayPublisher(ctx, "publisher", &netskopepublisher.ScalewayPublisherArgs{
	NamePrefix: pulumi.String("pub-fr"),
	Replicas:   pulumi.Int(2),
	TenantUrl:  pulumi.String(netskope.Require("tenantUrl")),
	ApiToken:   netskope.RequireSecret("apiToken"),
	Zone:       pulumi.String("fr-par-1"),
	Type:       pulumi.String("DEV1-M"),
})
```

## Java

```java
var publisher = new ScalewayPublisher("publisher", ScalewayPublisherArgs.builder()
    .namePrefix("pub-fr")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .apiToken(netskope.requireSecret("apiToken"))
    .zone("fr-par-1")
    .type("DEV1-M")
    .build());
```

## Rust

```rust
let publisher = netskope::scaleway_publisher::create(
    ctx,
    "publisher",
    netskope::scaleway_publisher::ScalewayPublisherArgs::builder()
        .name_prefix("pub-fr")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .api_token("secret-token")
        .zone("fr-par-1")
        .type_("DEV1-M")
        .build_struct(),
);
```
