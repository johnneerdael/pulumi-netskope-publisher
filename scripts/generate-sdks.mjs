import { copyFileSync, existsSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { spawnSync } from "node:child_process";

const languages = process.argv.slice(2);
const selectedLanguages = languages.length > 0 ? languages : ["python", "dotnet", "go"];

const supportedLanguages = new Set(["python", "dotnet", "go"]);
for (const language of selectedLanguages) {
  if (!supportedLanguages.has(language)) {
    console.error(`Unsupported SDK language: ${language}`);
    process.exit(1);
  }
}

for (const language of selectedLanguages) {
  rmSync(`sdk/${language}`, { recursive: true, force: true });

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
}
