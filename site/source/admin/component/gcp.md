---
title: GCP Component
---

# GCP Component

`GcpPublisher` creates one Compute Engine instance per publisher name.

Required inputs: `project`, `zone`, `network`, `subnetwork`, and `image`.

Optional inputs include `machineType`, `assignPublicIp`, `networkTags`,
`serviceAccount`, `tags`, `namePrefix`, `names`, and `replicas`.

Outputs: `publisherNames` and secret `publishers`.
