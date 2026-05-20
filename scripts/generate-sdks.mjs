import { copyFileSync, existsSync, mkdirSync, readFileSync, readdirSync, renameSync, rmSync, statSync, writeFileSync } from "node:fs";
import { spawnSync } from "node:child_process";

const languages = process.argv.slice(2);
const selectedLanguages = languages.length > 0 ? languages : ["python", "dotnet", "go", "java", "rust"];

const supportedLanguages = new Set(["python", "dotnet", "go", "java", "rust"]);
for (const language of selectedLanguages) {
  if (!supportedLanguages.has(language)) {
    console.error(`Unsupported SDK language: ${language}`);
    process.exit(1);
  }
}

for (const language of selectedLanguages) {
  rmSync(`sdk/${language}`, { recursive: true, force: true });

  if (language === "rust") {
    generateRustSdk();
    continue;
  }

  const result = spawnSync(
    "npm",
    ["exec", "--", "pulumi", "package", "gen-sdk", "schema.json", "--language", language, "-o", "sdk"],
    {
      stdio: "inherit"
    }
  );

  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }

  if (language === "python" && existsSync("sdk/python/pulumi_netskope_publisher/README.md")) {
    copyFileSync("sdk/python/pulumi_netskope_publisher/README.md", "sdk/python/README.md");
    const pyproject = "sdk/python/pyproject.toml";
    const contents = readFileSync(pyproject, "utf8")
      .replace(/\n  \[project\.license\]\n    text = "Apache-2\.0"\n/, '\n  license = "Apache-2.0"\n');
    writeFileSync(pyproject, contents);
  }

  if (language === "dotnet") {
    normalizeDotnetProviderCase();
    normalizeDotnetProviderNamespaces();
    removeDotnetComponentSecretOutputOptions();

    const project = "sdk/dotnet/Pulumi.NetskopePublisher.csproj";
    let contents = readFileSync(project, "utf8");
    if (!contents.includes("<PackageId>JohninNL.Pulumi.NetskopePublisher</PackageId>")) {
      contents = contents.replace(
        "    <PackageReadmeFile>README.md</PackageReadmeFile>\n",
        "    <PackageReadmeFile>README.md</PackageReadmeFile>\n    <PackageId>JohninNL.Pulumi.NetskopePublisher</PackageId>\n"
      );
    }
    if (!contents.includes("<PackageReadmeFile>README.md</PackageReadmeFile>")) {
      contents = contents.replace(
        "    <PackageIcon>logo.png</PackageIcon>\n",
        "    <PackageIcon>logo.png</PackageIcon>\n    <PackageReadmeFile>README.md</PackageReadmeFile>\n    <PackageId>JohninNL.Pulumi.NetskopePublisher</PackageId>\n"
      );
    }
    if (!contents.includes('<None Include="README.md"')) {
      contents = contents.replace(
        "    <None Include=\"logo.png\">\n",
        "    <None Include=\"README.md\" Pack=\"True\" PackagePath=\"\" />\n    <None Include=\"logo.png\">\n"
      );
    }
    writeFileSync(project, contents);
  }

  if (language === "go") {
    removeGoPointerSecretAssertions();
  }

  if (language === "java") {
    removeJavaComponentSecretOutputOptions();

    const build = "sdk/java/build.gradle";
    let contents = readFileSync(build, "utf8");
    contents = contents
      .replace(
        "plugins {\n",
        "import org.gradle.api.credentials.HttpHeaderCredentials\nimport org.gradle.authentication.http.HttpHeaderAuthentication\n\nplugins {\n"
      )
      .replace(
        'group = "com.pulumi"',
        'def javaMavenGroupId = System.getenv("JAVA_MAVEN_GROUP_ID") ?: "com.pulumi"\n\ngroup = javaMavenGroupId'
      )
      .replace(
        'def publishRepoPassword = System.getenv("PUBLISH_REPO_PASSWORD")',
        'def publishRepoPassword = System.getenv("PUBLISH_REPO_PASSWORD")\ndef publishRepoBearerToken = System.getenv("PUBLISH_REPO_BEARER_TOKEN")'
      )
      .replace('                inceptionYear = ""', '                inceptionYear = "2026"')
      .replace('                name = ""', '                name = "Netskope Publisher Pulumi Java SDK"')
      .replace('                        id = ""', '                        id = "johnneerdael"')
      .replace('                        name = ""', '                        name = "John Neerdael"')
      .replace('                        email = ""', '                        email = "johnneerdael@users.noreply.github.com"')
      .replace('            groupId = "com.pulumi"', '            groupId = javaMavenGroupId')
      .replace(
        `                credentials {
                    username = publishRepoUsername
                    password = publishRepoPassword
                }`,
        `                if (publishRepoBearerToken) {
                    credentials(HttpHeaderCredentials) {
                        name = "Authorization"
                        value = "Bearer \${publishRepoBearerToken}"
                    }
                    authentication {
                        header(HttpHeaderAuthentication)
                    }
                } else {
                    credentials {
                        username = publishRepoUsername
                        password = publishRepoPassword
                    }
                }`
      );
    writeFileSync(build, contents);
  }
}

