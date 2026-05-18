import * as gcp from "@pulumi/gcp";
import * as pulumi from "@pulumi/pulumi";
import { renderUserData } from "./cloudInit";
import {
  createPublisherOutput,
  createRegistrations,
  resolvePublisherNames,
} from "./componentCore";
import { GcpPublisherArgs, PublisherOutput } from "./types";

export class GcpPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: GcpPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:GcpPublisher", name, {}, opts);

    const parentOpts = { parent: this };
    const publisherNames = resolvePublisherNames(args);
    this.publisherNames = pulumi.output(publisherNames);
    const registrations = createRegistrations(name, publisherNames, args, parentOpts);
    const publisherOutputs: Record<string, pulumi.Output<PublisherOutput>> = {};

    for (const publisherName of publisherNames) {
      const registration = registrations.apply((allRegistrations) => allRegistrations[publisherName]);
      const userData = pulumi.all([registration, args.wizardPath]).apply(([record, wizardPath]) =>
        renderUserData({
          publisherName,
          registrationToken: record.registrationToken,
          wizardPath,
        }),
      );

      const instance = new gcp.compute.Instance(`${name}-${publisherName}`, {
        name: publisherName,
        project: args.project,
        zone: args.zone,
        machineType: args.machineType ?? "e2-medium",
        tags: args.networkTags ?? [],
        labels: args.tags,
        bootDisk: {
          initializeParams: {
            image: args.image,
          },
        },
        networkInterfaces: [{
          network: args.network,
          subnetwork: args.subnetwork,
          accessConfigs: pulumi.output(args.assignPublicIp ?? false).apply((enabled) => enabled ? [{}] : []),
        }],
        metadata: {
          "user-data": userData,
        },
        serviceAccount: args.serviceAccount === undefined ? undefined : pulumi.output(args.serviceAccount).apply((serviceAccount) => ({
          email: serviceAccount.email,
          scopes: serviceAccount.scopes ?? ["https://www.googleapis.com/auth/cloud-platform"],
        })),
      }, parentOpts);

      const firstInterface = pulumi.output(instance.networkInterfaces).apply((interfaces) => interfaces[0]);

      publisherOutputs[publisherName] = createPublisherOutput({
        registration,
        vmId: instance.instanceId,
        privateIp: firstInterface.apply((networkInterface) => networkInterface.networkIp ?? ""),
        publicIp: firstInterface.apply((networkInterface) => networkInterface.accessConfigs?.[0]?.natIp),
      });
    }

    this.publishers = pulumi.secret(pulumi.all(publisherOutputs));
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
