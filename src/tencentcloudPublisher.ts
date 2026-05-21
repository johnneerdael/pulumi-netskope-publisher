import * as pulumi from "@pulumi/pulumi";
import { createCatalogRawVmPublishers, userDataProperty } from "./catalogVmFactory";
import { providerCatalog } from "./providerCatalog";
import { PublisherOutput, TencentcloudPublisherArgs } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class TencentcloudPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: TencentcloudPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:TencentcloudPublisher", name, {}, opts);
    validateComponentArgs("TencentcloudPublisher", args);

    const provider = providerCatalog.TencentcloudPublisher;
    const outputs = createCatalogRawVmPublishers({
      parent: this,
      componentName: name,
      provider,
      args,
      mapInputs: (input, currentArgs) => ({
        instanceName: input.publisherName,
        hostname: input.publisherName,
        availabilityZone: currentArgs.availabilityZone,
        imageId: currentArgs.imageId,
        instanceType: currentArgs.instanceType ?? "S5.MEDIUM4",
        subnetId: currentArgs.subnetId,
        vpcId: currentArgs.vpcId,
        keyName: currentArgs.keyName,
        securityGroups: currentArgs.securityGroups,
        systemDiskType: currentArgs.systemDiskType,
        systemDiskSize: currentArgs.systemDiskSize,
        ...userDataProperty(provider, input),
        userDataReplaceOnChange: true,
        tags: currentArgs.tags,
      }),
      mapOutputs: (resource) => ({
        vmId: resource.id,
        privateIp: resource.output<string>("privateIp"),
        publicIp: resource.output<string>("publicIp"),
      }),
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
