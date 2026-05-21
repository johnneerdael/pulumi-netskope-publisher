const userDataModesWithoutSingleProperty = new Set(["none", "scalewayDual", "proxmoxSnippet"]);

export interface RegistryProviderEntry {
  componentName: string;
  resourceToken?: string;
  providerPackage?: string;
  registrySchemaChecks?: Array<{
    resourceToken: string;
    propertyPath: string[];
    description: string;
    propertyKind?: "input" | "output";
  }>;
  upstreamPropertyChecks?: Array<{
    resourceToken: string;
    propertyPath: string[];
    description: string;
    propertyKind?: "input" | "output";
  }>;
  userData?: {
    mode: string;
    property?: string;
  };
}

export interface RegistrySchemaResource {
  inputProperties?: Record<string, unknown>;
  properties?: Record<string, unknown>;
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
      const error = checkRegistryPath(provider, schema, check);
      if (error) {
        errors.push(error);
      }
    }
  }

  if (provider.upstreamPropertyChecks && provider.upstreamPropertyChecks.length > 0) {
    for (const check of provider.upstreamPropertyChecks) {
      const error = checkRegistryPath(provider, schema, check);
      if (error) {
        errors.push(error);
      }
    }
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

function checkRegistryPath(
  provider: RegistryProviderEntry,
  schema: RegistrySchema,
  check: {
    resourceToken: string;
    propertyPath: string[];
    description: string;
    propertyKind?: "input" | "output";
  },
): string | undefined {
  const checkedResource = schema.resources?.[check.resourceToken];
  if (!checkedResource) {
    return `${provider.componentName} upstream schema missing resource token ${check.resourceToken}`;
  }

  const propertyKind = check.propertyKind ?? "input";
  const properties = propertyKind === "output"
    ? checkedResource.properties ?? {}
    : checkedResource.inputProperties ?? {};

  if (!schemaHasPath(schema, properties, check.propertyPath)) {
    return `${provider.componentName} upstream resource ${check.resourceToken} missing ${propertyKind} ${check.description} path ${check.propertyPath.join(".")}`;
  }

  return undefined;
}

function schemaHasPath(schema: RegistrySchema, properties: Record<string, unknown>, path: string[]): boolean {
  let currentProperties: Record<string, unknown> | undefined = properties;
  for (const [index, segment] of path.entries()) {
    const property = currentProperties?.[segment] as { $ref?: string; properties?: Record<string, unknown>; items?: { $ref?: string; properties?: Record<string, unknown> } } | undefined;
    if (!property) {
      return false;
    }
    if (index === path.length - 1) {
      return true;
    }
    currentProperties = property.properties
      ?? resolveRefProperties(schema, property.$ref)
      ?? property.items?.properties
      ?? resolveRefProperties(schema, property.items?.$ref);
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
