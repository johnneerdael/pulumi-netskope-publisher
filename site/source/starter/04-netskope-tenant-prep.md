---
title: Netskope Tenant Prep
---

# Netskope Tenant Prep

Create or obtain an API token that can list publishers, create
publishers, and generate publisher registration tokens. Store it in
Pulumi config as a secret.

The token is used by the Pulumi provider during deployment. It is not
written into the VM after the registration token has been generated.

**Next:** [Configure the Pulumi stack](/pulumi-netskope-publisher/starter/05-configure-stack/).
