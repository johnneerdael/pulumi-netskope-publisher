import * as oci from "@pulumi/oci";
import * as pulumi from "@pulumi/pulumi";
import { base64UserData } from "./userDataAdapters";
import { createVmPublishers } from "./vmPublisherCore";
import { OciPublisherArgs, PublisherOutput } from "./types";

export class OciPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: OciPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:OciPublisher", name, {}, opts);

    const outputs = createVmPublishers({
      parent: this,
      componentName: name,
      args,
      forceBootstrap: true,
    }, ({ publisherName, userData }) => {
      const instance = new oci.core.Instance(`${name}-${publisherName}`, {
        displayName: publisherName,
        compartmentId: args.compartmentId,
        availabilityDomain: args.availabilityDomain,
        shape: args.shape ?? "VM.Standard.E4.Flex",
        createVnicDetails: {
          subnetId: args.subnetId,
          assignPublicIp: pulumi.output(args.assignPublicIp ?? false).apply((value) => String(value)),
          displayName: `${publisherName}-vnic`,
        },
        sourceDetails: {
          sourceType: "image",
          sourceId: args.imageId,
        },
        metadata: pulumi.all([base64UserData(userData), args.sshPublicKey]).apply(([encodedUserData, sshPublicKey]) => ({
          userData: encodedUserData,
          ...(sshPublicKey === undefined ? {} : { ssh_authorized_keys: sshPublicKey }),
        })),
        freeformTags: args.tags,
      }, { parent: this });

      return {
        vmId: instance.id,
        privateIp: instance.privateIp,
        publicIp: instance.publicIp,
      };
    });

    this.publisherNames = outputs.publisherNames;
    this.publishers = outputs.publishers;
    this.registerOutputs({ publisherNames: this.publisherNames, publishers: this.publishers });
  }
}
