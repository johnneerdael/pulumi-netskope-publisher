import * as pulumi from "@pulumi/pulumi";
import { NetskopeAuthMode, NetskopeOAuth2Args } from "./types";

export interface RealtimeProtectionPolicyArgs {
  tenantUrl: pulumi.Input<string>;
  bearerToken?: pulumi.Input<string>;
  authMode?: pulumi.Input<NetskopeAuthMode>;
  oauth2?: pulumi.Input<NetskopeOAuth2Args>;
  /** @deprecated Use bearerToken instead. */
  apiToken?: pulumi.Input<string>;
  name: pulumi.Input<string>;
  policyGroupId?: pulumi.Input<number>;
  policyGroupName?: pulumi.Input<string>;
  appIds?: pulumi.Input<pulumi.Input<number>[]>;
  appTags?: pulumi.Input<pulumi.Input<string>[]>;
  users?: pulumi.Input<pulumi.Input<string>[]>;
  groups?: pulumi.Input<pulumi.Input<string>[]>;
  action: pulumi.Input<string>;
  enabled: pulumi.Input<boolean>;
}

export class RealtimeProtectionPolicy extends pulumi.CustomResource {
  public readonly policyId!: pulumi.Output<number>;
  public readonly resolvedPolicyGroupId!: pulumi.Output<number>;

  constructor(name: string, args: RealtimeProtectionPolicyArgs, opts?: pulumi.CustomResourceOptions) {
    super("netskope-publisher:index:RealtimeProtectionPolicy", name, args, opts);
  }
}
