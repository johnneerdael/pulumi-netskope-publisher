## Pulumi YAML

```yaml
name: netskope-publisher-oci
runtime: yaml
config:
  tenantUrl:
    type: String
  bearerToken:
    type: String
    secret: true
resources:
  publisher:
    type: netskope-publisher:index:OciPublisher
    properties:
      tenantUrl: ${tenantUrl}
      bearerToken: ${bearerToken}
      namePrefix: pub
      replicas: 2
      compartmentId: compartmentId-example
      availabilityDomain: availabilityDomain-example
      subnetId: subnetId-example
      imageId: imageId-example
outputs:
  publisherNames: ${publisher.publisherNames}
  publishers: ${publisher.publishers}
```
