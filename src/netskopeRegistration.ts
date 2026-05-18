import * as pulumi from "@pulumi/pulumi";
import { NetskopeClient } from "./netskopeClient";

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
  apiToken: pulumi.Input<string>;
}

interface NetskopeRegistrationProviderInputs {
  publisherNames: string[];
  tenantUrl: string;
  apiToken: string;
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
