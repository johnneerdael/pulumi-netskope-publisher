import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { DigitaloceanPublisher } from "../src/digitaloceanPublisher";
import { EquinixPublisher } from "../src/equinixPublisher";
import { ExoscalePublisher } from "../src/exoscalePublisher";
import { OpentelekomcloudPublisher } from "../src/opentelekomcloudPublisher";
import { OutscalePublisher } from "../src/outscalePublisher";
import { StackitPublisher } from "../src/stackitPublisher";
import { TencentcloudPublisher } from "../src/tencentcloudPublisher";
import { UpcloudPublisher } from "../src/upcloudPublisher";
import { VultrPublisher } from "../src/vultrPublisher";
import { YandexPublisher } from "../src/yandexPublisher";
import { PublisherOutput } from "../src/types";

const createdResources: Record<string, Record<string, any>> = {};

const mockedResourceOutputs: Record<string, Record<string, any>> = {
  "digitalocean:index/droplet:Droplet": {
    ipv4Address: "203.0.113.51",
    ipv4AddressPrivate: "10.0.0.51",
  },
  "vultr:index/instance:Instance": {
    mainIp: "203.0.113.52",
    internalIp: "10.0.0.52",
  },
  "exoscale:index/computeInstance:ComputeInstance": {
    publicIpAddress: "203.0.113.53",
  },
  "equinix:metal/device:Device": {
    accessPublicIpv4: "203.0.113.54",
    accessPrivateIpv4: "10.0.0.54",
  },
  "outscale:index/vm:Vm": {
    publicIp: "203.0.113.55",
    privateIp: "10.0.0.55",
  },
  "tencentcloud:index/instance:Instance": {
    publicIp: "203.0.113.56",
    privateIp: "10.0.0.56",
  },
  "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2": {
    accessIpV4: "203.0.113.57",
  },
  "ovh:CloudProject/instance:Instance": {
    addresses: [{ ip: "203.0.113.58", version: 4 }],
  },
};

pulumi.runtime.setMocks({
  newResource(args) {
    createdResources[args.type] ??= {};
    createdResources[args.type][args.name] = args.inputs;
    return {
      id: `${args.name}-id`,
      state: {
        ...args.inputs,
        ...(mockedResourceOutputs[args.type] ?? {}),
      },
    };
  },
  call(args) {
    return args.inputs;
  },
});

test("DigitaloceanPublisher creates Ubuntu droplet with plain userData", async () => {
  const component = new DigitaloceanPublisher("digitalocean", baseArgs({
    region: "ams3",
  }));

  const publishers = await outputValue<Record<string, PublisherOutput>>(component.publishers);
  const droplet = createdResources["digitalocean:index/droplet:Droplet"]["digitalocean-pub-1"];

  assert.equal(droplet.name, "pub-1");
  assert.equal(droplet.image, "ubuntu-22-04-x64");
  assert.equal(droplet.size, "s-2vcpu-4gb");
  assert.match(await inputValue<string>(droplet.userData), /bootstrap\.sh/);
  assert.equal(publishers["pub-1"].vmId, "digitalocean-pub-1-id");
});

test("VultrPublisher creates instance with plain userData", async () => {
  const component = new VultrPublisher("vultr", baseArgs({
    region: "ams",
    plan: "vc2-2c-4gb",
    osId: 1743,
  }));

  await outputValue(component.publishers);
  const instance = createdResources["vultr:index/instance:Instance"]["vultr-pub-1"];

  assert.equal(instance.osId, 1743);
  assert.match(await inputValue<string>(instance.userData), /bootstrap\.sh/);
});

test("ExoscalePublisher creates compute instance with plain userData", async () => {
  const component = new ExoscalePublisher("exoscale", baseArgs({
    zone: "ch-gva-2",
    type: "standard.medium",
    templateId: "template-id",
    diskSize: 50,
  }));

  await outputValue(component.publishers);
  const instance = createdResources["exoscale:index/computeInstance:ComputeInstance"]["exoscale-pub-1"];

  assert.equal(instance.templateId, "template-id");
  assert.match(await inputValue<string>(instance.userData), /bootstrap\.sh/);
});

