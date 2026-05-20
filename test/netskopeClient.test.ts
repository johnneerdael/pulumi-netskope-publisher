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
  assert.equal((requests[0].init.headers as Record<string, string>)["Authorization"], "Bearer secret");
});

test("client uses bearerToken as the preferred static bearer credential", async () => {
  const requests: Array<{ url: string; init: RequestInit }> = [];
  const client = new NetskopeClient({
    tenantUrl: "https://tenant.goskope.com",
    bearerToken: "bearer-secret",
    fetchImpl: async (url, init) => {
      requests.push({ url: String(url), init: init ?? {} });
      return response(200, { data: { publishers: [] } });
    },
  });

  await client.listPublishers();

  assert.equal((requests[0].init.headers as Record<string, string>)["Authorization"], "Bearer bearer-secret");
  assert.equal((requests[0].init.headers as Record<string, string>)["Netskope-Api-Token"], undefined);
});

test("client fetches one OAuth2 access token and reuses it for registration API calls", async () => {
  const requests: Array<{ url: string; init: RequestInit }> = [];
  const client = new NetskopeClient({
    tenantUrl: "https://tenant.goskope.com",
    authMode: "oauth2",
    oauth2: {
      tokenUrl: "https://tenant.goskope.com/oauth2/token",
      clientId: "client-id",
      clientSecret: "client-secret",
      scope: "npa.publisher",
    },
    fetchImpl: async (url, init) => {
      requests.push({ url: String(url), init: init ?? {} });
      if (String(url).endsWith("/oauth2/token")) {
        return response(200, { access_token: "oauth-access-token" });
      }
      if (String(url).endsWith("/registration_token")) {
        return response(200, { data: { token: "registration-token" } });
      }
      return response(200, { data: { publishers: [] } });
    },
  });

  await client.listPublishers();
  await client.generateRegistrationToken(123);

  const tokenRequests = requests.filter((request) => request.url.endsWith("/oauth2/token"));
  assert.equal(tokenRequests.length, 1);
  const tokenBody = tokenRequests[0].init.body as URLSearchParams;
  assert.equal(tokenBody.get("grant_type"), "client_credentials");
  assert.equal(tokenBody.get("client_id"), "client-id");
  assert.equal(tokenBody.get("client_secret"), "client-secret");
  assert.equal(tokenBody.get("scope"), "npa.publisher");

  const apiRequests = requests.filter((request) => !request.url.endsWith("/oauth2/token"));
  assert.equal(apiRequests.length, 2);
  for (const request of apiRequests) {
    assert.equal((request.init.headers as Record<string, string>)["Authorization"], "Bearer oauth-access-token");
  }
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
