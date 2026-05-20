## Pulumi YAML

```yaml
name: netskope-publisher-exoscale
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:ExoscalePublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      zone: zone-example
      type: type-example
      templateId: templateId-example
      diskSize: 2
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
