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
  propertyKind?: "input" | "output";
}

export type ProviderUpstreamPropertyCheck = ProviderRegistrySchemaCheck;

export interface ProviderCatalogEntry {
  displayName: string;
  componentName: string;
  token: string;
  support: ProviderSupportStatus;
  providerPackage?: string;
  resourceToken?: string;
  registrySchemaUrl?: string;
  registrySchemaChecks?: ProviderRegistrySchemaCheck[];
  upstreamPropertyChecks?: ProviderUpstreamPropertyCheck[];
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
  upstreamPropertyChecks?: ProviderUpstreamPropertyCheck[];
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
    upstreamPropertyChecks: definition.upstreamPropertyChecks,
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

function upstreamChecks(resourceToken: string, checks: Array<[string[], string]>): ProviderUpstreamPropertyCheck[] {
  return checks.map(([propertyPath, description]) => ({
    resourceToken,
    propertyPath,
    description,
  }));
}

function upstreamOutputChecks(resourceToken: string, checks: Array<[string[], string]>): ProviderUpstreamPropertyCheck[] {
  return checks.map(([propertyPath, description]) => ({
    resourceToken,
    propertyPath,
    description,
    propertyKind: "output",
  }));
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
  provider({ displayName: "Hcloud", componentName: "HcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "hcloud", required: [], resourceToken: "hcloud:index/server:Server", providerPackage: "@pulumi/hcloud", upstreamPropertyChecks: [
    ...upstreamChecks("hcloud:index/server:Server", [
      [["publicNets"], "public network controls"],
      [["networks"], "private network attachments"],
    ]),
    ...upstreamOutputChecks("hcloud:index/server:Server", [
      [["ipv4Address"], "public IPv4 output"],
      [["networks"], "private network output"],
    ]),
  ] }),
  provider({ displayName: "Nutanix", componentName: "NutanixPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "base64", userDataProperty: "guestCustomizationCloudInitUserData", slug: "nutanix", required: ["clusterUuid"], resourceToken: "nutanix:index/virtualMachine:VirtualMachine", providerPackage: "@pierskarsenbarg/nutanix", upstreamPropertyChecks: [
    ...upstreamChecks("nutanix:index/virtualMachine:VirtualMachine", [
      [["clusterUuid"], "cluster placement"],
      [["diskLists"], "image disk list"],
      [["nicLists"], "network interface list"],
    ]),
    ...upstreamOutputChecks("nutanix:index/virtualMachine:VirtualMachine", [
      [["nicListStatuses"], "network status outputs"],
    ]),
  ] }),
  provider({ displayName: "OpenStack", componentName: "OpenstackPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "openstack", required: ["imageName", "flavorName", "networkName"], resourceToken: "openstack:compute/instance:Instance", providerPackage: "@pulumi/openstack", upstreamPropertyChecks: [
    ...upstreamChecks("openstack:compute/instance:Instance", [
      [["networks"], "instance network attachments"],
      [["imageName"], "image selection"],
      [["flavorName"], "flavor selection"],
    ]),
    ...upstreamOutputChecks("openstack:compute/instance:Instance", [
      [["accessIpV4"], "public IPv4 output"],
      [["networks"], "network output"],
    ]),
  ] }),
  provider({ displayName: "OVH", componentName: "OvhPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "ovh", required: ["serviceName", "region", "imageId", "flavorId"], resourceToken: "ovh:CloudProject/instance:Instance", providerPackage: "@ovhcloud/pulumi-ovh", upstreamPropertyChecks: [
    ...upstreamChecks("ovh:CloudProject/instance:Instance", [
      [["bootFrom", "imageId"], "boot image"],
      [["flavor", "flavorId"], "flavor selection"],
      [["network"], "network placement"],
    ]),
    ...upstreamOutputChecks("ovh:CloudProject/instance:Instance", [
      [["addresses"], "instance address outputs"],
    ]),
  ] }),
  provider({ displayName: "Scaleway", componentName: "ScalewayPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "scalewayDual", slug: "scaleway", required: [], resourceToken: "scaleway:index/instanceServer:InstanceServer", providerPackage: "@pulumiverse/scaleway", upstreamPropertyChecks: [
    ...upstreamChecks("scaleway:index/instanceServer:InstanceServer", [
      [["cloudInit"], "cloud-init content"],
      [["userData"], "user data map"],
      [["enableDynamicIp"], "dynamic IP flag"],
    ]),
    ...upstreamOutputChecks("scaleway:index/instanceServer:InstanceServer", [
      [["privateIps"], "private IP outputs"],
      [["publicIps"], "public IP outputs"],
    ]),
  ] }),
  provider({
    displayName: "OCI",
    componentName: "OciPublisher",
    implementation: "catalogRawVm",
    bootstrapModel: "bootstrapOnly",
    userDataMode: "ociMetadata",
    slug: "oci",
    required: ["compartmentId", "availabilityDomain", "subnetId", "imageId"],
    resourceToken: "oci:Core/instance:Instance",
    providerPackage: "@pulumi/oci",
    upstreamPropertyChecks: [{
      resourceToken: "oci:Core/instance:Instance",
      propertyPath: ["createVnicDetails", "subnetId"],
      description: "primary VNIC subnet",
    }, {
      resourceToken: "oci:Core/instance:Instance",
      propertyPath: ["sourceDetails", "sourceId"],
      description: "image source ID",
    }, {
      resourceToken: "oci:Core/instance:Instance",
      propertyPath: ["metadata"],
      description: "cloud-init metadata map",
    }],
  }),
  provider({ displayName: "Alicloud", componentName: "AlicloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "base64", slug: "alicloud", required: ["imageId", "vswitchId", "securityGroupIds"], resourceToken: "alicloud:ecs/instance:Instance", providerPackage: "@pulumi/alicloud", upstreamPropertyChecks: [
    ...upstreamChecks("alicloud:ecs/instance:Instance", [
      [["imageId"], "image selection"],
      [["vswitchId"], "VSwitch placement"],
      [["securityGroups"], "security group attachments"],
    ]),
    ...upstreamOutputChecks("alicloud:ecs/instance:Instance", [
      [["primaryIpAddress"], "primary private IP output"],
      [["publicIp"], "public IP output"],
    ]),
  ] }),
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
  provider({ displayName: "DigitalOcean", componentName: "DigitaloceanPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "digitalocean", required: ["region"], resourceToken: "digitalocean:index/droplet:Droplet", providerPackage: "@pulumi/digitalocean", upstreamPropertyChecks: [
    ...upstreamChecks("digitalocean:index/droplet:Droplet", [
      [["region"], "region placement"],
      [["image"], "image selection"],
      [["vpcUuid"], "VPC placement"],
    ]),
    ...upstreamOutputChecks("digitalocean:index/droplet:Droplet", [
      [["ipv4Address"], "public IPv4 output"],
      [["ipv4AddressPrivate"], "private IPv4 output"],
    ]),
  ], yamlProperties: [["namePrefix", "pub"], ["replicas", 2], ["region", "ams3"], ["size", "s-2vcpu-4gb"], ["image", "ubuntu-22-04-x64"], ["bootstrap", true]] }),
  provider({ displayName: "Vultr", componentName: "VultrPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "vultr", required: ["region", "plan"], resourceToken: "vultr:index/instance:Instance", providerPackage: "@ediri/vultr", upstreamPropertyChecks: [
    ...upstreamChecks("vultr:index/instance:Instance", [
      [["region"], "region placement"],
      [["plan"], "instance plan"],
      [["osId"], "Ubuntu OS selection"],
      [["imageId"], "custom image selection"],
    ]),
    ...upstreamOutputChecks("vultr:index/instance:Instance", [
      [["internalIp"], "private IP output"],
      [["mainIp"], "public IP output"],
    ]),
  ], validation: { requiredOneOf: [["osId", "imageId"]], mutuallyExclusive: [["osId", "imageId"]] }, yamlProperties: [["namePrefix", "pub"], ["replicas", 2], ["region", "ams"], ["plan", "vc2-2c-4gb"], ["osId", 1743], ["bootstrap", true]] }),
  provider({ displayName: "Exoscale", componentName: "ExoscalePublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "exoscale", required: ["zone", "type", "templateId", "diskSize"], resourceToken: "exoscale:index/computeInstance:ComputeInstance", providerPackage: "@pulumiverse/exoscale", upstreamPropertyChecks: [
    ...upstreamChecks("exoscale:index/computeInstance:ComputeInstance", [
      [["zone"], "zone placement"],
      [["templateId"], "template image"],
      [["networkInterfaces"], "network attachments"],
    ]),
    ...upstreamOutputChecks("exoscale:index/computeInstance:ComputeInstance", [
      [["publicIpAddress"], "public IP output"],
    ]),
  ] }),
  provider({ displayName: "UpCloud", componentName: "UpcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "upcloud", required: ["zone"], resourceToken: "upcloud:index/server:Server", providerPackage: "@upcloud/pulumi-upcloud", upstreamPropertyChecks: upstreamChecks("upcloud:index/server:Server", [
    [["zone"], "zone placement"],
    [["template"], "Ubuntu template selection"],
    [["networkInterfaces"], "network interface configuration"],
    [["metadata"], "metadata service enablement"],
  ]) }),
  provider({ displayName: "Stackit", componentName: "StackitPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "stackit", required: ["projectId", "machineType", "imageId"], resourceToken: "stackit:index/server:Server", providerPackage: "@stackitcloud/pulumi-stackit", upstreamPropertyChecks: upstreamChecks("stackit:index/server:Server", [
    [["projectId"], "project placement"],
    [["machineType"], "machine type"],
    [["imageId"], "image selection"],
    [["networkInterfaces"], "network interface configuration"],
  ]) }),
  provider({ displayName: "Equinix Metal", componentName: "EquinixPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "equinix", required: ["projectId", "metro", "plan"], resourceToken: "equinix:metal/device:Device", providerPackage: "@equinix-labs/pulumi-equinix", upstreamPropertyChecks: [
    ...upstreamChecks("equinix:metal/device:Device", [
      [["projectId"], "project placement"],
      [["metro"], "metro placement"],
      [["operatingSystem"], "Ubuntu operating system"],
    ]),
    ...upstreamOutputChecks("equinix:metal/device:Device", [
      [["accessPrivateIpv4"], "private IPv4 output"],
      [["accessPublicIpv4"], "public IPv4 output"],
    ]),
  ] }),
  provider({ displayName: "Outscale", componentName: "OutscalePublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "outscale", required: ["imageId"], resourceToken: "outscale:index/vm:Vm", providerPackage: "terraform-provider:outscale/outscale", upstreamPropertyChecks: [
    ...upstreamChecks("outscale:index/vm:Vm", [
      [["imageId"], "image selection"],
      [["subnetId"], "subnet placement"],
      [["securityGroupIds"], "security group attachments"],
    ]),
    ...upstreamOutputChecks("outscale:index/vm:Vm", [
      [["privateIp"], "private IP output"],
      [["publicIp"], "public IP output"],
    ]),
  ] }),
  provider({ displayName: "OpenTelekomCloud", componentName: "OpentelekomcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "plain", slug: "opentelekomcloud", required: ["networks"], resourceToken: "opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", providerPackage: "terraform-provider:opentelekomcloud/opentelekomcloud", upstreamPropertyChecks: [
    ...upstreamChecks("opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", [
      [["networks"], "network attachments"],
      [["imageName"], "Ubuntu image selection"],
      [["flavorName"], "flavor selection"],
    ]),
    ...upstreamOutputChecks("opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", [
      [["accessIpV4"], "public IPv4 output"],
    ]),
  ] }),
  provider({ displayName: "TencentCloud", componentName: "TencentcloudPublisher", implementation: "catalogRawVm", bootstrapModel: "bootstrapOnly", userDataMode: "raw", slug: "tencentcloud", required: ["availabilityZone", "imageId"], resourceToken: "tencentcloud:index/instance:Instance", providerPackage: "terraform-provider:tencentcloudstack/tencentcloud", upstreamPropertyChecks: [
    ...upstreamChecks("tencentcloud:index/instance:Instance", [
      [["availabilityZone"], "availability zone placement"],
      [["imageId"], "image selection"],
      [["subnetId"], "subnet placement"],
      [["userDataRaw"], "raw user-data content"],
    ]),
    ...upstreamOutputChecks("tencentcloud:index/instance:Instance", [
      [["privateIp"], "private IP output"],
      [["publicIp"], "public IP output"],
    ]),
  ] }),
  provider({
    displayName: "Yandex Cloud",
    componentName: "YandexPublisher",
    implementation: "catalogRawVm",
    bootstrapModel: "bootstrapOnly",
    userDataMode: "metadata",
    slug: "yandex",
    required: ["imageId", "subnetId"],
    resourceToken: "yandex:index/computeInstance:ComputeInstance",
    providerPackage: "pulumi/yandex",
    upstreamPropertyChecks: [{
      resourceToken: "yandex:index/computeInstance:ComputeInstance",
      propertyPath: ["bootDisk", "initializeParams", "imageId"],
      description: "boot disk image",
    }, {
      resourceToken: "yandex:index/computeInstance:ComputeInstance",
      propertyPath: ["networkInterfaces", "subnetId"],
      description: "network interface subnet",
    }, {
      resourceToken: "yandex:index/computeInstance:ComputeInstance",
      propertyPath: ["metadata"],
      description: "cloud-init metadata map",
    }],
  }),
  provider({ displayName: "Hyper-V", componentName: "HypervPublisher", implementation: "bespoke", bootstrapModel: "experimental", userDataMode: "none", slug: "hyperv", required: ["switchName", "hardDrives"], validation: { experimentalOptInField: "enableExperimentalHyperv" } }),
  provider({ displayName: "Netskope Registration", componentName: "NetskopeRegistration", implementation: "bespoke", bootstrapModel: "registrationOnly", userDataMode: "none", slug: "registration", required: ["publisherNames", "tenantUrl"] }),
] satisfies ProviderCatalogEntry[];

export const catalogProviders = providerDefinitions;

export const providerCatalog = Object.fromEntries(catalogProviders.map((entry) => [entry.componentName, entry])) as Record<string, ProviderCatalogEntry>;

export const catalogDrivenProviders = catalogProviders.filter((entry) => entry.implementation !== "bespoke");
export const bespokeProviders = catalogProviders.filter((entry) => entry.implementation === "bespoke");
