package provider

import (
	"encoding/base64"
	"fmt"
	"strings"

	awsec2 "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	azurecompute "github.com/pulumi/pulumi-azure-native-sdk/compute/v3"
	azurenetwork "github.com/pulumi/pulumi-azure-native-sdk/network/v3"
	gcpcompute "github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-vsphere/sdk/v4/go/vsphere"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AwsPublisherArgs struct {
	NamePrefix    *string                               `pulumi:"namePrefix,optional"`
	Names         []string                              `pulumi:"names,optional"`
	Replicas      *int                                  `pulumi:"replicas,optional"`
	TenantURL     *string                               `pulumi:"tenantUrl,optional"`
	APIToken      *string                               `pulumi:"apiToken,optional" provider:"secret"`
	WizardPath    *string                               `pulumi:"wizardPath,optional"`
	Tags          map[string]string                     `pulumi:"tags,optional"`
	Registrations map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	SubnetID                 string           `pulumi:"subnetId"`
	SecurityGroupIDs         []string         `pulumi:"securityGroupIds"`
	KeyName                  *string          `pulumi:"keyName,optional"`
	InstanceType             *string          `pulumi:"instanceType,optional"`
	AMIID                    *string          `pulumi:"amiId,optional"`
	AssociatePublicIPAddress *bool            `pulumi:"associatePublicIpAddress,optional"`
	IAMInstanceProfile       *string          `pulumi:"iamInstanceProfile,optional"`
	EBSOptimized             *bool            `pulumi:"ebsOptimized,optional"`
	Monitoring               *bool            `pulumi:"monitoring,optional"`
	MetadataOptions          *MetadataOptions `pulumi:"metadataOptions,optional"`
}

type AwsPublisher struct {
	pulumi.ResourceState
	AwsPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*AwsPublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "AwsPublisher")
}

func NewAwsPublisher(ctx *pulumi.Context, name string, args AwsPublisherArgs, opts ...pulumi.ResourceOption) (*AwsPublisher, error) {
	component := &AwsPublisher{AwsPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}

	publisherNames, registrations, err := resolvePublisherInputs(args.common())
	if err != nil {
		return nil, err
	}

	amiID := args.AMIID
	if amiID == nil {
		ami, err := awsec2.LookupAmi(ctx, &awsec2.LookupAmiArgs{
			MostRecent: pulumi.BoolRef(true),
			Owners:     []string{"679593333241"},
			Filters: []awsec2.GetAmiFilter{{
				Name:   "name",
				Values: []string{"Netskope Private Access Publisher*"},
			}},
		}, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}
		amiID = &ami.Id
	}

	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		instance, err := awsec2.NewInstance(ctx, name+"-"+publisherName, &awsec2.InstanceArgs{
			Ami:                      pulumi.StringPtr(*amiID),
			InstanceType:             pulumi.StringPtr(defaultString(args.InstanceType, "t3.medium")),
			SubnetId:                 pulumi.StringPtr(args.SubnetID),
			VpcSecurityGroupIds:      toStringArray(args.SecurityGroupIDs),
			KeyName:                  stringPtrInput(args.KeyName),
			AssociatePublicIpAddress: pulumi.BoolPtr(defaultBool(args.AssociatePublicIPAddress, false)),
			IamInstanceProfile:       stringPtrInput(args.IAMInstanceProfile),
			EbsOptimized:             pulumi.BoolPtr(defaultBool(args.EBSOptimized, true)),
			Monitoring:               pulumi.BoolPtr(defaultBool(args.Monitoring, true)),
			UserDataBase64:           pulumi.StringPtr(renderUserDataBase64(publisherName, registration.RegistrationToken, args.WizardPath)),
			MetadataOptions: &awsec2.InstanceMetadataOptionsArgs{
				HttpEndpoint: pulumi.StringPtr(defaultMetadataValue(args.MetadataOptions, "endpoint", "enabled")),
				HttpTokens:   pulumi.StringPtr(defaultMetadataValue(args.MetadataOptions, "tokens", "required")),
			},
			Tags: toStringMap(nameTag(args.Tags, publisherName)),
		}, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}

		outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), instance.PrivateIp, instance.PublicIp)
	}

	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type AzurePublisherArgs struct {
	NamePrefix    *string                               `pulumi:"namePrefix,optional"`
	Names         []string                              `pulumi:"names,optional"`
	Replicas      *int                                  `pulumi:"replicas,optional"`
	TenantURL     *string                               `pulumi:"tenantUrl,optional"`
	APIToken      *string                               `pulumi:"apiToken,optional" provider:"secret"`
	WizardPath    *string                               `pulumi:"wizardPath,optional"`
	Tags          map[string]string                     `pulumi:"tags,optional"`
	Registrations map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	ResourceGroupName      string                 `pulumi:"resourceGroupName"`
	Location               string                 `pulumi:"location"`
	SubnetID               string                 `pulumi:"subnetId"`
	VMSize                 *string                `pulumi:"vmSize,optional"`
	AdminUsername          *string                `pulumi:"adminUsername,optional"`
	AdminSSHPublicKey      string                 `pulumi:"adminSshPublicKey"`
	NetworkSecurityGroupID *string                `pulumi:"networkSecurityGroupId,optional"`
	AssignPublicIP         *bool                  `pulumi:"assignPublicIp,optional"`
	OSDisk                 *AzureOsDisk           `pulumi:"osDisk,optional"`
	ImageID                *string                `pulumi:"imageId,optional"`
	Marketplace            *AzureMarketplaceImage `pulumi:"marketplace,optional"`
	AcceptMarketplaceTerms *bool                  `pulumi:"acceptMarketplaceTerms,optional"`
}

