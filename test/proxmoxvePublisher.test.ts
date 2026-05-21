import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { ProxmoxvePublisher } from "../src/proxmoxvePublisher";
import { PublisherOutput } from "../src/types";

const createdFiles: Record<string, any> = {};
const createdVms: Record<string, any> = {};

pulumi.runtime.setMocks({
  newResource(args) {
    if (args.type === "proxmoxve:index/fileLegacy:FileLegacy") {
      createdFiles[args.name] = args.inputs;
      return {
        id: `${args.inputs.nodeName}/${args.inputs.datastoreId}:snippets/${args.inputs.sourceRaw.fileName}`,
        state: args.inputs,
      };
    }

    if (args.type === "proxmoxve:index/vmLegacy:VmLegacy") {
      createdVms[args.name] = args.inputs;
      return {
        id: `${args.inputs.nodeName}/${args.inputs.vmId ?? 100}`,
        state: {
          ...args.inputs,
          ipv4Addresses: [["10.10.0.50"]],
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

test("ProxmoxvePublisher creates cloud-init snippet and cloned VM", async () => {
  const component = new ProxmoxvePublisher("publisher", {
    names: ["pub-1"],
    tenantUrl: "https://tenant.goskope.com",
    apiToken: pulumi.secret("api-token"),
    nodeName: "pve-1",
    datastoreId: "local",
    templateVmId: 9000,
    vmId: 101,
    networkBridge: "vmbr1",
    ipAddress: "10.10.0.50/24",
    gateway: "10.10.0.1",
    tags: {
      role: "netskope-publisher",
    },
  });

  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);
  const file = createdFiles["publisher-pub-1-user-data"];
  const vm = createdVms["publisher-pub-1"];

  assert.equal(file.contentType, "snippets");
  assert.equal(file.datastoreId, "local");
  assert.equal(file.sourceRaw.fileName, "pub-1-user-data.yaml");
  assert.match(file.sourceRaw.data, /bootstrap\.sh/);
  assert.equal(vm.clone.vmId, 9000);
  assert.equal(vm.clone.datastoreId, "local");
  assert.equal(vm.initialization.userDataFileId, "pve-1/local:snippets/pub-1-user-data.yaml");
  assert.equal(vm.initialization.ipConfigs[0].ipv4.address, "10.10.0.50/24");
  assert.equal(vm.networkDevices[0].bridge, "vmbr1");
  assert.deepEqual(vm.tags, ["role=netskope-publisher"]);
  assert.equal(publishers["pub-1"].privateIp, "10.10.0.50");
});

test("ProxmoxvePublisher rejects missing catalog-required templateVmId", () => {
  assert.throws(
    () => new ProxmoxvePublisher("missing-template", {
      names: ["pub-1"],
      tenantUrl: "https://tenant.goskope.com",
      apiToken: pulumi.secret("api-token"),
      nodeName: "pve-1",
      datastoreId: "local",
    } as any),
    /ProxmoxvePublisher requires input templateVmId/,
  );
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
