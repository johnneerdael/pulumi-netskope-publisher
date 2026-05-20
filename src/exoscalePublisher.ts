import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { ExoscalePublisherArgs, PublisherOutput } from "./types";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";

export class ExoscalePublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: ExoscalePublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:ExoscalePublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userData }) => {
      const instance = new RawResource(`${name}-${publisherName}`, "exoscale:index/computeInstance:ComputeInstance", {
        name: publisherName,
        zone: args.zone,
        type: args.type,
        templateId: args.templateId,
        diskSize: args.diskSize,
        sshKeys: args.sshKeys,
        securityGroupIds: args.securityGroupIds,
        networkInterfaces: args.networkInterfaces,
        userData: plainUserData(userData),
        labels: args.tags,
      }, { parent: this });

      return { vmId: instance.id, privateIp: pulumi.output(""), publicIp: pulumi.output("") };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
