import { readdirSync, statSync } from "node:fs";
import { join } from "node:path";
import { spawnSync } from "node:child_process";

const testDir = "dist/test";

function collectTests(dir) {
  return readdirSync(dir)
    .flatMap((entry) => {
      const path = join(dir, entry);
      const stat = statSync(path);
      if (stat.isDirectory()) {
        return collectTests(path);
      }
      return entry.endsWith(".test.js") ? [path] : [];
    })
    .sort();
}

const tests = collectTests(testDir);
if (tests.length === 0) {
  console.error(`No compiled tests found under ${testDir}`);
  process.exit(1);
}

const result = spawnSync(process.execPath, ["--test", ...tests], {
  stdio: "inherit"
});

process.exit(result.status ?? 1);
