import * as pulumi from "@pulumi/pulumi";
import { NetskopeClient, NetskopeOAuth2ClientArgs } from "./netskopeClient";
import { NetskopeAuthMode, NetskopeOAuth2Args } from "./types";

export interface RegistrationRecord {
  publisherId: number;
  registrationToken: string;
  existedBefore: boolean;
}

export interface RegistrationClient {
  listPublishers(): Promise<Record<string, number>>;
  createPublisher(name: string): Promise<number>;
  generateRegistrationToken(publisherId: number): Promise<string>;
}

export interface ResolveRegistrationsArgs {
  publisherNames: string[];
  client: RegistrationClient;
}

export async function resolveRegistrations(args: ResolveRegistrationsArgs): Promise<Record<string, RegistrationRecord>> {
  const existingByName = await args.client.listPublishers();
  const result: Record<string, RegistrationRecord> = {};

  for (const publisherName of args.publisherNames) {
    const existingId = existingByName[publisherName];
    const existedBefore = existingId !== undefined;
    const publisherId = existedBefore ? existingId : await args.client.createPublisher(publisherName);
    const registrationToken = await args.client.generateRegistrationToken(publisherId);

    result[publisherName] = {
      publisherId,
      registrationToken,
      existedBefore,
    };
  }

  return result;
}

export interface NetskopeRegistrationArgs {
  publisherNames: pulumi.Input<string[]>;
  tenantUrl: pulumi.Input<string>;
  bearerToken?: pulumi.Input<string>;
  authMode?: pulumi.Input<NetskopeAuthMode>;
  oauth2?: pulumi.Input<NetskopeOAuth2Args>;
  /** @deprecated Use bearerToken instead. */
  apiToken?: pulumi.Input<string>;
}

interface NetskopeRegistrationProviderInputs {
  publisherNames: string[];
  tenantUrl: string;
  bearerToken?: string;
  authMode?: NetskopeAuthMode;
  oauth2?: NetskopeOAuth2ClientArgs;
  apiToken?: string;
}

interface NetskopeRegistrationProviderOutputs extends NetskopeRegistrationProviderInputs {
  registrations: Record<string, RegistrationRecord>;
}

class NetskopeRegistrationProvider implements pulumi.dynamic.ResourceProvider {
  async create(inputs: NetskopeRegistrationProviderInputs): Promise<pulumi.dynamic.CreateResult> {
    const registrations = await resolveRegistrations({
      publisherNames: inputs.publisherNames,
      client: new NetskopeClient({
        tenantUrl: inputs.tenantUrl,
        bearerToken: inputs.bearerToken,
        authMode: inputs.authMode,
        oauth2: inputs.oauth2,
        apiToken: inputs.apiToken,
      }),
    });

    return {
      id: inputs.publisherNames.join(","),
      outs: {
        ...inputs,
        registrations,
      },
    };
  }

  async diff(
    id: string,
    oldOutputs: NetskopeRegistrationProviderOutputs,
    newInputs: NetskopeRegistrationProviderInputs,
  ): Promise<pulumi.dynamic.DiffResult> {
    return {
      changes: JSON.stringify(oldOutputs.publisherNames) !== JSON.stringify(newInputs.publisherNames)
        || oldOutputs.tenantUrl !== newInputs.tenantUrl,
      replaces: ["publisherNames", "tenantUrl"],
    };
  }
}

export class NetskopeRegistration extends pulumi.dynamic.Resource {
  declare readonly registrations: pulumi.Output<Record<string, RegistrationRecord>>;

  constructor(name: string, args: NetskopeRegistrationArgs, opts?: pulumi.CustomResourceOptions) {
    super(new NetskopeRegistrationProvider(), name, {
      registrations: undefined,
      ...args,
    }, opts);
  }
}
