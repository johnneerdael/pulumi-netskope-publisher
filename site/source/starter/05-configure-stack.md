---
title: Configure Stack
---

# Configure Stack

## AWS

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set subnetId subnet-1234567890abcdef0
pulumi config set securityGroupIds '["sg-1234567890abcdef0"]'
pulumi config set replicas 1
```

## GCP

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set project my-project
pulumi config set zone europe-west4-a
pulumi config set network default
pulumi config set subnetwork default
pulumi config set image projects/ubuntu-os-cloud/global/images/family/ubuntu-2204-lts
pulumi config set replicas 1
```

For GCP, leave `bootstrap` unset unless you are booting a custom
pre-baked publisher image. The default is `bootstrap: true`.

**Next:** [Deploy the first publisher](/pulumi-netskope-publisher/starter/06-first-publisher/).
