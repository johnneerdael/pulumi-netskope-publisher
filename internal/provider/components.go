package provider

import (
	"fmt"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AwsPublisherArgs struct {
	CommonPublisherArgs

	SubnetID                 pulumi.StringInput      `pulumi:"subnetId"`
	SecurityGroupIDs         pulumi.StringArrayInput `pulumi:"securityGroupIds"`
	KeyName                  pulumi.StringPtrInput   `pulumi:"keyName,optional"`
	InstanceType             pulumi.StringPtrInput   `pulumi:"instanceType,optional"`
	AMIID                    pulumi.StringPtrInput   `pulumi:"amiId,optional"`
	AssociatePublicIPAddress pulumi.BoolPtrInput     `pulumi:"associatePublicIpAddress,optional"`
	IAMInstanceProfile       pulumi.StringPtrInput   `pulumi:"iamInstanceProfile,optional"`
	EBSOptimized             pulumi.BoolPtrInput     `pulumi:"ebsOptimized,optional"`
	Monitoring               pulumi.BoolPtrInput     `pulumi:"monitoring,optional"`
	MetadataOptions          *MetadataOptions        `pulumi:"metadataOptions,optional"`
}

type AwsPublisher struct {
	PublisherComponent
	AwsPublisherArgs
}

func (*AwsPublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "AwsPublisher")
}

