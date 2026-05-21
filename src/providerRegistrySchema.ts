const userDataModesWithoutSingleProperty = new Set(["none", "scalewayDual", "proxmoxSnippet"]);

export interface RegistryProviderEntry {
  componentName: string;
  resourceToken?: string;
  providerPackage?: string;
  registrySchemaChecks?: Array<{
    resourceToken: string;
    propertyPath: string[];
    description: string;
  }>;
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
  types?: Record<string, {
    properties?: Record<string, unknown>;
  }>;
}

export function validateProviderAgainstRegistrySchema(provider: RegistryProviderEntry, schema: RegistrySchema): string[] {
  const errors: string[] = [];

  const expectedPackage = provider.providerPackage;
  const nodePackage = schema.language?.nodejs?.packageName;
  if (expectedPackage && expectedPackage.startsWith("@") && nodePackage && nodePackage !== expectedPackage) {
    errors.push(`${provider.componentName} providerPackage ${expectedPackage} does not match upstream node package ${nodePackage}`);
  }

  if (provider.registrySchemaChecks && provider.registrySchemaChecks.length > 0) {
    for (const check of provider.registrySchemaChecks) {
      const checkedResource = schema.resources?.[check.resourceToken];
      if (!checkedResource) {
        errors.push(`${provider.componentName} upstream schema missing resource token ${check.resourceToken}`);
        continue;
      }
      if (!schemaHasPath(schema, checkedResource.inputProperties ?? {}, check.propertyPath)) {
        errors.push(`${provider.componentName} upstream resource ${check.resourceToken} missing ${check.description} path ${check.propertyPath.join(".")}`);
      }
    }
    return errors;
  }

  if (!provider.resourceToken) {
    return errors;
  }

  const resource = schema.resources?.[provider.resourceToken];
  if (!resource) {
    errors.push(`${provider.componentName} upstream schema missing resource token ${provider.resourceToken}`);
    return errors;
  }

  const userDataProperty = provider.userData?.property;
  const userDataMode = provider.userData?.mode ?? "none";
  if (userDataProperty && !userDataModesWithoutSingleProperty.has(userDataMode) && !resource.inputProperties?.[userDataProperty]) {
    errors.push(`${provider.componentName} upstream resource ${provider.resourceToken} missing user-data property ${userDataProperty}`);
  }

  return errors;
}

function schemaHasPath(schema: RegistrySchema, properties: Record<string, unknown>, path: string[]): boolean {
  let currentProperties: Record<string, unknown> | undefined = properties;
  for (const [index, segment] of path.entries()) {
    const property = currentProperties?.[segment] as { $ref?: string; properties?: Record<string, unknown> } | undefined;
    if (!property) {
      return false;
    }
    if (index === path.length - 1) {
      return true;
    }
    currentProperties = property.properties ?? resolveRefProperties(schema, property.$ref);
  }
  return false;
}

function resolveRefProperties(schema: RegistrySchema, ref: string | undefined): Record<string, unknown> | undefined {
  if (!ref?.startsWith("#/types/")) {
    return undefined;
  }
  const typeToken = ref.slice("#/types/".length);
  return schema.types?.[typeToken]?.properties;
}
