import * as azure from "@pulumi/azure-native";
import * as pulumi from "@pulumi/pulumi";
import { renderUserDataBase64 } from "./cloudInit";
import {
  buildNameTag,
  createPublisherOutput,
  createRegistrations,
  resolvePublisherNames,
} from "./componentCore";
import { AzurePublisherArgs, PublisherOutput } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class AzurePublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: AzurePublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:AzurePublisher", name, {}, opts);
    validateComponentArgs("AzurePublisher", args);

    const bootstrap = args.bootstrap ?? false;
    if (args.imageId === undefined && args.marketplace === undefined && bootstrap !== true) {
      throw new Error("Provide imageId, marketplace, or set bootstrap: true.");
    }

    const parentOpts = { parent: this };
    const publisherNames = resolvePublisherNames(args);
    this.publisherNames = pulumi.output(publisherNames);
    const registrations = createRegistrations(name, publisherNames, args, parentOpts);
    const publisherOutputs: Record<string, pulumi.Output<PublisherOutput>> = {};

    const adminUsername = args.adminUsername ?? args.installUser ?? "ubuntu";
    const effectiveMarketplace = args.marketplace ?? (bootstrap === true ? {
      publisher: "Canonical",
      offer: "0001-com-ubuntu-minimal-jammy",
      sku: "minimal-22_04-lts-gen2",
      version: "latest",
    } : undefined);
    const vmSize = args.vmSize ?? "Standard_D2s_v5";
    const assignPublicIp = args.assignPublicIp ?? false;
    const osDisk = pulumi.output(args.osDisk ?? {}).apply((disk) => ({
      type: disk.type ?? "Premium_LRS",
      sizeGb: disk.sizeGb ?? 64,
    }));

    for (const publisherName of publisherNames) {
      const registration = registrations.apply((allRegistrations) => allRegistrations[publisherName]);
      const customData = pulumi.all({
        registration,
        wizardPath: args.wizardPath,
        bootstrap,
        bootstrapUrl: args.bootstrapUrl,
        nonat: args.nonat ?? false,
        installUser: args.installUser,
        installUserPassword: args.installUserPassword,
        installUserPasswordIsHash: args.installUserPasswordIsHash,
        installUserSshAuthorizedKeys: args.installUserSshAuthorizedKeys,
        deleteDefaultUser: args.deleteDefaultUser,
        guestNetworkInterface: args.guestNetworkInterface,
      }).apply((options: any) =>
        renderUserDataBase64({
          publisherName,
          registrationToken: options.registration.registrationToken,
          wizardPath: options.wizardPath,
          bootstrap: options.bootstrap,
          bootstrapUrl: options.bootstrapUrl,
          nonat: options.nonat,
          installUser: options.installUser,
          installUserPassword: options.installUserPassword,
          installUserPasswordIsHash: options.installUserPasswordIsHash,
          installUserSshAuthorizedKeys: options.installUserSshAuthorizedKeys,
          deleteDefaultUser: options.deleteDefaultUser,
          guestNetworkInterface: options.guestNetworkInterface,
        }),
      );

      const publicIp = assignPublicIp
        ? new azure.network.PublicIPAddress(`${name}-${publisherName}-pip`, {
          publicIpAddressName: `${publisherName}-pip`,
          resourceGroupName: args.resourceGroupName,
          location: args.location,
          publicIPAllocationMethod: "Static",
          sku: { name: "Standard" },
          tags: args.tags,
        }, parentOpts)
        : undefined;

      const nic = new azure.network.NetworkInterface(`${name}-${publisherName}-nic`, {
        networkInterfaceName: `${publisherName}-nic`,
        resourceGroupName: args.resourceGroupName,
        location: args.location,
        tags: args.tags,
        ipConfigurations: [{
          name: "primary",
          subnet: { id: args.subnetId },
          privateIPAllocationMethod: "Dynamic",
          publicIPAddress: publicIp ? { id: publicIp.id } : undefined,
        }],
        networkSecurityGroup: args.networkSecurityGroupId ? { id: args.networkSecurityGroupId } : undefined,
      }, parentOpts);

      const vm = new azure.compute.VirtualMachine(`${name}-${publisherName}`, {
        vmName: publisherName,
        resourceGroupName: args.resourceGroupName,
        location: args.location,
        tags: buildNameTag(args.tags, publisherName),
        hardwareProfile: { vmSize },
        networkProfile: {
          networkInterfaces: [{ id: nic.id, primary: true }],
        },
        osProfile: {
          computerName: publisherName,
          adminUsername,
          customData,
          linuxConfiguration: {
            disablePasswordAuthentication: true,
            ssh: {
              publicKeys: [{
                path: pulumi.interpolate`/home/${adminUsername}/.ssh/authorized_keys`,
                keyData: args.adminSshPublicKey,
              }],
            },
          },
        },
        storageProfile: {
          imageReference: pulumi.output(effectiveMarketplace).apply((marketplace) =>
            args.imageId ? { id: args.imageId } : marketplace ? {
              publisher: marketplace.publisher,
              offer: marketplace.offer,
              sku: marketplace.sku,
              version: marketplace.version ?? "latest",
            } : undefined,
          ),
          osDisk: osDisk.apply((disk) => ({
            createOption: "FromImage",
            caching: "ReadWrite",
            managedDisk: { storageAccountType: disk.type },
            diskSizeGB: disk.sizeGb,
          })),
        },
        plan: pulumi.output(args.marketplace).apply((marketplace) =>
          args.imageId ? undefined : marketplace ? {
            publisher: marketplace.publisher,
            product: marketplace.offer,
            name: marketplace.sku,
          } : undefined,
        ),
      } as azure.compute.VirtualMachineArgs, parentOpts);

      publisherOutputs[publisherName] = createPublisherOutput({
        registration,
        vmId: vm.id,
        privateIp: pulumi.output(nic.ipConfigurations).apply((configs) => configs?.[0]?.privateIPAddress ?? ""),
        publicIp: publicIp?.ipAddress,
        placementLabels: args.placementLabels,
      });
    }

    this.publishers = pulumi.secret(pulumi.all(publisherOutputs));
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
