---
title: pulumi-netskope-publisher
---

# pulumi-netskope-publisher

Provision Netskope Private Access Publishers on AWS, Azure, GCP,
Kubernetes, vSphere, ESXi, Hcloud, Nutanix, OpenStack, OVH, Scaleway,
OCI, and Alicloud with Pulumi SDKs for TypeScript, Python, C#, Go,
Java, and Rust.

Start with the [starter walkthrough](/pulumi-netskope-publisher/starter/)
or read the [provider matrix](/pulumi-netskope-publisher/reference/provider-matrix/).

Hyper-V is exposed as an experimental gate because the upstream Pulumi
Hyper-V provider is not published to npm.

The repository also includes a Go executable component provider,
generated SDKs, and Registry metadata for Pulumi Registry publication.
The Go provider constructs cloud, virtualization, Kubernetes, and
registration child resources and includes a stateful
`NetskopeRegistration` resource.
