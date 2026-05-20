import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { PublisherOutput, UpcloudPublisherArgs } from "./types";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";

const ubuntu2204Template = "01000000-0000-4000-8000-000030220200";

export class UpcloudPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: UpcloudPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:UpcloudPublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userData }) => {
      const server = new RawResource(`${name}-${publisherName}`, "upcloud:index/server:Server", {
        hostname: args.hostname ?? publisherName,
        title: publisherName,
        zone: args.zone,
        plan: args.plan ?? "2xCPU-4GB",
        template: args.template ?? ubuntu2204Template,
        networkInterfaces: args.networkInterfaces,
        metadata: true,
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
