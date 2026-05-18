# Pulumi Registry Publication Checklist

The current package is TypeScript source-based. Public Pulumi Registry
publication requires an executable-based package track with generated
SDKs.

Before requesting public Registry listing:

- Decide whether to rewrite the package as a Go executable provider.
- Generate schema and SDKs for Node.js, Python, Go, .NET, and Java.
- Publish SDKs to public language package feeds.
- Publish provider binaries for supported platforms.
- Add Registry metadata, examples, and API docs.
- Keep the TypeScript source-based package as either the canonical
  package or a compatibility package, but do not publish two packages
  with conflicting APIs.
