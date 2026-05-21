export type ProviderImplementationMode = "catalogRawVm" | "catalogTypedVm" | "catalogSpecializedVm" | "bespoke";
export type ProviderSupportStatus = "supported" | "experimental";
export type BootstrapModel = "bootstrapOnly" | "prebakedSupported" | "marketplaceSupported" | "helm" | "registrationOnly" | "experimental";
export type UserDataMode =
  | "none"
  | "plain"
  | "base64"
  | "metadata"
  | "customData"
  | "ociMetadata"
  | "scalewayDual"
  | "guestInfo"
  | "raw"
  | "proxmoxSnippet";

export interface ProviderInputMetadata {
  name: string;
  required?: boolean;
  secret?: boolean;
  summary: string;
  example?: unknown;
}

export interface ProviderValidationMetadata {
  required?: string[];
  requiredOneOf?: string[][];
  mutuallyExclusive?: string[][];
  experimentalOptInField?: string;
}

export interface ProviderYamlExample {
  name: string;
  properties: Array<[string, unknown]>;
}

export interface ProviderRegistrySchemaCheck {
  resourceToken: string;
  propertyPath: string[];
  description: string;
}

export interface ProviderCatalogEntry {
  displayName: string;
  componentName: string;
  token: string;
  support: ProviderSupportStatus;
  providerPackage?: string;
  resourceToken?: string;
  registrySchemaUrl?: string;
  registrySchemaChecks?: ProviderRegistrySchemaCheck[];
  implementation: ProviderImplementationMode;
  bootstrapModel: BootstrapModel;
  userData: {
    mode: UserDataMode;
    property?: string;
    metadataKey?: string;
  };
  inputs: {
    required: ProviderInputMetadata[];
    optional: ProviderInputMetadata[];
  };
  validation: ProviderValidationMetadata;
  docs: {
    slug: string;
    summary: string;
    bootstrapNotes?: string;
  };
  yamlExample: ProviderYamlExample;
}

interface ProviderDefinition {
  displayName: string;
  componentName: string;
  implementation: ProviderImplementationMode;
  bootstrapModel: BootstrapModel;
  userDataMode: UserDataMode;
  userDataProperty?: string;
  slug: string;
  required: string[];
  resourceToken?: string;
  providerPackage?: string;
  registrySchemaUrl?: string;
  registrySchemaChecks?: ProviderRegistrySchemaCheck[];
  validation?: Omit<ProviderValidationMetadata, "required">;
  yamlProperties?: Array<[string, unknown]>;
}

function token(componentName: string): string {
  return `netskope-publisher:index:${componentName}`;
}

function provider(definition: ProviderDefinition): ProviderCatalogEntry {
  const yamlProperties = definition.yamlProperties ?? [
    ["namePrefix", "pub"],
    ["replicas", 2],
    ...definition.required.map((name) => [name, exampleValue(name)] as [string, unknown]),
  ];

  return {
    displayName: definition.displayName,
    componentName: definition.componentName,
    token: token(definition.componentName),
    support: definition.componentName === "HypervPublisher" ? "experimental" : "supported",
    providerPackage: definition.providerPackage,
    resourceToken: definition.resourceToken,
    registrySchemaUrl: definition.registrySchemaUrl ?? registrySchemaUrl(definition),
    registrySchemaChecks: definition.registrySchemaChecks,
    implementation: definition.implementation,
    bootstrapModel: definition.bootstrapModel,
    userData: {
      mode: definition.userDataMode,
      property: definition.userDataProperty ?? userDataProperty(definition.userDataMode),
      metadataKey: definition.userDataMode === "metadata" ? "user-data" : undefined,
    },
    inputs: {
      required: definition.required.map((name) => ({
        name,
        required: true,
        summary: `${definition.displayName} ${name} input.`,
        example: exampleValue(name),
      })),
      optional: [],
    },
    validation: { required: definition.required, ...(definition.validation ?? {}) },
    docs: { slug: definition.slug, summary: `${definition.displayName} publisher component.` },
    yamlExample: {
      name: `netskope-publisher-${definition.slug}`,
      properties: yamlProperties,
    },
  };
}

function registrySchemaUrl(definition: ProviderDefinition): string | undefined {
  if (!definition.resourceToken) {
    return undefined;
  }
  return `https://www.pulumi.com/registry/packages/${definition.slug}/schema.json`;
}

function userDataProperty(mode: UserDataMode): string | undefined {
  if (mode === "plain") return "userData";
  if (mode === "base64") return "userData";
  if (mode === "raw") return "userDataRaw";
  if (mode === "metadata") return "metadata";
  if (mode === "customData") return "customData";
  if (mode === "ociMetadata") return "metadata";
  if (mode === "scalewayDual") return "cloudInit";
  if (mode === "guestInfo") return "guestinfo.userdata";
  if (mode === "proxmoxSnippet") return "content";
  return undefined;
}

