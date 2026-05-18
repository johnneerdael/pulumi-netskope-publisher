import * as pulumi from "@pulumi/pulumi";
import { HypervPublisherArgs, PublisherOutput } from "./types";

export class HypervPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: HypervPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope:index:HypervPublisher", name, {}, opts);

    if (args.enableExperimentalHyperv !== true) {
      throw new Error("Hyper-V support is experimental and requires enableExperimentalHyperv: true");
    }

    throw new Error(
      "Hyper-V support requires @pulumi/hyperv from pulumi/pulumi-hyperv because it is not published to npm.",
    );
  }
}
