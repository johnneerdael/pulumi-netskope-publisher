import * as pulumi from "@pulumi/pulumi";
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
import { OutscalePublisherArgs, PublisherOutput } from "./types";

export class OutscalePublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: OutscalePublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:OutscalePublisher", name, {}, opts);

    const provider = providerCatalog.OutscalePublisher;
    const outputs = createCatalogRawVmPublishers({
      parent: this,
      componentName: name,
      provider,
      args,
      mapInputs: (input, currentArgs) => ({
        imageId: currentArgs.imageId,
        vmType: currentArgs.vmType ?? "tinav5.c2r4p1",
        subnetId: currentArgs.subnetId,
        keypairName: currentArgs.keypairName,
        securityGroupIds: currentArgs.securityGroupIds,
        placementSubregionName: currentArgs.placementSubregionName,
        ...userDataProperty(provider, input),
      }),
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
