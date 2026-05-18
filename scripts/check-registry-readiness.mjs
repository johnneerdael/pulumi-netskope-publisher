import { existsSync, readFileSync } from "node:fs";
import { spawnSync } from "node:child_process";

const requiredFiles = [
  "README.md",
  "PulumiPlugin.yaml",
  "package.json",
  "schema.json",
  "cmd/pulumi-resource-netskope-publisher/main.go",
  "internal/provider/provider.go",
  "go.sum",
  "scripts/build-plugin-archives.mjs",
  "scripts/bump-release-version.mjs",
  "scripts/run-node-tests.mjs",
  "docs/_index.md",
  "docs/installation-configuration.md",
  "docs/registry-submission.md",
  "docs/registry-publication-checklist.md"
];

const expectedResourceTokens = [
  "netskope-publisher:index:AwsPublisher",
  "netskope-publisher:index:AzurePublisher",
  "netskope-publisher:index:GcpPublisher",
  "netskope-publisher:index:VspherePublisher",
  "netskope-publisher:index:HypervPublisher",
  "netskope-publisher:index:NetskopeRegistration"
];

const sourceTokens = {
  "src/awsPublisher.ts": "netskope-publisher:index:AwsPublisher",
  "src/azurePublisher.ts": "netskope-publisher:index:AzurePublisher",
  "src/gcpPublisher.ts": "netskope-publisher:index:GcpPublisher",
  "src/vspherePublisher.ts": "netskope-publisher:index:VspherePublisher",
  "src/hypervPublisher.ts": "netskope-publisher:index:HypervPublisher"
};

const requiredSchemaFields = [
  "name",
  "version",
  "displayName",
  "description",
  "publisher",
  "keywords",
  "homepage",
  "repository",
  "license",
  "logoUrl",
  "language",
  "resources"
];

const errors = [];
const warnings = [];

for (const file of requiredFiles) {
  if (!existsSync(file)) {
    errors.push(`Missing required registry file: ${file}`);
  }
}

let schema;
if (existsSync("schema.json")) {
  try {
    schema = JSON.parse(readFileSync("schema.json", "utf8"));
  } catch (error) {
    errors.push(`schema.json is not valid JSON: ${error.message}`);
  }
}

if (schema) {
  for (const field of requiredSchemaFields) {
    if (schema[field] === undefined || schema[field] === "") {
      errors.push(`schema.json is missing required metadata field: ${field}`);
    }
  }

  if (schema.name !== "netskope-publisher") {
    errors.push(`schema.json name must be netskope-publisher, got ${schema.name}`);
  }

  if (!Array.isArray(schema.keywords)) {
    errors.push("schema.json keywords must be an array");
  } else {
    for (const keyword of ["category/network", "kind/component"]) {
      if (!schema.keywords.includes(keyword)) {
        errors.push(`schema.json keywords must include ${keyword}`);
      }
    }
  }

  for (const token of expectedResourceTokens) {
    if (!schema.resources?.[token]) {
      errors.push(`schema.json is missing resource token: ${token}`);
    }
  }

  if (!schema.language?.nodejs?.packageName) {
    errors.push("schema.json must declare language.nodejs.packageName");
  }

  if (schema.pluginDownloadURL !== "github://api.github.com/johnneerdael/pulumi-netskope-publisher") {
    errors.push("schema.json pluginDownloadURL must point to the GitHub Releases plugin asset host");
  }
}

for (const [file, token] of Object.entries(sourceTokens)) {
  if (!existsSync(file)) {
    errors.push(`Missing component source file: ${file}`);
    continue;
  }

  const contents = readFileSync(file, "utf8");
  if (!contents.includes(token)) {
    errors.push(`${file} must use component token ${token}`);
  }
  if (contents.includes("netskope:index:")) {
    errors.push(`${file} still uses the old netskope:index:* package token`);
  }
}

const packageJson = JSON.parse(readFileSync("package.json", "utf8"));
for (const file of [
  "schema.json",
  "docs/_index.md",
  "docs/installation-configuration.md",
  "docs/registry-submission.md",
  "PulumiPlugin.yaml",
  "README.md",
  "cmd/pulumi-resource-netskope-publisher/main.go",
  "internal/provider/components.go",
  "internal/provider/provider.go",
  "internal/provider/registration.go",
  "internal/provider/types.go",
  "go.mod",
  "go.sum",
  "scripts/build-plugin-archives.mjs",
  "scripts/bump-release-version.mjs",
  "scripts/check-registry-readiness.mjs",
  "scripts/run-node-tests.mjs"
]) {
  if (!packageJson.files?.includes(file)) {
    errors.push(`package.json files must include ${file}`);
  }
}

if (!packageJson.scripts?.["registry:check"]) {
  errors.push("package.json must expose npm run registry:check");
}

if (!packageJson.scripts?.["plugin:dist"]) {
  errors.push("package.json must expose npm run plugin:dist");
}

if (!packageJson.scripts?.["release:bump"]) {
  errors.push("package.json must expose npm run release:bump");
}

const pluginYaml = readFileSync("PulumiPlugin.yaml", "utf8").trim();
if (pluginYaml !== "runtime: go") {
  errors.push("PulumiPlugin.yaml must declare runtime: go for the executable provider package");
}

const schemaCommand = spawnSync("go", ["run", "./cmd/pulumi-resource-netskope-publisher", "--schema"], {
  encoding: "utf8"
});
if (schemaCommand.status !== 0) {
  errors.push(`Go provider schema command failed: ${schemaCommand.stderr || schemaCommand.stdout}`);
} else {
  try {
    const providerSchema = JSON.parse(schemaCommand.stdout);
    for (const token of expectedResourceTokens) {
      if (!providerSchema.resources?.[token]) {
        errors.push(`Go provider schema is missing resource token: ${token}`);
      }
    }
  } catch (error) {
    errors.push(`Go provider schema command did not emit valid JSON: ${error.message}`);
  }
}

if (errors.length > 0) {
  console.error("Registry readiness check failed:");
  for (const error of errors) {
    console.error(`- ${error}`);
  }
  process.exit(1);
}

for (const warning of warnings) {
  console.warn(`Warning: ${warning}`);
}

console.log("Registry readiness check passed.");
