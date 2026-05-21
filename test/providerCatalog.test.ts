import assert from "node:assert/strict";
import test from "node:test";
import { providerCatalog, catalogProviders, catalogDrivenProviders, bespokeProviders } from "../src/providerCatalog";

test("provider catalog has unique names and tokens", () => {
  const names = new Set<string>();
  const tokens = new Set<string>();

  for (const provider of catalogProviders) {
    assert.equal(names.has(provider.componentName), false, `duplicate component ${provider.componentName}`);
    assert.equal(tokens.has(provider.token), false, `duplicate token ${provider.token}`);
    names.add(provider.componentName);
    tokens.add(provider.token);
  }
});

test("provider catalog includes current public components", () => {
  for (const componentName of [
    "AwsPublisher",
    "AzurePublisher",
    "GcpPublisher",
    "KubernetesPublisher",
    "VspherePublisher",
    "EsxiPublisher",
    "HcloudPublisher",
    "NutanixPublisher",
    "OpenstackPublisher",
    "OvhPublisher",
    "ScalewayPublisher",
    "OciPublisher",
    "AlicloudPublisher",
    "ProxmoxvePublisher",
    "DigitaloceanPublisher",
    "VultrPublisher",
    "ExoscalePublisher",
    "UpcloudPublisher",
    "StackitPublisher",
    "EquinixPublisher",
    "OutscalePublisher",
    "OpentelekomcloudPublisher",
    "TencentcloudPublisher",
    "YandexPublisher",
    "HypervPublisher",
    "NetskopeRegistration",
  ]) {
    assert.ok(providerCatalog[componentName], `${componentName} missing from provider catalog`);
    assert.equal(providerCatalog[componentName].token, `netskope-publisher:index:${componentName}`);
  }
});

test("catalog-driven providers declare resource token, adapter, docs, and yaml example", () => {
  for (const provider of catalogDrivenProviders) {
    assert.ok(provider.resourceToken, `${provider.componentName} missing resourceToken`);
    assert.ok(provider.registrySchemaUrl, `${provider.componentName} missing registrySchemaUrl`);
    assert.notEqual(provider.userData.mode, "none", `${provider.componentName} missing user-data mode`);
    assert.ok(provider.docs.summary, `${provider.componentName} missing docs summary`);
    assert.ok(provider.yamlExample.name, `${provider.componentName} missing yaml example name`);
    assert.ok(provider.yamlExample.properties.length > 0, `${provider.componentName} missing yaml properties`);
  }
});

test("bespoke providers are metadata-only", () => {
  for (const provider of bespokeProviders) {
    assert.equal(provider.implementation, "bespoke");
    assert.ok(provider.docs.summary, `${provider.componentName} missing docs summary`);
  }
});

test("provider catalog uses installable upstream package metadata", () => {
  const expectedPackages: Record<string, string> = {
    UpcloudPublisher: "@upcloud/pulumi-upcloud",
    StackitPublisher: "@stackitcloud/pulumi-stackit",
    EquinixPublisher: "@equinix-labs/pulumi-equinix",
    OpentelekomcloudPublisher: "terraform-provider:opentelekomcloud/opentelekomcloud",
    OutscalePublisher: "terraform-provider:outscale/outscale",
    TencentcloudPublisher: "terraform-provider:tencentcloudstack/tencentcloud",
    YandexPublisher: "pulumi/yandex",
  };

  for (const [componentName, packageName] of Object.entries(expectedPackages)) {
    assert.equal(providerCatalog[componentName].providerPackage, packageName, `${componentName} providerPackage mismatch`);
  }
});

test("catalog-driven providers declare upstream registry schema URLs", () => {
  const expectedUrls: Record<string, string> = {
    HcloudPublisher: "https://www.pulumi.com/registry/packages/hcloud/schema.json",
    NutanixPublisher: "https://www.pulumi.com/registry/packages/nutanix/schema.json",
    OpenstackPublisher: "https://www.pulumi.com/registry/packages/openstack/schema.json",
    OvhPublisher: "https://www.pulumi.com/registry/packages/ovh/schema.json",
    ScalewayPublisher: "https://www.pulumi.com/registry/packages/scaleway/schema.json",
    OciPublisher: "https://www.pulumi.com/registry/packages/oci/schema.json",
    AlicloudPublisher: "https://www.pulumi.com/registry/packages/alicloud/schema.json",
    ProxmoxvePublisher: "https://www.pulumi.com/registry/packages/proxmoxve/schema.json",
    DigitaloceanPublisher: "https://www.pulumi.com/registry/packages/digitalocean/schema.json",
    VultrPublisher: "https://www.pulumi.com/registry/packages/vultr/schema.json",
    ExoscalePublisher: "https://www.pulumi.com/registry/packages/exoscale/schema.json",
    UpcloudPublisher: "https://www.pulumi.com/registry/packages/upcloud/schema.json",
    StackitPublisher: "https://www.pulumi.com/registry/packages/stackit/schema.json",
    EquinixPublisher: "https://www.pulumi.com/registry/packages/equinix/schema.json",
    OutscalePublisher: "https://www.pulumi.com/registry/packages/outscale/schema.json",
    OpentelekomcloudPublisher: "https://www.pulumi.com/registry/packages/opentelekomcloud/schema.json",
    TencentcloudPublisher: "https://www.pulumi.com/registry/packages/tencentcloud/schema.json",
    YandexPublisher: "https://www.pulumi.com/registry/packages/yandex/schema.json",
  };

  for (const [componentName, registrySchemaUrl] of Object.entries(expectedUrls)) {
    assert.equal(providerCatalog[componentName].registrySchemaUrl, registrySchemaUrl, `${componentName} registrySchemaUrl mismatch`);
  }
});
