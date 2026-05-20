---
title: Installation & Configuration
meta_desc: Install and configure the Netskope Publisher Pulumi package.
layout: package
---

Install the TypeScript SDK from npm:

```bash
npm install @johninnl/pulumi-netskope-publisher
```

Published TypeScript SDK:
[`@johninnl/pulumi-netskope-publisher`](https://www.npmjs.com/package/@johninnl/pulumi-netskope-publisher).

Install the Python SDK from PyPI:

```bash
pip install pulumi-netskope-publisher
```

Published Python SDK:
[`pulumi-netskope-publisher`](https://pypi.org/project/pulumi-netskope-publisher/).

Install the C# SDK from NuGet:

```bash
dotnet add package JohninNL.Pulumi.NetskopePublisher
```

Published C# SDK:
[`JohninNL.Pulumi.NetskopePublisher`](https://www.nuget.org/packages/JohninNL.Pulumi.NetskopePublisher).

Install the Go SDK from the tagged GitHub module:

```bash
go get github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher
```

Published Go SDK:
[`github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher`](https://pkg.go.dev/github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher).

Install the Java SDK from GitHub Packages or the configured
Maven-compatible repository:

```kotlin
implementation("com.pulumi:netskope-publisher:<version>")
```

Published Java SDK:
[`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages).

The release workflow publishes Java packages to GitHub Packages by
default, or to the Maven-compatible repository configured through
`JAVA_MAVEN_REPOSITORY_URL`. Configure the matching Maven repository in
Gradle or Maven before resolving this dependency.

Install the cloud provider packages used by the component you deploy:

```bash
npm install @pulumi/aws @pulumi/azure-native @pulumi/gcp @pulumi/kubernetes @pulumi/vsphere
```

## Netskope configuration

For automatic publisher registration, configure the Netskope tenant URL
and API token as Pulumi stack configuration:

```bash
pulumi config set tenantUrl https://example.goskope.com
pulumi config set --secret apiToken ns-api-token
```

The token must be allowed to create or look up publisher registration
records in the tenant.

To avoid automatic registration, pass `registrations` to the component.
Each entry is keyed by the publisher name and must include
`publisherId` and `registrationToken`.

## Provider configuration

Configure the cloud provider used by the selected component with the
standard Pulumi provider configuration for AWS, Azure Native, Google
Cloud, Kubernetes, or vSphere.

Each component also requires provider-specific network and image inputs.
See the component API docs and the examples directory for complete
programs. On GCP, use a standard Linux image such as Ubuntu 22.04; the
`GcpPublisher` component runs the Netskope generic bootstrap script from
cloud-init by default because there is no public Netskope Publisher GCE
image.

On Kubernetes, configure the Pulumi Kubernetes provider for the target
cluster. `KubernetesPublisher` creates the namespace, required Secrets,
and Helm release resources for the
`kubernetes-netskope-publisher` chart.

## Publisher images

The package exports `officialImageSources` with the official Netskope
VHDX and OVA download URLs:

```typescript
import { officialImageSources } from "@johninnl/pulumi-netskope-publisher";

export const ovaUrl = officialImageSources.ova;
export const vhdxUrl = officialImageSources.vhdx;
```
