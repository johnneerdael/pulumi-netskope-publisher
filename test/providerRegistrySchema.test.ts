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

test("validateProviderAgainstRegistrySchema accepts explicit nested schema checks", () => {
  const errors = validateProviderAgainstRegistrySchema({
    componentName: "CompositePublisher",
    providerPackage: "@example/provider",
    resourceToken: "example:index/vm:Vm",
    userData: {
      mode: "proxmoxSnippet",
    },
    registrySchemaChecks: [{
      resourceToken: "example:index/file:File",
      propertyPath: ["sourceRaw", "data"],
      description: "cloud-init snippet content",
    }, {
      resourceToken: "example:index/vm:Vm",
      propertyPath: ["initialization", "userDataFileId"],
      description: "VM cloud-init user-data file reference",
    }],
  }, {
    name: "example",
    language: { nodejs: { packageName: "@example/provider" } },
    resources: {
      "example:index/file:File": {
        inputProperties: {
          sourceRaw: { "$ref": "#/types/example:index/FileSourceRaw:FileSourceRaw" },
        },
      },
      "example:index/vm:Vm": {
        inputProperties: {
          initialization: { "$ref": "#/types/example:index/VmInitialization:VmInitialization" },
        },
      },
    },
    types: {
      "example:index/FileSourceRaw:FileSourceRaw": {
        properties: {
          data: { type: "string" },
        },
      },
      "example:index/VmInitialization:VmInitialization": {
        properties: {
          userDataFileId: { type: "string" },
        },
      },
    },
  });

  assert.deepEqual(errors, []);
});

test("validateProviderAgainstRegistrySchema rejects missing nested schema check path", () => {
  const errors = validateProviderAgainstRegistrySchema({
    componentName: "CompositePublisher",
    resourceToken: "example:index/vm:Vm",
    userData: {
      mode: "proxmoxSnippet",
    },
    registrySchemaChecks: [{
      resourceToken: "example:index/file:File",
      propertyPath: ["sourceRaw", "data"],
      description: "cloud-init snippet content",
    }],
  }, {
    name: "example",
    resources: {
      "example:index/file:File": {
        inputProperties: {
          sourceRaw: { "$ref": "#/types/example:index/FileSourceRaw:FileSourceRaw" },
        },
      },
    },
    types: {
      "example:index/FileSourceRaw:FileSourceRaw": {
        properties: {
          fileName: { type: "string" },
        },
      },
    },
  });

  assert.match(errors.join("\n"), /CompositePublisher upstream resource example:index\/file:File missing cloud-init snippet content path sourceRaw\.data/);
});

test("validateProviderAgainstRegistrySchema accepts declared upstream property paths", () => {
  const errors = validateProviderAgainstRegistrySchema({
    componentName: "NestedPublisher",
    resourceToken: "example:index/server:Server",
    providerPackage: "@example/provider",
    upstreamPropertyChecks: [{
      resourceToken: "example:index/server:Server",
      propertyPath: ["network", "subnetId"],
      description: "server subnet placement",
    }],
    userData: {
      mode: "plain",
      property: "userData",
    },
  }, {
    name: "example",
    language: { nodejs: { packageName: "@example/provider" } },
    resources: {
      "example:index/server:Server": {
        inputProperties: {
          userData: { type: "string" },
          network: { "$ref": "#/types/example:index/ServerNetwork:ServerNetwork" },
        },
      },
    },
    types: {
      "example:index/ServerNetwork:ServerNetwork": {
        properties: {
          subnetId: { type: "string" },
        },
      },
    },
  });

  assert.deepEqual(errors, []);
});

test("validateProviderAgainstRegistrySchema rejects missing declared upstream property paths", () => {
  const errors = validateProviderAgainstRegistrySchema({
    componentName: "NestedPublisher",
    resourceToken: "example:index/server:Server",
    upstreamPropertyChecks: [{
      resourceToken: "example:index/server:Server",
      propertyPath: ["network", "subnetId"],
      description: "server subnet placement",
    }],
    userData: {
      mode: "plain",
      property: "userData",
    },
  }, {
    name: "example",
    resources: {
      "example:index/server:Server": {
        inputProperties: {
          userData: { type: "string" },
          network: { "$ref": "#/types/example:index/ServerNetwork:ServerNetwork" },
        },
      },
    },
    types: {
      "example:index/ServerNetwork:ServerNetwork": {
        properties: {
          networkId: { type: "string" },
        },
      },
    },
  });

  assert.match(errors.join("\n"), /NestedPublisher upstream resource example:index\/server:Server missing server subnet placement path network\.subnetId/);
});
