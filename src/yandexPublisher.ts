import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { PublisherOutput, YandexPublisherArgs } from "./types";
import { metadataUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";

export class YandexPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: YandexPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:YandexPublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userData }) => {
      const instance = new RawResource(`${name}-${publisherName}`, "yandex:index/computeInstance:ComputeInstance", {
        name: publisherName,
        hostname: publisherName,
        zone: args.zone,
        platformId: args.platformId ?? "standard-v3",
        bootDisk: {
          initializeParams: {
            imageId: args.imageId,
          },
        },
        resources: {
          cores: args.cores ?? 2,
          memory: args.memory ?? 4,
          coreFraction: args.coreFraction,
        },
        networkInterfaces: [{
          subnetId: args.subnetId,
          nat: args.nat ?? false,
        }],
        metadata: pulumi.all([metadataUserData(userData), args.sshKeys]).apply(([metadata, sshKeys]) => ({
          ...metadata,
          ...(sshKeys === undefined ? {} : { "ssh-keys": sshKeys.join("\n") }),
        })),
        labels: args.tags,
      }, { parent: this });

      return { vmId: instance.id, privateIp: pulumi.output(""), publicIp: pulumi.output("") };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
