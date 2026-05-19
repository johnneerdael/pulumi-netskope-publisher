import * as nutanix from "@pierskarsenbarg/nutanix";
import * as pulumi from "@pulumi/pulumi";
import { createVmPublishers } from "./vmPublisherCore";
import { NutanixPublisherArgs, PublisherOutput } from "./types";

export class NutanixPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: NutanixPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:NutanixPublisher", name, {}, opts);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userDataBase64 }) => {
      const vm = new nutanix.VirtualMachine(`${name}-${publisherName}`, {
        name: publisherName,
        clusterUuid: args.clusterUuid,
        numSockets: args.numVCpus ?? 2,
        numVcpusPerSocket: args.numCoresPerVcpu ?? 1,
        memorySizeMib: args.memorySizeMib ?? 4096,
        diskLists: args.imageUuid === undefined ? undefined : [{
          dataSourceReference: {
            kind: "image",
            uuid: args.imageUuid,
          },
        }],
        nicLists: args.subnetUuid === undefined ? undefined : [{
          subnetUuid: args.subnetUuid,
          nicType: "NORMAL_NIC",
          model: "VIRTIO",
        }],
        guestCustomizationCloudInitUserData: userDataBase64,
      }, { parent: this });

      return {
        vmId: vm.id,
        privateIp: vm.nicListStatuses.apply((statuses) => statuses?.[0]?.ipEndpointLists?.[0]?.ip ?? ""),
        publicIp: vm.nicListStatuses.apply((statuses) => statuses?.[0]?.floatingIp),
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
