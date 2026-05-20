## Pulumi YAML

```yaml
name: netskope-publisher-openstack
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:OpenstackPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      imageName: imageName-example
      flavorName: flavorName-example
      networkName: networkName-example
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
