---
title: Bring your own networking
---

# Bring your own networking

The package does not create VPCs, VNets, subnets, route tables, NAT
gateways, firewalls, or security groups. Pass existing network
identifiers to the component.

| Platform | Networking inputs |
|---|---|
| AWS | `subnetId`, `securityGroupIds`, `associatePublicIpAddress` |
| Azure | `subnetId`, `networkSecurityGroupId`, `assignPublicIp` |
| GCP | `network`, `subnetwork`, `assignPublicIp`, `networkTags` |
| vSphere | `networkName` |
| Kubernetes | provider context, namespace, `chartValues` |

The workload needs outbound TCP/443 to Netskope and DNS resolution for
tenant and gateway hostnames.

## Guest OS interface override

AWS, Azure, and GCP accept `guestNetworkInterface` to write a netplan
override before the publisher install runs.

```ts
new AwsPublisher("publisher", {
  tenantUrl,
  bearerToken,
  subnetId,
  securityGroupIds: [securityGroupId],
  bootstrap: true,
  guestNetworkInterface: {
    name: "ens5",
    dhcp4: false,
    addresses: ["10.0.0.50/24"],
    gateway4: "10.0.0.1",
    nameservers: ["8.8.8.8", "1.1.1.1"],
    mtu: 9001,
  },
});
```

Static addresses must still be valid for the cloud subnet and must not
collide with addresses managed by the platform.
