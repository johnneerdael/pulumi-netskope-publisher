import { spawnSync } from "node:child_process";
import { get } from "node:https";

export async function validateProvidersAgainstRegistrySchemas(catalogProviders) {
  const { validateProviderAgainstRegistrySchema } = await import("../dist/src/providerRegistrySchema.js");
  const errors = [];

  for (const provider of catalogProviders) {
    if (!provider.registrySchemaUrl || !provider.resourceToken) {
      continue;
    }

    const schema = await fetchJson(provider.registrySchemaUrl);
    errors.push(...validateProviderAgainstRegistrySchema(provider, schema));
  }

  return errors;
}

async function fetchJson(url, redirectsRemaining = 3) {
  return new Promise((resolve, reject) => {
    get(url, { headers: { accept: "application/json" } }, (response) => {
      const location = response.headers.location;
      if (location && response.statusCode && response.statusCode >= 300 && response.statusCode < 400) {
        if (redirectsRemaining === 0) {
          reject(new Error(`${url} returned too many redirects`));
          return;
        }
        response.resume();
        const nextUrl = new URL(location, url).toString();
        resolve(fetchJson(nextUrl, redirectsRemaining - 1));
        return;
      }

      let body = "";
      response.setEncoding("utf8");
      response.on("data", (chunk) => {
        body += chunk;
      });
      response.on("end", () => {
        if (response.statusCode !== 200) {
          reject(new Error(`${url} returned HTTP ${response.statusCode}`));
          return;
        }
        try {
          resolve(JSON.parse(body));
        } catch (error) {
          reject(new Error(`${url} did not return valid JSON: ${error.message}`));
        }
      });
    }).on("error", reject);
  });
}

async function main() {
  const build = spawnSync("npm", ["run", "build"], { stdio: "inherit" });
  if (build.status !== 0) {
    process.exit(build.status ?? 1);
  }

  const { catalogDrivenProviders } = await import("../dist/src/providerCatalog.js");
  const errors = await validateProvidersAgainstRegistrySchemas(catalogDrivenProviders);
  if (errors.length > 0) {
    for (const error of errors) {
      console.error(`- ${error}`);
    }
    process.exit(1);
  }

  console.log("Provider registry schema check passed.");
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch((error) => {
    console.error(error.message);
    process.exit(1);
  });
}
