## Pulumi YAML

```yaml
name: netskope-publisher-yandex
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:YandexPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      imageId: imageId-example
      subnetId: subnetId-example
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
