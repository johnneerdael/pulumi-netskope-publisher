import * as pulumi from "@pulumi/pulumi";
import { AwsPublisher } from "@johninnl/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new AwsPublisher("publisher", {
  namePrefix: config.get("namePrefix") ?? "npa-publisher",
  replicas: config.getNumber("replicas") ?? 1,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  subnetId: config.require("subnetId"),
  securityGroupIds: config.requireObject<string[]>("securityGroupIds"),
  keyName: config.get("keyName") ?? undefined,
  amiId: config.get("amiId") ?? undefined,
  associatePublicIpAddress: config.getBoolean("associatePublicIpAddress") ?? false,
  tags: {
    Project: pulumi.getProject(),
    Stack: pulumi.getStack(),
  },
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
