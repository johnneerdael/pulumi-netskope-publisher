import * as pulumi from "@pulumi/pulumi";
import * as vsphere from "@pulumi/vsphere";
import { renderMetadata, renderUserDataBase64 } from "./cloudInit";
import {
  createPublisherOutput,
  createRegistrations,
  resolvePublisherNames,
} from "./componentCore";
import { PublisherOutput, VspherePublisherArgs } from "./types";

export class VspherePublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: VspherePublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:VspherePublisher", name, {}, opts);

    if (args.cluster === undefined && args.host === undefined) {
      throw new Error("Provide either vsphere.cluster or vsphere.host.");
    }

    const parentOpts = { parent: this };
    const publisherNames = resolvePublisherNames(args);
    this.publisherNames = pulumi.output(publisherNames);
    const registrations = createRegistrations(name, publisherNames, args, parentOpts);
    const publisherOutputs: Record<string, pulumi.Output<PublisherOutput>> = {};

    const datacenter = vsphere.getDatacenterOutput({ name: args.datacenter }, parentOpts);
    const datastore = vsphere.getDatastoreOutput({
      name: args.datastore,
      datacenterId: datacenter.id,
    }, parentOpts);
    const network = vsphere.getNetworkOutput({
      name: args.networkName,
      datacenterId: datacenter.id,
    }, parentOpts);
    const template = vsphere.getVirtualMachineOutput({
      name: args.templateName,
      datacenterId: datacenter.id,
    }, parentOpts);
    const resourcePoolId = args.cluster !== undefined
      ? vsphere.getComputeClusterOutput({
        name: args.cluster,
        datacenterId: datacenter.id,
      }, parentOpts).resourcePoolId
      : vsphere.getHostOutput({
        name: args.host!,
        datacenterId: datacenter.id,
      }, parentOpts).resourcePoolId;

    for (const publisherName of publisherNames) {
      const registration = registrations.apply((allRegistrations) => allRegistrations[publisherName]);
      const userData = pulumi.all([registration, args.wizardPath]).apply(([record, wizardPath]) =>
        renderUserDataBase64({
          publisherName,
          registrationToken: record.registrationToken,
          wizardPath,
        }),
      );
      const metadata = Buffer.from(renderMetadata(publisherName), "utf8").toString("base64");
      const firstDisk = template.disks.apply((disks) => disks[0]);

      const vm = new vsphere.VirtualMachine(`${name}-${publisherName}`, {
        name: publisherName,
        resourcePoolId,
        datastoreId: datastore.id,
        folder: args.folder,
        numCpus: args.numCpus ?? 2,
        memory: args.memory ?? 4096,
        guestId: template.guestId,
        networkInterfaces: [{
          networkId: network.id,
          adapterType: template.networkInterfaceTypes.apply((types) => types?.[0]),
        }],
        disks: [{
          label: "disk0",
          size: firstDisk.size,
          eagerlyScrub: firstDisk.eagerlyScrub,
          thinProvisioned: firstDisk.thinProvisioned,
        }],
        clone: {
          templateUuid: template.id,
        },
        extraConfig: {
          "guestinfo.userdata": userData,
          "guestinfo.userdata.encoding": "base64",
          "guestinfo.metadata": metadata,
          "guestinfo.metadata.encoding": "base64",
        },
      }, parentOpts);

      publisherOutputs[publisherName] = createPublisherOutput({
        registration,
        vmId: vm.id,
        privateIp: vm.defaultIpAddress,
        publicIp: pulumi.output(undefined),
      });
    }

    this.publishers = pulumi.secret(pulumi.all(publisherOutputs));
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
