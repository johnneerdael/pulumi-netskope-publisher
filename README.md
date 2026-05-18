# pulumi-netskope-publisher

Pulumi components for provisioning Netskope Private Access Publishers.

The package mirrors the Terraform modules from
`terraform-netskope-publisher`: register or reuse Netskope publisher
records, generate per-publisher cloud-init, and create platform VMs.

## Current scope

- AWS publisher component: `AwsPublisher`
- Azure publisher component: `AzurePublisher`
- GCP publisher component: `GcpPublisher`
- vSphere publisher component: `VspherePublisher`
- Experimental Hyper-V gate: `HypervPublisher`
- Netskope publisher registration by name
- Bring-your-own registration data escape hatch
- GitHub Pages documentation

Hyper-V depends on the upstream Pulumi Hyper-V provider becoming
consumable from a stable package source. The official Netskope publisher
VHDX and OVA download URLs are exported from `officialImageSources`.

## Development

```bash
npm install
npm run typecheck
npm test
```

## Quick start

```ts
import * as pulumi from "@pulumi/pulumi";
import { AwsPublisher } from "@johnneerdael/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new AwsPublisher("publisher", {
  namePrefix: "pub-eu",
  replicas: 2,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  subnetId: config.require("subnetId"),
  securityGroupIds: config.requireObject<string[]>("securityGroupIds"),
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
```

## Examples

See `examples/aws-single`, `examples/azure-single`,
`examples/gcp-single`, and `examples/vsphere-single` for Pulumi programs
that deploy one or more publishers.

## Documentation

Full guides are published with GitHub Pages:

https://johnneerdael.github.io/pulumi-netskope-publisher/
