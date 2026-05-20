## Pulumi YAML

```yaml
name: netskope-publisher-proxmoxve
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:ProxmoxvePublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      nodeName: nodeName-example
      datastoreId: datastoreId-example
      templateVmId: 2
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
