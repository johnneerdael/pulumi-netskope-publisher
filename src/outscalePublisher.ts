import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { OutscalePublisherArgs, PublisherOutput } from "./types";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";

export class OutscalePublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: OutscalePublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:OutscalePublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userData }) => {
      const vm = new RawResource(`${name}-${publisherName}`, "outscale:index/vm:Vm", {
        imageId: args.imageId,
        vmType: args.vmType ?? "tinav5.c2r4p1",
        subnetId: args.subnetId,
        keypairName: args.keypairName,
        securityGroupIds: args.securityGroupIds,
        placementSubregionName: args.placementSubregionName,
        userData: plainUserData(userData),
      }, { parent: this });

      return { vmId: vm.id, privateIp: pulumi.output(""), publicIp: pulumi.output("") };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
