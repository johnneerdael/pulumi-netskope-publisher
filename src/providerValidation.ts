import { providerCatalog } from "./providerCatalog";

export function validateProviderArgs(componentName: string, args: Record<string, unknown>): void {
  const provider = providerCatalog[componentName];
  if (!provider) {
    throw new Error(`Unknown provider component ${componentName}`);
  }

  for (const field of provider.validation.required ?? []) {
    if (isMissing(args[field])) {
      throw new Error(`${componentName} requires input ${field}`);
    }
  }

  for (const group of provider.validation.requiredOneOf ?? []) {
    if (!group.some((field) => !isMissing(args[field]))) {
      throw new Error(`${componentName} requires one of: ${group.join(", ")}`);
    }
  }

  for (const group of provider.validation.mutuallyExclusive ?? []) {
    const present = group.filter((field) => !isMissing(args[field]));
    if (present.length > 1) {
      throw new Error(`${componentName} accepts only one of: ${group.join(", ")}`);
    }
  }

  const optInField = provider.validation.experimentalOptInField;
  if (optInField && args[optInField] !== true) {
    throw new Error(`${componentName} requires ${optInField}: true`);
  }
}

function isMissing(value: unknown): boolean {
  return value === undefined || value === null || value === "";
}
