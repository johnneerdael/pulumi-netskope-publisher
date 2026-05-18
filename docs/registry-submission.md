# Pulumi Registry Submission Notes

This repository now carries the local artifacts expected by the Pulumi
Registry publishing flow:

- `schema.json` contains package metadata and component API docs.
- `docs/_index.md` provides the package overview.
- `docs/installation-configuration.md` provides install and setup docs.
- `npm run registry:check` validates the local Registry-facing files.

Public Registry submission is still blocked on the executable package
track. The current package is a source-based TypeScript component
package; Pulumi's public Registry publishing flow expects a
`pluginDownloadURL` that points to compiled plugin binaries and package
SDKs published to the relevant language package feeds.

When the executable package track is available:

1. Publish the npm package and any generated SDK packages.
2. Publish compiled plugin binaries to GitHub Releases or another
   supported host.
3. Add `pluginDownloadURL` to `schema.json`.
4. Run `npm run registry:check`.
5. Open a PR against `pulumi/registry` and add:

   ```json
   {
     "repo": "johnneerdael/pulumi-netskope-publisher",
     "schema": "schema.json"
   }
   ```
