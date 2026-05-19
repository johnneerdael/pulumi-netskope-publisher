---
title: Upgrading publisher software
---

# Upgrading publisher software

Publisher software upgrades are managed by Netskope, not by this Pulumi
package. After enrollment, the publisher follows the upgrade profile
assigned in the Netskope admin console.

Manage profiles in **Settings -> Security Cloud Platform -> Netskope
Private Access -> Publishers -> Upgrade Profiles**.

Upgrade profiles control:

- Maintenance windows.
- Release channel.
- Automatic upgrade or approval behavior.

## What Pulumi does not do

- Trigger publisher software upgrades.
- Assign Netskope upgrade profiles.
- Roll back failed publisher upgrades.
- Replace a VM when only publisher software changes.

## When to replace infrastructure

Replace a publisher VM or chart release only when the underlying image,
bootstrap path, network placement, or registration token needs to change.
Use `pulumi preview` first so planned replacements are explicit.
