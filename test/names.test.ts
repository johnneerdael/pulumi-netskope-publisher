import assert from "node:assert/strict";
import test from "node:test";
import { derivePublisherNames } from "../src/names";

test("derivePublisherNames returns explicit names unchanged", () => {
  assert.deepEqual(
    derivePublisherNames({ namePrefix: "pub", names: ["alpha", "beta"], replicas: 3 }),
    ["alpha", "beta"],
  );
});

test("derivePublisherNames creates Terraform-compatible numbered names", () => {
  assert.deepEqual(
    derivePublisherNames({ namePrefix: "npa-publisher", replicas: 3 }),
    ["npa-publisher-1", "npa-publisher-2", "npa-publisher-3"],
  );
});

test("derivePublisherNames defaults to one npa-publisher", () => {
  assert.deepEqual(
    derivePublisherNames({}),
    ["npa-publisher-1"],
  );
});

test("derivePublisherNames rejects replicas below one", () => {
  assert.throws(
    () => derivePublisherNames({ replicas: 0 }),
    /replicas must be >= 1/,
  );
});

test("derivePublisherNames rejects an empty names list", () => {
  assert.throws(
    () => derivePublisherNames({ names: [] }),
    /names must contain at least one publisher name/,
  );
});
