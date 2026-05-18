---
title: Installation & Configuration
meta_desc: Install and configure the Netskope Publisher Pulumi package.
layout: package
---

Install the package from npm:

```bash
npm install @johninnl/pulumi-netskope-publisher
```

Install the cloud provider packages used by the component you deploy:

```bash
npm install @pulumi/aws @pulumi/azure-native @pulumi/gcp @pulumi/vsphere
```

## Netskope configuration

For automatic publisher registration, configure the Netskope tenant URL
and API token as Pulumi stack configuration:

```bash
pulumi config set tenantUrl https://example.goskope.com
pulumi config set --secret apiToken ns-api-token
```

The token must be allowed to create or look up publisher registration
records in the tenant.

To avoid automatic registration, pass `registrations` to the component.
Each entry is keyed by the publisher name and must include
`publisherId` and `registrationToken`.

## Provider configuration

Configure the cloud provider used by the selected component with the
standard Pulumi provider configuration for AWS, Azure Native, Google
Cloud, or vSphere.

Each component also requires provider-specific network and image inputs.
See the component API docs and the examples directory for complete
programs.

## Publisher images

The package exports `officialImageSources` with the official Netskope
VHDX and OVA download URLs:

```typescript
import { officialImageSources } from "@johninnl/pulumi-netskope-publisher";

export const ovaUrl = officialImageSources.ova;
export const vhdxUrl = officialImageSources.vhdx;
```
