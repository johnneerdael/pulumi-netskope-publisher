---
title: Rotate the registration token
---

# Rotate the registration token

Registration tokens are single-use enrollment tokens. To rotate one,
replace the Pulumi registration output and the workload that consumes it.

## Component-owned registration

Use targeted replacement for the publisher component, then apply during a
maintenance window:

```bash
pulumi preview --target-replace urn:pulumi:prod::stack::netskope-publisher:index:AwsPublisher::publisher
pulumi up --target-replace urn:pulumi:prod::stack::netskope-publisher:index:AwsPublisher::publisher
```

Get the exact URN from the preview output or `pulumi stack --show-urns`.

## Manually supplied registrations

If `registrations` is passed directly, generate a fresh token in
Netskope, update the secret source, and replace the VM, chart release, or
component using that token.

The tenant-side publisher record can stay in place; the consumed token
and workload are the pieces that need replacement.
