import * as aws from "@pulumi/aws";
import * as pulumi from "@pulumi/pulumi";
import { renderUserDataBase64 } from "./cloudInit";
import { derivePublisherNames } from "./names";
import { NetskopeRegistration, RegistrationRecord } from "./netskopeRegistration";
import { AwsPublisherArgs, PublisherOutput, PublisherRegistrationInput } from "./types";

export class AwsPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: AwsPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope:index:AwsPublisher", name, {}, opts);

    const parentOpts = { parent: this };
    const publisherNames = derivePublisherNames({
      namePrefix: args.namePrefix,
      names: args.names,
      replicas: args.replicas,
    });

    this.publisherNames = pulumi.output(publisherNames);

    const registrations = args.registrations !== undefined
      ? pulumi.output(args.registrations).apply((byoRegistrations) =>
        normalizeByoRegistrations(publisherNames, byoRegistrations),
      )
      : createManagedRegistrations(name, publisherNames, args, parentOpts);

    const ami = args.amiId
      ? pulumi.output(args.amiId)
      : aws.ec2.getAmiOutput({
        mostRecent: true,
        owners: ["679593333241"],
        filters: [{
          name: "name",
          values: ["Netskope Private Access Publisher*"],
        }],
      }, parentOpts).id;

    const publisherOutputs: Record<string, pulumi.Output<PublisherOutput>> = {};

    for (const publisherName of publisherNames) {
      const registration = registrations.apply((allRegistrations) => allRegistrations[publisherName]);
      const userDataBase64 = pulumi.all([registration, args.wizardPath]).apply(([record, wizardPath]) =>
        renderUserDataBase64({
          publisherName,
          registrationToken: record.registrationToken,
          wizardPath,
        }),
      );

      const tags = pulumi.output(args.tags ?? {}).apply((inputTags) => ({
        ...inputTags,
        Name: publisherName,
      }));

      const instance = new aws.ec2.Instance(`${name}-${publisherName}`, {
        ami,
        instanceType: args.instanceType ?? "t3.medium",
        subnetId: args.subnetId,
        vpcSecurityGroupIds: args.securityGroupIds,
        keyName: args.keyName,
        associatePublicIpAddress: args.associatePublicIpAddress ?? false,
        iamInstanceProfile: args.iamInstanceProfile,
        ebsOptimized: args.ebsOptimized ?? true,
        monitoring: args.monitoring ?? true,
        userDataBase64,
        metadataOptions: pulumi.output(args.metadataOptions ?? {}).apply((metadataOptions) => ({
          httpEndpoint: metadataOptions.httpEndpoint ?? "enabled",
          httpTokens: metadataOptions.httpTokens ?? "required",
        })),
        tags,
      }, parentOpts);

      publisherOutputs[publisherName] = pulumi.all([
        registration,
        instance.id,
        instance.privateIp,
        instance.publicIp,
      ]).apply(([registration, instanceId, privateIp, publicIp]) => ({
        publisherId: registration.publisherId,
        registrationToken: registration.registrationToken,
        instanceId,
        privateIp,
        publicIp,
      }));
    }

    this.publishers = pulumi.secret(pulumi.all(publisherOutputs));

    this.registerOutputs({
      publisherNames: this.publisherNames,
      publishers: this.publishers,
    });
  }
}

function createManagedRegistrations(
  name: string,
  publisherNames: string[],
  args: AwsPublisherArgs,
  opts: pulumi.CustomResourceOptions,
): pulumi.Output<Record<string, RegistrationRecord>> {
  if (args.tenantUrl === undefined || args.apiToken === undefined) {
    throw new Error("tenantUrl and apiToken are required when registrations are not provided");
  }

  return new NetskopeRegistration(`${name}-registration`, {
    publisherNames,
    tenantUrl: args.tenantUrl,
    apiToken: args.apiToken,
  }, opts).registrations;
}

function normalizeByoRegistrations(
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
