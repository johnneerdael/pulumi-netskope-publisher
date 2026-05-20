---
title: OCI Component
toc: true
---

# OCI Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`OciPublisher` creates one Oracle Cloud Infrastructure compute instance
per publisher name.

## Inputs

Required: `compartmentId`, `availabilityDomain`, `subnetId`, `imageId`,
and `tenantUrl` plus `bearerToken` unless `registrations` is provided.

Optional platform inputs: `shape`, `sshPublicKey`, and `assignPublicIp`.

## Image and bootstrap behavior

OCI requires an Ubuntu 22.04 image OCID through `imageId`. The component
uses bootstrap mode and passes base64 cloud-init through instance
metadata.

## Outputs

`publisherNames` and secret `publishers`, keyed by publisher name.

## Pulumi YAML

```yaml
name: netskope-publisher-oci
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:OciPublisher
    properties:
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      compartmentId: ocid1.compartment.oc1..example
      availabilityDomain: '<availability-domain>'
      subnetId: ocid1.subnet.oc1..example
      imageId: ocid1.image.oc1..ubuntu2204
      shape: VM.Standard.E4.Flex
      bootstrap: true
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```

## TypeScript

```ts
const publisher = new OciPublisher("publisher", {
  names: ["pub-fra-1", "pub-fra-2"],
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  compartmentId: config.require("compartmentId"),
  availabilityDomain: config.require("availabilityDomain"),
  subnetId: config.require("subnetId"),
  imageId: config.require("imageId"),
  shape: "VM.Standard.E4.Flex",
});
```

## Python

```python
publisher = OciPublisher(
    "publisher",
    names=["pub-fra-1", "pub-fra-2"],
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    compartment_id=config.require("compartmentId"),
    availability_domain=config.require("availabilityDomain"),
    subnet_id=config.require("subnetId"),
    image_id=config.require("imageId"),
    shape="VM.Standard.E4.Flex",
)
```

## C#

```csharp
var publisher = new OciPublisher("publisher", new OciPublisherArgs
{
    Names = { "pub-fra-1", "pub-fra-2" },
    TenantUrl = netskope.Require("tenantUrl"),
    BearerToken = netskope.RequireSecret("bearerToken"),
    CompartmentId = config.Require("compartmentId"),
    AvailabilityDomain = config.Require("availabilityDomain"),
    SubnetId = config.Require("subnetId"),
    ImageId = config.Require("imageId"),
    Shape = "VM.Standard.E4.Flex",
});
```

## Go

```go
publisher, err := netskopepublisher.NewOciPublisher(ctx, "publisher", &netskopepublisher.OciPublisherArgs{
	Names:              pulumi.StringArray{pulumi.String("pub-fra-1"), pulumi.String("pub-fra-2")},
	TenantUrl:          pulumi.String(netskope.Require("tenantUrl")),
	BearerToken:           netskope.RequireSecret("bearerToken"),
	CompartmentId:      pulumi.String(cfg.Require("compartmentId")),
	AvailabilityDomain: pulumi.String(cfg.Require("availabilityDomain")),
	SubnetId:           pulumi.String(cfg.Require("subnetId")),
	ImageId:             pulumi.String(cfg.Require("imageId")),
	Shape:               pulumi.String("VM.Standard.E4.Flex"),
})
```

## Java

```java
var publisher = new OciPublisher("publisher", OciPublisherArgs.builder()
    .names("pub-fra-1", "pub-fra-2")
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .compartmentId(config.require("compartmentId"))
    .availabilityDomain(config.require("availabilityDomain"))
    .subnetId(config.require("subnetId"))
    .imageId(config.require("imageId"))
    .shape("VM.Standard.E4.Flex")
    .build());
```

## Rust

```rust
let publisher = netskope::oci_publisher::create(
    ctx,
    "publisher",
    netskope::oci_publisher::OciPublisherArgs::builder()
        .names(vec!["pub-fra-1".to_string(), "pub-fra-2".to_string()])
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .compartment_id("ocid1.compartment.oc1..example")
        .availability_domain("AD-1")
        .subnet_id("ocid1.subnet.oc1..example")
        .image_id("ocid1.image.oc1..ubuntu")
        .shape("VM.Standard.E4.Flex")
        .build_struct(),
);
```
