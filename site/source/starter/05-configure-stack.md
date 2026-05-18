---
title: Configure Stack
---

# Configure Stack

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set subnetId subnet-1234567890abcdef0
pulumi config set securityGroupIds '["sg-1234567890abcdef0"]'
pulumi config set replicas 1
```
