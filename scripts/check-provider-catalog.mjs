import { existsSync, readFileSync } from "node:fs";
import { spawnSync } from "node:child_process";

const build = spawnSync("npm", ["run", "build"], { stdio: "inherit" });
if (build.status !== 0) {
  process.exit(build.status ?? 1);
}

const { catalogProviders } = await import("../dist/src/providerCatalog.js");
const schema = JSON.parse(readFileSync("schema.json", "utf8"));
const indexSource = readFileSync("src/index.ts", "utf8");
const goProviderSource = readFileSync("internal/provider/provider.go", "utf8");
const errors = [];

for (const provider of catalogProviders) {
  if (!schema.resources?.[provider.token]) {
    errors.push(`schema.json missing catalog token ${provider.token}`);
  }

  if (provider.componentName !== "NetskopeRegistration" && !indexSource.includes(`./${lowerFirst(provider.componentName)}`)) {
    errors.push(`src/index.ts missing export for ${provider.componentName}`);
  }

  if (
    provider.componentName !== "NetskopeRegistration" &&
    !goProviderSource.includes(`New${provider.componentName}`) &&
    !goProviderSource.includes(provider.componentName)
  ) {
    errors.push(`internal/provider/provider.go missing Go registration for ${provider.componentName}`);
  }

  const docsPath = `site/source/admin/component/${provider.docs.slug}.md`;
  if (!existsSync(docsPath)) {
    errors.push(`Missing component docs page ${docsPath}`);
  } else {
    const docs = readFileSync(docsPath, "utf8");
    if (!docs.includes("## Pulumi YAML")) {
      errors.push(`${docsPath} missing Pulumi YAML example`);
    }
  }
}

if (errors.length > 0) {
  for (const error of errors) {
    console.error(`- ${error}`);
  }
  process.exit(1);
}

console.log("Provider catalog parity check passed.");

function lowerFirst(value) {
  return value.charAt(0).toLowerCase() + value.slice(1);
}
