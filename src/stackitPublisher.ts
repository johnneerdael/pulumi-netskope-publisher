import * as pulumi from "@pulumi/pulumi";
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
import { PublisherOutput, StackitPublisherArgs } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class StackitPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: StackitPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:StackitPublisher", name, {}, opts);
    validateComponentArgs("StackitPublisher", args);

    const provider = providerCatalog.StackitPublisher;
    const outputs = createCatalogRawVmPublishers({
      parent: this,
      componentName: name,
      provider,
      args,
      mapInputs: (input, currentArgs) => ({
        name: input.publisherName,
        projectId: currentArgs.projectId,
        machineType: currentArgs.machineType,
        imageId: currentArgs.imageId,
        availabilityZone: currentArgs.availabilityZone,
        keypairName: currentArgs.keypairName,
        networkInterfaces: currentArgs.networkInterfaces,
        ...userDataProperty(provider, input),
        labels: currentArgs.tags,
      }),
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
