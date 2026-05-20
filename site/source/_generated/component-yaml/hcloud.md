## Pulumi YAML

```yaml
name: netskope-publisher-hcloud
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:HcloudPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
