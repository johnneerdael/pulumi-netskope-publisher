import assert from "node:assert/strict";
import test from "node:test";
import { HypervPublisher } from "../src/hypervPublisher";

test("HypervPublisher requires explicit experimental opt-in", () => {
  assert.throws(
    () => new HypervPublisher("publisher", {
      names: ["pub-1"],
      switchName: "Default Switch",
      hardDrives: [{ path: "C:\\VMs\\pub-1\\disk.vhdx" }],
    }),
    /HypervPublisher requires enableExperimentalHyperv: true/,
  );
});
