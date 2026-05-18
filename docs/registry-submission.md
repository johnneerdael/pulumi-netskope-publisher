# Pulumi Registry Submission Notes

This repository now carries the local artifacts expected by the Pulumi
Registry publishing flow:

- `schema.json` contains package metadata and component API docs.
- `docs/_index.md` provides the package overview.
- `docs/installation-configuration.md` provides install and setup docs.
- `npm run registry:check` validates the local Registry-facing files.
- `npm run plugin:dist` builds GitHub Release plugin archives matching
  Pulumi's expected executable plugin asset naming.

`schema.json` sets `pluginDownloadURL` to
`github://api.github.com/johnneerdael/pulumi-netskope-publisher`.
Tagged releases upload archives named
`pulumi-resource-netskope-publisher-v<version>-<os>-<arch>.tar.gz`.

Before opening the public Registry PR:

1. Publish the npm package and any generated SDK packages.
2. Confirm the tag release attached plugin archives for supported
   platforms.
3. Run `npm run registry:check`.
4. Open a PR against `pulumi/registry` and add:

   ```json
   {
     "repo": "johnneerdael/pulumi-netskope-publisher",
     "schema": "schema.json"
   }
   ```
