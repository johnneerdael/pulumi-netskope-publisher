import { mkdirSync, readFileSync, rmSync } from "node:fs";
import { basename, join } from "node:path";
import { spawnSync } from "node:child_process";

const packageJson = JSON.parse(readFileSync("package.json", "utf8"));
const version = packageJson.version;
const pluginName = "pulumi-resource-netskope-publisher";
const outDir = join("dist", "plugins");
const targets = [
  ["linux", "amd64"],
  ["linux", "arm64"],
  ["darwin", "amd64"],
  ["darwin", "arm64"],
  ["windows", "amd64"]
];

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    stdio: "inherit",
    ...options,
    env: {
      ...process.env,
      ...options.env
    }
  });

  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
}

rmSync(outDir, { recursive: true, force: true });
mkdirSync(outDir, { recursive: true });

for (const [goos, goarch] of targets) {
  const targetName = `${goos}-${goarch}`;
  const targetDir = join(outDir, targetName);
  const binaryName = goos === "windows" ? `${pluginName}.exe` : pluginName;
  const binaryPath = join(targetDir, binaryName);
  const archiveName = `${pluginName}-v${version}-${targetName}.tar.gz`;
  const archivePath = join(outDir, archiveName);

  mkdirSync(targetDir, { recursive: true });

  run("go", [
    "build",
    "-trimpath",
    "-ldflags",
    `-s -w -X main.version=${version}`,
    "-o",
    binaryPath,
    "./cmd/pulumi-resource-netskope-publisher"
  ], {
    env: {
      CGO_ENABLED: "0",
      GOOS: goos,
      GOARCH: goarch
    }
  });

  run("tar", ["-czf", archivePath, "-C", targetDir, basename(binaryPath)]);
  console.log(`Created ${archivePath}`);
}
