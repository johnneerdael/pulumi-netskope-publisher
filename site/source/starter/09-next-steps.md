---
title: Next Steps
---

# Next Steps

You have a working publisher. From here:

- Review the [provider matrix](/pulumi-netskope-publisher/reference/provider-matrix/)
  before choosing a production platform.
- Read the [AWS component](/pulumi-netskope-publisher/admin/component/aws/)
  and [GCP component](/pulumi-netskope-publisher/admin/component/gcp/)
  inputs for provider-specific options.
- For cluster deployments, read the
  [Kubernetes component](/pulumi-netskope-publisher/admin/component/kubernetes/)
  page and choose `token` or `api` enrollment deliberately.
- Use `names` for stable, explicit publisher names in production.
- Use `registrations` when publisher records and registration tokens are
  created outside Pulumi.
- On GCP, keep using a standard Linux image with `bootstrap: true`
  unless you maintain your own pre-baked publisher image.

For deeper operational guidance, start with
[secret handling](/pulumi-netskope-publisher/admin/operations/secrets/).
