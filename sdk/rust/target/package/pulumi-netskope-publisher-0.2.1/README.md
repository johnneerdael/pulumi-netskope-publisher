# pulumi-netskope-publisher Rust SDK

Rust SDK for `pulumi-netskope-publisher` built with Pulumi Gestalt.

This crate generates provider glue from the packaged Pulumi schema during
`cargo build`. Pulumi programs that use it must install the Pulumi Gestalt
Rust language plugin:

```bash
pulumi plugin install language rust "0.0.10" --server github://api.github.com/andrzejressel/pulumi-gestalt
```

Use `runtime: rust` in `Pulumi.yaml`.
