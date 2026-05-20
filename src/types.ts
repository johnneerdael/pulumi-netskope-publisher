import * as pulumi from "@pulumi/pulumi";

export interface NameArgs {
  namePrefix?: string;
  names?: string[];
  replicas?: number;
}

export interface MetadataOptions {
  httpEndpoint?: pulumi.Input<string>;
  httpTokens?: pulumi.Input<string>;
}

export interface GuestNetworkInterface {
  name: pulumi.Input<string>;
  dhcp4?: pulumi.Input<boolean>;
  addresses?: pulumi.Input<pulumi.Input<string>[]>;
  gateway4?: pulumi.Input<string>;
  nameservers?: pulumi.Input<pulumi.Input<string>[]>;
  mtu?: pulumi.Input<number>;
}

export interface PublisherRegistrationInput {
  publisherId: pulumi.Input<number>;
  registrationToken: pulumi.Input<string>;
}

export type NetskopeAuthMode = "token" | "oauth2";

export interface NetskopeOAuth2Args {
  tokenUrl: pulumi.Input<string>;
  clientId: pulumi.Input<string>;
  clientSecret: pulumi.Input<string>;
  scope?: pulumi.Input<string>;
}

export interface CommonPublisherArgs extends NameArgs {
  namePrefix?: string;
  names?: string[];
  replicas?: number;
  placementLabels?: pulumi.Input<pulumi.Input<string>[]>;
  tenantUrl?: pulumi.Input<string>;
  bearerToken?: pulumi.Input<string>;
  authMode?: pulumi.Input<NetskopeAuthMode>;
  oauth2?: pulumi.Input<NetskopeOAuth2Args>;
  /** @deprecated Use bearerToken instead. */
  apiToken?: pulumi.Input<string>;
  wizardPath?: pulumi.Input<string>;
  tags?: pulumi.Input<Record<string, pulumi.Input<string>>>;
  registrations?: pulumi.Input<Record<string, PublisherRegistrationInput>>;
  bootstrap?: pulumi.Input<boolean>;
  bootstrapUrl?: pulumi.Input<string>;
  nonat?: pulumi.Input<boolean>;
  installUser?: pulumi.Input<string>;
  installUserPassword?: pulumi.Input<string>;
  installUserPasswordIsHash?: pulumi.Input<boolean>;
  installUserSshAuthorizedKeys?: pulumi.Input<pulumi.Input<string>[]>;
  deleteDefaultUser?: pulumi.Input<boolean>;
  guestNetworkInterface?: pulumi.Input<GuestNetworkInterface>;
}

export interface AwsPublisherArgs extends CommonPublisherArgs {
  subnetId: pulumi.Input<string>;
  securityGroupIds: pulumi.Input<pulumi.Input<string>[]>;
  keyName?: pulumi.Input<string>;
  instanceType?: pulumi.Input<string>;
  amiId?: pulumi.Input<string>;
  associatePublicIpAddress?: pulumi.Input<boolean>;
  iamInstanceProfile?: pulumi.Input<string>;
  ebsOptimized?: pulumi.Input<boolean>;
  monitoring?: pulumi.Input<boolean>;
  metadataOptions?: pulumi.Input<MetadataOptions>;
}

export interface AzureMarketplaceImage {
  publisher: pulumi.Input<string>;
  offer: pulumi.Input<string>;
  sku: pulumi.Input<string>;
  version?: pulumi.Input<string>;
}

export interface AzureOsDisk {
  type?: pulumi.Input<string>;
  sizeGb?: pulumi.Input<number>;
}

export interface AzurePublisherArgs extends CommonPublisherArgs {
  resourceGroupName: pulumi.Input<string>;
  location: pulumi.Input<string>;
  subnetId: pulumi.Input<string>;
  vmSize?: pulumi.Input<string>;
  adminUsername?: pulumi.Input<string>;
  adminSshPublicKey: pulumi.Input<string>;
  networkSecurityGroupId?: pulumi.Input<string>;
  assignPublicIp?: pulumi.Input<boolean>;
  osDisk?: pulumi.Input<AzureOsDisk>;
  imageId?: pulumi.Input<string>;
  marketplace?: pulumi.Input<AzureMarketplaceImage>;
  acceptMarketplaceTerms?: pulumi.Input<boolean>;
}

