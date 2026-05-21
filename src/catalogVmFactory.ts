import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "./rawResource";
import { userDataAdapters } from "./userDataAdapters";
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
  if (provider.userData.mode === "scalewayDual" || provider.userData.mode === "guestInfo" || provider.userData.mode === "ociMetadata" || provider.userData.mode === "customData") {
    throw new Error(`${provider.componentName} cannot use catalog raw VM factory with user-data mode ${provider.userData.mode}`);
  }

  const adapter = userDataAdapters[provider.userData.mode];
  if (!adapter) {
    throw new Error(`${provider.componentName} cannot use catalog raw VM factory with user-data mode ${provider.userData.mode}`);
  }

  const property = provider.userData.property;
  const rendered = adapter(input.userData, provider.userData.metadataKey);

  if (provider.userData.mode === "metadata") {
    return { [property ?? "metadata"]: rendered };
  }

  return { [property ?? "userData"]: rendered };
}
