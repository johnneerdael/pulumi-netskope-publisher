---
title: Provider Matrix
---

# Provider Matrix

| Platform | Component | Status |
|---|---|---|
| AWS | `AwsPublisher` | Supported |
| Azure | `AzurePublisher` | Supported |
| GCP | `GcpPublisher` | Supported |
| vSphere | `VspherePublisher` | Supported |
| Hyper-V | `HypervPublisher` | Experimental gate |

All supported providers share name derivation, Netskope registration,
cloud-init generation, and secret output conventions.

Official Netskope image sources:

- Hyper-V VHDX:
  `https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.vhdx`
- vSphere OVA:
  `https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.ova`
