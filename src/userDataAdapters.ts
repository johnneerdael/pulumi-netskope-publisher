import * as pulumi from "@pulumi/pulumi";
import type { UserDataMode } from "./providerCatalog";

export type { UserDataMode } from "./providerCatalog";

export interface PublisherUserDataAdapter {
  mode: UserDataMode;
  maxBytes?: number;
  render(payload: pulumi.Output<string>): pulumi.Input<string> | Record<string, pulumi.Input<unknown>>;
}

export function plainUserData(payload: pulumi.Output<string>): pulumi.Output<string> {
  return payload;
}

export function base64UserData(payload: pulumi.Output<string>): pulumi.Output<string> {
  return payload.apply((value) => Buffer.from(value, "utf8").toString("base64"));
}

export function metadataUserData(payload: pulumi.Output<string>, key = "user-data"): Record<string, pulumi.Input<string>> {
  return { [key]: payload };
}

export function base64MetadataUserData(payload: pulumi.Output<string>, key = "userData"): Record<string, pulumi.Input<string>> {
  return { [key]: base64UserData(payload) };
}

export function customData(payload: pulumi.Output<string>): pulumi.Output<string> {
  return base64UserData(payload);
}

export function guestInfoUserData(payload: pulumi.Output<string>): Record<string, pulumi.Input<string>> {
  return {
    "guestinfo.userdata": base64UserData(payload),
    "guestinfo.userdata.encoding": "base64",
  };
}

export function scalewayUserData(payload: pulumi.Output<string>): Record<string, pulumi.Input<unknown>> {
  return {
    cloudInit: payload,
    userData: {
      "cloud-init": payload,
    },
  };
}

export const userDataAdapters: Partial<Record<UserDataMode, (payload: pulumi.Output<string>, key?: string) => pulumi.Input<string> | Record<string, pulumi.Input<unknown>>>> = {
  plain: (payload) => plainUserData(payload),
  base64: (payload) => base64UserData(payload),
  metadata: (payload, key = "user-data") => metadataUserData(payload, key),
  raw: (payload) => plainUserData(payload),
  customData: (payload) => customData(payload),
  guestInfo: (payload) => guestInfoUserData(payload),
  scalewayDual: (payload) => scalewayUserData(payload),
  ociMetadata: (payload, key = "userData") => base64MetadataUserData(payload, key),
};
