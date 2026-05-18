import * as pulumi from "@pulumi/pulumi";
import { VspherePublisher } from "@johninnl/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new VspherePublisher("publisher", {
  namePrefix: config.get("namePrefix") ?? "npa-publisher",
  replicas: config.getNumber("replicas") ?? 1,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  datacenter: config.require("datacenter"),
  cluster: config.get("cluster") ?? undefined,
  host: config.get("host") ?? undefined,
  datastore: config.require("datastore"),
  networkName: config.require("networkName"),
  templateName: config.require("templateName"),
});

export const publisherNames = publisher.publisherNames;
export const publishers = pulumi.secret(publisher.publishers);
