---
title: Components
---

# Components

The package exposes provider-specific publisher components:

- `AwsPublisher`
- `AzurePublisher`
- `GcpPublisher`
- `VspherePublisher`
- `HypervPublisher` behind an experimental gate

The TypeScript package supports managed Netskope registration or
pre-created registration tokens. The Go executable provider path
currently requires pre-created `registrations` while Netskope
registration is promoted into a stateful provider resource.
