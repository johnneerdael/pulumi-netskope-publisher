# Pulumi Registry Publication Checklist

The repository now includes the Registry-facing files that can be
reviewed locally:

- `schema.json`
- `docs/_index.md`
- `docs/installation-configuration.md`
- `docs/registry-submission.md`
- `npm run registry:check`

Before requesting public Registry listing:

- Confirm the package name, publisher name, logo URL, and package
  description in `schema.json`.
- Keep the component resource tokens aligned with the schema package
  name: `netskope-publisher:index:*`.
- Run `npm run registry:check`.
- Run `npm run plugin:dist` and confirm the release archives are named
  `pulumi-resource-netskope-publisher-v<version>-<os>-<arch>.tar.gz`.
- Publish the npm package.
- Confirm the tag release uploaded the plugin archives to GitHub
  Releases, matching `pluginDownloadURL` in `schema.json`.
- If generating multi-language SDKs, publish the SDK packages to the
  relevant public package feeds.
- Open a PR against `pulumi/registry` and add the community package
  entry for `johnneerdael/pulumi-netskope-publisher` with schema path
  `schema.json`.
- Keep the TypeScript source-based package as either the canonical
  package or a compatibility package, but do not publish two packages
  with conflicting APIs.
