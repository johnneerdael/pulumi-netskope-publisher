# GCP Single Example

Deploy one or more Netskope Private Access Publishers on GCP.

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set project my-project
pulumi config set zone europe-west4-a
pulumi config set network default
pulumi config set subnetwork default
pulumi config set image projects/my-project/global/images/npa
npm install
pulumi preview
pulumi up
```
