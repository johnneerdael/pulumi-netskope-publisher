import assert from "node:assert/strict";
import test from "node:test";
import { renderMetadata, renderUserData, renderUserDataBase64 } from "../src/cloudInit";

test("renderUserData bootstraps publisher software by default", () => {
  const userData = renderUserData({
    publisherName: "pub-1",
    registrationToken: "token-123",
  });

  assert.match(userData, /system_info:\n  default_user:\n    name: ubuntu/);
  assert.match(userData, /install -d -o ubuntu -g ubuntu -m 0755 \/home\/ubuntu\/resources/);
  assert.match(userData, /install -o ubuntu -g ubuntu -m 0644 \/dev\/null \/home\/ubuntu\/resources\/\.nonat/);
  assert.match(userData, /curl -fsSL https:\/\/s3-us-west-2\.amazonaws\.com\/publisher\.netskope\.com\/latest\/generic\/bootstrap\.sh \| sudo bash/);
  assert.match(userData, /sudo \/home\/ubuntu\/npa_publisher_wizard -token "token-123"/);
});

test("renderUserData can skip bootstrap for pre-baked publisher images", () => {
  const userData = renderUserData({
    publisherName: "pub-1",
    registrationToken: "token-123",
    bootstrap: false,
    nonat: false,
  });

  assert.doesNotMatch(userData, /bootstrap\.sh/);
  assert.doesNotMatch(userData, /\.nonat/);
  assert.match(userData, /\[ \/home\/ubuntu\/npa_publisher_wizard, -token, "token-123" \]/);
});

test("renderMetadata matches Terraform metadata structure", () => {
  assert.equal(
    renderMetadata("pub-1"),
    [
      "instance-id: pub-1",
      "local-hostname: pub-1",
      "",
    ].join("\n"),
  );
});

test("renderUserDataBase64 encodes rendered user data", () => {
  const raw = renderUserData({
    publisherName: "pub-1",
    registrationToken: "token-123",
    wizardPath: "/home/ubuntu/npa_publisher_wizard",
  });

  assert.equal(renderUserDataBase64({
    publisherName: "pub-1",
    registrationToken: "token-123",
    wizardPath: "/home/ubuntu/npa_publisher_wizard",
  }), Buffer.from(raw, "utf8").toString("base64"));
});