type AzurePublisher struct {
	pulumi.ResourceState
	AzurePublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*AzurePublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "AzurePublisher")
}

func NewAzurePublisher(ctx *pulumi.Context, name string, args AzurePublisherArgs, opts ...pulumi.ResourceOption) (*AzurePublisher, error) {
	if args.ImageID == nil && args.Marketplace == nil {
		return nil, fmt.Errorf("provide either imageId or marketplace")
	}

	component := &AzurePublisher{AzurePublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}

	publisherNames, registrations, err := resolvePublisherInputs(args.common())
	if err != nil {
		return nil, err
	}

	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		var publicIP *azurenetwork.PublicIPAddress
		if defaultBool(args.AssignPublicIP, false) {
			publicIP, err = azurenetwork.NewPublicIPAddress(ctx, name+"-"+publisherName+"-pip", &azurenetwork.PublicIPAddressArgs{
				PublicIpAddressName:      pulumi.StringPtr(publisherName + "-pip"),
				ResourceGroupName:        pulumi.String(args.ResourceGroupName),
				Location:                 pulumi.StringPtr(args.Location),
				PublicIPAllocationMethod: pulumi.StringPtr("Static"),
				Sku: &azurenetwork.PublicIPAddressSkuArgs{
					Name: pulumi.StringPtr("Standard"),
				},
				Tags: toStringMap(args.Tags),
			}, pulumi.Parent(component))
			if err != nil {
				return nil, err
			}
		}

		ipConfig := azurenetwork.NetworkInterfaceIPConfigurationArgs{
			Name:                      pulumi.StringPtr("primary"),
			PrivateIPAllocationMethod: pulumi.StringPtr("Dynamic"),
			Subnet: &azurenetwork.SubnetTypeArgs{
				Id: pulumi.StringPtr(args.SubnetID),
			},
		}
		if publicIP != nil {
			ipConfig.PublicIPAddress = &azurenetwork.PublicIPAddressTypeArgs{Id: publicIP.ID().ToStringPtrOutput()}
		}

		nicArgs := &azurenetwork.NetworkInterfaceArgs{
			NetworkInterfaceName: pulumi.StringPtr(publisherName + "-nic"),
			ResourceGroupName:    pulumi.String(args.ResourceGroupName),
			Location:             pulumi.StringPtr(args.Location),
			Tags:                 toStringMap(args.Tags),
			IpConfigurations: azurenetwork.NetworkInterfaceIPConfigurationArray{
				ipConfig,
			},
		}
		if args.NetworkSecurityGroupID != nil {
			nicArgs.NetworkSecurityGroup = &azurenetwork.NetworkSecurityGroupTypeArgs{
				Id: pulumi.StringPtr(*args.NetworkSecurityGroupID),
			}
		}
		nic, err := azurenetwork.NewNetworkInterface(ctx, name+"-"+publisherName+"-nic", nicArgs, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}

		vm, err := azurecompute.NewVirtualMachine(ctx, name+"-"+publisherName, &azurecompute.VirtualMachineArgs{
			VmName:            pulumi.StringPtr(publisherName),
			ResourceGroupName: pulumi.String(args.ResourceGroupName),
			Location:          pulumi.StringPtr(args.Location),
			Tags:              toStringMap(nameTag(args.Tags, publisherName)),
			HardwareProfile: &azurecompute.HardwareProfileArgs{
				VmSize: pulumi.StringPtr(defaultString(args.VMSize, "Standard_D2s_v5")),
			},
			NetworkProfile: &azurecompute.NetworkProfileArgs{
				NetworkInterfaces: azurecompute.NetworkInterfaceReferenceArray{
					azurecompute.NetworkInterfaceReferenceArgs{
						Id:      nic.ID().ToStringPtrOutput(),
						Primary: pulumi.BoolPtr(true),
					},
				},
			},
			OsProfile: &azurecompute.OSProfileArgs{
				ComputerName:  pulumi.StringPtr(publisherName),
				AdminUsername: pulumi.StringPtr(defaultString(args.AdminUsername, "ubuntu")),
				CustomData:    pulumi.StringPtr(renderUserDataBase64(publisherName, registration.RegistrationToken, args.WizardPath)),
				LinuxConfiguration: &azurecompute.LinuxConfigurationArgs{
					DisablePasswordAuthentication: pulumi.BoolPtr(true),
					Ssh: &azurecompute.SshConfigurationArgs{
						PublicKeys: azurecompute.SshPublicKeyTypeArray{
							azurecompute.SshPublicKeyTypeArgs{
								Path:    pulumi.StringPtr("/home/" + defaultString(args.AdminUsername, "ubuntu") + "/.ssh/authorized_keys"),
								KeyData: pulumi.StringPtr(args.AdminSSHPublicKey),
							},
						},
					},
				},
			},
			StorageProfile: azureStorageProfile(args),
			Plan:           azurePlan(args.Marketplace, args.ImageID),
		}, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}

		publicIPOutput := pulumi.String("").ToStringOutput()
		if publicIP != nil {
			publicIPOutput = publicIP.IpAddress.Elem()
		}
		outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), publicIPOutput)
	}

	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type GcpPublisherArgs struct {
	NamePrefix    *string                               `pulumi:"namePrefix,optional"`
	Names         []string                              `pulumi:"names,optional"`
	Replicas      *int                                  `pulumi:"replicas,optional"`
	TenantURL     *string                               `pulumi:"tenantUrl,optional"`
	APIToken      *string                               `pulumi:"apiToken,optional" provider:"secret"`
	WizardPath    *string                               `pulumi:"wizardPath,optional"`
	Tags          map[string]string                     `pulumi:"tags,optional"`
	Registrations map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Project        string             `pulumi:"project"`
	Zone           string             `pulumi:"zone"`
	Network        string             `pulumi:"network"`
	Subnetwork     string             `pulumi:"subnetwork"`
	MachineType    *string            `pulumi:"machineType,optional"`
	Image          string             `pulumi:"image"`
	AssignPublicIP *bool              `pulumi:"assignPublicIp,optional"`
	NetworkTags    []string           `pulumi:"networkTags,optional"`
	ServiceAccount *GcpServiceAccount `pulumi:"serviceAccount,optional"`
}

