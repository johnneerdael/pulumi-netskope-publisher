import * as pulumi from "@pulumi/pulumi";
import { HypervPublisherArgs, PublisherOutput } from "./types";
import { validateComponentArgs } from "./providerValidation";

export class HypervPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, PublisherOutput>>;

  constructor(name: string, args: HypervPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:HypervPublisher", name, {}, opts);
    validateComponentArgs("HypervPublisher", args);

    throw new Error(
      "Hyper-V support requires @pulumi/hyperv from pulumi/pulumi-hyperv because it is not published to npm.",
    );
  }
}
