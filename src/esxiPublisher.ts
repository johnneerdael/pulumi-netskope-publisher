import * as pulumi from "@pulumi/pulumi";
import * as esxi from "@pulumiverse/esxi-native";
import { createVmPublishers } from "./vmPublisherCore";
import { EsxiPublisherArgs, PublisherOutput } from "./types";

export class EsxiPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: EsxiPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:EsxiPublisher", name, {}, opts);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userDataBase64 }) => {
      const vm = new esxi.VirtualMachine(`${name}-${publisherName}`, {
        name: publisherName,
        diskStore: args.diskStore,
        os: args.os ?? "ubuntu-64",
        memSize: args.memory ?? 4096,
        numVCpus: args.numVCpus ?? 2,
        bootDiskSize: args.diskSize ?? 64,
        networkInterfaces: [{
          virtualNetwork: args.virtualNetwork,
        }],
        info: [{
          key: "guestinfo.userdata",
          value: userDataBase64,
        }, {
          key: "guestinfo.userdata.encoding",
          value: "base64",
        }],
        power: "on",
      }, { parent: this });

      return {
        vmId: vm.id,
        privateIp: pulumi.output(""),
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
