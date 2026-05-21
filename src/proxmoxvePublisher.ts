import * as proxmoxve from "@muhlba91/pulumi-proxmoxve";
import * as pulumi from "@pulumi/pulumi";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";
import { ProxmoxvePublisherArgs, PublisherOutput } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class ProxmoxvePublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: ProxmoxvePublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:ProxmoxvePublisher", name, {}, opts);
    validateComponentArgs("ProxmoxvePublisher", args);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userData }) => {
      const userDataFile = new proxmoxve.FileLegacy(`${name}-${publisherName}-user-data`, {
        contentType: "snippets",
        datastoreId: args.datastoreId,
        nodeName: args.nodeName,
        sourceRaw: {
          data: plainUserData(userData),
          fileName: `${publisherName}-user-data.yaml`,
        },
      }, { parent: this });

      const vm = new proxmoxve.VmLegacy(`${name}-${publisherName}`, {
        name: publisherName,
        nodeName: args.nodeName,
        vmId: args.vmId,
        clone: {
          vmId: args.templateVmId,
          nodeName: args.cloneNodeName,
          datastoreId: args.datastoreId,
          full: args.fullClone ?? true,
        },
        agent: { enabled: true },
        cpu: { cores: args.cpuCores ?? 2 },
        memory: { dedicated: args.memory ?? 4096 },
        disks: args.diskSize === undefined ? undefined : [{
          datastoreId: args.datastoreId,
          interface: "scsi0",
          size: args.diskSize,
        }],
        networkDevices: [{
          bridge: args.networkBridge ?? "vmbr0",
          model: args.networkModel ?? "virtio",
          vlanId: args.vlanId,
        }],
        initialization: {
          datastoreId: args.datastoreId,
          userDataFileId: userDataFile.id,
          ipConfigs: [{
            ipv4: {
              address: args.ipAddress ?? "dhcp",
              gateway: args.gateway,
            },
          }],
          dns: args.nameservers === undefined ? undefined : {
            servers: args.nameservers,
          },
        },
        onBoot: args.onBoot ?? true,
        operatingSystem: { type: "l26" },
        poolId: args.poolId,
        started: args.started ?? true,
        tags: pulumi.output(args.tags ?? {}).apply((tags) =>
          Object.entries(tags).map(([key, value]) => `${key}=${value}`),
        ),
      }, { parent: this });

      return {
        vmId: vm.id,
        privateIp: vm.ipv4Addresses.apply((addresses) => addresses?.[0]?.[0] ?? ""),
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
