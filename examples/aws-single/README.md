# AWS Single Example

This example deploys one or more Netskope Private Access Publishers on
AWS using the local Pulumi component package.

## Configure

```bash
pulumi stack init dev
pulumi config set tenantUrl https://tenant.goskope.com
pulumi config set apiToken --secret
pulumi config set subnetId subnet-1234567890abcdef0
pulumi config set securityGroupIds '["sg-1234567890abcdef0"]'
pulumi config set replicas 1
```

Optional:

```bash
pulumi config set keyName my-key
pulumi config set amiId ami-1234567890abcdef0
pulumi config set associatePublicIpAddress false
```

## Deploy

```bash
npm install
pulumi preview
pulumi up
```

## Destroy

```bash
pulumi destroy
```
