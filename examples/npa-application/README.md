# NPA Application Example

This example deploys a placement-labeled publisher pool, registers a private
application, assigns all apps tagged `vpc-a` to publishers labeled `vpc-a`, and
creates an NPA realtime protection policy for the app tag.

Required config:

```bash
pulumi config set tenantUrl https://example.goskope.com
pulumi config set --secret bearerToken <token>
pulumi config set subnetId subnet-123
pulumi config set --path 'securityGroupIds[0]' sg-123
pulumi config set amiId ami-123
pulumi config set policyGroupName default
pulumi config set --path 'users[0]' user@example.com
```
