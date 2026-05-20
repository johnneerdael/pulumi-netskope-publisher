import { mkdirSync, writeFileSync } from "node:fs";
import { spawnSync } from "node:child_process";

const build = spawnSync("npm", ["run", "build"], { stdio: "inherit" });
if (build.status !== 0) {
  process.exit(build.status ?? 1);
}

const { catalogProviders } = await import("../dist/src/providerCatalog.js");

mkdirSync("site/source/_generated/component-yaml", { recursive: true });

const publisherProviders = catalogProviders.filter((provider) => provider.componentName.endsWith("Publisher"));

const providerRows = [
  "| Platform | Component | Status | User-data mode |",
  "|---|---|---|---|",
  ...publisherProviders.map(
    (provider) => `| ${provider.displayName} | \`${provider.componentName}\` | ${provider.support} | ${provider.userData.mode} |`,
  ),
].join("\n");
writeFileSync("site/source/_generated/provider-matrix.md", `${providerRows}\n`);

const componentLinks = publisherProviders
  .map((provider) => `- [${provider.displayName}](/pulumi-netskope-publisher/admin/component/${provider.docs.slug}/)`)
  .join("\n");
writeFileSync("site/source/_generated/component-links.md", `${componentLinks}\n`);

const adapterRows = [
  "| Component | User-data mode | Property |",
  "|---|---|---|",
  ...catalogProviders
    .filter((provider) => provider.userData.mode !== "none")
    .map(
      (provider) =>
        `| \`${provider.componentName}\` | ${provider.userData.mode} | ${provider.userData.property ?? provider.userData.metadataKey ?? ""} |`,
    ),
].join("\n");
writeFileSync("site/source/_generated/shared-cloud-init-table.md", `${adapterRows}\n`);

for (const provider of publisherProviders) {
  writeFileSync(`site/source/_generated/component-yaml/${provider.docs.slug}.md`, renderYaml(provider));
}

function renderYaml(provider) {
  const lines = [
    "## Pulumi YAML",
    "",
    "```yaml",
    `name: ${provider.yamlExample.name}`,
    "runtime: yaml",
    "config:",
    "  tenantUrl:",
    "    type: String",
    "  bearerToken:",
    "    type: String",
    "    secret: true",
    "resources:",
    "  publisher:",
    `    type: ${provider.token}`,
    "    properties:",
    "      tenantUrl: ${tenantUrl}",
    "      bearerToken: ${bearerToken}",
  ];

  for (const [key, value] of provider.yamlExample.properties) {
    appendProperty(lines, key, value, "      ");
  }

  lines.push("outputs:");
  lines.push("  publisherNames: ${publisher.publisherNames}");
  lines.push("  publishers: ${publisher.publishers}");
  lines.push("```");
  return `${lines.join("\n")}\n`;
}

function appendProperty(lines, key, value, indent) {
  if (Array.isArray(value)) {
    lines.push(`${indent}${key}:`);
    for (const item of value) {
      if (item && typeof item === "object") {
        const entries = Object.entries(item);
        if (entries.length === 0) {
          lines.push(`${indent}  - {}`);
          continue;
        }
        const [[firstKey, firstValue], ...rest] = entries;
        lines.push(`${indent}  - ${firstKey}: ${firstValue}`);
        for (const [childKey, childValue] of rest) {
          appendProperty(lines, childKey, childValue, `${indent}    `);
        }
        continue;
      }
      lines.push(`${indent}  - ${item}`);
    }
    return;
  }

  if (value && typeof value === "object") {
    lines.push(`${indent}${key}:`);
    for (const [childKey, childValue] of Object.entries(value)) {
      appendProperty(lines, childKey, childValue, `${indent}  `);
    }
    return;
  }

  lines.push(`${indent}${key}: ${value}`);
}
