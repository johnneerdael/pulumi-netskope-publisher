---
title: Registration flow
---

# Registration flow

For each resolved publisher name:

1. **List** publishers with `GET {tenantUrl}/api/v2/infrastructure/publishers`.
2. **Create if missing** with `POST {tenantUrl}/api/v2/infrastructure/publishers`.
3. **Generate token** with `POST {tenantUrl}/api/v2/infrastructure/publishers/{id}/registration_token`.
4. **Render cloud-init** with the token and optional bootstrap settings.
5. **Provision the workload** so first boot or pod startup registers the
   publisher with the tenant.

Existing publisher records with the same name are reused. The
`publishers` output includes `existedBefore` so automation can see
whether Pulumi created the tenant-side object or attached to an
existing one.

## Pre-baked image runcmd

```yaml
runcmd:
  - su - ubuntu -c 'sudo /home/ubuntu/npa_publisher_wizard -token "<TOKEN>"'
```

The wizard binary is already present, so cloud-init only runs the
registration command.

## Bootstrap runcmd

```yaml
runcmd:
  - chmod 0600 /etc/netplan/60-cloudinit-override.yaml  # when guestNetworkInterface is set
  - netplan apply                                       # when guestNetworkInterface is set
  - pkill -KILL -u ubuntu || true                       # when installUser != "ubuntu"
  - userdel -r ubuntu 2>/dev/null || true               # when installUser != "ubuntu"
  - chmod 1777 /tmp
  - install -d -o <user> -g <user> -m 0755 /home/<user>/resources   # when nonat is true
  - install -o <user> -g <user> -m 0644 /dev/null /home/<user>/resources/.nonat
  - su - <user> -c 'curl -fsSL <bootstrapUrl> | sudo bash'
  - su - <user> -c 'sudo /home/<user>/npa_publisher_wizard -token "<TOKEN>"'
```

The default `<user>` is `ubuntu`. Override `installUser`, SSH keys,
password handling, and guest networking when the base image or platform
requires it.

## What is not handled automatically

- Publisher software upgrades are managed by Netskope upgrade profiles.
- Destroying Pulumi resources does not delete tenant-side publisher
  records.
- Token rotation requires replacing the relevant registration and VM or
  chart resources.

See [Rotate the registration token](/pulumi-netskope-publisher/admin/how-to/rotate-token/)
and [Delete a publisher cleanly](/pulumi-netskope-publisher/admin/how-to/delete-publisher/).
