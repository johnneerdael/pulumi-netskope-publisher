export interface RenderUserDataArgs {
  publisherName: string;
  registrationToken: string;
  wizardPath?: string;
  bootstrap?: boolean;
  bootstrapUrl?: string;
  nonat?: boolean;
  installUser?: string;
  installUserPassword?: string;
  installUserPasswordIsHash?: boolean;
  installUserSshAuthorizedKeys?: string[];
  deleteDefaultUser?: boolean;
  guestNetworkInterface?: CloudInitGuestNetworkInterface;
}

export interface CloudInitGuestNetworkInterface {
  name: string;
  dhcp4?: boolean;
  addresses?: string[];
  gateway4?: string;
  nameservers?: string[];
  mtu?: number;
}

export function renderUserData(args: RenderUserDataArgs): string {
  const installUser = args.installUser ?? "ubuntu";
  const wizardPath = args.wizardPath ?? `/home/${installUser}/npa_publisher_wizard`;
  const bootstrap = args.bootstrap ?? true;
  const bootstrapUrl = args.bootstrapUrl ?? "https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/generic/bootstrap.sh";
  const nonat = args.nonat ?? true;
  const installUserSshAuthorizedKeys = args.installUserSshAuthorizedKeys ?? [];
  const deleteDefaultUser = args.deleteDefaultUser ?? true;

  if (!bootstrap && !nonat && installUser === "ubuntu" && !args.installUserPassword && installUserSshAuthorizedKeys.length === 0 && !args.guestNetworkInterface) {
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
    `    name: ${installUser}`,
    "",
    "users:",
    `  - name: ${installUser}`,
    "    groups: [sudo]",
    '    sudo: "ALL=(ALL) NOPASSWD:ALL"',
    "    shell: /bin/bash",
    `    lock_passwd: ${args.installUserPassword === undefined}`,
  ];

  if (installUserSshAuthorizedKeys.length > 0) {
    lines.push(
      "    ssh_authorized_keys:",
      ...installUserSshAuthorizedKeys.map((key) => `      - "${escapeDoubleQuoted(key)}"`),
    );
  }

  if (args.installUserPassword !== undefined) {
    lines.push(
      "",
      "chpasswd:",
      "  expire: false",
      "  users:",
      `    - name: ${installUser}`,
      `      password: "${escapeDoubleQuoted(args.installUserPassword)}"`,
      `      type: ${args.installUserPasswordIsHash ? "hash" : "text"}`,
      "ssh_pwauth: true",
    );
  }

  if (args.guestNetworkInterface) {
    lines.push("", ...renderNetplan(args.guestNetworkInterface));
  }

  lines.push("", "runcmd:");

  if (args.guestNetworkInterface) {
    lines.push(
      "  - chmod 0600 /etc/netplan/60-cloudinit-override.yaml",
      "  - netplan apply",
    );
  }

  if (deleteDefaultUser && installUser !== "ubuntu") {
    lines.push(
      "  - pkill -KILL -u ubuntu || true",
      "  - userdel -r ubuntu 2>/dev/null || true",
    );
  }

  if (bootstrap || nonat) {
    lines.push("  - chmod 1777 /tmp");
  }

  if (nonat) {
    lines.push(
      `  - install -d -o ${installUser} -g ${installUser} -m 0755 /home/${installUser}/resources`,
      `  - install -o ${installUser} -g ${installUser} -m 0644 /dev/null /home/${installUser}/resources/.nonat`,
    );
  }

  if (bootstrap) {
    lines.push(`  - su - ${installUser} -c 'curl -fsSL ${bootstrapUrl} | sudo bash'`);
  }
  lines.push(`  - su - ${installUser} -c 'sudo ${wizardPath} -token "${escapeSingleQuoted(args.registrationToken)}"'`, "");

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

function renderNetplan(networkInterface: CloudInitGuestNetworkInterface): string[] {
  const lines = [
    "write_files:",
    "  - path: /etc/netplan/60-cloudinit-override.yaml",
    "    owner: root:root",
    '    permissions: "0600"',
    "    content: |",
    "      network:",
    "        version: 2",
    "        ethernets:",
    `          ${networkInterface.name}:`,
    `            dhcp4: ${networkInterface.dhcp4 ?? false}`,
  ];

  if (networkInterface.addresses && networkInterface.addresses.length > 0) {
    lines.push(
      "            addresses:",
      ...networkInterface.addresses.map((address) => `              - ${address}`),
    );
  }
  if (networkInterface.gateway4) {
    lines.push(`            gateway4: ${networkInterface.gateway4}`);
  }
  if (networkInterface.nameservers && networkInterface.nameservers.length > 0) {
    lines.push(
      "            nameservers:",
      "              addresses:",
      ...networkInterface.nameservers.map((nameserver) => `                - ${nameserver}`),
    );
  }
  if (networkInterface.mtu !== undefined) {
    lines.push(`            mtu: ${networkInterface.mtu}`);
  }

  return lines;
}