type GcpPublisher struct {
	pulumi.ResourceState
	GcpPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*GcpPublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "GcpPublisher")
}

func NewGcpPublisher(ctx *pulumi.Context, name string, args GcpPublisherArgs, opts ...pulumi.ResourceOption) (*GcpPublisher, error) {
	component := &GcpPublisher{GcpPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}

	publisherNames, registrations, err := resolvePublisherInputs(args.common())
	if err != nil {
		return nil, err
	}

	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		networkInterface := &gcpcompute.InstanceNetworkInterfaceArgs{
			Network:    pulumi.StringPtr(args.Network),
			Subnetwork: pulumi.StringPtr(args.Subnetwork),
		}
		if defaultBool(args.AssignPublicIP, false) {
			networkInterface.AccessConfigs = gcpcompute.InstanceNetworkInterfaceAccessConfigArray{
				gcpcompute.InstanceNetworkInterfaceAccessConfigArgs{},
			}
		}
		instanceArgs := &gcpcompute.InstanceArgs{
			Name:        pulumi.StringPtr(publisherName),
			Project:     pulumi.StringPtr(args.Project),
			Zone:        pulumi.StringPtr(args.Zone),
			MachineType: pulumi.String(defaultString(args.MachineType, "e2-medium")),
			Tags:        toStringArray(args.NetworkTags),
			Labels:      toStringMap(args.Tags),
			BootDisk: &gcpcompute.InstanceBootDiskArgs{
				InitializeParams: &gcpcompute.InstanceBootDiskInitializeParamsArgs{
					Image: pulumi.StringPtr(args.Image),
				},
			},
			NetworkInterfaces: gcpcompute.InstanceNetworkInterfaceArray{networkInterface},
			Metadata: pulumi.StringMap{
				"user-data": pulumi.String(renderUserData(publisherName, registration.RegistrationToken, args.WizardPath)),
			},
		}
		if args.ServiceAccount != nil {
			scopes := args.ServiceAccount.Scopes
			if len(scopes) == 0 {
				scopes = []string{"https://www.googleapis.com/auth/cloud-platform"}
			}
			instanceArgs.ServiceAccount = &gcpcompute.InstanceServiceAccountArgs{
				Email:  pulumi.StringPtr(args.ServiceAccount.Email),
				Scopes: toStringArray(scopes),
			}
		}
		instance, err := gcpcompute.NewInstance(ctx, name+"-"+publisherName, instanceArgs, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}

		outputs[publisherName] = publisherOutput(registration, instance.InstanceId, pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput())
	}

	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type VspherePublisherArgs struct {
	NamePrefix    *string                               `pulumi:"namePrefix,optional"`
	Names         []string                              `pulumi:"names,optional"`
	Replicas      *int                                  `pulumi:"replicas,optional"`
	TenantURL     *string                               `pulumi:"tenantUrl,optional"`
	APIToken      *string                               `pulumi:"apiToken,optional" provider:"secret"`
	WizardPath    *string                               `pulumi:"wizardPath,optional"`
	Tags          map[string]string                     `pulumi:"tags,optional"`
	Registrations map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Datacenter   string  `pulumi:"datacenter"`
	Cluster      *string `pulumi:"cluster,optional"`
	Host         *string `pulumi:"host,optional"`
	Datastore    string  `pulumi:"datastore"`
	NetworkName  string  `pulumi:"networkName"`
	TemplateName string  `pulumi:"templateName"`
	Folder       *string `pulumi:"folder,optional"`
	NumCPUs      *int    `pulumi:"numCpus,optional"`
	Memory       *int    `pulumi:"memory,optional"`
}

