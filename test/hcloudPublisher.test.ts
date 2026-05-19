import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { HcloudPublisher } from "../src/hcloudPublisher";
import { PublisherOutput } from "../src/types";

const createdServers: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "hcloud:index/server:Server") {
      createdServers[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          ipv4Address: "198.51.100.20",
          networks: [{ ip: "10.0.0.20" }],
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

test("HcloudPublisher creates Ubuntu bootstrap server", async () => {
  const component = new HcloudPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    networkId: 123,
  });

  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);
  const server = createdServers["publisher-pub-1"];

  assert.equal(server.image, "ubuntu-22.04");
  assert.equal(server.serverType, "cx22");
  assert.equal(server.networks[0].networkId, 123);
  assert.match(server.userData, /bootstrap\.sh/);
  assert.equal(publishers["pub-1"].vmId, "publisher-pub-1-id");
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
