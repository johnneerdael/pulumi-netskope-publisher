import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { NutanixPublisher } from "../src/nutanixPublisher";
import { PublisherOutput } from "../src/types";

const createdVms: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "nutanix:index/virtualMachine:VirtualMachine") {
      createdVms[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          nicListStatuses: [{
            ipEndpointLists: [{ ip: "10.4.0.10" }],
            floatingIp: "198.51.100.24",
          }],
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

test("NutanixPublisher creates VM with cloud-init bootstrap data", async () => {
  const component = new NutanixPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    clusterUuid: "cluster-uuid",
    imageUuid: "image-uuid",
    subnetUuid: "subnet-uuid",
  });

  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);
  const vm = createdVms["publisher-pub-1"];

  assert.equal(vm.clusterUuid, "cluster-uuid");
  assert.equal(vm.diskLists[0].dataSourceReference.uuid, "image-uuid");
  assert.equal(vm.nicLists[0].subnetUuid, "subnet-uuid");
  assert.match(Buffer.from(vm.guestCustomizationCloudInitUserData, "base64").toString("utf8"), /bootstrap\.sh/);
  assert.equal(publishers["pub-1"].privateIp, "10.4.0.10");
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
