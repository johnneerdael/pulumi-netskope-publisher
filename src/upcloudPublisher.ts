import * as pulumi from "@pulumi/pulumi";
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
import { PublisherOutput, UpcloudPublisherArgs } from "./types";

const ubuntu2204Template = "01000000-0000-4000-8000-000030220200";

export class UpcloudPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: UpcloudPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:UpcloudPublisher", name, {}, opts);

    const provider = providerCatalog.UpcloudPublisher;
    const outputs = createCatalogRawVmPublishers({
      parent: this,
      componentName: name,
      provider,
      args,
      mapInputs: (input, currentArgs) => ({
        hostname: currentArgs.hostname ?? input.publisherName,
        title: input.publisherName,
        zone: currentArgs.zone,
        plan: currentArgs.plan ?? "2xCPU-4GB",
        template: currentArgs.template ?? ubuntu2204Template,
        networkInterfaces: currentArgs.networkInterfaces,
        metadata: true,
        ...userDataProperty(provider, input),
        labels: currentArgs.tags,
      }),
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
