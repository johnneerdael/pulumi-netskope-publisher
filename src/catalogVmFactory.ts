import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { validateProviderArgs } from "./providerValidation";
import { base64UserData, metadataUserData, plainUserData } from "./userDataAdapters";
import { CommonPublisherArgs, PublisherOutput } from "./types";
import { ProviderCatalogEntry } from "./providerCatalog";
import { createVmPublishers, VmPublisherBuildInput, VmPublisherBuildResult } from "./vmPublisherCore";

export interface CatalogRawVmComponentArgs<TArgs extends CommonPublisherArgs> {
  parent: pulumi.ComponentResource;
  componentName: string;
  provider: ProviderCatalogEntry;
  args: TArgs;
  mapInputs: (input: VmPublisherBuildInput, args: TArgs) => pulumi.Inputs;
  mapOutputs?: (resource: RawResource) => VmPublisherBuildResult;
}

export function createCatalogRawVmPublishers<TArgs extends CommonPublisherArgs>(
  options: CatalogRawVmComponentArgs<TArgs>,
): {
  publisherNames: pulumi.Output<string[]>;
  publishers: pulumi.Output<Record<string, PublisherOutput>>;
} {
  validateProviderArgs(options.provider.componentName, options.args as Record<string, unknown>);

  return createVmPublishers({
    parent: options.parent,
    componentName: options.componentName,
    args: options.args,
    forceBootstrap: options.provider.bootstrapModel === "bootstrapOnly",
  }, (input) => {
    const resource = new RawResource(
      `${options.componentName}-${input.publisherName}`,
      options.provider.resourceToken!,
      options.mapInputs(input, options.args),
      { parent: options.parent },
    );
    return options.mapOutputs?.(resource) ?? {
      vmId: resource.id,
      privateIp: pulumi.output(""),
      publicIp: pulumi.output(""),
    };
  });
}

export function userDataProperty(provider: ProviderCatalogEntry, input: VmPublisherBuildInput): Record<string, pulumi.Input<unknown>> {
  if (provider.userData.mode === "plain") {
    return { [provider.userData.property ?? "userData"]: plainUserData(input.userData) };
  }
  if (provider.userData.mode === "base64") {
    return { [provider.userData.property ?? "userData"]: base64UserData(input.userData) };
  }
  if (provider.userData.mode === "metadata") {
    return {
      [provider.userData.property ?? "metadata"]: metadataUserData(input.userData, provider.userData.metadataKey ?? "user-data"),
    };
  }
  if (provider.userData.mode === "raw") {
    return { [provider.userData.property ?? "userDataRaw"]: plainUserData(input.userData) };
  }
  throw new Error(`${provider.componentName} cannot use catalog raw VM factory with user-data mode ${provider.userData.mode}`);
}
