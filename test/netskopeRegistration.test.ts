import assert from "node:assert/strict";
import test from "node:test";
import { resolveRegistrations } from "../src/netskopeRegistration";

test("resolveRegistrations reuses existing publisher and creates missing publisher", async () => {
  const created: string[] = [];
  const tokens: number[] = [];

  const result = await resolveRegistrations({
    publisherNames: ["pub-a", "pub-b"],
    client: {
      listPublishers: async () => ({ "pub-a": 101 }),
      createPublisher: async (name: string) => {
        created.push(name);
        return 202;
      },
      generateRegistrationToken: async (publisherId: number) => {
        tokens.push(publisherId);
        return `token-${publisherId}`;
      },
    },
  });

  assert.deepEqual(created, ["pub-b"]);
  assert.deepEqual(tokens, [101, 202]);
  assert.deepEqual(result, {
    "pub-a": { publisherId: 101, registrationToken: "token-101", existedBefore: true },
    "pub-b": { publisherId: 202, registrationToken: "token-202", existedBefore: false },
  });
});