type VspherePublisher struct {
	pulumi.ResourceState
	VspherePublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*VspherePublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "VspherePublisher")
}

func NewVspherePublisher(ctx *pulumi.Context, name string, args VspherePublisherArgs, opts ...pulumi.ResourceOption) (*VspherePublisher, error) {
	if args.Cluster == nil && args.Host == nil {
		return nil, fmt.Errorf("provide either cluster or host")
	}

	component := &VspherePublisher{VspherePublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}

	publisherNames, registrations, err := resolvePublisherInputs(args.common())
	if err != nil {
		return nil, err
	}

	datacenter := vsphere.LookupDatacenterOutput(ctx, vsphere.LookupDatacenterOutputArgs{Name: pulumi.String(args.Datacenter)}, pulumi.Parent(component))
	datastore := vsphere.GetDatastoreOutput(ctx, vsphere.GetDatastoreOutputArgs{Name: pulumi.String(args.Datastore), DatacenterId: datacenter.Id()}, pulumi.Parent(component))
	network := vsphere.GetNetworkOutput(ctx, vsphere.GetNetworkOutputArgs{Name: pulumi.String(args.NetworkName), DatacenterId: datacenter.Id()}, pulumi.Parent(component))
	template := vsphere.LookupVirtualMachineOutput(ctx, vsphere.LookupVirtualMachineOutputArgs{Name: pulumi.String(args.TemplateName), DatacenterId: datacenter.Id()}, pulumi.Parent(component))

	var resourcePoolID pulumi.StringOutput
	if args.Cluster != nil {
		cluster := vsphere.LookupComputeClusterOutput(ctx, vsphere.LookupComputeClusterOutputArgs{Name: pulumi.String(*args.Cluster), DatacenterId: datacenter.Id()}, pulumi.Parent(component))
		resourcePoolID = cluster.ResourcePoolId()
	} else {
		host := vsphere.LookupHostOutput(ctx, vsphere.LookupHostOutputArgs{Name: pulumi.String(*args.Host), DatacenterId: datacenter.Id()}, pulumi.Parent(component))
		resourcePoolID = host.ResourcePoolId()
	}

	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		vm, err := vsphere.NewVirtualMachine(ctx, name+"-"+publisherName, &vsphere.VirtualMachineArgs{
			Name:           pulumi.StringPtr(publisherName),
			ResourcePoolId: resourcePoolID,
			DatastoreId:    datastore.Id(),
			Folder:         stringPtrInput(args.Folder),
			NumCpus:        pulumi.IntPtr(defaultInt(args.NumCPUs, 2)),
			Memory:         pulumi.IntPtr(defaultInt(args.Memory, 4096)),
			GuestId:        template.GuestId(),
			NetworkInterfaces: vsphere.VirtualMachineNetworkInterfaceArray{
				vsphere.VirtualMachineNetworkInterfaceArgs{
					NetworkId:   network.Id(),
					AdapterType: template.NetworkInterfaceTypes().Index(pulumi.Int(0)),
				},
			},
			Disks: vsphere.VirtualMachineDiskArray{
				vsphere.VirtualMachineDiskArgs{
					Label: pulumi.String("disk0"),
					Size:  template.Disks().Index(pulumi.Int(0)).Size(),
				},
			},
			Clone: &vsphere.VirtualMachineCloneArgs{
				TemplateUuid: template.Id(),
			},
			ExtraConfig: pulumi.StringMap{
				"guestinfo.userdata":          pulumi.String(renderUserDataBase64(publisherName, registration.RegistrationToken, args.WizardPath)),
				"guestinfo.userdata.encoding": pulumi.String("base64"),
				"guestinfo.metadata":          pulumi.String(base64.StdEncoding.EncodeToString([]byte(renderMetadata(publisherName)))),
				"guestinfo.metadata.encoding": pulumi.String("base64"),
			},
		}, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}

		outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), vm.DefaultIpAddress, pulumi.String("").ToStringOutput())
	}

	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type HypervPublisherArgs struct {
	NamePrefix    *string                               `pulumi:"namePrefix,optional"`
	Names         []string                              `pulumi:"names,optional"`
	Replicas      *int                                  `pulumi:"replicas,optional"`
	TenantURL     *string                               `pulumi:"tenantUrl,optional"`
	APIToken      *string                               `pulumi:"apiToken,optional" provider:"secret"`
	WizardPath    *string                               `pulumi:"wizardPath,optional"`
	Tags          map[string]string                     `pulumi:"tags,optional"`
	Registrations map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	SwitchName               string            `pulumi:"switchName"`
	HardDrives               []HypervHardDrive `pulumi:"hardDrives"`
	Generation               *int              `pulumi:"generation,optional"`
	ProcessorCount           *int              `pulumi:"processorCount,optional"`
	MemorySize               *int              `pulumi:"memorySize,optional"`
	DynamicMemory            *bool             `pulumi:"dynamicMemory,optional"`
	MinimumMemory            *int              `pulumi:"minimumMemory,optional"`
	MaximumMemory            *int              `pulumi:"maximumMemory,optional"`
	AutoStartAction          *string           `pulumi:"autoStartAction,optional"`
	AutoStopAction           *string           `pulumi:"autoStopAction,optional"`
	EnableExperimentalHyperv bool              `pulumi:"enableExperimentalHyperv,optional"`
}

