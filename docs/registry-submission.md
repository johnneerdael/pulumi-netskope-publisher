# Pulumi Registry Submission Notes

This repository now carries the local artifacts expected by the Pulumi
Registry publishing flow:

- `schema.json` contains package metadata and component API docs.
- `docs/_index.md` provides the package overview.
- `docs/installation-configuration.md` provides install and setup docs.
- `npm run registry:check` validates the local Registry-facing files.
- `npm run plugin:dist` builds GitHub Release plugin archives matching
  Pulumi's expected executable plugin asset naming.
- `npm run go:test` validates the Go executable component provider and
  its schema command.

`schema.json` sets `pluginDownloadURL` to
`github://api.github.com/johnneerdael/pulumi-netskope-publisher`.
Tagged releases upload archives named
`pulumi-resource-netskope-publisher-v<version>-<os>-<arch>.tar.gz`.

The executable provider is implemented with `pulumi-go-provider` under
`internal/provider` and served by
`cmd/pulumi-resource-netskope-publisher`.

The Go provider exposes the component schema, executable provider entry
point, a stateful `NetskopeRegistration` resource, and child-resource
graphs for AWS, Azure, GCP, and vSphere.

Before opening the public Registry PR:

1. Publish the npm package and any generated SDK packages.
2. Confirm the tag release attached plugin archives for supported
   platforms.
3. Run `npm run go:test`.
4. Run `npm run registry:check`.
5. Open a PR against `pulumi/registry` and add:

   ```json
   {
     "repo": "johnneerdael/pulumi-netskope-publisher",
     "schema": "schema.json"
   }
   ```
