## Pulumi YAML

```yaml
name: netskope-publisher-opentelekomcloud
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:OpentelekomcloudPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      networks:
        - name: private
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
