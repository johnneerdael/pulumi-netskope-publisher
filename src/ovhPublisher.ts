import * as ovh from "@ovhcloud/pulumi-ovh";
import * as pulumi from "@pulumi/pulumi";
import { plainUserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";
import { OvhPublisherArgs, PublisherOutput } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class OvhPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: OvhPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:OvhPublisher", name, {}, opts);
    validateComponentArgs("OvhPublisher", args);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userData }) => {
      const instance = new ovh.cloudproject.Instance(`${name}-${publisherName}`, {
        serviceName: args.serviceName,
        name: publisherName,
        region: args.region,
        billingPeriod: "hourly",
        bootFrom: {
          imageId: args.imageId,
        },
        flavor: {
          flavorId: args.flavorId,
        },
        network: args.networkId === undefined ? {
          public: true,
        } : {
          public: true,
          private: {
            network: {
              id: args.networkId,
            },
          },
        },
        sshKey: args.sshKeyName === undefined ? undefined : {
          name: args.sshKeyName,
        },
        userData: plainUserData(userData),
      }, { parent: this });

      return {
        vmId: instance.id,
        privateIp: pulumi.output(""),
        publicIp: instance.addresses.apply((addresses) => addresses?.[0]?.ip),
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
