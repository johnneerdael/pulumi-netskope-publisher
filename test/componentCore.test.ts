import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import {
  buildNameTag,
  createPublisherOutput,
  normalizeByoRegistrations,
  requireManagedRegistrationInputs,
} from "../src/componentCore";

test("buildNameTag merges provider tags with Name", async () => {
  const tags = await outputValue(buildNameTag({ Env: "dev" }, "pub-1"));
  assert.deepEqual(tags, { Env: "dev", Name: "pub-1" });
});

test("normalizeByoRegistrations requires every publisher name", () => {
  assert.throws(
    () => normalizeByoRegistrations(["pub-1"], {}),
    /registrations is missing data for publisher pub-1/,
  );
});

test("requireManagedRegistrationInputs rejects missing tenantUrl", () => {
  assert.throws(
    () => requireManagedRegistrationInputs({ apiToken: "token" }),
    /tenantUrl and a bearer token or oauth2 credentials are required when registrations are not provided/,
  );
});

test("requireManagedRegistrationInputs accepts bearerToken without legacy apiToken", () => {
  const result = requireManagedRegistrationInputs({
    tenantUrl: "https://tenant.goskope.com",
    bearerToken: "bearer-token",
  });

  assert.equal(result.bearerToken, "bearer-token");
});

test("requireManagedRegistrationInputs accepts oauth2 credentials", () => {
  const result = requireManagedRegistrationInputs({
    tenantUrl: "https://tenant.goskope.com",
    authMode: "oauth2",
    oauth2: {
      tokenUrl: "https://tenant.goskope.com/oauth2/token",
      clientId: "client-id",
      clientSecret: "client-secret",
    },
  });

  assert.equal(result.authMode, "oauth2");
  assert.equal((result.oauth2 as { tokenUrl: string }).tokenUrl, "https://tenant.goskope.com/oauth2/token");
});

test("createPublisherOutput preserves provider IDs and token", async () => {
  const output = await outputValue(createPublisherOutput({
    registration: pulumi.output({
      publisherId: 101,
      registrationToken: "token-101",
      existedBefore: true,
    }),
    vmId: pulumi.output("vm-1"),
    privateIp: pulumi.output("10.0.0.10"),
    publicIp: pulumi.output(undefined),
  }));

  assert.deepEqual(output, {
    publisherId: 101,
    registrationToken: "token-101",
    vmId: "vm-1",
    privateIp: "10.0.0.10",
    publicIp: undefined,
  });
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
