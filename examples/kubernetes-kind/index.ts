import * as pulumi from "@pulumi/pulumi";
import { KubernetesPublisher } from "@johninnl/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publisher = new KubernetesPublisher("publisher", {
  namePrefix: config.get("namePrefix") ?? "npa-publisher",
  replicas: config.getNumber("replicas") ?? 1,
  tenantUrl: config.require("tenantUrl"),
  apiToken: config.requireSecret("apiToken"),
  namespace: config.get("namespace") ?? "netskope",
  enrollmentMode: config.get("enrollmentMode") as "token" | "api" | undefined,
  workloadType: config.get("workloadType") as "daemonset" | "statefulset" | undefined,
  hpaEnabled: config.getBoolean("hpaEnabled") ?? false,
});

export const publisherNames = publisher.publisherNames;
export const helmReleaseNames = publisher.helmReleaseNames;
export const publishers = pulumi.secret(publisher.publishers);
