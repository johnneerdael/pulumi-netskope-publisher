---
title: Troubleshooting
toc: true
---

# Troubleshooting

## Registration flow failures

### `List publishers failed (status=401)`

The API token is missing, expired, or lacks publisher read/write scope.
Create a new Netskope REST API v2 token with access to
`/infrastructure/publishers`, then update the Pulumi stack secret.

### `List publishers failed (status=404)`

`tenantUrl` is probably wrong. Open the URL in a browser and confirm it
is the tenant admin console URL, then rerun `pulumi preview`.

### `Create publisher ... failed (status=409)`

A publisher with that name already exists but the list endpoint did not
return it. Use a token with tenant-wide visibility, or pass
`registrations` for pre-created publishers.

## VM provisioning failures

### AWS AMI lookup found nothing

When `bootstrap` is true and `amiId` is omitted, the component resolves a
Canonical Ubuntu 22.04 Minimal AMI. Make sure the deployment identity can
call `ec2:DescribeImages`, or pass `amiId` explicitly.

### Azure marketplace image failures

When using a marketplace image, make sure the subscription has accepted
the image terms or set the component inputs needed for marketplace use.
With `bootstrap: true`, the component uses Canonical Ubuntu Minimal and
does not need Netskope marketplace terms.

## Cloud-init and wizard failures

### Publisher never goes online

The most common cause is blocked outbound TCP/443. Confirm the instance,
VM, or pod has egress to the Netskope tenant URL and regional gateway
infrastructure.

SSH to a VM and inspect cloud-init:

```bash
sudo cloud-init status --long
sudo journalctl -u cloud-final --no-pager | tail -100
sudo tail -100 /var/log/cloud-init-output.log
```

Common log patterns:

| Symptom | Likely cause |
|---|---|
| Cannot connect to Netskope gateway | Outbound 443, DNS, proxy, or routing issue. |
| Authentication failed from wizard | Registration token was already consumed. Rotate the token and replace the workload. |
| `npa_publisher_wizard: command not found` | Wrong pre-baked image or bootstrap failed before installing the wizard. |
| Cannot download `bootstrap.sh` | Egress to S3 is blocked; set `bootstrapUrl` to an internal mirror or allow HTTPS to the bucket. |

## Kubernetes chart failures

Use standard Pulumi and Kubernetes diagnostics:

```bash
pulumi stack output publisherNames
kubectl get pods -n npa
kubectl describe pod -n npa -l app.kubernetes.io/name=kubernetes-netskope-publisher
kubectl logs -n npa -l app.kubernetes.io/name=kubernetes-netskope-publisher
```

Check image pull access, NetworkPolicies, chart values, and Secret names.

## Getting help

Open an issue on the project repository with platform, package version,
Pulumi CLI version, provider versions, and a redacted error.
