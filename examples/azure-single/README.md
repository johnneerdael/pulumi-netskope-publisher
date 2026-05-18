# Azure Single Example

Deploy one or more Netskope Private Access Publishers on Azure.

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set resourceGroupName rg-npa
pulumi config set location westeurope
pulumi config set subnetId /subscriptions/.../subnets/default
pulumi config set adminSshPublicKey "ssh-rsa AAAA..."
pulumi config set imageId /subscriptions/.../providers/Microsoft.Compute/images/npa
npm install
pulumi preview
pulumi up
```
