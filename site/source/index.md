---
title: pulumi-netskope-publisher
---

# pulumi-netskope-publisher

Provision Netskope Private Access Publishers on AWS, Azure, GCP,
Kubernetes, and vSphere with Pulumi.

Start with the [starter walkthrough](/pulumi-netskope-publisher/starter/)
or read the [provider matrix](/pulumi-netskope-publisher/reference/provider-matrix/).

Hyper-V is exposed as an experimental gate because the upstream Pulumi
Hyper-V provider is not published to npm.

The repository also includes a Go executable component provider and
Registry metadata for Pulumi Registry publication. The Go provider
constructs AWS, Azure, GCP, Kubernetes, and vSphere child resources and
includes a stateful `NetskopeRegistration` resource.
