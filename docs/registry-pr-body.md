# Add pulumi-netskope-publisher to the Pulumi Registry

This adds `johnneerdael/pulumi-netskope-publisher` to the Pulumi
community package list with schema path `schema.json`.

## Package release checklist

- [x] Released package with a `v`-prefixed semver tag:
  `{{TAG}}`
- [x] Published TypeScript SDK:
  [`@johninnl/pulumi-netskope-publisher`](https://www.npmjs.com/package/@johninnl/pulumi-netskope-publisher)
- [x] Published Python SDK:
  [`pulumi-netskope-publisher`](https://pypi.org/project/pulumi-netskope-publisher/)
- [x] Published Go SDK:
  [`github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher`](https://pkg.go.dev/github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher)
- [x] Published C# SDK:
  [`JohninNL.Pulumi.NetskopePublisher`](https://www.nuget.org/packages/JohninNL.Pulumi.NetskopePublisher)
- [x] Published Java SDK:
  [`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages)
  in the configured Maven-compatible package repository
- [x] Published Rust SDK:
  [`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher)
- [x] Checked in `schema.json` at the same path used by the Registry package entry
- [x] Added `docs/_index.md`
- [x] Added `docs/installation-configuration.md`
- [x] `docs/installation-configuration.md` documents the published TypeScript, Python, Go, C#, Java, and Rust SDKs
- [x] `docs/_index.md` shows usage for the published TypeScript, Python, Go, C#, Java, and Rust SDKs

## Notes for reviewers

The package publishes TypeScript, Python, Go, C#, Java, and Rust SDKs
plus compiled executable provider plugin archives through GitHub
Releases. The provider schema is implemented by the Go executable
provider.