function generateRustSdk() {
  const schema = JSON.parse(readFileSync("schema.json", "utf8"));
  const version = schema.version;
  const sdkDir = "sdk/rust";
  const srcDir = `${sdkDir}/src`;

  mkdirSync(srcDir, { recursive: true });
  copyFileSync("schema.json", `${sdkDir}/schema.json`);

  writeFileSync(`${sdkDir}/Cargo.toml`, `[package]
name = "pulumi-netskope-publisher"
version = "${version}"
edition = "2021"
description = "Pulumi Gestalt Rust SDK for Netskope Private Access Publishers."
license = "Apache-2.0"
repository = "https://github.com/johnneerdael/pulumi-netskope-publisher"
include = [
  "Cargo.toml",
  "README.md",
  "build.rs",
  "schema.json",
  "src/**/*.rs"
]

[dependencies]
anyhow = "1"
bon = "3"
pulumi_gestalt_rust = "0.0.10"
serde = { version = "1", features = ["derive"] }
wit-bindgen = "0.46"

[build-dependencies]
pulumi_gestalt_build = "0.0.10"
`);

  writeFileSync(`${sdkDir}/build.rs`, `use std::error::Error;
use std::path::PathBuf;

fn main() -> Result<(), Box<dyn Error>> {
    let schema = PathBuf::from(std::env::var("CARGO_MANIFEST_DIR")?).join("schema.json");
    pulumi_gestalt_build::generate_from_schema(&schema)?;
    Ok(())
}
`);

  writeFileSync(`${srcDir}/lib.rs`, `//! Pulumi Gestalt Rust SDK for Netskope Private Access Publishers.
//!
//! The provider glue is generated at build time from the packaged Pulumi schema.

pub mod netskope_publisher {
    pulumi_gestalt_rust::include_provider!("netskope-publisher");
}
`);

  writeFileSync(`${sdkDir}/README.md`, `# pulumi-netskope-publisher Rust SDK

Rust SDK for \`pulumi-netskope-publisher\` built with Pulumi Gestalt.

This crate generates provider glue from the packaged Pulumi schema during
\`cargo build\`. Pulumi programs that use it must install the Pulumi Gestalt
Rust language plugin:

\`\`\`bash
pulumi plugin install language rust "0.0.10" --server github://api.github.com/andrzejressel/pulumi-gestalt
\`\`\`

Use \`runtime: rust\` in \`Pulumi.yaml\`.
`);
}

function normalizeDotnetProviderCase() {
  const lowerProvider = "sdk/dotnet/provider.cs";
  const canonicalProvider = "sdk/dotnet/Provider.cs";
  const temporaryProvider = "sdk/dotnet/Provider.cs.tmp";

  if (!existsSync(lowerProvider)) {
    return;
  }

  rmSync(temporaryProvider, { force: true });
  renameSync(lowerProvider, temporaryProvider);
  rmSync(canonicalProvider, { force: true });
  renameSync(temporaryProvider, canonicalProvider);
}

function normalizeDotnetProviderNamespaces() {
  for (const file of listFiles("sdk/dotnet", ".cs")) {
    const contents = readFileSync(file, "utf8");
    const updated = contents.replaceAll(
      "Pulumi.NetskopePublisher.Provider.",
      "Pulumi.NetskopePublisher.Types."
    ).replaceAll(
      "namespace Pulumi.NetskopePublisher.Provider.",
      "namespace Pulumi.NetskopePublisher.Types."
    );
    if (updated !== contents) {
      writeFileSync(file, updated);
    }
  }
}

function removeJavaComponentSecretOutputOptions() {
  for (const file of listFiles("sdk/java/src/main/java/com/pulumi/netskopepublisher", ".java")) {
    const contents = readFileSync(file, "utf8");
    if (!contents.includes("extends com.pulumi.resources.ComponentResource")) {
      continue;
    }
    const updated = contents.replace(/\n\s*\.additionalSecretOutputs\(List\.of\([\s\S]*?\n\s*\)\)/g, "");
    if (updated !== contents) {
      writeFileSync(file, updated);
    }
  }
}

function removeDotnetComponentSecretOutputOptions() {
  for (const file of listFiles("sdk/dotnet", ".cs")) {
    const contents = readFileSync(file, "utf8");
    if (!contents.includes(": global::Pulumi.ComponentResource")) {
      continue;
    }
    const updated = contents.replace(/\n\s*AdditionalSecretOutputs =\n\s*\{[\s\S]*?\n\s*\},/g, "");
    if (updated !== contents) {
      writeFileSync(file, updated);
    }
  }
}

function removeGoPointerSecretAssertions() {
  for (const file of listFiles("sdk/go/netskopepublisher", ".go")) {
    const contents = readFileSync(file, "utf8");
    const updated = contents.replace(
      /\n\tif args\.(ApiToken|BearerToken|InstallUserPassword) != nil \{\n\t\targs\.\1 = pulumi\.ToSecret\(args\.\1\)\.\(\*string\)\n\t\}/g,
      ""
    );
    if (updated !== contents) {
      writeFileSync(file, updated);
    }
  }
}

function listFiles(dir, extension) {
  const files = [];
  for (const entry of readdirSync(dir)) {
    const path = `${dir}/${entry}`;
    if (statSync(path).isDirectory()) {
      files.push(...listFiles(path, extension));
    } else if (path.endsWith(extension)) {
      files.push(path);
    }
  }
  return files;
}
