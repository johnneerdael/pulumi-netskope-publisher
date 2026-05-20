---
title: Netskope Registration
toc: true
---

# NetskopeRegistration

Java examples use the published
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
SDK. Rust examples use the published
[`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
crate plus Pulumi Gestalt.

`NetskopeRegistration` is the stateful provider resource that registers
or reuses Netskope publisher records and generates registration tokens.

Provider-specific publisher components create this resource
automatically when `tenantUrl` and either `bearerToken` or OAuth2
client credentials are supplied. Use it directly when you want to
separate tenant registration from platform infrastructure.

## Inputs

- `publisherNames`
- `tenantUrl`
- `bearerToken`, or `authMode: "oauth2"` plus `oauth2`

`apiToken` remains accepted as a backwards-compatible alias for
`bearerToken`.

## Outputs

- `registrations`, keyed by publisher name
- `publisherId`
- `registrationToken`
- `existedBefore`

## Pulumi YAML

Static bearer token enrollment:

```yaml
name: netskope-registration
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  registration:
    type: netskope-publisher:index:NetskopeRegistration
    properties:
      publisherNames:
        - pub-eu-1
        - pub-eu-2
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
outputs:
  registrations: ${registration.registrations}
```

OAuth2 client credentials enrollment:

```yaml
name: netskope-registration-oauth2
runtime: yaml
config:
  tenantUrl:
    type: String
  oauthClientSecret:
    type: String
    secret: true
resources:
  registration:
    type: netskope-publisher:index:NetskopeRegistration
    properties:
      publisherNames:
        - pub-eu-1
        - pub-eu-2
      tenantUrl: ${tenantUrl}
      authMode: oauth2
      oauth2:
        tokenUrl: https://tenant.goskope.com/oauth2/token
        clientId: <client-id>
        clientSecret: ${oauthClientSecret}
outputs:
  registrations: ${registration.registrations}
```

## TypeScript

```ts
import * as pulumi from "@pulumi/pulumi";
import { NetskopeRegistration, AwsPublisher } from "@johninnl/pulumi-netskope-publisher";

const netskope = new pulumi.Config("netskope");

const registration = new NetskopeRegistration("registration", {
  publisherNames: ["pub-eu-1", "pub-eu-2"],
  tenantUrl: netskope.require("tenantUrl"),
  bearerToken: netskope.requireSecret("bearerToken"),
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

OAuth2:

```ts
const registration = new NetskopeRegistration("registration", {
  publisherNames: ["pub-eu-1", "pub-eu-2"],
  tenantUrl: netskope.require("tenantUrl"),
  authMode: "oauth2",
  oauth2: {
    tokenUrl: netskope.require("oauthTokenUrl"),
    clientId: netskope.require("oauthClientId"),
    clientSecret: netskope.requireSecret("oauthClientSecret"),
  },
});
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
    bearer_token=netskope.require_secret("bearerToken"),
)

pulumi.export("registrations", pulumi.Output.secret(registration.registrations))
```

OAuth2:

```python
registration = NetskopeRegistration(
    "registration",
    publisher_names=["pub-eu-1", "pub-eu-2"],
    tenant_url=netskope.require("tenantUrl"),
    auth_mode="oauth2",
    oauth2={
        "token_url": netskope.require("oauthTokenUrl"),
        "client_id": netskope.require("oauthClientId"),
        "client_secret": netskope.require_secret("oauthClientSecret"),
    },
)
```

## C#

```csharp
var registration = new NetskopeRegistration("registration", new NetskopeRegistrationArgs
{
    PublisherNames = { "pub-eu-1", "pub-eu-2" },
    TenantUrl = netskope.Require("tenantUrl"),
    BearerToken = netskope.RequireSecret("bearerToken"),
});
```

OAuth2:

```csharp
var registration = new NetskopeRegistration("registration", new NetskopeRegistrationArgs
{
    PublisherNames = { "pub-eu-1", "pub-eu-2" },
    TenantUrl = netskope.Require("tenantUrl"),
    AuthMode = "oauth2",
    Oauth2 = new Pulumi.NetskopePublisher.Provider.Inputs.NetskopeOAuth2ArgsArgs
    {
        TokenUrl = netskope.Require("oauthTokenUrl"),
        ClientId = netskope.Require("oauthClientId"),
        ClientSecret = netskope.RequireSecret("oauthClientSecret"),
    },
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
	BearerToken: netskope.RequireSecret("bearerToken"),
})
if err != nil {
	return err
}
ctx.Export("registrations", pulumi.ToSecret(registration.Registrations))
```

OAuth2:

```go
registration, err := netskopepublisher.NewNetskopeRegistration(ctx, "registration", &netskopepublisher.NetskopeRegistrationArgs{
	PublisherNames: pulumi.StringArray{
		pulumi.String("pub-eu-1"),
		pulumi.String("pub-eu-2"),
	},
	TenantUrl: pulumi.String(netskope.Require("tenantUrl")),
	AuthMode:  pulumi.StringPtr("oauth2"),
	Oauth2: &provider.NetskopeOAuth2ArgsArgs{
		TokenUrl:     pulumi.String(netskope.Require("oauthTokenUrl")),
		ClientId:     pulumi.String(netskope.Require("oauthClientId")),
		ClientSecret: netskope.RequireSecret("oauthClientSecret"),
	},
})
```

## Java

```java
var registration = new NetskopeRegistration("registration", NetskopeRegistrationArgs.builder()
    .publisherNames("pub-eu-1", "pub-eu-2")
    .tenantUrl(netskope.require("tenantUrl"))
    .bearerToken(netskope.requireSecret("bearerToken"))
    .build());

ctx.export("registrations", Output.secret(registration.registrations()));
```

OAuth2:

```java
var registration = new NetskopeRegistration("registration", NetskopeRegistrationArgs.builder()
    .publisherNames("pub-eu-1", "pub-eu-2")
    .tenantUrl(netskope.require("tenantUrl"))
    .authMode("oauth2")
    .oauth2(NetskopeOAuth2ArgsArgs.builder()
        .tokenUrl(netskope.require("oauthTokenUrl"))
        .clientId(netskope.require("oauthClientId"))
        .clientSecret(netskope.requireSecret("oauthClientSecret"))
        .build())
    .build());
```

## Rust

```rust
let registration = netskope::netskope_registration::create(
    ctx,
    "registration",
    netskope::netskope_registration::NetskopeRegistrationArgs::builder()
        .publisher_names(vec!["pub-eu-1".to_string(), "pub-eu-2".to_string()])
        .tenant_url("https://tenant.goskope.com")
        .bearer_token("secret-token")
        .build_struct(),
);

add_export("registrations", &registration.registrations);
```

OAuth2:

```rust
let registration = netskope::netskope_registration::create(
    ctx,
    "registration",
    netskope::netskope_registration::NetskopeRegistrationArgs::builder()
        .publisher_names(vec!["pub-eu-1".to_string(), "pub-eu-2".to_string()])
        .tenant_url("https://tenant.goskope.com")
        .auth_mode("oauth2")
        .oauth2(netskope::types::NetskopeOAuth2Args::builder()
            .token_url("https://tenant.goskope.com/oauth2/token")
            .client_id("client-id")
            .client_secret("client-secret")
            .build_struct())
        .build_struct(),
);
```
