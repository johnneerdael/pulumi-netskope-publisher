import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { AwsPublisher } from "../src/awsPublisher";
import { PublisherOutput } from "../src/types";

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "aws:ec2/instance:Instance") {
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          id: `${args.name}-id`,
          privateIp: "10.0.0.10",
          publicIp: "198.51.100.10",
        },
      };
    }

    if (args.type === "pulumi-nodejs:dynamic:Resource") {
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

    return { id: `${args.name}-id`, state: args.inputs };
  },
  call(args) {
    if (args.token === "aws:ec2/getAmi:getAmi") {
      return { id: "ami-123" };
    }

    return args.inputs;
  },
});

test("AwsPublisher creates outputs keyed by publisher name", async () => {
  const component = new AwsPublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    subnetId: "subnet-123",
    securityGroupIds: ["sg-123"],
  });

  const publisherNames = await outputValue(component.publisherNames);
  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);

  assert.deepEqual(publisherNames, ["pub-1"]);
  assert.equal(publishers["pub-1"].publisherId, 101);
  assert.equal(publishers["pub-1"].instanceId, "publisher-pub-1-id");
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
