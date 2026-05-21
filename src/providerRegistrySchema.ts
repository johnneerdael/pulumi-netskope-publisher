const userDataModesWithoutSingleProperty = new Set(["none", "scalewayDual", "proxmoxSnippet"]);

export interface RegistryProviderEntry {
  componentName: string;
  resourceToken?: string;
  providerPackage?: string;
  userData?: {
    mode: string;
    property?: string;
  };
}

export interface RegistrySchemaResource {
  inputProperties?: Record<string, unknown>;
}

export interface RegistrySchema {
  name?: string;
  language?: {
    nodejs?: {
      packageName?: string;
    };
  };
  resources?: Record<string, RegistrySchemaResource>;
}

export function validateProviderAgainstRegistrySchema(provider: RegistryProviderEntry, schema: RegistrySchema): string[] {
  const errors: string[] = [];

  if (!provider.resourceToken) {
    return errors;
  }

  const resource = schema.resources?.[provider.resourceToken];
  if (!resource) {
    errors.push(`${provider.componentName} upstream schema missing resource token ${provider.resourceToken}`);
    return errors;
  }

  const expectedPackage = provider.providerPackage;
  const nodePackage = schema.language?.nodejs?.packageName;
  if (expectedPackage && expectedPackage.startsWith("@") && nodePackage && nodePackage !== expectedPackage) {
    errors.push(`${provider.componentName} providerPackage ${expectedPackage} does not match upstream node package ${nodePackage}`);
  }

  const userDataProperty = provider.userData?.property;
  const userDataMode = provider.userData?.mode ?? "none";
  if (userDataProperty && !userDataModesWithoutSingleProperty.has(userDataMode) && !resource.inputProperties?.[userDataProperty]) {
    errors.push(`${provider.componentName} upstream resource ${provider.resourceToken} missing user-data property ${userDataProperty}`);
  }

  return errors;
}
