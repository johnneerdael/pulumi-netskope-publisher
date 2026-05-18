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

export interface PublisherRegistrationInput {
  publisherId: pulumi.Input<number>;
  registrationToken: pulumi.Input<string>;
}

export interface CommonPublisherArgs extends NameArgs {
  namePrefix?: string;
  names?: string[];
  replicas?: number;
  tenantUrl?: pulumi.Input<string>;
  apiToken?: pulumi.Input<string>;
  wizardPath?: pulumi.Input<string>;
  tags?: pulumi.Input<Record<string, pulumi.Input<string>>>;
  registrations?: pulumi.Input<Record<string, PublisherRegistrationInput>>;
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
  bootstrap?: pulumi.Input<boolean>;
  bootstrapUrl?: pulumi.Input<string>;
  nonat?: pulumi.Input<boolean>;
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
}
