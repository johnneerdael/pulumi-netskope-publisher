import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { ScalewayPublisher } from "../src/scalewayPublisher";
import { PublisherOutput } from "../src/types";

const createdServers: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "scaleway:instance/server:Server") {
      createdServers[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          privateIps: [{ address: "10.1.0.10" }],
          publicIps: [{ address: "198.51.100.21" }],
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

test("ScalewayPublisher creates Ubuntu bootstrap instance", async () => {
  const component = new ScalewayPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
  });

  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);
  const server = createdServers["publisher-pub-1"];

  assert.equal(server.image, "ubuntu_jammy");
  assert.equal(server.type, "DEV1-M");
  assert.match(server.cloudInit, /bootstrap\.sh/);
  assert.match(server.userData["cloud-init"], /bootstrap\.sh/);
  assert.equal(publishers["pub-1"].privateIp, "10.1.0.10");
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
