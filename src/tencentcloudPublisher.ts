import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { PublisherOutput, TencentcloudPublisherArgs } from "./types";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";

export class TencentcloudPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: TencentcloudPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:TencentcloudPublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userData }) => {
      const instance = new RawResource(`${name}-${publisherName}`, "tencentcloud:index/instance:Instance", {
        instanceName: publisherName,
        hostname: publisherName,
        availabilityZone: args.availabilityZone,
        imageId: args.imageId,
        instanceType: args.instanceType ?? "S5.MEDIUM4",
        subnetId: args.subnetId,
        vpcId: args.vpcId,
        keyName: args.keyName,
        securityGroups: args.securityGroups,
        systemDiskType: args.systemDiskType,
        systemDiskSize: args.systemDiskSize,
        userDataRaw: plainUserData(userData),
        userDataReplaceOnChange: true,
        tags: args.tags,
      }, { parent: this });

      return { vmId: instance.id, privateIp: pulumi.output(""), publicIp: pulumi.output("") };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
