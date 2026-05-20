---
title: Provision an HA pair
---

# Provision an HA pair

Set `replicas: 2` or pass two explicit names. Each publisher gets its
own Netskope record, token, and workload.

```ts
const publisher = new AwsPublisher("publisher", {
  namePrefix: "pub-eu",
  replicas: 2,
  tenantUrl,
  bearerToken,
  subnetId,
  securityGroupIds: [securityGroupId],
  keyName,
});
```

This creates `pub-eu-1` and `pub-eu-2` in the same subnet.

For multi-AZ HA, create one component per subnet:

```ts
new AwsPublisher("publisher-a", {
  names: ["pub-eu-a"],
  tenantUrl,
  bearerToken,
  subnetId: subnetA,
  securityGroupIds: [securityGroupId],
});

new AwsPublisher("publisher-b", {
  names: ["pub-eu-b"],
  tenantUrl,
  bearerToken,
  subnetId: subnetB,
  securityGroupIds: [securityGroupId],
});
```

Attach both publishers to the same Netskope private apps in the Netskope
admin console.
