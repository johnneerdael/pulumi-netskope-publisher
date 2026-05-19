import { existsSync, rmSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { homedir } from "node:os";

const localDotnet = `${homedir()}/.dotnet/dotnet`;
const dotnetCommand = existsSync(localDotnet) ? localDotnet : "dotnet";

const sdkPackages = [
  {
    name: "python",
    outputDir: "dist/sdks/python",
    commands: [
      {
        command: "python3",
        args: ["-m", "build", "sdk/python", "--outdir", "dist/sdks/python"]
      }
    ],
    missingToolHint: "Install Python build tooling with: python3 -m pip install build twine"
  },
  {
    name: "dotnet",
    outputDir: "dist/sdks/dotnet",
    commands: [
      {
        command: dotnetCommand,
        args: ["build", "sdk/dotnet/Pulumi.NetskopePublisher.csproj", "--configuration", "Release", "/p:GeneratePackageOnBuild=false"]
      },
      {
        command: dotnetCommand,
        args: ["pack", "Pulumi.NetskopePublisher.csproj", "--configuration", "Release", "--no-build", "--output", "../../dist/sdks/dotnet"],
        cwd: "sdk/dotnet"
      }
    ],
    missingToolHint: "Install the .NET SDK before packaging the C# SDK."
  },
  {
    name: "go",
    commands: [
      {
        command: "go",
        args: ["test", "./sdk/go/..."]
      }
    ],
    missingToolHint: "Install Go before validating the generated Go SDK."
  },
  {
    name: "java",
    commands: [
      {
        command: "gradle",
        args: ["build"],
        cwd: "sdk/java"
      }
    ],
    missingToolHint: "Install Gradle before validating the generated Java SDK."
  },
  {
    name: "rust",
    commands: [
      {
        command: "cargo",
        args: ["check"],
        cwd: "sdk/rust"
      }
    ],
    missingToolHint: "Install Rust before validating the generated Rust SDK."
  }
];

for (const sdkPackage of sdkPackages) {
  if (!existsSync(`sdk/${sdkPackage.name}`)) {
    console.error(`Missing generated ${sdkPackage.name} SDK. Run npm run sdk:gen first.`);
    process.exit(1);
  }

  if (sdkPackage.outputDir) {
    rmSync(sdkPackage.outputDir, { recursive: true, force: true });
  }
  for (const command of sdkPackage.commands) {
    const result = spawnSync(command.command, command.args, {
      cwd: command.cwd,
      stdio: "inherit"
    });

    if (result.error?.code === "ENOENT") {
      console.error(sdkPackage.missingToolHint);
      process.exit(1);
    }

    if (result.status !== 0) {
      process.exit(result.status ?? 1);
    }
  }
}