export interface GcpServiceAccount {
  email: pulumi.Input<string>;
  scopes?: pulumi.Input<pulumi.Input<string>[]>;
}

export interface GcpPublisherArgs extends CommonPublisherArgs {
  project: pulumi.Input<string>;
  zone: pulumi.Input<string>;
  network: pulumi.Input<string>;
  subnetwork: pulumi.Input<string>;
  machineType?: pulumi.Input<string>;
  image: pulumi.Input<string>;
  assignPublicIp?: pulumi.Input<boolean>;
  networkTags?: pulumi.Input<pulumi.Input<string>[]>;
  serviceAccount?: pulumi.Input<GcpServiceAccount>;
}

export interface VspherePublisherArgs extends CommonPublisherArgs {
  datacenter: pulumi.Input<string>;
  cluster?: pulumi.Input<string>;
  host?: pulumi.Input<string>;
  datastore: pulumi.Input<string>;
  networkName: pulumi.Input<string>;
  templateName: pulumi.Input<string>;
  folder?: pulumi.Input<string>;
  numCpus?: pulumi.Input<number>;
  memory?: pulumi.Input<number>;
}

export interface EsxiPublisherArgs extends CommonPublisherArgs {
  diskStore: pulumi.Input<string>;
  virtualNetwork: pulumi.Input<string>;
  os?: pulumi.Input<string>;
  memory?: pulumi.Input<number>;
  numVCpus?: pulumi.Input<number>;
  diskSize?: pulumi.Input<number>;
}

export interface HcloudPublisherArgs extends CommonPublisherArgs {
  serverType?: pulumi.Input<string>;
  image?: pulumi.Input<string>;
  location?: pulumi.Input<string>;
  datacenter?: pulumi.Input<string>;
  sshKeys?: pulumi.Input<pulumi.Input<string>[]>;
  firewallIds?: pulumi.Input<pulumi.Input<number>[]>;
  networkId?: pulumi.Input<number>;
  assignPublicIp?: pulumi.Input<boolean>;
}

export interface NutanixPublisherArgs extends CommonPublisherArgs {
  clusterUuid: pulumi.Input<string>;
  imageUuid?: pulumi.Input<string>;
  subnetUuid?: pulumi.Input<string>;
  numVCpus?: pulumi.Input<number>;
  numCoresPerVcpu?: pulumi.Input<number>;
  memorySizeMib?: pulumi.Input<number>;
}

export interface OpenstackPublisherArgs extends CommonPublisherArgs {
  imageName: pulumi.Input<string>;
  flavorName: pulumi.Input<string>;
  networkName: pulumi.Input<string>;
  keyPair?: pulumi.Input<string>;
  securityGroups?: pulumi.Input<pulumi.Input<string>[]>;
  availabilityZone?: pulumi.Input<string>;
  assignFloatingIp?: pulumi.Input<boolean>;
  floatingIpPool?: pulumi.Input<string>;
}

export interface OvhPublisherArgs extends CommonPublisherArgs {
  serviceName: pulumi.Input<string>;
  region: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  flavorId: pulumi.Input<string>;
  sshKeyName?: pulumi.Input<string>;
  networkId?: pulumi.Input<string>;
}

export interface ScalewayPublisherArgs extends CommonPublisherArgs {
  type?: pulumi.Input<string>;
  image?: pulumi.Input<string>;
  zone?: pulumi.Input<string>;
  securityGroupId?: pulumi.Input<string>;
  enableDynamicIp?: pulumi.Input<boolean>;
}

export interface OciPublisherArgs extends CommonPublisherArgs {
  compartmentId: pulumi.Input<string>;
  availabilityDomain: pulumi.Input<string>;
  shape?: pulumi.Input<string>;
  subnetId: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  sshPublicKey?: pulumi.Input<string>;
  assignPublicIp?: pulumi.Input<boolean>;
}

