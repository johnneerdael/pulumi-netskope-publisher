---
title: Cloud Account Prep
---

# Cloud Account Prep

## AWS

Prepare an existing VPC subnet and security group for the publisher EC2
instance. `AwsPublisher` expects `subnetId` and `securityGroupIds`.

The AWS default image lookup searches for the latest AMI named
`Netskope Private Access Publisher*` owned by AWS account `679593333241`.
Use `amiId` to override this lookup.

## GCP

Prepare an existing VPC network and subnetwork for the publisher Compute
Engine instance. `GcpPublisher` expects `project`, `zone`, `network`,
`subnetwork`, and `image`.

Use a normal Linux image such as Ubuntu 22.04:

```text
projects/ubuntu-os-cloud/global/images/family/ubuntu-2204-lts
```

GCP does not have a public Netskope Publisher image. The component runs
Netskope's generic bootstrap script from cloud-init by default and then
registers the publisher with the generated token.

**Next:** [Prepare Netskope](/pulumi-netskope-publisher/starter/04-netskope-tenant-prep/).
