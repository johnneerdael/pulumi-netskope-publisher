import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { EquinixPublisherArgs, PublisherOutput } from "./types";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";

export class EquinixPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: EquinixPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:EquinixPublisher", name, {}, opts);

    const outputs = createVmPublishers({ parent: this, componentName: name, args, forceBootstrap: true }, ({ publisherName, userData }) => {
      const device = new RawResource(`${name}-${publisherName}`, "equinix:metal/device:Device", {
        hostname: publisherName,
        projectId: args.projectId,
        metro: args.metro,
        plan: args.plan,
        operatingSystem: args.operatingSystem ?? "ubuntu_22_04",
        billingCycle: args.billingCycle ?? "hourly",
        projectSshKeyIds: args.projectSshKeyIds,
        userSshKeyIds: args.userSshKeyIds,
        userData: plainUserData(userData),
        tags: args.tags === undefined ? undefined : pulumi.output(args.tags).apply((tags) => Object.entries(tags).map(([key, value]) => `${key}:${value}`)),
      }, { parent: this });

      return { vmId: device.id, privateIp: pulumi.output(""), publicIp: pulumi.output("") };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
