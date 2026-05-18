import assert from "node:assert/strict";
import test from "node:test";
import { NetskopeClient } from "../src/netskopeClient";

test("listPublishers parses publisher IDs by name", async () => {
  const requests: Array<{ url: string; init: RequestInit }> = [];
  const client = new NetskopeClient({
    tenantUrl: "https://tenant.goskope.com/",
    apiToken: "secret",
    fetchImpl: async (url, init) => {
      requests.push({ url: String(url), init: init ?? {} });
      return response(200, {
        data: {
          publishers: [
            { publisher_name: "pub-a", publisher_id: "101" },
            { publisher_name: "pub-b", publisher_id: 102 },
          ],
        },
      });
    },
  });

  assert.deepEqual(await client.listPublishers(), {
    "pub-a": 101,
    "pub-b": 102,
  });
  assert.equal(requests[0].url, "https://tenant.goskope.com/api/v2/infrastructure/publishers");
  assert.equal((requests[0].init.headers as Record<string, string>)["Netskope-Api-Token"], "secret");
});

test("createPublisher returns created publisher ID", async () => {
  const client = new NetskopeClient({
    tenantUrl: "https://tenant.goskope.com",
    apiToken: "secret",
    fetchImpl: async () => response(201, { data: { id: "123" } }),
  });

  assert.equal(await client.createPublisher("pub-a"), 123);
});

test("generateRegistrationToken returns token", async () => {
  const client = new NetskopeClient({
    tenantUrl: "https://tenant.goskope.com",
    apiToken: "secret",
    fetchImpl: async () => response(200, { data: { token: "registration-token" } }),
  });

  assert.equal(await client.generateRegistrationToken(123), "registration-token");
});

test("client errors include operation and status without token", async () => {
  const client = new NetskopeClient({
    tenantUrl: "https://tenant.goskope.com",
    apiToken: "super-secret-token",
    fetchImpl: async () => response(403, { message: "forbidden" }),
  });

  await assert.rejects(
    () => client.listPublishers(),
    (error: unknown) => {
      assert.match(String(error), /List publishers failed \(status=403\)/);
      assert.doesNotMatch(String(error), /super-secret-token/);
      return true;
    },
  );
});

function response(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}
