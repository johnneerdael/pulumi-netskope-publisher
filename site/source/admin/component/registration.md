---
title: Netskope Registration
toc: true
---

# NetskopeRegistration

`NetskopeRegistration` is the stateful provider resource that registers
or reuses Netskope publisher records and generates registration tokens.

Provider-specific publisher components create this resource
automatically when `tenantUrl` and `apiToken` are supplied. Use it
directly when you want to separate tenant registration from platform
infrastructure.

## Inputs

- `publisherNames`
- `tenantUrl`
- `apiToken`

## Outputs

- `registrations`, keyed by publisher name
- `publisherId`
- `registrationToken`
- `existedBefore`

## Pulumi CLI

```bash
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:apiToken --secret
pulumi up
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { NetskopeRegistration, AwsPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");

const registration = new NetskopeRegistration("registration", {
  publisherNames: ["pub-eu-1", "pub-eu-2"],
  tenantUrl: netskope.require("tenantUrl"),
  apiToken: netskope.requireSecret("apiToken"),
});

const publisher = new AwsPublisher("publisher", {
  names: ["pub-eu-1", "pub-eu-2"],
  registrations: registration.registrations,
  subnetId,
  securityGroupIds: [securityGroupId],
});

export const registrations = pulumi.secret(registration.registrations);
export const publisherNames = publisher.publisherNames;
```

## Python

```python
import pulumi
from pulumi_netskope_publisher import NetskopeRegistration

netskope = pulumi.Config("netskope")

registration = NetskopeRegistration(
    "registration",
    publisher_names=["pub-eu-1", "pub-eu-2"],
    tenant_url=netskope.require("tenantUrl"),
    api_token=netskope.require_secret("apiToken"),
)

pulumi.export("registrations", pulumi.Output.secret(registration.registrations))
```

## C#

```csharp
var registration = new NetskopeRegistration("registration", new NetskopeRegistrationArgs
{
    PublisherNames = { "pub-eu-1", "pub-eu-2" },
    TenantUrl = netskope.Require("tenantUrl"),
    ApiToken = netskope.RequireSecret("apiToken"),
});
```

## Go

```go
registration, err := netskopepublisher.NewNetskopeRegistration(ctx, "registration", &netskopepublisher.NetskopeRegistrationArgs{
	PublisherNames: pulumi.StringArray{
		pulumi.String("pub-eu-1"),
		pulumi.String("pub-eu-2"),
	},
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	ApiToken:  netskope.RequireSecret("apiToken"),
})
if err != nil {
	return err
}
ctx.Export("registrations", pulumi.ToSecret(registration.Registrations))
```
