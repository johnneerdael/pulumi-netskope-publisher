---
title: AWS Account Prep
---

# AWS Account Prep

Prepare an existing VPC subnet and security group for the publisher EC2
instance. The component expects `subnetId` and `securityGroupIds`.

The default image lookup searches for the latest AMI named
`Netskope Private Access Publisher*` owned by AWS account `679593333241`.
Use `amiId` to override this lookup.
