# Pulumi Registry Submission Notes

This repository now carries the local artifacts expected by the Pulumi
Registry publishing flow:

- `schema.json` contains package metadata and component API docs.
- `docs/_index.md` provides the package overview.
- `docs/installation-configuration.md` provides install and setup docs.
- `docs/registry-pr-body.md` provides a copy/paste Pulumi Registry PR
  checklist with the current published artifact status.
- `npm run registry:check` validates the local Registry-facing files.
- `npm run plugin:dist` builds GitHub Release plugin archives matching
  Pulumi's expected executable plugin asset naming.
- `npm run sdk:gen` generates Python, Go, C#, Java, and Rust SDKs from
  `schema.json`.
- `npm run sdk:pack` builds the Python distribution artifacts, validates
  the generated Go, Java, and Rust SDKs, and builds the NuGet package.
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
graphs for AWS, Azure, GCP, Kubernetes, and vSphere.

Before opening the public Registry PR:

1. Publish the npm package, Python SDK, Go SDK, C# SDK, Java SDK
   ([`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)),
   and Rust SDK
   ([`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)).
2. Confirm the tag release attached plugin archives for supported
   platforms.
3. Run `npm run sdk:gen`.
4. Run `npm run sdk:pack`.
5. Run `npm run go:test`.
6. Run `npm run registry:check`.
7. Open a PR against `pulumi/registry`, use
   `docs/registry-pr-body.md` for the checklist details, and add:

   ```json
   {
     "repo": "johnneerdael/pulumi-netskope-publisher",
     "schema": "schema.json"
   }
   ```
