import assert from "node:assert/strict";
import test from "node:test";
import { renderMetadata, renderUserData, renderUserDataBase64 } from "../src/cloudInit";

test("renderUserData matches Terraform cloud-init structure", () => {
  assert.equal(
    renderUserData({
      publisherName: "pub-1",
      registrationToken: "token-123",
      wizardPath: "/home/ubuntu/npa_publisher_wizard",
    }),
    [
      "#cloud-config",
      "hostname: pub-1",
      "preserve_hostname: false",
      "runcmd:",
      '  - [ /home/ubuntu/npa_publisher_wizard, -token, "token-123" ]',
      "",
    ].join("\n"),
  );
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
