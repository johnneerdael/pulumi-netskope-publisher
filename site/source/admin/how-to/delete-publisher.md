---
title: Delete a publisher cleanly
---

# Delete a publisher cleanly

`pulumi destroy` deletes the cloud or Kubernetes resources that Pulumi
owns. It does not delete the Netskope publisher record from the tenant.

After destroying infrastructure, remove the tenant-side record only when
you are sure it is no longer referenced by policy:

1. Open the Netskope admin console.
2. Go to **Settings -> Security Cloud Platform -> Netskope Private Access -> Publishers**.
3. Select the publisher and delete it.

Or use the API:

```bash
curl -X DELETE \
  -H "Netskope-Api-Token: $NETSKOPE_API_TOKEN" \
  "$NETSKOPE_TENANT_URL/api/v2/infrastructure/publishers/<publisher_id>"
```

Read the `publisherId` from the secret `publishers` output before
destroying:

```bash
pulumi stack output publishers --show-secrets
```
