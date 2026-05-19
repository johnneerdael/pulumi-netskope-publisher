---
title: Secret handling
---

# Secret handling

## What is secret

| Value | Type |
|---|---|
| Netskope API token | Long-lived tenant API credential |
| Registration tokens | Short-lived publisher enrollment tokens |
| Rendered cloud-init | Contains registration tokens |
| `publishers` output | Secret output because it includes registration tokens |

## Store the API token as a Pulumi secret

```bash
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:apiToken --secret
```

Read it in code and pass it to the component:

```ts
const config = new pulumi.Config("netskope");
const tenantUrl = config.require("tenantUrl");
const apiToken = config.requireSecret("apiToken");
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

## Rotating the API token

After rotating the upstream Netskope API token:

1. Update the Pulumi stack secret.
2. Run `pulumi preview`.
3. Apply only if the preview matches the infrastructure change you
   expected.

The publisher VMs do not use the API token after enrollment.

For per-publisher token rotation, see
[Rotate the registration token](/pulumi-netskope-publisher/admin/how-to/rotate-token/).