test("UpcloudPublisher creates server with plain userData", async () => {
  const component = new UpcloudPublisher("upcloud", baseArgs({
    zone: "nl-ams1",
  }));

  await outputValue(component.publishers);
  const server = createdResources["upcloud:index/server:Server"]["upcloud-pub-1"];

  assert.equal(server.template, "01000000-0000-4000-8000-000030220200");
  assert.match(await inputValue<string>(server.userData), /bootstrap\.sh/);
});

test("StackitPublisher creates server with plain userData", async () => {
  const component = new StackitPublisher("stackit", baseArgs({
    projectId: "project-id",
    machineType: "g1.2",
    imageId: "image-id",
  }));

  await outputValue(component.publishers);
  const server = createdResources["stackit:index/server:Server"]["stackit-pub-1"];

  assert.equal(server.imageId, "image-id");
  assert.match(await inputValue<string>(server.userData), /bootstrap\.sh/);
});

test("EquinixPublisher creates metal device with plain userData", async () => {
  const component = new EquinixPublisher("equinix", baseArgs({
    projectId: "project-id",
    metro: "AM",
    plan: "c3.small.x86",
  }));

  await outputValue(component.publishers);
  const device = createdResources["equinix:metal/device:Device"]["equinix-pub-1"];

  assert.equal(device.operatingSystem, "ubuntu_22_04");
  assert.match(await inputValue<string>(device.userData), /bootstrap\.sh/);
});

test("OutscalePublisher creates VM with plain userData", async () => {
  const component = new OutscalePublisher("outscale", baseArgs({
    imageId: "ami-123",
  }));

  await outputValue(component.publishers);
  const vm = createdResources["outscale:index/vm:Vm"]["outscale-pub-1"];

  assert.equal(vm.imageId, "ami-123");
  assert.match(await inputValue<string>(vm.userData), /bootstrap\.sh/);
});

test("OpentelekomcloudPublisher creates compute instance with plain userData", async () => {
  const component = new OpentelekomcloudPublisher("otc", baseArgs({
    networks: [{ name: "private" }],
  }));

  await outputValue(component.publishers);
  const instance = createdResources["opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2"]["otc-pub-1"];

  assert.equal(instance.imageName, "Ubuntu 22.04");
  assert.match(await inputValue<string>(instance.userData), /bootstrap\.sh/);
});

test("OpentelekomcloudPublisher uses imageId and flavorId without default name selectors", async () => {
  const component = new OpentelekomcloudPublisher("otc-id-selectors", baseArgs({
    networks: [{ name: "private" }],
    imageId: "image-id",
    flavorId: "flavor-id",
  }));

  await outputValue(component.publishers);
  const instance = createdResources["opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2"]["otc-id-selectors-pub-1"];

  assert.equal(instance.imageId, "image-id");
  assert.equal(instance.flavorId, "flavor-id");
  assert.equal(instance.imageName, undefined);
  assert.equal(instance.flavorName, undefined);
});

test("OpentelekomcloudPublisher rejects conflicting image and flavor selectors", () => {
  assert.throws(
    () => new OpentelekomcloudPublisher("otc-conflicting-image", baseArgs({
      networks: [{ name: "private" }],
      imageName: "Ubuntu 22.04",
      imageId: "image-id",
    })),
    /OpentelekomcloudPublisher accepts only one of: imageName, imageId/,
  );

  assert.throws(
    () => new OpentelekomcloudPublisher("otc-conflicting-flavor", baseArgs({
      networks: [{ name: "private" }],
      flavorName: "s3.medium.2",
      flavorId: "flavor-id",
    })),
    /OpentelekomcloudPublisher accepts only one of: flavorName, flavorId/,
  );
});

