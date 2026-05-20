## Pulumi YAML

```yaml
name: netskope-publisher-kubernetes
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:KubernetesPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
