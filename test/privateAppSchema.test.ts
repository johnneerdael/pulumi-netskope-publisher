import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const schema = JSON.parse(readFileSync("schema.json", "utf8"));
const privateAppSource = readFileSync("src/privateApp.ts", "utf8");

test("PrivateApp schema does not expose unsupported hosts input", () => {
  const privateApp = schema.resources["netskope-publisher:index:PrivateApp"];
  assert.ok(privateApp, "PrivateApp resource must exist in schema");
  assert.equal(privateApp.inputProperties.hosts, undefined);
  assert.equal(privateApp.properties.hosts, undefined);
  assert.equal(privateAppSource.includes("hosts?"), false);
});
