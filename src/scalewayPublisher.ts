import * as pulumi from "@pulumi/pulumi";
import * as scaleway from "@pulumiverse/scaleway";
import { createVmPublishers } from "./vmPublisherCore";
import { PublisherOutput, ScalewayPublisherArgs } from "./types";
import { scalewayUserData } from "./userDataAdapters";
import { validateComponentArgs } from "./providerValidation";

export class ScalewayPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: ScalewayPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:ScalewayPublisher", name, {}, opts);
    validateComponentArgs("ScalewayPublisher", args);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userData }) => {
      const userDataPlacement = scalewayUserData(userData);
      const server = new scaleway.instance.Server(`${name}-${publisherName}`, {
        name: publisherName,
        type: args.type ?? "DEV1-M",
        image: args.image ?? "ubuntu_jammy",
        zone: args.zone,
        securityGroupId: args.securityGroupId,
        enableDynamicIp: args.enableDynamicIp ?? true,
        cloudInit: userDataPlacement.cloudInit as pulumi.Input<string>,
        userData: userDataPlacement.userData as pulumi.Input<Record<string, pulumi.Input<string>>>,
        tags: pulumi.output(args.tags ?? {}).apply((tags) =>
          Object.entries(tags).map(([key, value]) => `${key}=${value}`),
        ),
      }, { parent: this });

      return {
        vmId: server.id,
        privateIp: server.privateIps.apply((ips) => ips?.[0]?.address ?? ""),
        publicIp: server.publicIps.apply((ips) => ips?.[0]?.address),
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
