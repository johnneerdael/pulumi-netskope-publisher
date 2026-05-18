import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { GcpPublisher } from "../src/gcpPublisher";
import { PublisherOutput } from "../src/types";

const createdInstances: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "gcp:compute/instance:Instance") {
      createdInstances[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          instanceId: `${args.name}-numeric-id`,
          networkInterfaces: [{
            networkIp: "10.2.0.10",
            accessConfigs: [{ natIp: "203.0.113.20" }],
          }],
        },
      };
    }

    if (args.type === "pulumi-nodejs:dynamic:Resource") {
      return {
        id: "pub-1",
        state: {
          ...args.inputs,
          registrations: {
            "pub-1": { publisherId: 101, registrationToken: "token-101", existedBefore: true },
          },
        },
      };
    }

    return { id: `${args.name}-id`, state: args.inputs };
  },
  call(args) {
    return args.inputs;
  },
});

test("GcpPublisher creates outputs keyed by publisher name", async () => {
  const component = new GcpPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    project: "project",
    zone: "europe-west4-a",
    network: "default",
    subnetwork: "default",
    image: "projects/example/global/images/npa",
    assignPublicIp: true,
  });

  const publisherNames = await outputValue(component.publisherNames);
  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);

  assert.deepEqual(publisherNames, ["pub-1"]);
  assert.equal(publishers["pub-1"].publisherId, 101);
  assert.equal(publishers["pub-1"].vmId, "publisher-pub-1-numeric-id");
  assert.match(createdInstances["publisher-pub-1"].metadata["user-data"], /bootstrap\.sh/);
  assert.match(createdInstances["publisher-pub-1"].metadata["user-data"], /resources\/\.nonat/);
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
