import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { AlicloudPublisher } from "../src/alicloudPublisher";
import { PublisherOutput } from "../src/types";

const createdInstances: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "alicloud:ecs/instance:Instance") {
      createdInstances[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          primaryIpAddress: "10.3.0.10",
          publicIp: "198.51.100.23",
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

test("AlicloudPublisher creates Ubuntu bootstrap instance", async () => {
  const component = new AlicloudPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    imageId: "ubuntu_22_04_x64_20G_alibase_20240130.vhd",
    vswitchId: "vsw-123",
    securityGroupIds: ["sg-123"],
    allocatePublicIp: true,
  });

  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);
  const instance = createdInstances["publisher-pub-1"];

  assert.equal(instance.imageId, "ubuntu_22_04_x64_20G_alibase_20240130.vhd");
  assert.match(Buffer.from(instance.userData, "base64").toString("utf8"), /bootstrap\.sh/);
  assert.equal(instance.internetMaxBandwidthOut, 10);
  assert.equal(publishers["pub-1"].privateIp, "10.3.0.10");
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
