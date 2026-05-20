---
title: SDK Installation
---

# SDK Installation

The package publishes Pulumi SDKs for TypeScript, Python, C#, Go, Java,
and Rust.

| Language | Package | Registry |
|---|---|---|
| TypeScript | [`@johninnl/pulumi-netskope-publisher`](https://www.npmjs.com/package/@johninnl/pulumi-netskope-publisher) | npm |
| Python | [`pulumi-netskope-publisher`](https://pypi.org/project/pulumi-netskope-publisher/) | PyPI |
| C# | [`JohninNL.Pulumi.NetskopePublisher`](https://www.nuget.org/packages/JohninNL.Pulumi.NetskopePublisher) | NuGet |
| Go | [`github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher`](https://pkg.go.dev/github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher) | tagged GitHub module |
| Java | [`com.pulumi:netskope-publisher`](https://github.com/johnneerdael/pulumi-netskope-publisher/packages) | GitHub Packages by default |
| Rust | [`pulumi-netskope-publisher`](https://crates.io/crates/pulumi-netskope-publisher) | crates.io |

## TypeScript

```bash
npm install @johninnl/pulumi-netskope-publisher
```

## Python

```bash
pip install pulumi-netskope-publisher
```

## C#

```bash
dotnet add package JohninNL.Pulumi.NetskopePublisher
```

## Go

```bash
go get github.com/johnneerdael/pulumi-netskope-publisher/sdk/go/netskopepublisher
```

## Java

The release workflow publishes the Java SDK as
`com.pulumi:netskope-publisher` to GitHub Packages by default.

Gradle dependency:

```kotlin
implementation("com.pulumi:netskope-publisher:0.1.11")
```

Gradle repository:

```kotlin
repositories {
    mavenCentral()
    maven {
        url = uri("https://maven.pkg.github.com/johnneerdael/pulumi-netskope-publisher")
        credentials {
            username = providers.gradleProperty("gpr.user").orNull
                ?: System.getenv("GITHUB_ACTOR")
            password = providers.gradleProperty("gpr.key").orNull
                ?: System.getenv("GITHUB_TOKEN")
        }
    }
}
```

Use the same coordinates with Maven:

```xml
<dependency>
  <groupId>com.pulumi</groupId>
  <artifactId>netskope-publisher</artifactId>
  <version>0.1.11</version>
</dependency>
```

## Rust

Add the published crate:

```toml
pulumi-netskope-publisher = "0.1.11"
```

Rust programs also need Pulumi Gestalt:

```toml
pulumi_gestalt_rust = "0.0.10"
```

Install the Pulumi Gestalt language plugin:

```bash
pulumi plugin install language rust "0.0.10" --server github://api.github.com/andrzejressel/pulumi-gestalt
```
