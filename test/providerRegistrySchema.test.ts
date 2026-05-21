import assert from "node:assert/strict";
import test from "node:test";
import { validateProviderAgainstRegistrySchema } from "../src/providerRegistrySchema";

const provider = {
  componentName: "ExamplePublisher",
  resourceToken: "example:index/server:Server",
  providerPackage: "@example/provider",
  userData: {
    mode: "plain",
    property: "userData",
  },
};

test("validateProviderAgainstRegistrySchema accepts matching token, package, and user-data property", () => {
  const errors = validateProviderAgainstRegistrySchema(provider, {
    name: "example",
    language: { nodejs: { packageName: "@example/provider" } },
    resources: {
      "example:index/server:Server": {
        inputProperties: {
          userData: { type: "string" },
        },
      },
    },
  });

  assert.deepEqual(errors, []);
});

test("validateProviderAgainstRegistrySchema rejects missing resource token", () => {
  const errors = validateProviderAgainstRegistrySchema(provider, {
    name: "example",
    language: { nodejs: { packageName: "@example/provider" } },
    resources: {},
  });

  assert.match(errors.join("\n"), /ExamplePublisher upstream schema missing resource token example:index\/server:Server/);
});

test("validateProviderAgainstRegistrySchema rejects missing user-data property", () => {
  const errors = validateProviderAgainstRegistrySchema(provider, {
    name: "example",
    language: { nodejs: { packageName: "@example/provider" } },
    resources: {
      "example:index/server:Server": {
        inputProperties: {
          metadata: { type: "object" },
        },
      },
    },
  });

  assert.match(errors.join("\n"), /ExamplePublisher upstream resource example:index\/server:Server missing user-data property userData/);
});

test("validateProviderAgainstRegistrySchema accepts terraform-provider package markers without node package comparison", () => {
  const errors = validateProviderAgainstRegistrySchema({
    ...provider,
    providerPackage: "terraform-provider:example/example",
  }, {
    name: "example",
    language: {},
    resources: {
      "example:index/server:Server": {
        inputProperties: {
          userData: { type: "string" },
        },
      },
    },
  });

  assert.deepEqual(errors, []);
});