type HypervPublisher struct {
	pulumi.ResourceState
	HypervPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*HypervPublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "HypervPublisher")
}

func NewHypervPublisher(ctx *pulumi.Context, name string, args HypervPublisherArgs, opts ...pulumi.ResourceOption) (*HypervPublisher, error) {
	if !args.EnableExperimentalHyperv {
		return nil, fmt.Errorf("Hyper-V support is experimental and requires enableExperimentalHyperv: true")
	}

	return nil, fmt.Errorf("Hyper-V support requires a stable Pulumi Hyper-V provider SDK")
}

func (args AwsPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		TenantURL: args.TenantURL, APIToken: args.APIToken, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
	}
}

func (args AzurePublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		TenantURL: args.TenantURL, APIToken: args.APIToken, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
	}
}

func (args GcpPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		TenantURL: args.TenantURL, APIToken: args.APIToken, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
	}
}

func (args VspherePublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		TenantURL: args.TenantURL, APIToken: args.APIToken, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
	}
}

func resolvePublisherInputs(args CommonPublisherArgs) ([]string, map[string]PublisherRegistrationInput, error) {
	names, err := derivePublisherNames(args)
	if err != nil {
		return nil, nil, err
	}
	if len(args.Registrations) == 0 {
		return nil, nil, fmt.Errorf("registrations are required by the Go provider until Netskope registration is implemented as a provider resource")
	}
	for _, name := range names {
		if _, ok := args.Registrations[name]; !ok {
			return nil, nil, fmt.Errorf("registrations is missing data for publisher %s", name)
		}
	}
	return names, args.Registrations, nil
}

