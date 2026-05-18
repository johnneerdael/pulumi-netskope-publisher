---
title: Azure Component
---

# Azure Component

`AzurePublisher` creates one Linux virtual machine per publisher name.

Required inputs: `resourceGroupName`, `location`, `subnetId`,
`adminSshPublicKey`, and either `imageId` or `marketplace`.

Optional inputs include `vmSize`, `adminUsername`,
`networkSecurityGroupId`, `assignPublicIp`, `osDisk`, `tags`,
`namePrefix`, `names`, and `replicas`.

Outputs: `publisherNames` and secret `publishers`.
