export interface RenderUserDataArgs {
  publisherName: string;
  registrationToken: string;
  wizardPath?: string;
  bootstrap?: boolean;
  bootstrapUrl?: string;
  nonat?: boolean;
}

export function renderUserData(args: RenderUserDataArgs): string {
  const wizardPath = args.wizardPath ?? "/home/ubuntu/npa_publisher_wizard";
  const bootstrap = args.bootstrap ?? true;
  const bootstrapUrl = args.bootstrapUrl ?? "https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/generic/bootstrap.sh";
  const nonat = args.nonat ?? true;

  if (!bootstrap && !nonat) {
    return [
      "#cloud-config",
      `hostname: ${args.publisherName}`,
      "preserve_hostname: false",
      "runcmd:",
      `  - [ ${wizardPath}, -token, "${escapeDoubleQuoted(args.registrationToken)}" ]`,
      "",
    ].join("\n");
  }

  const lines = [
    "#cloud-config",
    `hostname: ${args.publisherName}`,
    "preserve_hostname: false",
    "",
    "system_info:",
    "  default_user:",
    "    name: ubuntu",
    "",
    "users:",
    "  - name: ubuntu",
    "    groups: [sudo]",
    '    sudo: "ALL=(ALL) NOPASSWD:ALL"',
    "    shell: /bin/bash",
    "    lock_passwd: true",
    "",
    "runcmd:",
    "  - chmod 1777 /tmp",
  ];

  if (nonat) {
    lines.push(
      "  - install -d -o ubuntu -g ubuntu -m 0755 /home/ubuntu/resources",
      "  - install -o ubuntu -g ubuntu -m 0644 /dev/null /home/ubuntu/resources/.nonat",
    );
  }

  if (bootstrap) {
    lines.push(`  - su - ubuntu -c 'curl -fsSL ${bootstrapUrl} | sudo bash'`);
  }
  lines.push(`  - su - ubuntu -c 'sudo ${wizardPath} -token "${escapeSingleQuoted(args.registrationToken)}"'`, "");

  return lines.join("\n");
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

function escapeSingleQuoted(value: string): string {
  return value.replace(/'/g, "'\"'\"'");
}
