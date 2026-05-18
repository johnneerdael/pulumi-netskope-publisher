import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { AzurePublisher } from "../src/azurePublisher";
import { PublisherOutput } from "../src/types";

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "azure-native:network:NetworkInterface") {
      return { id: `${args.name}-id`, state: { ...args.inputs, ipConfigurations: [{ privateIPAddress: "10.1.0.10" }] } };
    }

    if (args.type === "azure-native:network:PublicIPAddress") {
      return { id: `${args.name}-id`, state: { ...args.inputs, ipAddress: "203.0.113.10" } };
    }

    if (args.type === "azure-native:compute:VirtualMachine") {
      return { id: `${args.name}-id`, state: args.inputs };
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

test("AzurePublisher creates outputs keyed by publisher name", async () => {
  const component = new AzurePublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    resourceGroupName: "rg",
    location: "westeurope",
    subnetId: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/default",
    adminSshPublicKey: "ssh-rsa AAAA",
    imageId: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/images/npa",
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