func derivePublisherNames(args CommonPublisherArgs) ([]string, error) {
	if len(args.Names) > 0 {
		return args.Names, nil
	}
	replicas := defaultInt(args.Replicas, 1)
	if replicas < 1 {
		return nil, fmt.Errorf("replicas must be >= 1")
	}
	prefix := defaultString(args.NamePrefix, "npa-publisher")
	names := make([]string, replicas)
	for i := range names {
		names[i] = fmt.Sprintf("%s-%d", prefix, i+1)
	}
	return names, nil
}

func publisherOutput(registration PublisherRegistrationInput, vmID pulumi.StringOutput, privateIP pulumi.StringOutput, publicIP pulumi.StringOutput) pulumi.MapOutput {
	return pulumi.Map{
		"publisherId":       pulumi.Int(registration.PublisherID),
		"registrationToken": pulumi.ToSecret(pulumi.String(registration.RegistrationToken)),
		"vmId":              vmID,
		"privateIp":         privateIP,
		"publicIp":          publicIP,
	}.ToMapOutput()
}

func renderUserData(publisherName string, registrationToken string, wizardPath *string) string {
	path := defaultString(wizardPath, "/home/ubuntu/npa_publisher_wizard")
	return strings.Join([]string{
		"#cloud-config",
		"hostname: " + publisherName,
		"preserve_hostname: false",
		"runcmd:",
		fmt.Sprintf("  - [ %s, -token, \"%s\" ]", path, escapeDoubleQuoted(registrationToken)),
		"",
	}, "\n")
}

