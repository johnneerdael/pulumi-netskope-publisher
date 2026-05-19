---
title: Kubernetes Component
---

# Kubernetes Component

`KubernetesPublisher` installs the
`kubernetes-netskope-publisher` Helm chart into an existing Kubernetes
cluster.

Required inputs: `tenantUrl` and `apiToken`, unless token-mode
`registrations` are provided.

Common optional inputs include `namespace`, `enrollmentMode`,
`chartRepository`, `chartVersion`, `chartValues`, `workloadType`,
`hpaEnabled`, `hpaMinReplicas`, `hpaMaxReplicas`, `imageRepository`,
`imageTag`, `tags`, `namePrefix`, `names`, and `replicas`.

## Enrollment modes

`token` mode is the default. Pulumi creates or reuses Netskope publisher
records, creates one Kubernetes Secret per registration token, and
installs one Helm release per publisher name.

`api` mode creates one `npa-api-token` Secret and one Helm release named
`npa-publisher`. The chart registers Publisher pods with the Netskope
API during startup.

Outputs: `publisherNames`, `helmReleaseNames`, and secret `publishers`.
