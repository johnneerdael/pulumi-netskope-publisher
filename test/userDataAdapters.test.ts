import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { providerCatalog } from "../src/providerCatalog";
import {
  base64UserData,
  guestInfoUserData,
  metadataUserData,
  plainUserData,
  scalewayUserData,
  userDataAdapters,
} from "../src/userDataAdapters";

async function outputValue<T>(value: pulumi.Output<T>): Promise<T> {
  return await new Promise<T>((resolve) => value.apply((resolved) => {
    resolve(resolved);
    return resolved;
  }));
}

test("plainUserData returns the payload unchanged", async () => {
  const result = plainUserData(pulumi.output("#cloud-config"));
  assert.equal(await outputValue(result), "#cloud-config");
});

test("base64UserData encodes the payload", async () => {
  const result = base64UserData(pulumi.output("#cloud-config"));
  assert.equal(await outputValue(result), Buffer.from("#cloud-config", "utf8").toString("base64"));
});

test("metadataUserData places payload under the requested key", async () => {
  const result = metadataUserData(pulumi.output("#cloud-config"), "user-data");
  assert.equal(await outputValue(result["user-data"] as pulumi.Output<string>), "#cloud-config");
});

test("guestInfoUserData emits base64 guestinfo keys", async () => {
  const result = guestInfoUserData(pulumi.output("#cloud-config"));
  assert.equal(await outputValue(result["guestinfo.userdata"] as pulumi.Output<string>), Buffer.from("#cloud-config", "utf8").toString("base64"));
  assert.equal(result["guestinfo.userdata.encoding"], "base64");
});

test("scalewayUserData emits both cloudInit and userData cloud-init keys", async () => {
  const result = scalewayUserData(pulumi.output("#cloud-config"));
  assert.equal(await outputValue(result.cloudInit as pulumi.Output<string>), "#cloud-config");
  const map = result.userData as Record<string, pulumi.Input<string>>;
  assert.equal(await outputValue(map["cloud-init"] as pulumi.Output<string>), "#cloud-config");
});

test("userDataAdapters cover raw VM factory provider modes", () => {
  for (const componentName of [
    "DigitaloceanPublisher",
    "VultrPublisher",
    "ExoscalePublisher",
    "UpcloudPublisher",
    "StackitPublisher",
    "EquinixPublisher",
    "OutscalePublisher",
    "OpentelekomcloudPublisher",
    "TencentcloudPublisher",
    "YandexPublisher",
  ]) {
    const mode = providerCatalog[componentName].userData.mode;
    assert.ok(userDataAdapters[mode], `${componentName} mode ${mode} has no adapter`);
  }
});
