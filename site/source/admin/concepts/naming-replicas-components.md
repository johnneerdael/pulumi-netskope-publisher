---
title: Naming, replicas, and component instances
---

# Naming, replicas, and component instances

Each platform component resolves names the same way:

```text
names ?? [namePrefix-1, namePrefix-2, ... namePrefix-replicas]
```

- Set `names` for stable explicit publisher names.
- Otherwise set `namePrefix` and `replicas`.
- `replicas` defaults to one publisher.
- `namePrefix` defaults to `npa-publisher`.

The resolved list is exposed as the `publisherNames` output.

## Explicit names

```ts
new AwsPublisher("publisher", {
  names: ["pub-eu-1", "pub-eu-2"],
  tenantUrl,
  bearerToken,
  subnetId,
  securityGroupIds: [securityGroupId],
});
```

Explicit names are safest for production because removing one name does
not force unrelated publishers to be renamed.

## Derived names

```ts
new AwsPublisher("publisher", {
  namePrefix: "pub-eu",
  replicas: 2,
  tenantUrl,
  bearerToken,
  subnetId,
  securityGroupIds: [securityGroupId],
});
```

This creates `pub-eu-1` and `pub-eu-2`.

## Multi-AZ and multi-region naming

Use one component instance per placement boundary when networking differs
by zone, region, or provider:

```ts
new AwsPublisher("publisher-eu-a", {
  namePrefix: "pub-eu-a",
  replicas: 1,
  tenantUrl,
  bearerToken,
  subnetId: euSubnetA,
  securityGroupIds: [euSecurityGroup],
});

new AwsPublisher("publisher-eu-b", {
  namePrefix: "pub-eu-b",
  replicas: 1,
  tenantUrl,
  bearerToken,
  subnetId: euSubnetB,
  securityGroupIds: [euSecurityGroup],
});
```
