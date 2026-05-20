import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { PublisherOutput, VultrPublisherArgs } from "./types";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";

export class VultrPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: VultrPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:VultrPublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userData }) => {
      const instance = new RawResource(`${name}-${publisherName}`, "vultr:index/instance:Instance", {
        label: publisherName,
        hostname: publisherName,
        region: args.region,
        plan: args.plan,
        osId: args.osId,
        imageId: args.imageId,
        sshKeyIds: args.sshKeyIds,
        vpc2Ids: args.vpc2Ids,
        enableIpv6: args.enableIpv6,
        firewallGroupId: args.firewallGroupId,
        userData: plainUserData(userData),
        tags: args.tags === undefined ? undefined : pulumi.output(args.tags).apply((tags) => Object.entries(tags).map(([key, value]) => `${key}:${value}`)),
      }, { parent: this });

      return { vmId: instance.id, privateIp: pulumi.output(""), publicIp: pulumi.output("") };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
