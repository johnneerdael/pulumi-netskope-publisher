import * as pulumi from "@pulumi/pulumi";
import { NetskopeAuthMode, NetskopeOAuth2Args } from "./types";

export interface PrivateAppProtocol {
  type: pulumi.Input<string>;
  ports: pulumi.Input<string>;
}

export interface PrivateAppArgs {
  tenantUrl: pulumi.Input<string>;
  bearerToken?: pulumi.Input<string>;
  authMode?: pulumi.Input<NetskopeAuthMode>;
  oauth2?: pulumi.Input<NetskopeOAuth2Args>;
  /** @deprecated Use bearerToken instead. */
  apiToken?: pulumi.Input<string>;
  appName: pulumi.Input<string>;
  appType?: pulumi.Input<"client" | "clientless">;
  host: pulumi.Input<string>;
  protocols: pulumi.Input<pulumi.Input<PrivateAppProtocol>[]>;
  clientlessAccess: pulumi.Input<boolean>;
  isUserPortalApp: pulumi.Input<boolean>;
  usePublisherDns: pulumi.Input<boolean>;
  trustSelfSignedCerts: pulumi.Input<boolean>;
  tags?: pulumi.Input<pulumi.Input<string>[]>;
  adoptExisting?: pulumi.Input<boolean>;
}

export class PrivateApp extends pulumi.CustomResource {
  public readonly appId!: pulumi.Output<number>;

  constructor(name: string, args: PrivateAppArgs, opts?: pulumi.CustomResourceOptions) {
    super("netskope-publisher:index:PrivateApp", name, args, opts);
  }
}