func NewAwsPublisher(ctx *pulumi.Context, name string, args AwsPublisherArgs, opts ...pulumi.ResourceOption) (*AwsPublisher, error) {
	return nil, constructionUnsupported("AwsPublisher")
	component := &AwsPublisher{AwsPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNamesOutput(args.Names)
	component.Publishers = pulumi.Map{}.ToMapOutput()
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type AzurePublisherArgs struct {
	CommonPublisherArgs

	ResourceGroupName      pulumi.StringInput     `pulumi:"resourceGroupName"`
	Location               pulumi.StringInput     `pulumi:"location"`
	SubnetID               pulumi.StringInput     `pulumi:"subnetId"`
	VMSize                 pulumi.StringPtrInput  `pulumi:"vmSize,optional"`
	AdminUsername          pulumi.StringPtrInput  `pulumi:"adminUsername,optional"`
	AdminSSHPublicKey      pulumi.StringInput     `pulumi:"adminSshPublicKey"`
	NetworkSecurityGroupID pulumi.StringPtrInput  `pulumi:"networkSecurityGroupId,optional"`
	AssignPublicIP         pulumi.BoolPtrInput    `pulumi:"assignPublicIp,optional"`
	OSDisk                 *AzureOsDisk           `pulumi:"osDisk,optional"`
	ImageID                pulumi.StringPtrInput  `pulumi:"imageId,optional"`
	Marketplace            *AzureMarketplaceImage `pulumi:"marketplace,optional"`
	AcceptMarketplaceTerms pulumi.BoolPtrInput    `pulumi:"acceptMarketplaceTerms,optional"`
}

type AzurePublisher struct {
	PublisherComponent
	AzurePublisherArgs
}

func (*AzurePublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "AzurePublisher")
}

func NewAzurePublisher(ctx *pulumi.Context, name string, args AzurePublisherArgs, opts ...pulumi.ResourceOption) (*AzurePublisher, error) {
	return nil, constructionUnsupported("AzurePublisher")
	component := &AzurePublisher{AzurePublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNamesOutput(args.Names)
	component.Publishers = pulumi.Map{}.ToMapOutput()
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type GcpPublisherArgs struct {
	CommonPublisherArgs

	Project        pulumi.StringInput      `pulumi:"project"`
	Zone           pulumi.StringInput      `pulumi:"zone"`
	Network        pulumi.StringInput      `pulumi:"network"`
	Subnetwork     pulumi.StringInput      `pulumi:"subnetwork"`
	MachineType    pulumi.StringPtrInput   `pulumi:"machineType,optional"`
	Image          pulumi.StringInput      `pulumi:"image"`
	AssignPublicIP pulumi.BoolPtrInput     `pulumi:"assignPublicIp,optional"`
	NetworkTags    pulumi.StringArrayInput `pulumi:"networkTags,optional"`
	ServiceAccount *GcpServiceAccount      `pulumi:"serviceAccount,optional"`
}

type GcpPublisher struct {
	PublisherComponent
	GcpPublisherArgs
}

func (*GcpPublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "GcpPublisher")
}

func NewGcpPublisher(ctx *pulumi.Context, name string, args GcpPublisherArgs, opts ...pulumi.ResourceOption) (*GcpPublisher, error) {
	return nil, constructionUnsupported("GcpPublisher")
	component := &GcpPublisher{GcpPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNamesOutput(args.Names)
	component.Publishers = pulumi.Map{}.ToMapOutput()
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type VspherePublisherArgs struct {
	CommonPublisherArgs

	Datacenter   pulumi.StringInput    `pulumi:"datacenter"`
	Cluster      pulumi.StringPtrInput `pulumi:"cluster,optional"`
	Host         pulumi.StringPtrInput `pulumi:"host,optional"`
	Datastore    pulumi.StringInput    `pulumi:"datastore"`
	NetworkName  pulumi.StringInput    `pulumi:"networkName"`
	TemplateName pulumi.StringInput    `pulumi:"templateName"`
	Folder       pulumi.StringPtrInput `pulumi:"folder,optional"`
	NumCPUs      pulumi.IntPtrInput    `pulumi:"numCpus,optional"`
	Memory       pulumi.IntPtrInput    `pulumi:"memory,optional"`
}

type VspherePublisher struct {
	PublisherComponent
	VspherePublisherArgs
}

func (*VspherePublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "VspherePublisher")
}

func NewVspherePublisher(ctx *pulumi.Context, name string, args VspherePublisherArgs, opts ...pulumi.ResourceOption) (*VspherePublisher, error) {
	return nil, constructionUnsupported("VspherePublisher")
	component := &VspherePublisher{VspherePublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNamesOutput(args.Names)
	component.Publishers = pulumi.Map{}.ToMapOutput()
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type HypervPublisherArgs struct {
	CommonPublisherArgs

	SwitchName               pulumi.StringInput    `pulumi:"switchName"`
	HardDrives               []HypervHardDrive     `pulumi:"hardDrives"`
	Generation               pulumi.IntPtrInput    `pulumi:"generation,optional"`
	ProcessorCount           pulumi.IntPtrInput    `pulumi:"processorCount,optional"`
	MemorySize               pulumi.IntPtrInput    `pulumi:"memorySize,optional"`
	DynamicMemory            pulumi.BoolPtrInput   `pulumi:"dynamicMemory,optional"`
	MinimumMemory            pulumi.IntPtrInput    `pulumi:"minimumMemory,optional"`
	MaximumMemory            pulumi.IntPtrInput    `pulumi:"maximumMemory,optional"`
	AutoStartAction          pulumi.StringPtrInput `pulumi:"autoStartAction,optional"`
	AutoStopAction           pulumi.StringPtrInput `pulumi:"autoStopAction,optional"`
	EnableExperimentalHyperv bool                  `pulumi:"enableExperimentalHyperv,optional"`
}

type HypervPublisher struct {
	PublisherComponent
	HypervPublisherArgs
}

func (*HypervPublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "HypervPublisher")
}

func NewHypervPublisher(ctx *pulumi.Context, name string, args HypervPublisherArgs, opts ...pulumi.ResourceOption) (*HypervPublisher, error) {
	if !args.EnableExperimentalHyperv {
		return nil, fmt.Errorf("Hyper-V support is experimental and requires enableExperimentalHyperv: true")
	}

	component := &HypervPublisher{HypervPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNamesOutput(args.Names)
	component.Publishers = pulumi.Map{}.ToMapOutput()
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

func constructionUnsupported(component string) error {
	return fmt.Errorf("Go provider child-resource parity is not implemented for %s yet; use the TypeScript component package for deployments", component)
}

func publisherNamesOutput(names pulumi.StringArrayInput) pulumi.StringArrayOutput {
	if names == nil {
		return pulumi.StringArray{}.ToStringArrayOutput()
	}

	return names.ToStringArrayOutput()
}
