## Pulumi YAML

```yaml
name: netskope-publisher-azure
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:AzurePublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      resourceGroupName: resourceGroupName-example
      location: location-example
      subnetId: subnetId-example
      adminSshPublicKey: adminSshPublicKey-example
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
