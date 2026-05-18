package provider

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

type CommonPublisherArgs struct {
	NamePrefix    pulumi.StringPtrInput   `pulumi:"namePrefix,optional"`
	Names         pulumi.StringArrayInput `pulumi:"names,optional"`
	Replicas      pulumi.IntPtrInput      `pulumi:"replicas,optional"`
	TenantURL     pulumi.StringPtrInput   `pulumi:"tenantUrl,optional"`
	APIToken      pulumi.StringPtrInput   `pulumi:"apiToken,optional" provider:"secret"`
	WizardPath    pulumi.StringPtrInput   `pulumi:"wizardPath,optional"`
	Tags          pulumi.StringMapInput   `pulumi:"tags,optional"`
	Registrations pulumi.MapInput         `pulumi:"registrations,optional"`
}

type PublisherOutput struct {
	PublisherID       pulumi.IntOutput    `pulumi:"publisherId"`
	RegistrationToken pulumi.StringOutput `pulumi:"registrationToken" provider:"secret"`
	VMID              pulumi.StringOutput `pulumi:"vmId"`
	PrivateIP         pulumi.StringOutput `pulumi:"privateIp"`
	PublicIP          pulumi.StringOutput `pulumi:"publicIp"`
}

type PublisherComponent struct {
	pulumi.ResourceState

	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

type MetadataOptions struct {
	HTTPEndpoint pulumi.StringPtrInput `pulumi:"httpEndpoint,optional"`
	HTTPTokens   pulumi.StringPtrInput `pulumi:"httpTokens,optional"`
}

type AzureMarketplaceImage struct {
	Publisher pulumi.StringInput    `pulumi:"publisher"`
	Offer     pulumi.StringInput    `pulumi:"offer"`
	SKU       pulumi.StringInput    `pulumi:"sku"`
	Version   pulumi.StringPtrInput `pulumi:"version,optional"`
}

type AzureOsDisk struct {
	Type   pulumi.StringPtrInput `pulumi:"type,optional"`
	SizeGB pulumi.IntPtrInput    `pulumi:"sizeGb,optional"`
}

type GcpServiceAccount struct {
	Email  pulumi.StringInput      `pulumi:"email"`
	Scopes pulumi.StringArrayInput `pulumi:"scopes,optional"`
}

type HypervHardDrive struct {
	Path               pulumi.StringInput    `pulumi:"path"`
	ControllerType     pulumi.StringPtrInput `pulumi:"controllerType,optional"`
	ControllerNumber   pulumi.IntPtrInput    `pulumi:"controllerNumber,optional"`
	ControllerLocation pulumi.IntPtrInput    `pulumi:"controllerLocation,optional"`
}