func renderUserDataBase64(publisherName string, registrationToken string, wizardPath *string) string {
	return base64.StdEncoding.EncodeToString([]byte(renderUserData(publisherName, registrationToken, wizardPath)))
}

func renderMetadata(publisherName string) string {
	return "instance-id: " + publisherName + "\nlocal-hostname: " + publisherName + "\n"
}

func azureStorageProfile(args AzurePublisherArgs) azurecompute.StorageProfilePtrInput {
	osDiskType := "Premium_LRS"
	osDiskSize := 64
	if args.OSDisk != nil {
		osDiskType = defaultString(args.OSDisk.Type, osDiskType)
		osDiskSize = defaultInt(args.OSDisk.SizeGB, osDiskSize)
	}
	profile := &azurecompute.StorageProfileArgs{
		OsDisk: &azurecompute.OSDiskArgs{
			CreateOption: pulumi.String("FromImage"),
			Caching:      azurecompute.CachingTypesReadWrite,
			ManagedDisk: &azurecompute.ManagedDiskParametersArgs{
				StorageAccountType: pulumi.StringPtr(osDiskType),
			},
			DiskSizeGB: pulumi.IntPtr(osDiskSize),
		},
	}
	if args.ImageID != nil {
		profile.ImageReference = &azurecompute.ImageReferenceArgs{Id: pulumi.StringPtr(*args.ImageID)}
	} else if args.Marketplace != nil {
		profile.ImageReference = &azurecompute.ImageReferenceArgs{
			Publisher: pulumi.StringPtr(args.Marketplace.Publisher),
			Offer:     pulumi.StringPtr(args.Marketplace.Offer),
			Sku:       pulumi.StringPtr(args.Marketplace.SKU),
			Version:   pulumi.StringPtr(defaultString(args.Marketplace.Version, "latest")),
		}
	}
	return profile
}

func azurePlan(marketplace *AzureMarketplaceImage, imageID *string) azurecompute.PlanPtrInput {
	if imageID != nil || marketplace == nil {
		return nil
	}
	return &azurecompute.PlanArgs{
		Publisher: pulumi.StringPtr(marketplace.Publisher),
		Product:   pulumi.StringPtr(marketplace.Offer),
		Name:      pulumi.StringPtr(marketplace.SKU),
	}
}

func defaultString(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}
	return *value
}

func defaultInt(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func defaultBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func defaultMetadataValue(options *MetadataOptions, field string, fallback string) string {
	if options == nil {
		return fallback
	}
	if field == "endpoint" {
		return defaultString(options.HTTPEndpoint, fallback)
	}
	return defaultString(options.HTTPTokens, fallback)
}

func stringPtrInput(value *string) pulumi.StringPtrInput {
	if value == nil {
		return nil
	}
	return pulumi.StringPtr(*value)
}

func toStringArray(values []string) pulumi.StringArray {
	result := make(pulumi.StringArray, len(values))
	for i, value := range values {
		result[i] = pulumi.String(value)
	}
	return result
}

func toStringMap(values map[string]string) pulumi.StringMap {
	result := pulumi.StringMap{}
	for key, value := range values {
		result[key] = pulumi.String(value)
	}
	return result
}

func nameTag(tags map[string]string, publisherName string) map[string]string {
	result := map[string]string{}
	for key, value := range tags {
		result[key] = value
	}
	result["Name"] = publisherName
	return result
}

func escapeDoubleQuoted(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `"`, `\"`)
}
