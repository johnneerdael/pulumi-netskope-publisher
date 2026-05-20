## Pulumi YAML

```yaml
name: netskope-publisher-esxi
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:EsxiPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      diskStore: diskStore-example
      virtualNetwork: virtualNetwork-example
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
