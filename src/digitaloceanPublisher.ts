import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { DigitaloceanPublisherArgs, PublisherOutput } from "./types";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";

export class DigitaloceanPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: DigitaloceanPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:DigitaloceanPublisher", name, {}, opts);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userData }) => {
      const droplet = new RawResource(`${name}-${publisherName}`, "digitalocean:index/droplet:Droplet", {
        name: publisherName,
        region: args.region,
        size: args.size ?? "s-2vcpu-4gb",
        image: args.image ?? "ubuntu-22-04-x64",
        sshKeys: args.sshKeys,
        vpcUuid: args.vpcUuid,
        monitoring: args.monitoring,
        ipv6: args.ipv6,
        userData: plainUserData(userData),
        tags: args.tags === undefined ? undefined : pulumi.output(args.tags).apply((tags) => Object.entries(tags).map(([key, value]) => `${key}:${value}`)),
      }, { parent: this });

      return { vmId: droplet.id, privateIp: pulumi.output(""), publicIp: pulumi.output("") };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
