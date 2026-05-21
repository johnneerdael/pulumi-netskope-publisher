import * as pulumi from "@pulumi/pulumi";

const outputPropertyNames = [
  "accessIpV4",
  "accessPrivateIpv4",
  "accessPublicIpv4",
  "addresses",
  "internalIp",
  "ipv4Address",
  "ipv4AddressPrivate",
  "mainIp",
  "privateIp",
  "publicIp",
  "publicIpAddress",
];

export class RawResource extends pulumi.CustomResource {
  constructor(name: string, type: string, args: pulumi.Inputs, opts?: pulumi.CustomResourceOptions) {
    const props = { ...args };
    for (const propertyName of outputPropertyNames) {
      props[propertyName] ??= undefined;
    }
    super(type, name, props, opts);
  }

  output<T>(name: string): pulumi.Output<T> {
    return (this as unknown as Record<string, pulumi.Output<T>>)[name] ?? pulumi.output(undefined as T);
  }
}