test("TencentcloudPublisher creates instance with raw userData", async () => {
  const component = new TencentcloudPublisher("tencent", baseArgs({
    availabilityZone: "ap-guangzhou-6",
    imageId: "img-123",
  }));

  await outputValue(component.publishers);
  const instance = createdResources["tencentcloud:index/instance:Instance"]["tencent-pub-1"];

  assert.equal(instance.imageId, "img-123");
  assert.match(await inputValue<string>(instance.userDataRaw), /bootstrap\.sh/);
});

test("YandexPublisher creates compute instance with metadata user-data", async () => {
  const component = new YandexPublisher("yandex", baseArgs({
    imageId: "image-id",
    subnetId: "subnet-id",
  }));

  await outputValue(component.publishers);
  const instance = createdResources["yandex:index/computeInstance:ComputeInstance"]["yandex-pub-1"];
  const metadata = await inputValue<Record<string, string>>(instance.metadata);

  assert.equal(instance.bootDisk.initializeParams.imageId, "image-id");
  assert.match(metadata["user-data"], /bootstrap\.sh/);
});

test("catalog raw VM publishers expose registry-backed IP outputs", async () => {
  const cases: Array<{
    name: string;
    component: pulumi.ComponentResource & { publishers: pulumi.Output<Record<string, PublisherOutput>> };
    expectedPrivateIp: string;
    expectedPublicIp: string | undefined;
  }> = [{
    name: "DigitalOcean",
    component: new DigitaloceanPublisher("digitalocean-outputs", baseArgs({ region: "ams3" })),
    expectedPrivateIp: "10.0.0.51",
    expectedPublicIp: "203.0.113.51",
  }, {
    name: "Vultr",
    component: new VultrPublisher("vultr-outputs", baseArgs({ region: "ams", plan: "vc2-2c-4gb", osId: 1743 })),
    expectedPrivateIp: "10.0.0.52",
    expectedPublicIp: "203.0.113.52",
  }, {
    name: "Exoscale",
    component: new ExoscalePublisher("exoscale-outputs", baseArgs({ zone: "ch-gva-2", type: "standard.medium", templateId: "template-id", diskSize: 50 })),
    expectedPrivateIp: "",
    expectedPublicIp: "203.0.113.53",
  }, {
    name: "Equinix",
    component: new EquinixPublisher("equinix-outputs", baseArgs({ projectId: "project-id", metro: "AM", plan: "c3.small.x86" })),
    expectedPrivateIp: "10.0.0.54",
    expectedPublicIp: "203.0.113.54",
  }, {
    name: "Outscale",
    component: new OutscalePublisher("outscale-outputs", baseArgs({ imageId: "ami-123" })),
    expectedPrivateIp: "10.0.0.55",
    expectedPublicIp: "203.0.113.55",
  }, {
    name: "TencentCloud",
    component: new TencentcloudPublisher("tencent-outputs", baseArgs({ availabilityZone: "ap-guangzhou-6", imageId: "img-123" })),
    expectedPrivateIp: "10.0.0.56",
    expectedPublicIp: "203.0.113.56",
  }, {
    name: "OpenTelekomCloud",
    component: new OpentelekomcloudPublisher("otc-outputs", baseArgs({ networks: [{ name: "private" }] })),
    expectedPrivateIp: "",
    expectedPublicIp: "203.0.113.57",
  }];

  for (const tc of cases) {
    const publishers = await outputValue<Record<string, PublisherOutput>>(tc.component.publishers);
    assert.equal(publishers["pub-1"].privateIp, tc.expectedPrivateIp, `${tc.name} privateIp mismatch`);
    assert.equal(publishers["pub-1"].publicIp, tc.expectedPublicIp, `${tc.name} publicIp mismatch`);
  }
});

function baseArgs<T extends object>(args: T): T & {
  names: string[];
  registrations: Record<string, { publisherId: number; registrationToken: string }>;
} {
  return {
    names: ["pub-1"],
    registrations: {
      "pub-1": { publisherId: 101, registrationToken: "token-101" },
    },
    ...args,
  };
}

async function inputValue<T>(input: pulumi.Input<T>): Promise<T> {
  if (pulumi.Output.isInstance(input)) {
    return outputValue(input as pulumi.Output<T>);
  }
  return input as T;
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
