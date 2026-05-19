import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { OciPublisher } from "../src/ociPublisher";
import { PublisherOutput } from "../src/types";

const createdInstances: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "oci:Core/instance:Instance") {
      createdInstances[args.name] = args.inputs;
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          privateIp: "10.2.0.10",
          publicIp: "198.51.100.22",
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

test("OciPublisher creates Ubuntu bootstrap instance", async () => {
  const component = new OciPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    compartmentId: "ocid1.compartment.oc1..example",
    availabilityDomain: "AD-1",
    subnetId: "ocid1.subnet.oc1..example",
    imageId: "ocid1.image.oc1..ubuntu",
  });

  await outputValue<Record<string, PublisherOutput>>(component.publishers);
  const instance = createdInstances["publisher-pub-1"];
  assert.equal(instance.sourceDetails.sourceId, "ocid1.image.oc1..ubuntu");
  assert.match(Buffer.from(instance.metadata.userData, "base64").toString("utf8"), /bootstrap\.sh/);
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
