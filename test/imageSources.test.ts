import assert from "node:assert/strict";
import test from "node:test";
import { officialImageSources } from "../src/imageSources";

test("officialImageSources exposes Netskope publisher downloads", () => {
  assert.equal(
    officialImageSources.hypervVhdx,
    "https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.vhdx",
  );
  assert.equal(
    officialImageSources.vsphereOva,
    "https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/NetskopePrivateAccessPublisher.ova",
  );
});
