---
title: First Publisher
---

# First Publisher

```bash
npm install
pulumi preview
pulumi up
```

Pulumi creates the Netskope publisher registration and the AWS EC2
or GCP Compute Engine instance in one deployment.

On GCP, the instance first runs the Netskope generic bootstrap script
from cloud-init, then runs `npa_publisher_wizard` with the generated
registration token.

**Next:** [Verify it's online](/pulumi-netskope-publisher/starter/07-verify-online/).
