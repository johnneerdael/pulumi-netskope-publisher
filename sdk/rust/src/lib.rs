//! Pulumi Gestalt Rust SDK for Netskope Private Access Publishers.
//!
//! The provider glue is generated at build time from the packaged Pulumi schema.

pub mod netskope_publisher {
    pulumi_gestalt_rust::include_provider!("netskope-publisher");
}