function exampleValue(name: string): unknown {
  if (name.endsWith("Ids")) return [`${name}-example`];
  if (name === "diskSize" || name === "templateVmId") return 2;
  if (name === "hardDrives") return [{ path: "/var/lib/hyperv/npa.vhdx" }];
  if (name === "networks") return [{ name: "private" }];
  if (name === "securityGroupIds") return ["sg-example"];
  return `${name}-example`;
}

const providerDefinitions = [
  provider({
    displayName: "AWS",
    componentName: "AwsPublisher",
    implementation: "bespoke",
    bootstrapModel: "prebakedSupported",
    userDataMode: "base64",
    slug: "aws",
    required: ["subnetId", "securityGroupIds"],
    providerPackage: "@pulumi/aws",
    yamlProperties: [["namePrefix", "pub-eu"], ["replicas", 2], ["subnetId", "subnet-0123456789abcdef0"], ["securityGroupIds", ["sg-0123456789abcdef0"]], ["instanceType", "t3.medium"], ["bootstrap", true]],
  }),
  provider({ displayName: "Azure", componentName: "AzurePublisher", implementation: "bespoke", bootstrapModel: "marketplaceSupported", userDataMode: "customData", slug: "azure", required: ["resourceGroupName", "location", "subnetId", "adminSshPublicKey"], providerPackage: "@pulumi/azure-native" }),
  provider({ displayName: "GCP", componentName: "GcpPublisher", implementation: "bespoke", bootstrapModel: "bootstrapOnly", userDataMode: "metadata", slug: "gcp", required: ["project", "zone", "network", "subnetwork", "image"], providerPackage: "@pulumi/gcp" }),
  provider({ displayName: "Kubernetes", componentName: "KubernetesPublisher", implementation: "bespoke", bootstrapModel: "helm", userDataMode: "none", slug: "kubernetes", required: [], providerPackage: "@pulumi/kubernetes" }),
  provider({ displayName: "vSphere", componentName: "VspherePublisher", implementation: "bespoke", bootstrapModel: "prebakedSupported", userDataMode: "guestInfo", slug: "vsphere", required: ["datacenter", "datastore", "networkName", "templateName"], providerPackage: "@pulumi/vsphere" }),
  provider({ displayName: "ESXi Native", componentName: "EsxiPublisher", implementation: "bespoke", bootstrapModel: "prebakedSupported", userDataMode: "guestInfo", slug: "esxi", required: ["diskStore", "virtualNetwork"], providerPackage: "@pulumiverse/esxi-native" }),
  provider({ displayName: "Hcloud", componentName: "HcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "hcloud", required: [], resourceToken: "hcloud:index/server:Server", providerPackage: "@pulumi/hcloud" }),
  provider({ displayName: "Nutanix", componentName: "NutanixPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "base64", userDataProperty: "guestCustomizationCloudInitUserData", slug: "nutanix", required: ["clusterUuid"], resourceToken: "nutanix:index/virtualMachine:VirtualMachine", providerPackage: "@pierskarsenbarg/nutanix" }),
  provider({ displayName: "OpenStack", componentName: "OpenstackPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "openstack", required: ["imageName", "flavorName", "networkName"], resourceToken: "openstack:compute/instance:Instance", providerPackage: "@pulumi/openstack" }),
  provider({ displayName: "OVH", componentName: "OvhPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "ovh", required: ["serviceName", "region", "imageId", "flavorId"], resourceToken: "ovh:CloudProject/instance:Instance", providerPackage: "@ovhcloud/pulumi-ovh" }),
  provider({ displayName: "Scaleway", componentName: "ScalewayPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "scalewayDual", slug: "scaleway", required: [], resourceToken: "scaleway:index/instanceServer:InstanceServer", providerPackage: "@pulumiverse/scaleway" }),
  provider({ displayName: "OCI", componentName: "OciPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "ociMetadata", slug: "oci", required: ["compartmentId", "availabilityDomain", "subnetId", "imageId"], resourceToken: "oci:Core/instance:Instance", providerPackage: "@pulumi/oci" }),
  provider({ displayName: "Alicloud", componentName: "AlicloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "base64", slug: "alicloud", required: ["imageId", "vswitchId", "securityGroupIds"], resourceToken: "alicloud:ecs/instance:Instance", providerPackage: "@pulumi/alicloud" }),
  provider({
    displayName: "Proxmox VE",
    componentName: "ProxmoxvePublisher",
    implementation: "catalogSpecializedVm",
    bootstrapModel: "bootstrapOnly",
    userDataMode: "proxmoxSnippet",
    slug: "proxmoxve",
    required: ["nodeName", "datastoreId", "templateVmId"],
    resourceToken: "proxmoxve:index/vmLegacy:VmLegacy",
    providerPackage: "@muhlba91/pulumi-proxmoxve",
    registrySchemaChecks: [{
      resourceToken: "proxmoxve:index/fileLegacy:FileLegacy",
      propertyPath: ["sourceRaw", "data"],
      description: "cloud-init snippet content",
    }, {
      resourceToken: "proxmoxve:index/vmLegacy:VmLegacy",
      propertyPath: ["initialization", "userDataFileId"],
      description: "VM cloud-init user-data file reference",
    }],
  }),
  provider({ displayName: "DigitalOcean", componentName: "DigitaloceanPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "digitalocean", required: ["region"], resourceToken: "digitalocean:index/droplet:Droplet", providerPackage: "@pulumi/digitalocean", yamlProperties: [["namePrefix", "pub"], ["replicas", 2], ["region", "ams3"], ["size", "s-2vcpu-4gb"], ["image", "ubuntu-22-04-x64"], ["bootstrap", true]] }),
  provider({ displayName: "Vultr", componentName: "VultrPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "vultr", required: ["region", "plan"], resourceToken: "vultr:index/instance:Instance", providerPackage: "@ediri/vultr", validation: { requiredOneOf: [["osId", "imageId"]], mutuallyExclusive: [["osId", "imageId"]] }, yamlProperties: [["namePrefix", "pub"], ["replicas", 2], ["region", "ams"], ["plan", "vc2-2c-4gb"], ["osId", 1743], ["bootstrap", true]] }),
  provider({ displayName: "Exoscale", componentName: "ExoscalePublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "exoscale", required: ["zone", "type", "templateId", "diskSize"], resourceToken: "exoscale:index/computeInstance:ComputeInstance", providerPackage: "@pulumiverse/exoscale" }),
  provider({ displayName: "UpCloud", componentName: "UpcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "upcloud", required: ["zone"], resourceToken: "upcloud:index/server:Server", providerPackage: "@upcloud/pulumi-upcloud" }),
  provider({ displayName: "Stackit", componentName: "StackitPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "stackit", required: ["projectId", "machineType", "imageId"], resourceToken: "stackit:index/server:Server", providerPackage: "@stackitcloud/pulumi-stackit" }),
  provider({ displayName: "Equinix Metal", componentName: "EquinixPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "equinix", required: ["projectId", "metro", "plan"], resourceToken: "equinix:metal/device:Device", providerPackage: "@equinix-labs/pulumi-equinix" }),
  provider({ displayName: "Outscale", componentName: "OutscalePublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "outscale", required: ["imageId"], resourceToken: "outscale:index/vm:Vm", providerPackage: "terraform-provider:outscale/outscale" }),
  provider({ displayName: "OpenTelekomCloud", componentName: "OpentelekomcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "opentelekomcloud", required: ["networks"], resourceToken: "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", providerPackage: "terraform-provider:opentelekomcloud/opentelekomcloud" }),
  provider({ displayName: "TencentCloud", componentName: "TencentcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "raw", slug: "tencentcloud", required: ["availabilityZone", "imageId"], resourceToken: "tencentcloud:index/instance:Instance", providerPackage: "terraform-provider:tencentcloudstack/tencentcloud" }),
  provider({ displayName: "Yandex Cloud", componentName: "YandexPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "metadata", slug: "yandex", required: ["imageId", "subnetId"], resourceToken: "yandex:index/computeInstance:ComputeInstance", providerPackage: "pulumi/yandex" }),
  provider({ displayName: "Hyper-V", componentName: "HypervPublisher", implementation: "bespoke", bootstrapModel: "experimental", userDataMode: "none", slug: "hyperv", required: ["switchName", "hardDrives"], validation: { experimentalOptInField: "enableExperimentalHyperv" } }),
  provider({ displayName: "Netskope Registration", componentName: "NetskopeRegistration", implementation: "bespoke", bootstrapModel: "registrationOnly", userDataMode: "none", slug: "registration", required: ["publisherNames"] }),
] satisfies ProviderCatalogEntry[];

export const catalogProviders = providerDefinitions;

export const providerCatalog = Object.fromEntries(catalogProviders.map((entry) => [entry.componentName, entry])) as Record<string, ProviderCatalogEntry>;

export const catalogDrivenProviders = catalogProviders.filter((entry) => entry.implementation !== "bespoke");
export const bespokeProviders = catalogProviders.filter((entry) => entry.implementation === "bespoke");
