---
title: Components
---

# Components

The package exposes provider-specific publisher components:

- `AwsPublisher`
- `AzurePublisher`
- `GcpPublisher`
- `KubernetesPublisher`
- `VspherePublisher`
- `HypervPublisher` behind an experimental gate
- `NetskopeRegistration`

The TypeScript package and Go executable provider both support managed
Netskope registration or pre-created registration tokens. Kubernetes
deployments additionally support chart API enrollment.
