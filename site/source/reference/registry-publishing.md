---
title: Registry Publishing
---

# Registry Publishing

The repository carries both release paths needed for publication:

- the TypeScript component package published to npm
- the Go executable component provider used for Pulumi Registry plugin
  downloads

`schema.json` sets `pluginDownloadURL` to GitHub Releases. Tagged
releases build plugin archives named
`pulumi-resource-netskope-publisher-v<version>-<os>-<arch>.tar.gz` for
Linux, macOS, and Windows targets.

Before opening the public Registry PR:

1. Run `npm test`.
2. Run `npm run go:test`.
3. Run `npm run registry:check`.
4. Run `npm run plugin:dist`.
5. Publish the npm package.
6. Tag a GitHub release and confirm the plugin archives are attached.
7. Add the community package entry in `pulumi/registry`.

The Go provider constructs AWS, Azure, GCP, Kubernetes, and vSphere child
resources and includes the stateful `NetskopeRegistration` resource used
by those components when `tenantUrl` and `bearerToken` are provided.
Pre-created `registrations` remain available for BYO registration
workflows.
