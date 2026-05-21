import * as pulumi from "@pulumi/pulumi";
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
import { OpentelekomcloudPublisherArgs, PublisherOutput } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class OpentelekomcloudPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: OpentelekomcloudPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:OpentelekomcloudPublisher", name, {}, opts);
    validateComponentArgs("OpentelekomcloudPublisher", args);

    const provider = providerCatalog.OpentelekomcloudPublisher;
    const outputs = createCatalogRawVmPublishers({
      parent: this,
      componentName: name,
      provider,
      args,
      mapInputs: (input, currentArgs) => ({
        name: input.publisherName,
        imageName: currentArgs.imageName ?? "Ubuntu 22.04",
        imageId: currentArgs.imageId,
        flavorName: currentArgs.flavorName ?? "s3.medium.2",
        flavorId: currentArgs.flavorId,
        networks: currentArgs.networks,
        keyPair: currentArgs.keyPair,
        availabilityZone: currentArgs.availabilityZone,
        securityGroups: currentArgs.securityGroups,
        ...userDataProperty(provider, input),
      }),
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
