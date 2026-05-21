import * as pulumi from "@pulumi/pulumi";
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
import { ExoscalePublisherArgs, PublisherOutput } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class ExoscalePublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: ExoscalePublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:ExoscalePublisher", name, {}, opts);
    validateComponentArgs("ExoscalePublisher", args);

    const provider = providerCatalog.ExoscalePublisher;
    const outputs = createCatalogRawVmPublishers({
      parent: this,
      componentName: name,
      provider,
      args,
      mapInputs: (input, currentArgs) => ({
        name: input.publisherName,
        zone: currentArgs.zone,
        type: currentArgs.type,
        templateId: currentArgs.templateId,
        diskSize: currentArgs.diskSize,
        sshKeys: currentArgs.sshKeys,
        securityGroupIds: currentArgs.securityGroupIds,
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
