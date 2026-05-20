## Pulumi YAML

```yaml
name: netskope-publisher-hyperv
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:HypervPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      switchName: switchName-example
      hardDrives:
        - path: /var/lib/hyperv/npa.vhdx
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
