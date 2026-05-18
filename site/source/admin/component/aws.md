---
title: AWS Component
---

# AWS Component

`AwsPublisher` creates one EC2 instance per publisher name.

## Required inputs

- `subnetId`
- `securityGroupIds`
- `tenantUrl` and `apiToken`, unless `registrations` is provided

## Common optional inputs

- `namePrefix`
- `names`
- `replicas`
- `amiId`
- `instanceType`
- `keyName`
- `tags`

## Outputs

- `publisherNames`
- `publishers`

`publishers` is keyed by publisher name and contains publisher ID,
registration token, EC2 instance ID, private IP, and public IP.
