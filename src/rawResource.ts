import * as pulumi from "@pulumi/pulumi";

export class RawResource extends pulumi.CustomResource {
  constructor(name: string, type: string, args: pulumi.Inputs, opts?: pulumi.CustomResourceOptions) {
    super(type, name, args, opts);
  }
}
