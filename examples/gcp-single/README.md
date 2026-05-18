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
pulumi config set image projects/ubuntu-os-cloud/global/images/family/ubuntu-2204-lts
npm install
pulumi preview
pulumi up
```

GCP does not provide a public Netskope Publisher image. This example
boots Ubuntu 22.04 and uses cloud-init to run Netskope's generic
publisher bootstrap script before registering the publisher.