export interface AlicloudPublisherArgs extends CommonPublisherArgs {
  instanceType?: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  vswitchId: pulumi.Input<string>;
  securityGroupIds: pulumi.Input<pulumi.Input<string>[]>;
  keyName?: pulumi.Input<string>;
  allocatePublicIp?: pulumi.Input<boolean>;
}

export interface ProxmoxvePublisherArgs extends CommonPublisherArgs {
  nodeName: pulumi.Input<string>;
  datastoreId: pulumi.Input<string>;
  templateVmId: pulumi.Input<number>;
  cloneNodeName?: pulumi.Input<string>;
  vmId?: pulumi.Input<number>;
  poolId?: pulumi.Input<string>;
  cpuCores?: pulumi.Input<number>;
  memory?: pulumi.Input<number>;
  diskSize?: pulumi.Input<number>;
  networkBridge?: pulumi.Input<string>;
  networkModel?: pulumi.Input<string>;
  vlanId?: pulumi.Input<number>;
  started?: pulumi.Input<boolean>;
  onBoot?: pulumi.Input<boolean>;
  fullClone?: pulumi.Input<boolean>;
  ipAddress?: pulumi.Input<string>;
  gateway?: pulumi.Input<string>;
  nameservers?: pulumi.Input<pulumi.Input<string>[]>;
}

export interface DigitaloceanPublisherArgs extends CommonPublisherArgs {
  region: pulumi.Input<string>;
  size?: pulumi.Input<string>;
  image?: pulumi.Input<string>;
  sshKeys?: pulumi.Input<pulumi.Input<string>[]>;
  vpcUuid?: pulumi.Input<string>;
  monitoring?: pulumi.Input<boolean>;
  ipv6?: pulumi.Input<boolean>;
}

export interface VultrPublisherArgs extends CommonPublisherArgs {
  region: pulumi.Input<string>;
  plan: pulumi.Input<string>;
  osId?: pulumi.Input<number>;
  imageId?: pulumi.Input<string>;
  sshKeyIds?: pulumi.Input<pulumi.Input<string>[]>;
  vpc2Ids?: pulumi.Input<pulumi.Input<string>[]>;
  enableIpv6?: pulumi.Input<boolean>;
  firewallGroupId?: pulumi.Input<string>;
}

export interface ExoscalePublisherArgs extends CommonPublisherArgs {
  zone: pulumi.Input<string>;
  type: pulumi.Input<string>;
  templateId: pulumi.Input<string>;
  diskSize: pulumi.Input<number>;
  sshKeys?: pulumi.Input<pulumi.Input<string>[]>;
  securityGroupIds?: pulumi.Input<pulumi.Input<string>[]>;
  networkInterfaces?: pulumi.Input<pulumi.Input<Record<string, pulumi.Input<unknown>>>[]>;
}

export interface UpcloudPublisherArgs extends CommonPublisherArgs {
  zone: pulumi.Input<string>;
  hostname?: pulumi.Input<string>;
  plan?: pulumi.Input<string>;
  template?: pulumi.Input<string>;
  networkInterfaces?: pulumi.Input<pulumi.Input<Record<string, pulumi.Input<unknown>>>[]>;
}

export interface StackitPublisherArgs extends CommonPublisherArgs {
  projectId: pulumi.Input<string>;
  machineType: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  availabilityZone?: pulumi.Input<string>;
  keypairName?: pulumi.Input<string>;
  networkInterfaces?: pulumi.Input<pulumi.Input<Record<string, pulumi.Input<unknown>>>[]>;
}

export interface EquinixPublisherArgs extends CommonPublisherArgs {
  projectId: pulumi.Input<string>;
  metro: pulumi.Input<string>;
  plan: pulumi.Input<string>;
  operatingSystem?: pulumi.Input<string>;
  billingCycle?: pulumi.Input<string>;
  projectSshKeyIds?: pulumi.Input<pulumi.Input<string>[]>;
  userSshKeyIds?: pulumi.Input<pulumi.Input<string>[]>;
}

