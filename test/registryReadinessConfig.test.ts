import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { catalogProviders } from "../src/providerCatalog";

function objectLiteralValues(source: string, name: string): string[] {
  const match = source.match(new RegExp(`const ${name} = \\{([\\s\\S]*?)\\};`));
  assert.ok(match, `${name} object not found`);
  return Array.from(match[1].matchAll(/"[^"]+":\s*"([^"]+)"/g)).map((entry) => entry[1]);
}

test("registry readiness derives expectedResourceTokens from schema resources", () => {
  const script = readFileSync("scripts/check-registry-readiness.mjs", "utf8");

  assert.match(script, /const expectedResourceTokens = schema[\s\S]*Object\.keys\(schema\.resources \?\? \{\}\)[\s\S]*startsWith\("netskope-publisher:index:"\)/);
});

test("registry readiness sourceTokens covers every TypeScript component source", () => {
  const script = readFileSync("scripts/check-registry-readiness.mjs", "utf8");
  const sourceTokenValues = new Set(objectLiteralValues(script, "sourceTokens"));
  const missing = catalogProviders
    .filter((provider) => provider.componentName !== "NetskopeRegistration")
    .filter((provider) => provider.componentName !== "PrivateApp")
    .filter((provider) => provider.componentName !== "TagPublisherAssignment")
    .filter((provider) => provider.componentName !== "RealtimeProtectionPolicy")
    .map((provider) => provider.token)
    .filter((token) => !sourceTokenValues.has(token))
    .sort();

  assert.deepEqual(missing, []);
});
