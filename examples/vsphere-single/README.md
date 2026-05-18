# vSphere Single Example

Deploy one or more Netskope Private Access Publishers on vSphere.

Prepare a VM template from the official Netskope OVA:

```text
https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.ova
```

Then configure and deploy:

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set datacenter dc-01
pulumi config set cluster cluster-01
pulumi config set datastore datastore-01
pulumi config set networkName "VM Network"
pulumi config set templateName npa-publisher-template
npm install
pulumi preview
pulumi up
```
