---
title: Kubernetes Component
toc: true
---

# Kubernetes Component

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`KubernetesPublisher` installs the `kubernetes-netskope-publisher` Helm
chart into an existing Kubernetes cluster.

## Inputs

Required:

- `tenantUrl` and `bearerToken`, unless token-mode `registrations` are provided

Common optional inputs include `namespace`, `enrollmentMode`,
`chartRepository`, `chartVersion`, `chartValues`, `workloadType`,
`hpaEnabled`, `hpaMinReplicas`, `hpaMaxReplicas`, `imageRepository`,
`imageTag`, `tags`, `namePrefix`, `names`, and `replicas`.

## Enrollment modes

`token` mode is the default. Pulumi creates or reuses Netskope publisher
records, creates one Kubernetes Secret per registration token, and
installs one Helm release per publisher name.

`api` mode creates one `npa-api-token` Secret and one Helm release named
`npa-publisher`. The chart registers publisher pods with the Netskope API
during startup.

## Outputs

- `publisherNames`
- `helmReleaseNames`
- secret `publishers`

## Pulumi YAML

```yaml
name: netskope-publisher-kubernetes
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:KubernetesPublisher
    properties:
      namePrefix: pub
      replicas: 2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namespace: npa
      authMode: token
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
  helmReleaseNames: ${publisher.helmReleaseNames}
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { KubernetesPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const config = new pulumi.Config();

const publisher = new KubernetesPublisher("publisher", {
  namePrefix: "pub-k8s",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
  namespace: config.get("namespace") ?? "npa",
  enrollmentMode: "token",
  workloadType: "deployment",
  hpaEnabled: true,
  hpaMinReplicas: 1,
  hpaMaxReplicas: 3,
  chartValues: {
    extraEnv: [{ name: "HTTPS_PROXY", value: "http://proxy.internal:8080" }],
  },
});

export const publisherNames = publisher.publisherNames;
export const helmReleaseNames = publisher.helmReleaseNames;
export const publishers = pulumi.secret(publisher.publishers);
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import KubernetesPublisher

netskope = pulumi.Config("netskope")
config = pulumi.Config()

publisher = KubernetesPublisher(
    "publisher",
    name_prefix="pub-k8s",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    bearer_token=netskope.require_secret("bearerToken"),
    namespace=config.get("namespace") or "npa",
    enrollment_mode="token",
    workload_type="deployment",
    hpa_enabled=True,
    hpa_min_replicas=1,
    hpa_max_replicas=3,
    chart_values={
        "extraEnv": [{"name": "HTTPS_PROXY", "value": "http://proxy.internal:8080"}],
    },
)

pulumi.export("publisherNames", publisher.publisher_names)
pulumi.export("helmReleaseNames", publisher.helm_release_names)
pulumi.export("publishers", pulumi.Output.secret(publisher.publishers))
```

## C#

```csharp
var publisher = new KubernetesPublisher("publisher", new KubernetesPublisherArgs
{
    NamePrefix = "pub-k8s",
    Replicas = 2,
    TenantUrl = netskope.Require("tenantUrl"),
    BearerToken = netskope.RequireSecret("bearerToken"),
    Namespace = config.Get("namespace") ?? "npa",
    EnrollmentMode = "token",
    WorkloadType = "deployment",
    HpaEnabled = true,
    HpaMinReplicas = 1,
    HpaMaxReplicas = 3,
});
```

## Go

```go
publisher, err := netskopepublisher.NewKubernetesPublisher(ctx, "publisher", &netskopepublisher.KubernetesPublisherArgs{
	NamePrefix:      pulumi.String("pub-k8s"),
	Replicas:        pulumi.Int(2),
	TenantUrl:       pulumi.String(netskope.Require("tenantUrl")),
	BearerToken:        netskope.RequireSecret("bearerToken"),
	Namespace:       pulumi.String("npa"),
	EnrollmentMode:  pulumi.String("token"),
	WorkloadType:    pulumi.String("deployment"),
	HpaEnabled:      pulumi.Bool(true),
	HpaMinReplicas:  pulumi.Int(1),
	HpaMaxReplicas:  pulumi.Int(3),
})
if err != nil {
	return err
}
ctx.Export("publisherNames", publisher.PublisherNames)
ctx.Export("helmReleaseNames", publisher.HelmReleaseNames)
ctx.Export("publishers", pulumi.ToSecret(publisher.Publishers))
```

## Java

```java
var publisher = new KubernetesPublisher("publisher", KubernetesPublisherArgs.builder()
    .namePrefix("pub-k8s")
    .replicas(2)
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .namespace("npa")
    .enrollmentMode("token")
    .workloadType("deployment")
    .hpaEnabled(true)
    .hpaMinReplicas(1)
    .hpaMaxReplicas(3)
    .build());

ctx.export("publisherNames", publisher.publisherNames());
ctx.export("helmReleaseNames", publisher.helmReleaseNames());
ctx.export("publishers", Output.secret(publisher.publishers()));
```

## Rust

```rust
let publisher = netskope::kubernetes_publisher::create(
    ctx,
    "publisher",
    netskope::kubernetes_publisher::KubernetesPublisherArgs::builder()
        .name_prefix("pub-k8s")
        .replicas(2)
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .namespace("npa")
        .enrollment_mode("token")
        .workload_type("deployment")
        .hpa_enabled(true)
        .hpa_min_replicas(1)
        .hpa_max_replicas(3)
        .build_struct(),
);

add_export("publisherNames", &publisher.publisher_names);
add_export("helmReleaseNames", &publisher.helm_release_names);
add_export("publishers", &publisher.publishers);
```
