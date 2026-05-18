---
title: vSphere Component
---

# vSphere Component

`VspherePublisher` clones one VM per publisher name from an existing
template.

Required inputs: `datacenter`, `datastore`, `networkName`,
`templateName`, and either `cluster` or `host`.

Optional inputs include `folder`, `numCpus`, `memory`, `tags`,
`namePrefix`, `names`, and `replicas`.

Use the official Netskope OVA to prepare the template:

```text
https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.ova
```

Outputs: `publisherNames` and secret `publishers`.
