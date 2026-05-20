import * as pulumi from "@pulumi/pulumi";
import {
  AwsPublisher,
  PrivateApp,
  RealtimeProtectionPolicy,
  TagPublisherAssignment,
} from "@johninnl/pulumi-netskope-publisher";

const config = new pulumi.Config();

const publishers = new AwsPublisher("vpc-a-publishers", {
  names: ["vpc-a-pub-1"],
  placementLabels: ["vpc-a"],
  tenantUrl: config.require("tenantUrl"),
  bearerToken: config.requireSecret("bearerToken"),
  subnetId: config.require("subnetId"),
  securityGroupIds: config.requireObject<string[]>("securityGroupIds"),
  amiId: config.require("amiId"),
});

const app = new PrivateApp("orders", {
  tenantUrl: config.require("tenantUrl"),
  bearerToken: config.requireSecret("bearerToken"),
  appName: "orders",
  appType: "client",
  host: "orders.internal",
  protocols: [{ type: "tcp", ports: "443" }],
  clientlessAccess: false,
  isUserPortalApp: false,
  usePublisherDns: false,
  trustSelfSignedCerts: false,
  tags: ["vpc-a"],
});

const assignment = new TagPublisherAssignment("vpc-a-access", {
  tenantUrl: config.require("tenantUrl"),
  bearerToken: config.requireSecret("bearerToken"),
  appTags: ["vpc-a"],
  publisherPlacementLabels: ["vpc-a"],
  publishers: publishers.publishers,
});

const policy = new RealtimeProtectionPolicy("orders-access", {
  tenantUrl: config.require("tenantUrl"),
  bearerToken: config.requireSecret("bearerToken"),
  name: "orders-access",
  policyGroupName: config.require("policyGroupName"),
  appTags: ["vpc-a"],
  users: config.requireObject<string[]>("users"),
  action: "allow",
  enabled: true,
});

export const appId = app.appId;
export const matchedApps = assignment.matchedApps;
export const policyId = policy.policyId;
