import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { VspherePublisher } from "../src/vspherePublisher";
import { PublisherOutput } from "../src/types";

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "vsphere:index/virtualMachine:VirtualMachine") {
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          defaultIpAddress: "10.3.0.10",
          uuid: `${args.name}-uuid`,
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
    switch (args.token) {
      case "vsphere:index/getDatacenter:getDatacenter":
        return { id: "dc-1", name: args.inputs.name };
      case "vsphere:index/getDatastore:getDatastore":
        return { id: "ds-1", name: args.inputs.name };
      case "vsphere:index/getNetwork:getNetwork":
        return { id: "net-1", name: args.inputs.name };
      case "vsphere:index/getVirtualMachine:getVirtualMachine":
        return {
          id: "template-1",
          name: args.inputs.name,
          guestId: "ubuntu64Guest",
          networkInterfaceTypes: ["vmxnet3"],
          disks: [{ size: 64, eagerlyScrub: false, thinProvisioned: true }],
        };
      case "vsphere:index/getComputeCluster:getComputeCluster":
        return { id: "cluster-1", name: args.inputs.name, resourcePoolId: "pool-1" };
      case "vsphere:index/getHost:getHost":
        return { id: "host-1", name: args.inputs.name, resourcePoolId: "pool-host-1" };
      default:
        return args.inputs;
    }
  },
});

test("VspherePublisher creates outputs keyed by publisher name", async () => {
  const component = new VspherePublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    datacenter: "dc-01",
    cluster: "cluster-01",
    datastore: "datastore-01",
    networkName: "VM Network",
    templateName: "npa-template",
  });

  const publisherNames = await outputValue(component.publisherNames);
  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);

  assert.deepEqual(publisherNames, ["pub-1"]);
  assert.equal(publishers["pub-1"].publisherId, 101);
  assert.equal(publishers["pub-1"].vmId, "publisher-pub-1-id");
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
