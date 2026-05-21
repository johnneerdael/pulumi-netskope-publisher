import * as pulumi from "@pulumi/pulumi";
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
import { PublisherOutput, YandexPublisherArgs } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class YandexPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: YandexPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:YandexPublisher", name, {}, opts);
    validateComponentArgs("YandexPublisher", args);

    const provider = providerCatalog.YandexPublisher;
    const outputs = createCatalogRawVmPublishers({
      parent: this,
      componentName: name,
      provider,
      args,
      mapInputs: (input, currentArgs) => ({
        name: input.publisherName,
        hostname: input.publisherName,
        zone: currentArgs.zone,
        platformId: currentArgs.platformId ?? "standard-v3",
        bootDisk: {
          initializeParams: {
            imageId: currentArgs.imageId,
          },
        },
        resources: {
          cores: currentArgs.cores ?? 2,
          memory: currentArgs.memory ?? 4,
          coreFraction: currentArgs.coreFraction,
        },
        networkInterfaces: [{
          subnetId: currentArgs.subnetId,
          nat: currentArgs.nat ?? false,
        }],
        metadata: pulumi.all([userDataProperty(provider, input).metadata, currentArgs.sshKeys]).apply(([metadata, sshKeys]) => ({
          ...(metadata as Record<string, string>),
          ...(sshKeys === undefined ? {} : { "ssh-keys": sshKeys.join("\n") }),
        })),
        labels: currentArgs.tags,
      }),
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
