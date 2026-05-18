# Hyper-V Experimental Status

Hyper-V support is not enabled by default.

The upstream Pulumi Hyper-V provider exists at
`https://github.com/pulumi/pulumi-hyperv`, but `@pulumi/hyperv` is not
published to npm. The package exposes a generated Node SDK in
`sdk/nodejs` and marks itself as `1.0.0-alpha.0+dev`.

Use the official Netskope VHDX source when preparing Hyper-V images:

```text
https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.vhdx
```

The related official OVA source for vSphere template preparation is:

```text
https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.ova
```

This repository keeps `HypervPublisher` behind an explicit runtime gate
until the provider can be consumed through a stable package source.
