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
- [ ] Published Go SDK:
  not currently published as a language SDK; the repository includes the Go executable provider implementation and Git tag
- [x] Published C# SDK:
  [`Pulumi.NetskopePublisher`](https://www.nuget.org/packages/Pulumi.NetskopePublisher)
- [ ] Published Java SDK:
  optional and not currently published
- [x] Checked in `schema.json` at the same path used by the Registry package entry
- [x] Added `docs/_index.md`
- [x] Added `docs/installation-configuration.md`
- [x] `docs/installation-configuration.md` links to the published TypeScript, Python, and C# SDKs
- [x] `docs/_index.md` shows usage for the published TypeScript, Python, and C# SDKs

## Notes for reviewers

The package publishes TypeScript, Python, and C# SDKs plus compiled
executable provider plugin archives through GitHub Releases. The
provider schema is implemented by the Go executable provider.

The Go language SDK is not published yet. If the Registry requires Go
SDK publication before acceptance, this PR should remain open until that
artifact is generated, tagged, and documented.
