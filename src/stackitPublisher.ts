import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { PublisherOutput, StackitPublisherArgs } from "./types";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";

export class StackitPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: StackitPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:StackitPublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userData }) => {
      const server = new RawResource(`${name}-${publisherName}`, "stackit:index/server:Server", {
        name: publisherName,
        projectId: args.projectId,
        machineType: args.machineType,
        imageId: args.imageId,
        availabilityZone: args.availabilityZone,
        keypairName: args.keypairName,
        networkInterfaces: args.networkInterfaces,
        userData: plainUserData(userData),
        labels: args.tags,
      }, { parent: this });

      return { vmId: server.id, privateIp: pulumi.output(""), publicIp: pulumi.output("") };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
