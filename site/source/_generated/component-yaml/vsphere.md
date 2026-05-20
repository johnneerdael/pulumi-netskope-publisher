## Pulumi YAML

```yaml
name: netskope-publisher-vsphere
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:VspherePublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      datacenter: datacenter-example
      datastore: datastore-example
      networkName: networkName-example
      templateName: templateName-example
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
