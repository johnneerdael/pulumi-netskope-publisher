import * as alicloud from "@pulumi/alicloud";
import * as pulumi from "@pulumi/pulumi";
import { AlicloudPublisherArgs, PublisherOutput } from "./types";
import { base64UserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";
import { validateComponentArgs } from "./providerValidation";

export class AlicloudPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: AlicloudPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:AlicloudPublisher", name, {}, opts);
    validateComponentArgs("AlicloudPublisher", args);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userData }) => {
      const instance = new alicloud.ecs.Instance(`${name}-${publisherName}`, {
        instanceName: publisherName,
        instanceType: args.instanceType ?? "ecs.t6-c1m2.large",
        imageId: args.imageId,
        vswitchId: args.vswitchId,
        securityGroups: args.securityGroupIds,
        keyName: args.keyName,
        internetMaxBandwidthOut: args.allocatePublicIp === true ? 10 : 0,
        userData: base64UserData(userData),
        tags: args.tags,
      }, { parent: this });

      return {
        vmId: instance.id,
        privateIp: instance.primaryIpAddress,
        publicIp: instance.publicIp,
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
