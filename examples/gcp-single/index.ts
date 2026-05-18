import * as pulumi from "@pulumi/pulumi";
import { GcpPublisher } from "@johninnl/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new GcpPublisher("publisher", {
  namePrefix: config.get("namePrefix") ?? "npa-publisher",
  replicas: config.getNumber("replicas") ?? 1,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  project: config.require("project"),
  zone: config.require("zone"),
  network: config.require("network"),
  subnetwork: config.require("subnetwork"),
  image: config.require("image"),
  assignPublicIp: config.getBoolean("assignPublicIp") ?? false,
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
