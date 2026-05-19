---
title: Connectivity requirements
toc: true
---

# Connectivity requirements

Publisher workloads must make outbound TCP/443 connections to Netskope
regional gateway infrastructure and to the tenant API URL. If that path
is blocked, Pulumi can still create the VM, but the publisher remains
offline.

| Direction | Protocol | Destination | Purpose |
|---|---|---|---|
| Egress | TCP 443 | Netskope regional gateways | Publisher control and data plane |
| Egress | TCP 443 | `tenantUrl` | Initial registration |

The Pulumi deployment host also needs HTTPS access to `tenantUrl` so
`NetskopeRegistration` can list publishers, create records, and generate
tokens.

## AWS

`AwsPublisher` uses existing networking. Supply `subnetId` and
`securityGroupIds`, then choose how the instance gets egress:

| Subnet shape | `associatePublicIpAddress` | Notes |
|---|---|---|
| Public subnet with internet gateway route | `true` | Simplest path and allows direct SSH if the security group permits it. |
| Private subnet with NAT gateway | `false` or omitted | Recommended for private workloads. |
| Transit-routed private subnet | `false` or omitted | Verify firewall policy allows TCP/443 outbound and return traffic. |

The component does not create VPCs, gateways, route tables, or security
groups.

## Azure

`AzurePublisher` accepts an existing `subnetId` and can attach a public
IP when `assignPublicIp` is true.

| Subnet shape | `assignPublicIp` | Notes |
|---|---|---|
| Per-VM public IP | `true` | Creates one Standard public IP per publisher VM. |
| NAT gateway on subnet | `false` or omitted | Common production pattern. |
| Firewall or NVA default route | `false` or omitted | Allow outbound TCP/443 to Netskope. |

## GCP

`GcpPublisher` accepts existing `network` and `subnetwork` names.

| Connectivity shape | `assignPublicIp` | Notes |
|---|---|---|
| Ephemeral external IPv4 | `true` | Simple for tests. |
| Cloud NAT | `false` or omitted | Recommended for private fleets. |
| Interconnect or appliance egress | `false` or omitted | Ensure TCP/443 is allowed. |

GCP defaults `nonat` to true because the default 1460-byte MTU works
better with Netskope's No-NAT mode.

## vSphere and Hyper-V

The VM network or virtual switch must provide DNS and outbound HTTPS.
For vSphere, pass `networkName` for the port group connected to that
path. For Hyper-V, the virtual switch should normally be External or an
Internal switch with host NAT configured.

## Kubernetes

Publisher pods need outbound TCP/443 to Netskope and access to the
container registry that serves the publisher image. Override
`imageRepository` when the cluster pulls from a private mirror.

If the namespace has default-deny egress NetworkPolicies, allow TCP/443
from the publisher pods. If egress is proxied, pass the chart's proxy
environment variables through `chartValues`.
