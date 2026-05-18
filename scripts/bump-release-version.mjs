import { readFileSync, writeFileSync } from "node:fs";

const releaseType = process.argv[2] ?? "patch";
const validReleaseTypes = new Set(["patch", "minor", "major"]);

if (!validReleaseTypes.has(releaseType)) {
  console.error(`Release type must be one of: ${Array.from(validReleaseTypes).join(", ")}`);
  process.exit(1);
}

function readJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function writeJson(path, value) {
  writeFileSync(path, `${JSON.stringify(value, null, 2)}\n`);
}

function nextVersion(version, type) {
  const match = version.match(/^(\d+)\.(\d+)\.(\d+)$/);
  if (!match) {
    throw new Error(`Unsupported semver version: ${version}`);
  }

  const [, major, minor, patch] = match.map(Number);
  if (type === "major") {
    return `${major + 1}.0.0`;
  }
  if (type === "minor") {
    return `${major}.${minor + 1}.0`;
  }
  return `${major}.${minor}.${patch + 1}`;
}

const packageJson = readJson("package.json");
const version = nextVersion(packageJson.version, releaseType);

packageJson.version = version;
writeJson("package.json", packageJson);

const packageLock = readJson("package-lock.json");
packageLock.version = version;
if (packageLock.packages?.[""]) {
  packageLock.packages[""].version = version;
}
writeJson("package-lock.json", packageLock);

const schema = readJson("schema.json");
schema.version = version;
writeJson("schema.json", schema);

const providerPath = "internal/provider/provider.go";
const providerSource = readFileSync(providerPath, "utf8");
const updatedProviderSource = providerSource.replace(
  /semver\.MustParse\("\d+\.\d+\.\d+"\)/,
  `semver.MustParse("${version}")`,
);
if (updatedProviderSource === providerSource) {
  throw new Error(`Could not update provider schema version in ${providerPath}`);
}
writeFileSync(providerPath, updatedProviderSource);

console.log(version);
