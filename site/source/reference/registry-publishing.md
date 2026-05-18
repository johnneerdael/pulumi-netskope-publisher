---
title: Registry Publishing
---

# Registry Publishing

This package is currently a TypeScript source-based component package.
It can be consumed from Git references and published to npm for
TypeScript users.

Pulumi's public Registry path for broadly consumable components expects
an executable-based package with generated SDKs. Moving this package to
public Registry publication requires a separate provider packaging track.

Immediate supported release path:

1. Publish the TypeScript package to npm.
2. Tag GitHub releases.
3. Use Git references or npm for consumption.
4. Revisit executable-based packaging before public Pulumi Registry
   submission.
