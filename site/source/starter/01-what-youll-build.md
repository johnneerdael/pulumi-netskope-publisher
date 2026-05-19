---
title: What You'll Build
---

# What You'll Build

By the end of this guide you'll have one Netskope Private Access
Publisher deployed from Pulumi and visible in your Netskope tenant.

The package supports AWS, Azure, GCP, Kubernetes, and vSphere. The
starter path focuses on AWS and GCP:

- **AWS** boots a Netskope Publisher AMI by default, then passes the
  registration token to the publisher wizard through cloud-init.
- **GCP** boots a standard Ubuntu 22.04 Compute Engine image, then
  cloud-init runs Netskope's generic bootstrap script before running the
  publisher wizard. There is no public Netskope Publisher image in GCP.
- **Kubernetes** installs the `kubernetes-netskope-publisher` Helm chart
  into an existing cluster and supports token or API enrollment.

In both cases Pulumi registers or reuses the publisher record in
Netskope, generates a registration token, creates the VM, and returns
secret publisher outputs.

**Next:** [Install the tools](/pulumi-netskope-publisher/starter/02-install-tools/).
