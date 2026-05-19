import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { OpenstackPublisher } from "../src/openstackPublisher";
import { PublisherOutput } from "../src/types";

const createdInstances: Record<string, any> = {};
const createdFloatingIps: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "openstack:compute/instance:Instance") {
      createdInstances[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          accessIpV4: "198.51.100.25",
          networks: [{ fixedIpV4: "10.5.0.10", port: "port-123" }],
        },
      };
    }

    if (args.type === "openstack:networking/floatingIp:FloatingIp") {
      createdFloatingIps[args.name] = args.inputs;
      return { id: `${args.name}-id`, state: { ...args.inputs, address: "203.0.113.25" } };
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

test("OpenstackPublisher creates Ubuntu bootstrap instance", async () => {
  const component = new OpenstackPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    imageName: "Ubuntu 22.04",
    flavorName: "m1.medium",
    networkName: "private",
    assignFloatingIp: true,
    floatingIpPool: "public",
  });

  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);
  const instance = createdInstances["publisher-pub-1"];

  assert.equal(instance.imageName, "Ubuntu 22.04");
  assert.equal(instance.flavorName, "m1.medium");
  assert.match(instance.userData, /bootstrap\.sh/);
  assert.equal(createdFloatingIps["publisher-pub-1-fip"].pool, "public");
  assert.equal(publishers["pub-1"].publicIp, "203.0.113.25");
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
