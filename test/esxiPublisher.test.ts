import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { EsxiPublisher } from "../src/esxiPublisher";
import { PublisherOutput } from "../src/types";

const createdVms: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "esxi-native:index:VirtualMachine") {
      createdVms[args.name] = args.inputs;
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

test("EsxiPublisher creates VM with guestinfo bootstrap data", async () => {
  const component = new EsxiPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    diskStore: "datastore1",
    virtualNetwork: "VM Network",
  });

  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);
  const vm = createdVms["publisher-pub-1"];

  assert.equal(vm.diskStore, "datastore1");
  assert.equal(vm.networkInterfaces[0].virtualNetwork, "VM Network");
  assert.equal(vm.info.find((item: any) => item.key === "guestinfo.userdata.encoding").value, "base64");
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
