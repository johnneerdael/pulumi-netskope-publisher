---
title: pulumi-netskope-publisher
---

# pulumi-netskope-publisher

Provision Netskope Private Access Publishers on AWS, Azure, GCP,
Kubernetes, vSphere, ESXi, Hcloud, Nutanix, OpenStack, OVH, Scaleway,
OCI, Alicloud, and Proxmox VE with Pulumi SDKs for TypeScript, Python,
C#, Go, Java, and Rust.

Start with the [starter walkthrough](/pulumi-netskope-publisher/starter/)
or read the [provider matrix](/pulumi-netskope-publisher/reference/provider-matrix/).
Install published SDKs from
[npm](https://www.npmjs.com/package/@johninnl/pulumi-netskope-publisher),
[PyPI](https://pypi.org/project/pulumi-netskope-publisher/),
[NuGet](https://www.nuget.org/packages/JohninNL.Pulumi.NetskopePublisher),
[pkg.go.dev](https://pkg.go.dev/github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher),
[GitHub Packages](https://github.com/johnneerdael/pulumi-netskope-publisher/packages),
and [crates.io](https://crates.io/crates/pulumi-netskope-publisher).

Hyper-V is exposed as an experimental gate because the upstream Pulumi
Hyper-V provider is not published to npm.

The repository also includes a Go executable component provider,
published SDKs, and Registry metadata for Pulumi Registry publication.
The Go provider constructs cloud, virtualization, Kubernetes, and
registration child resources and includes a stateful
`NetskopeRegistration` resource.
