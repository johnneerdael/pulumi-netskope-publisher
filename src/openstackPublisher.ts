import * as openstack from "@pulumi/openstack";
import * as pulumi from "@pulumi/pulumi";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";
import { OpenstackPublisherArgs, PublisherOutput } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class OpenstackPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: OpenstackPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:OpenstackPublisher", name, {}, opts);
    validateComponentArgs("OpenstackPublisher", args);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userData }) => {
      const instance = new openstack.compute.Instance(`${name}-${publisherName}`, {
        name: publisherName,
        imageName: args.imageName,
        flavorName: args.flavorName,
        networks: [{ name: args.networkName }],
        keyPair: args.keyPair,
        securityGroups: args.securityGroups,
        availabilityZone: args.availabilityZone,
        userData: plainUserData(userData),
      }, { parent: this });

      const floatingIp = args.assignFloatingIp === true
        ? new openstack.networking.FloatingIp(`${name}-${publisherName}-fip`, {
          pool: args.floatingIpPool,
        }, { parent: this })
        : undefined;

      if (floatingIp) {
        new openstack.networking.FloatingIpAssociate(`${name}-${publisherName}-fip-association`, {
          floatingIp: floatingIp.address,
          portId: instance.networks.apply((networks) => networks[0]?.port ?? ""),
        }, { parent: this });
      }

      return {
        vmId: instance.id,
        privateIp: instance.networks.apply((networks) => networks[0]?.fixedIpV4 ?? ""),
        publicIp: floatingIp?.address ?? instance.accessIpV4,
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
