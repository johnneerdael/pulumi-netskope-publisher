---
title: Secret handling
---

# Secret handling

## What is secret

| Value | Type |
|---|---|
| Netskope bearer token | Tenant API credential for static-token auth |
| OAuth2 client secret | Secret used to fetch short-lived bearer tokens |
| Registration tokens | Short-lived publisher enrollment tokens |
| Rendered cloud-init | Contains registration tokens |
| `publishers` output | Secret output because it includes registration tokens |

## Store static auth as a Pulumi secret

```bash
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:bearerToken --secret
```

Read it in code and pass it to the component:

```ts
const config = new pulumi.Config("netskope");
const tenantUrl = config.require("tenantUrl");
const bearerToken = config.requireSecret("bearerToken");
```

`apiToken` remains accepted as a backwards-compatible alias, but new
programs should use `bearerToken`.

## Store OAuth2 credentials as Pulumi secrets

```bash
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:oauthTokenUrl https://tenant.goskope.com/oauth2/token
pulumi config set netskope:oauthClientId <client-id>
pulumi config set netskope:oauthClientSecret --secret
```

```ts
authMode: "oauth2",
oauth2: {
  tokenUrl: config.require("oauthTokenUrl"),
  clientId: config.require("oauthClientId"),
  clientSecret: config.requireSecret("oauthClientSecret"),
}
```

Use a Pulumi backend with encrypted state and scoped access. Secrets are
encrypted in state, but anyone with stack access and the secrets provider
can decrypt them.

## Registration tokens

Normally do not handle registration tokens directly. They flow from the
Netskope API to Pulumi state, then into cloud-init or Kubernetes Secrets,
and are consumed by the publisher at first boot.

When passing `registrations` manually, mark the token secret:

```ts
registrations: pulumi.secret({
  "pub-eu-1": {
    publisherId: 12345,
    registrationToken: "token-value",
    existedBefore: true,
  },
}),
```

## Rotating credentials

After rotating the upstream Netskope bearer token or OAuth2 client secret:

1. Update the Pulumi stack secret.
2. Run `pulumi preview`.
3. Apply only if the preview matches the infrastructure change you
   expected.

The publisher VMs do not use tenant API credentials after enrollment.

For per-publisher token rotation, see
[Rotate the registration token](/pulumi-netskope-publisher/admin/how-to/rotate-token/).