export interface OutscalePublisherArgs extends CommonPublisherArgs {
  imageId: pulumi.Input<string>;
  vmType?: pulumi.Input<string>;
  subnetId?: pulumi.Input<string>;
  keypairName?: pulumi.Input<string>;
  securityGroupIds?: pulumi.Input<pulumi.Input<string>[]>;
  placementSubregionName?: pulumi.Input<string>;
}

export interface OpentelekomcloudPublisherArgs extends CommonPublisherArgs {
  imageName?: pulumi.Input<string>;
  imageId?: pulumi.Input<string>;
  flavorName?: pulumi.Input<string>;
  flavorId?: pulumi.Input<string>;
  networks: pulumi.Input<pulumi.Input<Record<string, pulumi.Input<unknown>>>[]>;
  keyPair?: pulumi.Input<string>;
  availabilityZone?: pulumi.Input<string>;
  securityGroups?: pulumi.Input<pulumi.Input<string>[]>;
}

export interface TencentcloudPublisherArgs extends CommonPublisherArgs {
  availabilityZone: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  instanceType?: pulumi.Input<string>;
  subnetId?: pulumi.Input<string>;
  vpcId?: pulumi.Input<string>;
  keyName?: pulumi.Input<string>;
  securityGroups?: pulumi.Input<pulumi.Input<string>[]>;
  systemDiskType?: pulumi.Input<string>;
  systemDiskSize?: pulumi.Input<number>;
}

export interface YandexPublisherArgs extends CommonPublisherArgs {
  zone?: pulumi.Input<string>;
  platformId?: pulumi.Input<string>;
  imageId: pulumi.Input<string>;
  subnetId: pulumi.Input<string>;
  cores?: pulumi.Input<number>;
  memory?: pulumi.Input<number>;
  coreFraction?: pulumi.Input<number>;
  nat?: pulumi.Input<boolean>;
  sshKeys?: pulumi.Input<pulumi.Input<string>[]>;
}

export interface HypervHardDrive {
  path: pulumi.Input<string>;
  controllerType?: pulumi.Input<string>;
  controllerNumber?: pulumi.Input<number>;
  controllerLocation?: pulumi.Input<number>;
}

export interface HypervPublisherArgs extends CommonPublisherArgs {
  switchName: pulumi.Input<string>;
  hardDrives: pulumi.Input<pulumi.Input<HypervHardDrive>[]>;
  generation?: pulumi.Input<number>;
  processorCount?: pulumi.Input<number>;
  memorySize?: pulumi.Input<number>;
  dynamicMemory?: pulumi.Input<boolean>;
  minimumMemory?: pulumi.Input<number>;
  maximumMemory?: pulumi.Input<number>;
  autoStartAction?: pulumi.Input<string>;
  autoStopAction?: pulumi.Input<string>;
  enableExperimentalHyperv?: boolean;
}

export interface PublisherOutput {
  publisherId: number;
  registrationToken: string;
  vmId: string;
  privateIp: string;
  publicIp?: string;
  placementLabels: string[];
}

export type KubernetesEnrollmentMode = "token" | "api";
export type KubernetesWorkloadType = "daemonset" | "statefulset";

export interface KubernetesPublisherArgs extends CommonPublisherArgs {
  namespace?: pulumi.Input<string>;
  enrollmentMode?: KubernetesEnrollmentMode;
  chartRepository?: pulumi.Input<string>;
  chartVersion?: pulumi.Input<string>;
  chartValues?: pulumi.Input<Record<string, any>>;
  workloadType?: KubernetesWorkloadType;
  hpaEnabled?: pulumi.Input<boolean>;
  hpaMinReplicas?: pulumi.Input<number>;
  hpaMaxReplicas?: pulumi.Input<number>;
  imageRepository?: pulumi.Input<string>;
  imageTag?: pulumi.Input<string>;
}

export interface KubernetesPublisherOutput {
  publisherId?: number;
  registrationToken?: string;
  helmReleaseName: string;
  namespace: string;
  status: string;
  vmId?: string;
  privateIp?: string;
  publicIp?: string;
}
