export type FetchLike = typeof fetch;

export interface NetskopeClientArgs {
  tenantUrl: string;
  apiToken: string;
  fetchImpl?: FetchLike;
}

export class NetskopeClient {
  private readonly apiBase: string;
  private readonly apiToken: string;
  private readonly fetchImpl: FetchLike;

  constructor(args: NetskopeClientArgs) {
    this.apiBase = `${args.tenantUrl.replace(/\/+$/, "")}/api/v2/infrastructure/publishers`;
    this.apiToken = args.apiToken;
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
    const response = await this.fetchImpl(url, {
      ...init,
      headers: {
        "Netskope-Api-Token": this.apiToken,
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
}
