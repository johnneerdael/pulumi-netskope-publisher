import type { NetskopeAuthMode } from "./types";

export type FetchLike = typeof fetch;

export interface NetskopeOAuth2ClientArgs {
  tokenUrl: string;
  clientId: string;
  clientSecret: string;
  scope?: string;
}

export interface NetskopeClientArgs {
  tenantUrl: string;
  bearerToken?: string;
  authMode?: NetskopeAuthMode;
  oauth2?: NetskopeOAuth2ClientArgs;
  apiToken?: string;
  fetchImpl?: FetchLike;
}

export class NetskopeClient {
  private readonly apiBase: string;
  private readonly bearerToken?: string;
  private readonly authMode: NetskopeAuthMode;
  private readonly oauth2?: NetskopeOAuth2ClientArgs;
  private readonly fetchImpl: FetchLike;
  private accessToken?: Promise<string>;

  constructor(args: NetskopeClientArgs) {
    this.apiBase = `${args.tenantUrl.replace(/\/+$/, "")}/api/v2/infrastructure/publishers`;
    this.bearerToken = args.bearerToken ?? args.apiToken;
    this.authMode = args.authMode ?? "token";
    this.oauth2 = args.oauth2;
    this.fetchImpl = args.fetchImpl ?? fetch;
  }

  async listPublishers(): Promise<Record<string, number>> {
    const body = await this.request("List publishers", this.apiBase, { method: "GET" });
    const publishers = body?.data?.publishers ?? [];

    return Object.fromEntries(
      publishers.map((publisher: { publisher_name: string; publisher_id: string | number }) => [
        publisher.publisher_name,
        Number(publisher.publisher_id),
      ]),
    );
  }

  async createPublisher(name: string): Promise<number> {
    const body = await this.request(`Create publisher ${name}`, this.apiBase, {
      method: "POST",
      body: JSON.stringify({ name }),
    });

    return Number(body?.data?.id);
  }

  async generateRegistrationToken(publisherId: number): Promise<string> {
    const body = await this.request(
      `Generate registration token for publisher ${publisherId}`,
      `${this.apiBase}/${publisherId}/registration_token`,
      { method: "POST" },
    );

    return String(body?.data?.token);
  }

  private async request(operation: string, url: string, init: RequestInit): Promise<any> {
    const token = await this.resolveAccessToken();
    const response = await this.fetchImpl(url, {
      ...init,
      headers: {
        "Authorization": `Bearer ${token}`,
        "Accept": "application/json",
        "Content-Type": "application/json",
        ...(init.headers ?? {}),
      },
    });

    const text = await response.text();
    const body = text.length > 0 ? JSON.parse(text) : undefined;

    if (response.status < 200 || response.status >= 300) {
      throw new Error(`${operation} failed (status=${response.status})`);
    }

    return body;
  }

  private async resolveAccessToken(): Promise<string> {
    if (this.authMode === "token") {
      if (!this.bearerToken) {
        throw new Error("Netskope bearerToken or apiToken is required for token authentication");
      }
      return this.bearerToken;
    }

    if (this.authMode !== "oauth2") {
      throw new Error(`Unsupported Netskope authMode ${this.authMode}`);
    }

    this.accessToken ??= this.fetchOAuth2AccessToken();
    return this.accessToken;
  }

  private async fetchOAuth2AccessToken(): Promise<string> {
    if (!this.oauth2?.tokenUrl || !this.oauth2.clientId || !this.oauth2.clientSecret) {
      throw new Error("Netskope oauth2.tokenUrl, clientId, and clientSecret are required for OAuth2 authentication");
    }

    const body = new URLSearchParams();
    body.set("grant_type", "client_credentials");
    body.set("client_id", this.oauth2.clientId);
    body.set("client_secret", this.oauth2.clientSecret);
    if (this.oauth2.scope) {
      body.set("scope", this.oauth2.scope);
    }

    const response = await this.fetchImpl(this.oauth2.tokenUrl, {
      method: "POST",
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body,
    });
    const text = await response.text();
    const responseBody = text.length > 0 ? JSON.parse(text) : undefined;

    if (response.status < 200 || response.status >= 300) {
      throw new Error(`Fetch OAuth2 access token failed (status=${response.status})`);
    }

    const token = responseBody?.access_token;
    if (typeof token !== "string" || token.length === 0) {
      throw new Error("Fetch OAuth2 access token returned no access_token");
    }
    return token;
  }
}
