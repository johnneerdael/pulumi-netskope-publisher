import * as pulumi from "@pulumi/pulumi";
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
import { DigitaloceanPublisherArgs, PublisherOutput } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class DigitaloceanPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: DigitaloceanPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:DigitaloceanPublisher", name, {}, opts);
    validateComponentArgs("DigitaloceanPublisher", args);

    const provider = providerCatalog.DigitaloceanPublisher;
    const outputs = createCatalogRawVmPublishers({
      parent: this,
      componentName: name,
      provider,
      args,
      mapInputs: (input, currentArgs) => ({
        name: input.publisherName,
        region: currentArgs.region,
        size: currentArgs.size ?? "s-2vcpu-4gb",
        image: currentArgs.image ?? "ubuntu-22-04-x64",
        sshKeys: currentArgs.sshKeys,
        vpcUuid: currentArgs.vpcUuid,
        monitoring: currentArgs.monitoring,
        ipv6: currentArgs.ipv6,
        ...userDataProperty(provider, input),
        tags: currentArgs.tags === undefined ? undefined : pulumi.output(currentArgs.tags).apply((tags) => Object.entries(tags).map(([key, value]) => `${key}:${value}`)),
      }),
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
