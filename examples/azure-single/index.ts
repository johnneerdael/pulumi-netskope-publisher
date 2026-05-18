import * as pulumi from "@pulumi/pulumi";
import { AzurePublisher } from "@johninnl/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new AzurePublisher("publisher", {
  namePrefix: config.get("namePrefix") ?? "npa-publisher",
  replicas: config.getNumber("replicas") ?? 1,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  resourceGroupName: config.require("resourceGroupName"),
  location: config.require("location"),
  subnetId: config.require("subnetId"),
  adminSshPublicKey: config.require("adminSshPublicKey"),
  imageId: config.get("imageId") ?? undefined,
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
