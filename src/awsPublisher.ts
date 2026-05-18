import * as aws from "@pulumi/aws";
import * as pulumi from "@pulumi/pulumi";
import { renderUserDataBase64 } from "./cloudInit";
import {
  buildNameTag,
  createPublisherOutput,
  createRegistrations,
  resolvePublisherNames,
} from "./componentCore";
import { AwsPublisherArgs, PublisherOutput } from "./types";

export class AwsPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: AwsPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:AwsPublisher", name, {}, opts);

    const parentOpts = { parent: this };
    const publisherNames = resolvePublisherNames(args);

    this.publisherNames = pulumi.output(publisherNames);
    const registrations = createRegistrations(name, publisherNames, args, parentOpts);

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

      const tags = buildNameTag(args.tags, publisherName);

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

      publisherOutputs[publisherName] = createPublisherOutput({
        registration,
        vmId: instance.id,
        privateIp: instance.privateIp,
        publicIp: instance.publicIp,
      });
    }

    this.publishers = pulumi.secret(pulumi.all(publisherOutputs));

    this.registerOutputs({
      publisherNames: this.publisherNames,
      publishers: this.publishers,
    });
  }
}
