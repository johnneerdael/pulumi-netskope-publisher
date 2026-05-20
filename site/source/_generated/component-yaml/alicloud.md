## Pulumi YAML

```yaml
name: netskope-publisher-alicloud
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:AlicloudPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      imageId: imageId-example
      vswitchId: vswitchId-example
      securityGroupIds:
        - securityGroupIds-example
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
