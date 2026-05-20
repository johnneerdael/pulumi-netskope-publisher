import * as pulumi from "@pulumi/pulumi";
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
import { EquinixPublisherArgs, PublisherOutput } from "./types";

export class EquinixPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: EquinixPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:EquinixPublisher", name, {}, opts);

    const provider = providerCatalog.EquinixPublisher;
    const outputs = createCatalogRawVmPublishers({
      parent: this,
      componentName: name,
      provider,
      args,
      mapInputs: (input, currentArgs) => ({
        hostname: input.publisherName,
        projectId: currentArgs.projectId,
        metro: currentArgs.metro,
        plan: currentArgs.plan,
        operatingSystem: currentArgs.operatingSystem ?? "ubuntu_22_04",
        billingCycle: currentArgs.billingCycle ?? "hourly",
        projectSshKeyIds: currentArgs.projectSshKeyIds,
        userSshKeyIds: currentArgs.userSshKeyIds,
        ...userDataProperty(provider, input),
        tags: currentArgs.tags === undefined ? undefined : pulumi.output(currentArgs.tags).apply((tags) => Object.entries(tags).map(([key, value]) => `${key}:${value}`)),
      }),
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
