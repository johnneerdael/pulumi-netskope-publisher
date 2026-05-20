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
