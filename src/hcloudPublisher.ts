import * as hcloud from "@pulumi/hcloud";
import * as pulumi from "@pulumi/pulumi";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";
import { HcloudPublisherArgs, PublisherOutput } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class HcloudPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: HcloudPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:HcloudPublisher", name, {}, opts);
    validateComponentArgs("HcloudPublisher", args);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userData }) => {
      const server = new hcloud.Server(`${name}-${publisherName}`, {
        name: publisherName,
        serverType: args.serverType ?? "cx22",
        image: args.image ?? "ubuntu-22.04",
        location: args.location,
        datacenter: args.datacenter,
        sshKeys: args.sshKeys,
        firewallIds: args.firewallIds,
        publicNets: [{
          ipv4Enabled: args.assignPublicIp ?? true,
          ipv6Enabled: args.assignPublicIp ?? true,
        }],
        networks: args.networkId === undefined ? undefined : [{
          networkId: args.networkId,
        }],
        userData: plainUserData(userData),
        labels: args.tags,
      }, { parent: this });

      return {
        vmId: server.id,
        privateIp: server.networks.apply((networks) => networks?.[0]?.ip ?? ""),
        publicIp: server.ipv4Address,
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
