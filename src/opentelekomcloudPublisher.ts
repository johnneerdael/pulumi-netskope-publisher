import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { OpentelekomcloudPublisherArgs, PublisherOutput } from "./types";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";

export class OpentelekomcloudPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: OpentelekomcloudPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:OpentelekomcloudPublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userData }) => {
      const instance = new RawResource(`${name}-${publisherName}`, "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", {
        name: publisherName,
        imageName: args.imageName ?? "Ubuntu 22.04",
        imageId: args.imageId,
        flavorName: args.flavorName ?? "s3.medium.2",
        flavorId: args.flavorId,
        networks: args.networks,
        keyPair: args.keyPair,
        availabilityZone: args.availabilityZone,
        securityGroups: args.securityGroups,
        userData: plainUserData(userData),
      }, { parent: this });

      return { vmId: instance.id, privateIp: pulumi.output(""), publicIp: pulumi.output("") };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
