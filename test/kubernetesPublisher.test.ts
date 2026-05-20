import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { KubernetesPublisher } from "../src/kubernetesPublisher";
import { KubernetesPublisherOutput } from "../src/types";

const createdResources: Record<string, Record<string, any>> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    createdResources[args.type] ??= {};
    createdResources[args.type][args.name] = args.inputs;

    if (args.type === "pulumi-nodejs:dynamic:Resource") {
      return {
        id: "registration",
        state: {
          ...args.inputs,
          registrations: {
            "pub-1": {
              publisherId: 101,
              registrationToken: "token-101",
              existedBefore: true,
            },
          },
        },
      };
    }

    if (args.type === "kubernetes:helm.sh/v3:Release") {
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          status: { status: "deployed" },
        },
      };
    }

    return { id: `${args.name}-id`, state: args.inputs };
  },
  call(args) {
    return args.inputs;
  },
});

test("KubernetesPublisher token mode creates token secret and release per publisher", async () => {
  const component = new KubernetesPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    namespace: "netskope",
  });

  const publisherNames = await outputValue(component.publisherNames);
  const helmReleaseNames = await outputValue(component.helmReleaseNames);
  const publishers = await outputValue<Record<string, KubernetesPublisherOutput>>(component.publishers);

  assert.deepEqual(publisherNames, ["pub-1"]);
  assert.deepEqual(helmReleaseNames, ["pub-1"]);
  const tokenSecretData = await outputValue(pulumi.output(createdResources["kubernetes:core/v1:Secret"]["publisher-pub-1-registration-token"].stringData));
  assert.equal((tokenSecretData.value ?? tokenSecretData).token, "token-101");
  assert.equal(createdResources["kubernetes:helm.sh/v3:Release"]["publisher-pub-1"].name, "pub-1");
  assert.equal(publishers["pub-1"].publisherId, 101);
  assert.equal(publishers["pub-1"].helmReleaseName, "pub-1");
  assert.equal(publishers["pub-1"].status, "deployed");
});

test("KubernetesPublisher api mode creates shared api token secret and release", async () => {
  const component = new KubernetesPublisher("publisher-api", {
    enrollmentMode: "api",
    namePrefix: "api-pub",
    replicas: 2,
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    namespace: "netskope",
  });

  const publisherNames = await outputValue(component.publisherNames);
  const helmReleaseNames = await outputValue(component.helmReleaseNames);
  const publishers = await outputValue<Record<string, KubernetesPublisherOutput>>(component.publishers);

  assert.deepEqual(publisherNames, ["api-pub-1", "api-pub-2"]);
  assert.deepEqual(helmReleaseNames, ["npa-publisher"]);
  const apiSecretData = await outputValue(pulumi.output(createdResources["kubernetes:core/v1:Secret"]["publisher-api-api-token"].stringData));
  assert.equal((apiSecretData.value ?? apiSecretData)["api-token"], "api-token");
  assert.equal(createdResources["kubernetes:helm.sh/v3:Release"]["publisher-api-npa-publisher"].name, "npa-publisher");
  assert.equal(publishers["npa-publisher"].helmReleaseName, "npa-publisher");
  assert.equal(publishers["npa-publisher"].publisherId, undefined);
});

test("KubernetesPublisher api mode supports OAuth2 auth", async () => {
  const component = new KubernetesPublisher("publisher-api-oauth", {
    enrollmentMode: "api",
    namePrefix: "api-pub",
    tenantUrl: "https://tenant.goskope.com",
    authMode: "oauth2",
    oauth2: {
      tokenUrl: "https://tenant.goskope.com/oauth2/token",
      clientId: "client-id",
      clientSecret: pulumi.secret("client-secret"),
      scope: "npa.publisher",
    },
    namespace: "netskope",
  });

  await outputValue(component.publisherNames);
  await outputValue(component.helmReleaseNames);
  await outputValue(component.publishers);

  const oauthSecret = Object.values(createdResources["kubernetes:core/v1:Secret"])
    .find((resource) => resource.metadata?.name === "npa-api-oauth");
  assert.ok(oauthSecret, `created secrets: ${Object.keys(createdResources["kubernetes:core/v1:Secret"]).join(", ")}`);
  const secretDataOutput = await outputValue(pulumi.output(oauthSecret.stringData));
  const apiSecretData = secretDataOutput.value ?? secretDataOutput;
  assert.equal(apiSecretData["client-id"] ?? apiSecretData.clientId, "client-id", `secret keys: ${Object.keys(apiSecretData).join(", ")}`);
  const clientSecret = apiSecretData["client-secret"] ?? apiSecretData.clientSecret;
  assert.equal((clientSecret.value ?? clientSecret), "client-secret");

  const releaseValuesOutput = await outputValue(pulumi.output(createdResources["kubernetes:helm.sh/v3:Release"]["publisher-api-oauth-npa-publisher"].values));
  const releaseValues = releaseValuesOutput.value ?? releaseValuesOutput;
  assert.equal(releaseValues.enrollment.api.authMode, "oauth2");
  assert.equal(releaseValues.enrollment.api.oauth2.tokenUrl, "https://tenant.goskope.com/oauth2/token");
  assert.equal(releaseValues.enrollment.api.oauth2.existingSecret, "npa-api-oauth");
  assert.equal(releaseValues.enrollment.api.oauth2.scope, "npa.publisher");
});

async function outputValue<T>(output: pulumi.Output<T>): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    output.apply((value) => {
      resolve(value);
      return value;
    });
    setTimeout(() => reject(new Error("Timed out waiting for Pulumi output")), 5000);
  });
}
