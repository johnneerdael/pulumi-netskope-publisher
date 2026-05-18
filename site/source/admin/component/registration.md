---
title: Netskope Registration
---

# NetskopeRegistration

`NetskopeRegistration` is the stateful provider resource that registers
or reuses Netskope publisher records and generates registration tokens.

Inputs:

- `publisherNames`
- `tenantUrl`
- `apiToken`

Outputs:

- `registrations`, keyed by publisher name
- `publisherId`
- `registrationToken`
- `existedBefore`

The provider-specific publisher components create this resource
automatically when `tenantUrl` and `apiToken` are provided. Pass
`registrations` directly to use pre-created publisher records instead.
