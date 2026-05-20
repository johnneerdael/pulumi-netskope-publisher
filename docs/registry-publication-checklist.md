# Pulumi Registry Publication Checklist

The repository now includes the Registry-facing files that can be
reviewed locally:

- `schema.json`
- `docs/_index.md`
- `docs/installation-configuration.md`
- `docs/registry-submission.md`
- `npm run registry:check`
- `npm run sdk:gen`
- `npm run sdk:pack`

Before requesting public Registry listing:

- Confirm the package name, publisher name, logo URL, and package
  description in `schema.json`.
- Keep the component resource tokens aligned with the schema package
  name: `netskope-publisher:index:*`.
- Run `npm run go:test`.
- Run `npm run sdk:gen`.
- Run `npm run sdk:pack`.
- Run `npm run registry:check`.
- Run `npm run plugin:dist` and confirm the release archives are named
  `pulumi-resource-netskope-publisher-v<version>-<os>-<arch>.tar.gz`.
- Publish the npm package.
- Publish the Python SDK to PyPI.
- Publish the Go SDK through the tagged GitHub module.
- Publish the C# SDK to NuGet.
- Publish the Java SDK to the configured Maven-compatible repository.
  For Maven Central, confirm `JAVA_MAVEN_AUTH_BASE64`,
  `JAVA_MAVEN_GROUP_ID`, `JAVA_SIGNING_KEY`, and
  `JAVA_SIGNING_PASSWORD` are configured and the Maven namespace is
  verified. `JAVA_MAVEN_AUTH_BASE64` is sent as a bearer token for Maven
  Central's OSSRH staging API.
- Publish the Rust SDK to crates.io.
- Confirm the tag release uploaded the plugin archives to GitHub
  Releases, matching `pluginDownloadURL` in `schema.json`.
- Confirm the Go provider schema includes `NetskopeRegistration` plus
  the AWS, Azure, GCP, Kubernetes, vSphere, and Hyper-V resources.
- Confirm the generated Go SDK is available at
  `github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher`.
- Use `docs/registry-pr-body.md` for the public Pulumi Registry PR body.
- Open a PR against `pulumi/registry` and add the community package
  entry for `johnneerdael/pulumi-netskope-publisher` with schema path
  `schema.json`.
- Keep the TypeScript source-based package as either the canonical
  package or a compatibility package, but do not publish two packages
  with conflicting APIs.
