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
import { validateComponentArgs } from "./providerValidation";

export class AwsPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: AwsPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:AwsPublisher", name, {}, opts);
    validateComponentArgs("AwsPublisher", args);

    const parentOpts = { parent: this };
    const publisherNames = resolvePublisherNames(args);

    this.publisherNames = pulumi.output(publisherNames);
    const registrations = createRegistrations(name, publisherNames, args, parentOpts);

    const bootstrap = args.bootstrap ?? false;
    const ami = args.amiId
      ? pulumi.output(args.amiId)
      : aws.ec2.getAmiOutput(bootstrap === true ? {
        mostRecent: true,
        owners: ["099720109477"],
        filters: [{
          name: "name",
          values: ["ubuntu-minimal/images/hvm-ssd*/ubuntu-jammy-22.04-amd64-minimal-*"],
        }, {
          name: "architecture",
          values: ["x86_64"],
        }, {
          name: "virtualization-type",
          values: ["hvm"],
        }],
      } : {
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
      const userDataBase64 = pulumi.all({
        registration,
        wizardPath: args.wizardPath,
        bootstrap,
        bootstrapUrl: args.bootstrapUrl,
        nonat: args.nonat ?? false,
        installUser: args.installUser,
        installUserPassword: args.installUserPassword,
        installUserPasswordIsHash: args.installUserPasswordIsHash,
        installUserSshAuthorizedKeys: args.installUserSshAuthorizedKeys,
        deleteDefaultUser: args.deleteDefaultUser,
        guestNetworkInterface: args.guestNetworkInterface,
      }).apply((options: any) =>
        renderUserDataBase64({
          publisherName,
          registrationToken: options.registration.registrationToken,
          wizardPath: options.wizardPath,
          bootstrap: options.bootstrap,
          bootstrapUrl: options.bootstrapUrl,
          nonat: options.nonat,
          installUser: options.installUser,
          installUserPassword: options.installUserPassword,
          installUserPasswordIsHash: options.installUserPasswordIsHash,
          installUserSshAuthorizedKeys: options.installUserSshAuthorizedKeys,
          deleteDefaultUser: options.deleteDefaultUser,
          guestNetworkInterface: options.guestNetworkInterface,
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
        placementLabels: args.placementLabels,
      });
    }

    this.publishers = pulumi.secret(pulumi.all(publisherOutputs));

    this.registerOutputs({
      publisherNames: this.publisherNames,
      publishers: this.publishers,
    });
  }
}
