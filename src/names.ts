import { NameArgs } from "./types";

export function derivePublisherNames(args: NameArgs): string[] {
  if (args.names !== undefined) {
    if (args.names.length === 0) {
      throw new Error("names must contain at least one publisher name");
    }

    return args.names;
  }

  const replicas = args.replicas ?? 1;
  if (replicas < 1) {
    throw new Error("replicas must be >= 1");
  }

  const namePrefix = args.namePrefix ?? "npa-publisher";
  return Array.from({ length: replicas }, (_, index) => `${namePrefix}-${index + 1}`);
}
