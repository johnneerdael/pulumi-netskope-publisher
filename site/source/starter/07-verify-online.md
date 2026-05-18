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
