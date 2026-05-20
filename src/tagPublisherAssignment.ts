import * as pulumi from "@pulumi/pulumi";
import { NetskopeAuthMode, NetskopeOAuth2Args, PublisherOutput } from "./types";

export interface PublisherAssignmentInput {
  publisherId: pulumi.Input<number>;
  placementLabels?: pulumi.Input<pulumi.Input<string>[]>;
}

export interface TagPublisherAssignmentArgs {
  tenantUrl: pulumi.Input<string>;
  bearerToken?: pulumi.Input<string>;
  authMode?: pulumi.Input<NetskopeAuthMode>;
  oauth2?: pulumi.Input<NetskopeOAuth2Args>;
  /** @deprecated Use bearerToken instead. */
  apiToken?: pulumi.Input<string>;
  appTags: pulumi.Input<pulumi.Input<string>[]>;
  publisherPlacementLabels: pulumi.Input<pulumi.Input<string>[]>;
  publishers: pulumi.Input<Record<string, pulumi.Input<PublisherAssignmentInput | PublisherOutput>>>;
  matchMode?: pulumi.Input<"any" | "all">;
}

export class TagPublisherAssignment extends pulumi.CustomResource {
  public readonly matchedApps!: pulumi.Output<string[]>;
  public readonly selectedPublishers!: pulumi.Output<number[]>;

  constructor(name: string, args: TagPublisherAssignmentArgs, opts?: pulumi.CustomResourceOptions) {
    super("netskope-publisher:index:TagPublisherAssignment", name, args, opts);
  }
}
