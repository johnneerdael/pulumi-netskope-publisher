import * as pulumi from "@pulumi/pulumi";
import { derivePublisherNames } from "./names";
import { NetskopeRegistration, RegistrationRecord } from "./netskopeRegistration";
import { CommonPublisherArgs, PublisherOutput, PublisherRegistrationInput } from "./types";

export function resolvePublisherNames(args: CommonPublisherArgs): string[] {
  return derivePublisherNames({
    namePrefix: args.namePrefix,
    names: args.names,
    replicas: args.replicas,
  });
}

export function createRegistrations(
  componentName: string,
  publisherNames: string[],
  args: CommonPublisherArgs,
  opts: pulumi.CustomResourceOptions,
): pulumi.Output<Record<string, RegistrationRecord>> {
  if (args.registrations !== undefined) {
    return pulumi.output(args.registrations).apply((registrations) =>
      normalizeByoRegistrations(publisherNames, registrations),
    );
  }

  const required = requireManagedRegistrationInputs(args);
  return new NetskopeRegistration(`${componentName}-registration`, {
    publisherNames,
    tenantUrl: required.tenantUrl,
    apiToken: required.apiToken,
  }, opts).registrations;
}

export function requireManagedRegistrationInputs(args: CommonPublisherArgs): {
  tenantUrl: pulumi.Input<string>;
  apiToken: pulumi.Input<string>;
} {
  if (args.tenantUrl === undefined || args.apiToken === undefined) {
    throw new Error("tenantUrl and apiToken are required when registrations are not provided");
  }

  return {
    tenantUrl: args.tenantUrl,
    apiToken: args.apiToken,
  };
}

export function normalizeByoRegistrations(
  publisherNames: string[],
  registrations: Record<string, PublisherRegistrationInput>,
): Record<string, RegistrationRecord> {
  return Object.fromEntries(publisherNames.map((publisherName) => {
    const registration = registrations[publisherName];
    if (registration === undefined) {
      throw new Error(`registrations is missing data for publisher ${publisherName}`);
    }

    return [publisherName, {
      publisherId: Number(registration.publisherId),
      registrationToken: String(registration.registrationToken),
      existedBefore: true,
    }];
  }));
}

export function buildNameTag(
  tags: pulumi.Input<Record<string, pulumi.Input<string>>> | undefined,
  publisherName: string,
): pulumi.Output<Record<string, pulumi.Input<string>>> {
  return pulumi.output(tags ?? {}).apply((inputTags): Record<string, pulumi.Input<string>> => {
    return {
      ...inputTags,
      Name: publisherName,
    };
  });
}

export function createPublisherOutput(args: {
  registration: pulumi.Output<RegistrationRecord>;
  vmId: pulumi.Input<string>;
  privateIp: pulumi.Input<string>;
  publicIp: pulumi.Input<string | undefined>;
}): pulumi.Output<PublisherOutput> {
  return pulumi.all([
    args.registration,
    args.vmId,
    args.privateIp,
    args.publicIp,
  ]).apply(([registration, vmId, privateIp, publicIp]) => ({
    publisherId: registration.publisherId,
    registrationToken: registration.registrationToken,
    vmId,
    privateIp,
    publicIp,
  }));
}
