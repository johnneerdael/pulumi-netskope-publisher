import * as pulumi from "@pulumi/pulumi";
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
import { PublisherOutput, VultrPublisherArgs } from "./types";

export class VultrPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: VultrPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:VultrPublisher", name, {}, opts);

    const provider = providerCatalog.VultrPublisher;
    const outputs = createCatalogRawVmPublishers({
      parent: this,
      componentName: name,
      provider,
      args,
      mapInputs: (input, currentArgs) => ({
        label: input.publisherName,
        hostname: input.publisherName,
        region: currentArgs.region,
        plan: currentArgs.plan,
        osId: currentArgs.osId,
        imageId: currentArgs.imageId,
        sshKeyIds: currentArgs.sshKeyIds,
        vpc2Ids: currentArgs.vpc2Ids,
        enableIpv6: currentArgs.enableIpv6,
        firewallGroupId: currentArgs.firewallGroupId,
        ...userDataProperty(provider, input),
        tags: currentArgs.tags === undefined ? undefined : pulumi.output(currentArgs.tags).apply((tags) => Object.entries(tags).map(([key, value]) => `${key}:${value}`)),
      }),
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
