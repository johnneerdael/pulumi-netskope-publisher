import assert from "node:assert/strict";
import test from "node:test";
import { validateProviderArgs } from "../src/providerValidation";

test("validateProviderArgs rejects missing required fields", () => {
  assert.throws(
    () => validateProviderArgs("DigitaloceanPublisher", { replicas: 1 }),
    /DigitaloceanPublisher requires input region/,
  );
});

test("validateProviderArgs accepts present required fields", () => {
  assert.doesNotThrow(() => validateProviderArgs("DigitaloceanPublisher", { region: "ams3" }));
});

test("validateProviderArgs rejects missing required-one-of fields", () => {
  assert.throws(
    () => validateProviderArgs("VultrPublisher", { region: "ams", plan: "vc2-2c-4gb" }),
    /VultrPublisher requires one of: osId, imageId/,
  );
});

test("validateProviderArgs rejects mutually exclusive fields", () => {
  assert.throws(
    () => validateProviderArgs("VultrPublisher", { region: "ams", plan: "vc2-2c-4gb", osId: 1743, imageId: "img-123" }),
    /VultrPublisher accepts only one of: osId, imageId/,
  );
});

test("validateProviderArgs enforces experimental opt-in", () => {
  assert.throws(
    () => validateProviderArgs("HypervPublisher", { switchName: "Default Switch", hardDrives: [] }),
    /HypervPublisher requires enableExperimentalHyperv: true/,
  );
});
