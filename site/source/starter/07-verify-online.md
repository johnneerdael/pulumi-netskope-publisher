---
title: Verify Online
---

# Verify Online

Check the Netskope tenant for the publisher name emitted by Pulumi:

```bash
pulumi stack output publisherNames
```

The publisher should appear in Netskope after the instance boots and
cloud-init runs the registration wizard.

For GCP, allow extra time for the first boot because cloud-init downloads
and installs the publisher software before registration.

**Next:** [Tear it down](/pulumi-netskope-publisher/starter/08-tear-down/).
