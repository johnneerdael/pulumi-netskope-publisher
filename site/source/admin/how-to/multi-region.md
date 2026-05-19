---
title: Multi-region deployments
---

# Multi-region deployments

Create one provider instance and one publisher component per region,
zone, or cloud placement boundary.

```ts
const eu = new aws.Provider("eu", { region: "eu-west-1" });
const us = new aws.Provider("us", { region: "us-east-1" });

new AwsPublisher("publisher-eu", {
  namePrefix: "pub-eu",
  replicas: 2,
  tenantUrl,
  apiToken,
  subnetId: euSubnetId,
  securityGroupIds: [euSecurityGroupId],
}, { provider: eu });

new AwsPublisher("publisher-us", {
  namePrefix: "pub-us",
  replicas: 2,
  tenantUrl,
  apiToken,
  subnetId: usSubnetId,
  securityGroupIds: [usSecurityGroupId],
}, { provider: us });
```

Use distinct `namePrefix` values or explicit names so tenant-side
publisher records do not collide.

For mixed clouds, instantiate the matching Pulumi provider for each
component and keep network inputs local to that cloud.
