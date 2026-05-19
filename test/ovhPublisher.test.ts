import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { OvhPublisher } from "../src/ovhPublisher";
import { PublisherOutput } from "../src/types";

const createdInstances: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "ovh:CloudProject/instance:Instance") {
      createdInstances[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
        },
      };
    }

    if (args.type === "pulumi-nodejs:dynamic:Resource") {
      return registrationMock(args);
    }

    return { id: `${args.name}-id`, state: args.inputs };
  },
  call(args) {
    return args.inputs;
  },
});

test("OvhPublisher creates public cloud instance with bootstrap data", async () => {
  const component = new OvhPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    serviceName: "project-id",
    region: "GRA11",
    imageId: "ubuntu-image-id",
    flavorId: "flavor-id",
    sshKeyName: "publisher-key",
  });

  await outputValue<Record<string, PublisherOutput>>(component.publishers);
  const instance = createdInstances["publisher-pub-1"];
  assert.equal(instance.serviceName, "project-id");
  assert.equal(instance.bootFrom.imageId, "ubuntu-image-id");
  assert.equal(instance.flavor.flavorId, "flavor-id");
  assert.match(instance.userData, /bootstrap\.sh/);
});

function registrationMock(args: pulumi.runtime.MockResourceArgs) {
  return {
    id: "pub-1",
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

async function outputValue<T>(output: pulumi.Output<T>): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    output.apply((value) => {
      resolve(value);
      return value;
    });
    setTimeout(() => reject(new Error("Timed out waiting for Pulumi output")), 5000);
  });
}
