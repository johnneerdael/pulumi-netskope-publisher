---
title: GCP Component
toc: true
---

# GCP Component

`GcpPublisher` creates one Compute Engine instance per publisher name.

## Inputs

Required:

- `project`
- `zone`
- `network`
- `subnetwork`
- `image`
- `tenantUrl` and `apiToken`, unless `registrations` is provided

Use a stock Ubuntu image for bootstrap mode:

```text
projects/ubuntu-os-cloud/global/images/family/ubuntu-2204-lts
```

Common optional inputs include `machineType`, `assignPublicIp`,
`networkTags`, `serviceAccount`, `bootstrap`, `bootstrapUrl`, `nonat`,
`installUser`, `installUserPassword`, `installUserPasswordIsHash`,
`installUserSshAuthorizedKeys`, `deleteDefaultUser`,
`guestNetworkInterface`, `tags`, `namePrefix`, `names`, and `replicas`.

GCP has no public Netskope Publisher image, so the component normally
runs the bootstrap script from cloud-init. Set `bootstrap: false` only
when using a custom image that already has `npa_publisher_wizard`.

## Outputs

- `publisherNames`
- secret `publishers`

## Pulumi CLI

```bash
pulumi new typescript
pulumi config set gcp:project my-gcp-project
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:apiToken --secret
pulumi config set project my-gcp-project
pulumi config set zone europe-west4-a
pulumi config set network default
pulumi config set subnetwork default
pulumi config set image projects/ubuntu-os-cloud/global/images/family/ubuntu-2204-lts
pulumi up
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { GcpPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");
const config = new pulumi.Config();

const publisher = new GcpPublisher("publisher", {
  namePrefix: "pub-gcp",
  replicas: 2,
  tenantUrl: netskope.require("tenantUrl"),
  apiToken: netskope.requireSecret("apiToken"),
  project: config.require("project"),
  zone: config.require("zone"),
  network: config.require("network"),
  subnetwork: config.require("subnetwork"),
  image: config.require("image"),
  machineType: "e2-medium",
  assignPublicIp: false,
  nonat: true,
  tags: { service: "npa" },
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import GcpPublisher

netskope = pulumi.Config("netskope")
config = pulumi.Config()

publisher = GcpPublisher(
    "publisher",
    name_prefix="pub-gcp",
    replicas=2,
    tenant_url=netskope.require("tenantUrl"),
    api_token=netskope.require_secret("apiToken"),
    project=config.require("project"),
    zone=config.require("zone"),
    network=config.require("network"),
    subnetwork=config.require("subnetwork"),
    image=config.require("image"),
    machine_type="e2-medium",
    assign_public_ip=False,
    nonat=True,
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

    var publisher = new GcpPublisher("publisher", new GcpPublisherArgs
    {
        NamePrefix = "pub-gcp",
        Replicas = 2,
        TenantUrl = netskope.Require("tenantUrl"),
        ApiToken = netskope.RequireSecret("apiToken"),
        Project = config.Require("project"),
        Zone = config.Require("zone"),
        Network = config.Require("network"),
        Subnetwork = config.Require("subnetwork"),
        Image = config.Require("image"),
        MachineType = "e2-medium",
        AssignPublicIp = false,
        Nonat = true,
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
publisher, err := netskopepublisher.NewGcpPublisher(ctx, "publisher", &netskopepublisher.GcpPublisherArgs{
	NamePrefix:    pulumi.String("pub-gcp"),
	Replicas:      pulumi.Int(2),
	TenantUrl:     pulumi.String(netskope.Require("tenantUrl")),
	ApiToken:      netskope.RequireSecret("apiToken"),
	Project:       pulumi.String(cfg.Require("project")),
	Zone:          pulumi.String(cfg.Require("zone")),
	Network:       pulumi.String(cfg.Require("network")),
	Subnetwork:    pulumi.String(cfg.Require("subnetwork")),
	Image:         pulumi.String(cfg.Require("image")),
	MachineType:   pulumi.String("e2-medium"),
	AssignPublicIp: pulumi.Bool(false),
	Nonat:         pulumi.Bool(true),
	Tags: pulumi.StringMap{
		"service": pulumi.String("npa"),
	},
})
if err != nil {
	return err
}
ctx.Export("publisherNames", publisher.PublisherNames)
ctx.Export("publishers", pulumi.ToSecret(publisher.Publishers))
```
