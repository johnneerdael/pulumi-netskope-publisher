## Pulumi YAML

```yaml
name: netskope-publisher-aws
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:AwsPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub-eu
      replicas: 2
      subnetId: subnet-0123456789abcdef0
      securityGroupIds:
        - sg-0123456789abcdef0
      instanceType: t3.medium
      bootstrap: true
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
