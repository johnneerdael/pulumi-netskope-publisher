# pulumi-netskope-publisher

Pulumi components for provisioning Netskope Private Access Publishers.

The first version is AWS-first. It mirrors the Terraform AWS module from
`terraform-netskope-publisher`: register or reuse Netskope publisher
records, generate per-publisher cloud-init, and create EC2 instances.

## Current scope

- AWS publisher component: `AwsPublisher`
- Netskope publisher registration by name
- Bring-your-own registration data escape hatch
- GitHub Pages documentation

Azure, GCP, vSphere, and Hyper-V are planned follow-up providers.

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

See `examples/aws-single` for a Pulumi program that deploys one or more
AWS publishers.

## Documentation

Full guides are published with GitHub Pages:

https://johnneerdael.github.io/pulumi-netskope-publisher/
