---
title: Install Tools
---

# Install Tools

Install:

- Node.js 20 or newer
- npm
- Pulumi CLI
- Cloud credentials for the platform you are deploying to:
  - AWS credentials with permission to create EC2 instances
  - GCP Application Default Credentials with permission to create Compute
    Engine instances

Verify:

```bash
node --version
npm --version
pulumi version
aws sts get-caller-identity
gcloud auth list
```

Run the AWS or GCP credential check that matches your target platform.

**Next:** [Prepare the cloud account](/pulumi-netskope-publisher/starter/03-aws-account-prep/).
