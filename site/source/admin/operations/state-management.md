---
title: State management
---

# State management

## What ends up in state

Pulumi state includes:

- Netskope publisher IDs and registration tokens.
- Rendered cloud-init, including the registration token.
- Provider resource attributes such as VM IDs and IP addresses.
- Kubernetes Secret resource metadata and values, encrypted when passed
  as Pulumi secrets.

Registration tokens in state are unavoidable when Pulumi owns the
registration flow.

## Backend recommendation

Use a remote Pulumi backend with encryption and access control:

| Backend | Notes |
|---|---|
| Pulumi Cloud | Built-in state storage, RBAC, auditability, and secret management. |
| S3, Azure Blob, or GCS | Use bucket encryption, object versioning, and narrow IAM. |
| Self-managed file backend | Suitable only for local testing unless wrapped in secure storage. |

Never commit Pulumi state files or exported stack JSON to source control.

## Stack isolation

Use separate stacks for dev, staging, and production. Tenant-side
publisher names are global within the Netskope tenant, so namespace them
with `namePrefix` or explicit `names`.

```bash
pulumi stack init prod
pulumi config set netskope:tenantUrl https://tenant.goskope.com
pulumi config set netskope:apiToken --secret
```

## Moving state

Before changing publisher names, component logical names, or resource
parents, run `pulumi preview` and inspect replacement plans carefully.
Use Pulumi aliases or imports when preserving existing VMs matters.

Destroying the Pulumi stack removes infrastructure resources, but it does
not delete the Netskope publisher records. Delete those separately when
the tenant-side object should also go away.
