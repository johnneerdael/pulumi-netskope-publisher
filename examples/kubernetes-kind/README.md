# Kubernetes Kind Example

Deploy one or more Netskope Private Access Publishers on Kubernetes by
installing the `kubernetes-netskope-publisher` Helm chart.

```bash
kind create cluster
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set namespace netskope
pulumi config set enrollmentMode token
npm install
pulumi preview
pulumi up
```

Use `enrollmentMode api` when the chart should self-register pods with
the Netskope API token. Use `enrollmentMode token` when Pulumi should own
the Netskope publisher records and feed per-publisher registration
tokens to the chart.
