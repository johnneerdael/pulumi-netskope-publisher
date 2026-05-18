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

export interface PublisherOutput {
  publisherId: number;
  registrationToken: string;
  vmId: string;
  privateIp: string;
  publicIp?: string;
}
