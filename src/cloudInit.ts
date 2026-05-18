export interface RenderUserDataArgs {
  publisherName: string;
  registrationToken: string;
  wizardPath?: string;
}

export function renderUserData(args: RenderUserDataArgs): string {
  const wizardPath = args.wizardPath ?? "/home/ubuntu/npa_publisher_wizard";

  return [
    "#cloud-config",
    `hostname: ${args.publisherName}`,
    "preserve_hostname: false",
    "runcmd:",
    `  - [ ${wizardPath}, -token, "${escapeDoubleQuoted(args.registrationToken)}" ]`,
    "",
  ].join("\n");
}

export function renderUserDataBase64(args: RenderUserDataArgs): string {
  return Buffer.from(renderUserData(args), "utf8").toString("base64");
}

export function renderMetadata(publisherName: string): string {
  return [
    `instance-id: ${publisherName}`,
    `local-hostname: ${publisherName}`,
    "",
  ].join("\n");
}

function escapeDoubleQuoted(value: string): string {
  return value.replace(/\\/g, "\\\\").replace(/"/g, '\\"');
}
