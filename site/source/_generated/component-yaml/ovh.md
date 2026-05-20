## Pulumi YAML

```yaml
name: netskope-publisher-ovh
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:OvhPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      serviceName: serviceName-example
      region: region-example
      imageId: imageId-example
      flavorId: flavorId-example
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
