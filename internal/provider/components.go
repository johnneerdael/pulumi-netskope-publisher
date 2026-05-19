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
	k8score "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	k8shelm "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	k8smeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
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

	SubnetID                     string                 `pulumi:"subnetId"`
	SecurityGroupIDs             []string               `pulumi:"securityGroupIds"`
	KeyName                      *string                `pulumi:"keyName,optional"`
	InstanceType                 *string                `pulumi:"instanceType,optional"`
	AMIID                        *string                `pulumi:"amiId,optional"`
	AssociatePublicIPAddress     *bool                  `pulumi:"associatePublicIpAddress,optional"`
	IAMInstanceProfile           *string                `pulumi:"iamInstanceProfile,optional"`
	EBSOptimized                 *bool                  `pulumi:"ebsOptimized,optional"`
	Monitoring                   *bool                  `pulumi:"monitoring,optional"`
	MetadataOptions              *MetadataOptions       `pulumi:"metadataOptions,optional"`
	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
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

	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}

	amiID := args.AMIID
	if amiID == nil {
		owners := []string{"679593333241"}
		filters := []awsec2.GetAmiFilter{{
			Name:   "name",
			Values: []string{"Netskope Private Access Publisher*"},
		}}
		if defaultBool(args.Bootstrap, false) {
			owners = []string{"099720109477"}
			filters = []awsec2.GetAmiFilter{{
				Name:   "name",
				Values: []string{"ubuntu-minimal/images/hvm-ssd*/ubuntu-jammy-22.04-amd64-minimal-*"},
			}, {
				Name:   "architecture",
				Values: []string{"x86_64"},
			}, {
				Name:   "virtualization-type",
				Values: []string{"hvm"},
			}}
		}
		ami, err := awsec2.LookupAmi(ctx, &awsec2.LookupAmiArgs{
			MostRecent: pulumi.BoolRef(true),
			Owners:     owners,
			Filters:    filters,
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
			UserDataBase64:           renderUserDataBase64OutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromAws(args)).ToStringPtrOutput(),
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

	ResourceGroupName            string                 `pulumi:"resourceGroupName"`
	Location                     string                 `pulumi:"location"`
	SubnetID                     string                 `pulumi:"subnetId"`
	VMSize                       *string                `pulumi:"vmSize,optional"`
	AdminUsername                *string                `pulumi:"adminUsername,optional"`
	AdminSSHPublicKey            string                 `pulumi:"adminSshPublicKey"`
	NetworkSecurityGroupID       *string                `pulumi:"networkSecurityGroupId,optional"`
	AssignPublicIP               *bool                  `pulumi:"assignPublicIp,optional"`
	OSDisk                       *AzureOsDisk           `pulumi:"osDisk,optional"`
	ImageID                      *string                `pulumi:"imageId,optional"`
	Marketplace                  *AzureMarketplaceImage `pulumi:"marketplace,optional"`
	AcceptMarketplaceTerms       *bool                  `pulumi:"acceptMarketplaceTerms,optional"`
	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
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
	if args.ImageID == nil && args.Marketplace == nil && !defaultBool(args.Bootstrap, false) {
		return nil, fmt.Errorf("provide imageId, marketplace, or set bootstrap: true")
	}

	component := &AzurePublisher{AzurePublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}

	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}

	outputs := pulumi.Map{}
	adminUsername := defaultString(args.AdminUsername, defaultString(args.InstallUser, "ubuntu"))
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
				AdminUsername: pulumi.StringPtr(adminUsername),
				CustomData:    renderUserDataBase64OutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromAzure(args)).ToStringPtrOutput(),
				LinuxConfiguration: &azurecompute.LinuxConfigurationArgs{
					DisablePasswordAuthentication: pulumi.BoolPtr(true),
					Ssh: &azurecompute.SshConfigurationArgs{
						PublicKeys: azurecompute.SshPublicKeyTypeArray{
							azurecompute.SshPublicKeyTypeArgs{
								Path:    pulumi.StringPtr("/home/" + adminUsername + "/.ssh/authorized_keys"),
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

	Project                      string                 `pulumi:"project"`
	Zone                         string                 `pulumi:"zone"`
	Network                      string                 `pulumi:"network"`
	Subnetwork                   string                 `pulumi:"subnetwork"`
	MachineType                  *string                `pulumi:"machineType,optional"`
	Image                        string                 `pulumi:"image"`
	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	AssignPublicIP               *bool                  `pulumi:"assignPublicIp,optional"`
	NetworkTags                  []string               `pulumi:"networkTags,optional"`
	ServiceAccount               *GcpServiceAccount     `pulumi:"serviceAccount,optional"`
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

	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
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
				"user-data": renderUserDataOutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromGcp(args)),
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

type KubernetesPublisherArgs struct {
	NamePrefix    *string                               `pulumi:"namePrefix,optional"`
	Names         []string                              `pulumi:"names,optional"`
	Replicas      *int                                  `pulumi:"replicas,optional"`
	TenantURL     *string                               `pulumi:"tenantUrl,optional"`
	APIToken      *string                               `pulumi:"apiToken,optional" provider:"secret"`
	WizardPath    *string                               `pulumi:"wizardPath,optional"`
	Tags          map[string]string                     `pulumi:"tags,optional"`
	Registrations map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Namespace       *string                `pulumi:"namespace,optional"`
	EnrollmentMode  *string                `pulumi:"enrollmentMode,optional"`
	ChartRepository *string                `pulumi:"chartRepository,optional"`
	ChartVersion    *string                `pulumi:"chartVersion,optional"`
	ChartValues     map[string]interface{} `pulumi:"chartValues,optional"`
	WorkloadType    *string                `pulumi:"workloadType,optional"`
	HPAEnabled      *bool                  `pulumi:"hpaEnabled,optional"`
	HPAMinReplicas  *int                   `pulumi:"hpaMinReplicas,optional"`
	HPAMaxReplicas  *int                   `pulumi:"hpaMaxReplicas,optional"`
	ImageRepository *string                `pulumi:"imageRepository,optional"`
	ImageTag        *string                `pulumi:"imageTag,optional"`
}

type KubernetesPublisher struct {
	pulumi.ResourceState
	KubernetesPublisherArgs
	PublisherNames   pulumi.StringArrayOutput `pulumi:"publisherNames"`
	HelmReleaseNames pulumi.StringArrayOutput `pulumi:"helmReleaseNames"`
	Publishers       pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*KubernetesPublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "KubernetesPublisher")
}

func NewKubernetesPublisher(ctx *pulumi.Context, name string, args KubernetesPublisherArgs, opts ...pulumi.ResourceOption) (*KubernetesPublisher, error) {
	component := &KubernetesPublisher{KubernetesPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}

	mode := defaultString(args.EnrollmentMode, "token")
	if mode != "token" && mode != "api" {
		return nil, fmt.Errorf("enrollmentMode must be \"token\" or \"api\"")
	}
	workloadType := defaultString(args.WorkloadType, "daemonset")
	if workloadType != "daemonset" && workloadType != "statefulset" {
		return nil, fmt.Errorf("workloadType must be \"daemonset\" or \"statefulset\"")
	}

	publisherNames, err := derivePublisherNames(args.common())
	if err != nil {
		return nil, err
	}
	namespaceName := defaultString(args.Namespace, "netskope")

	namespace, err := k8score.NewNamespace(ctx, name+"-namespace", &k8score.NamespaceArgs{
		Metadata: &k8smeta.ObjectMetaArgs{
			Name: pulumi.StringPtr(namespaceName),
		},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	outputs := pulumi.Map{}
	releaseNames := []string{}

	if mode == "api" {
		if args.TenantURL == nil || args.APIToken == nil {
			return nil, fmt.Errorf("tenantUrl and apiToken are required in api enrollment mode")
		}
		apiSecret, err := k8score.NewSecret(ctx, name+"-api-token", &k8score.SecretArgs{
			Metadata: &k8smeta.ObjectMetaArgs{
				Name:      pulumi.StringPtr("npa-api-token"),
				Namespace: pulumi.StringPtr(namespaceName),
			},
			StringData: pulumi.StringMap{
				"api-token": pulumi.ToSecret(pulumi.String(*args.APIToken)).(pulumi.StringOutput),
			},
			Type: pulumi.StringPtr("Opaque"),
		}, pulumi.Parent(component), pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return nil, err
		}

		releaseName := "npa-publisher"
		release, err := newKubernetesRelease(ctx, component, name, releaseName, namespaceName, args, kubernetesValues(args, pulumi.Map{
			"enrollment": pulumi.Map{
				"mode": pulumi.String("api"),
				"api": pulumi.Map{
					"baseUrl":         pulumi.String(*args.TenantURL),
					"existingSecret":  pulumi.String("npa-api-token"),
					"tokenKey":        pulumi.String("api-token"),
					"cleanupOnDelete": pulumi.Bool(false),
				},
			},
		}), []pulumi.Resource{apiSecret})
		if err != nil {
			return nil, err
		}
		releaseNames = []string{releaseName}
		outputs[releaseName] = pulumi.Map{
			"helmReleaseName": pulumi.String(releaseName),
			"namespace":       pulumi.String(namespaceName),
			"status":          release.Status.Status(),
			"vmId":            pulumi.String(""),
			"privateIp":       pulumi.String(""),
			"publicIp":        pulumi.String(""),
		}.ToMapOutput()
	} else {
		publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
		if err != nil {
			return nil, err
		}
		releaseNames = publisherNames
		for _, publisherName := range publisherNames {
			registration := registrations[publisherName]
			secretName := publisherName + "-registration-token"
			tokenSecret, err := k8score.NewSecret(ctx, name+"-"+publisherName+"-registration-token", &k8score.SecretArgs{
				Metadata: &k8smeta.ObjectMetaArgs{
					Name:      pulumi.StringPtr(secretName),
					Namespace: pulumi.StringPtr(namespaceName),
				},
				StringData: pulumi.StringMap{
					"token": registration.RegistrationToken,
				},
				Type: pulumi.StringPtr("Opaque"),
			}, pulumi.Parent(component), pulumi.DependsOn([]pulumi.Resource{namespace}))
			if err != nil {
				return nil, err
			}

			release, err := newKubernetesRelease(ctx, component, name, publisherName, namespaceName, args, kubernetesValues(args, pulumi.Map{
				"enrollment": pulumi.Map{
					"mode":       pulumi.String("token"),
					"commonName": pulumi.String(publisherName),
				},
				"registrationToken": pulumi.Map{
					"existingSecret":    pulumi.String(secretName),
					"existingSecretKey": pulumi.String("token"),
				},
			}), []pulumi.Resource{tokenSecret})
			if err != nil {
				return nil, err
			}

			outputs[publisherName] = pulumi.Map{
				"publisherId":       registration.PublisherID,
				"registrationToken": registration.RegistrationToken,
				"helmReleaseName":   pulumi.String(publisherName),
				"namespace":         pulumi.String(namespaceName),
				"status":            release.Status.Status(),
				"vmId":              pulumi.String(""),
				"privateIp":         pulumi.String(""),
				"publicIp":          pulumi.String(""),
			}.ToMapOutput()
		}
	}

	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.HelmReleaseNames = toStringArray(releaseNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames":   component.PublisherNames,
		"helmReleaseNames": component.HelmReleaseNames,
		"publishers":       component.Publishers,
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

	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
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
				"guestinfo.userdata":          renderUserDataBase64OutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, preBakedCloudInitOptions()),
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

func (args KubernetesPublisherArgs) common() CommonPublisherArgs {
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

type publisherRegistrationOutput struct {
	PublisherID       pulumi.IntOutput
	RegistrationToken pulumi.StringOutput
	ExistedBefore     pulumi.BoolOutput
}

type NetskopeRegistrationResource struct {
	pulumi.CustomResourceState

	Registrations pulumi.MapOutput `pulumi:"registrations"`
}

func resolvePublisherInputs(
	ctx *pulumi.Context,
	parent pulumi.Resource,
	componentName string,
	args CommonPublisherArgs,
) ([]string, map[string]publisherRegistrationOutput, error) {
	names, err := derivePublisherNames(args)
	if err != nil {
		return nil, nil, err
	}

	if len(args.Registrations) > 0 {
		registrations := make(map[string]publisherRegistrationOutput, len(names))
		for _, name := range names {
			registration, ok := args.Registrations[name]
			if !ok {
				return nil, nil, fmt.Errorf("registrations is missing data for publisher %s", name)
			}
			registrations[name] = publisherRegistrationOutput{
				PublisherID:       pulumi.Int(registration.PublisherID).ToIntOutput(),
				RegistrationToken: pulumi.ToSecret(pulumi.String(registration.RegistrationToken)).(pulumi.StringOutput),
				ExistedBefore:     pulumi.Bool(registration.ExistedBefore).ToBoolOutput(),
			}
		}
		return names, registrations, nil
	}

	if args.TenantURL == nil || *args.TenantURL == "" || args.APIToken == nil || *args.APIToken == "" {
		return nil, nil, fmt.Errorf("tenantUrl and apiToken are required when registrations are not provided")
	}

	registrationResource := &NetskopeRegistrationResource{}
	err = ctx.RegisterResource("netskope-publisher:index:NetskopeRegistration", componentName+"-registration", pulumi.Map{
		"publisherNames": toStringArray(names),
		"tenantUrl":      pulumi.String(*args.TenantURL),
		"apiToken":       pulumi.ToSecret(pulumi.String(*args.APIToken)),
	}, registrationResource, pulumi.Parent(parent))
	if err != nil {
		return nil, nil, err
	}

	registrations := make(map[string]publisherRegistrationOutput, len(names))
	for _, name := range names {
		registrations[name] = registrationOutputFromMap(registrationResource.Registrations, name)
	}
	return names, registrations, nil
}

func registrationOutputFromMap(registrations pulumi.MapOutput, publisherName string) publisherRegistrationOutput {
	publisherID := registrations.ApplyT(func(values map[string]interface{}) int {
		return intFromRegistrationMap(values, publisherName, "publisherId")
	}).(pulumi.IntOutput)
	registrationToken := registrations.ApplyT(func(values map[string]interface{}) string {
		return stringFromRegistrationMap(values, publisherName, "registrationToken")
	}).(pulumi.StringOutput)
	existedBefore := registrations.ApplyT(func(values map[string]interface{}) bool {
		return boolFromRegistrationMap(values, publisherName, "existedBefore")
	}).(pulumi.BoolOutput)

	return publisherRegistrationOutput{
		PublisherID:       publisherID,
		RegistrationToken: pulumi.ToSecret(registrationToken).(pulumi.StringOutput),
		ExistedBefore:     existedBefore,
	}
}

func registrationRecord(values map[string]interface{}, publisherName string) map[string]interface{} {
	value, ok := values[publisherName]
	if !ok || value == nil {
		return map[string]interface{}{}
	}
	switch record := value.(type) {
	case map[string]interface{}:
		return record
	case map[string]string:
		result := make(map[string]interface{}, len(record))
		for key, value := range record {
			result[key] = value
		}
		return result
	case map[string]int:
		result := make(map[string]interface{}, len(record))
		for key, value := range record {
			result[key] = value
		}
		return result
	default:
		return map[string]interface{}{}
	}
}

func intFromRegistrationMap(values map[string]interface{}, publisherName string, field string) int {
	switch value := registrationRecord(values, publisherName)[field].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case string:
		parsed, _ := parsePublisherID(value)
		return parsed
	default:
		return 0
	}
}

func stringFromRegistrationMap(values map[string]interface{}, publisherName string, field string) string {
	value, ok := registrationRecord(values, publisherName)[field]
	if !ok || value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func boolFromRegistrationMap(values map[string]interface{}, publisherName string, field string) bool {
	value, ok := registrationRecord(values, publisherName)[field]
	if !ok || value == nil {
		return false
	}
	switch value := value.(type) {
	case bool:
		return value
	case string:
		return value == "true"
	default:
		return false
	}
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

func publisherOutput(registration publisherRegistrationOutput, vmID pulumi.StringOutput, privateIP pulumi.StringOutput, publicIP pulumi.StringOutput) pulumi.MapOutput {
	return pulumi.Map{
		"publisherId":       registration.PublisherID,
		"registrationToken": registration.RegistrationToken,
		"vmId":              vmID,
		"privateIp":         privateIP,
		"publicIp":          publicIP,
	}.ToMapOutput()
}

const defaultBootstrapURL = "https://s3-us-west-2.amazonaws.com/publisher.netskope.com/latest/generic/bootstrap.sh"

type cloudInitOptions struct {
	Bootstrap                    bool
	BootstrapURL                 string
	Nonat                        bool
	InstallUser                  string
	InstallUserPassword          *string
	InstallUserPasswordIsHash    bool
	InstallUserSSHAuthorizedKeys []string
	DeleteDefaultUser            bool
	GuestNetworkInterface        *GuestNetworkInterface
}

func preBakedCloudInitOptions() cloudInitOptions {
	return cloudInitOptions{Bootstrap: false, Nonat: false, InstallUser: "ubuntu", DeleteDefaultUser: true}
}

func cloudInitOptionsFromAws(args AwsPublisherArgs) cloudInitOptions {
	return cloudInitOptions{
		Bootstrap:                    defaultBool(args.Bootstrap, false),
		BootstrapURL:                 defaultString(args.BootstrapURL, defaultBootstrapURL),
		Nonat:                        defaultBool(args.Nonat, false),
		InstallUser:                  defaultString(args.InstallUser, "ubuntu"),
		InstallUserPassword:          args.InstallUserPassword,
		InstallUserPasswordIsHash:    defaultBool(args.InstallUserPasswordIsHash, false),
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            defaultBool(args.DeleteDefaultUser, true),
		GuestNetworkInterface:        args.GuestNetworkInterface,
	}
}

func cloudInitOptionsFromAzure(args AzurePublisherArgs) cloudInitOptions {
	return cloudInitOptions{
		Bootstrap:                    defaultBool(args.Bootstrap, false),
		BootstrapURL:                 defaultString(args.BootstrapURL, defaultBootstrapURL),
		Nonat:                        defaultBool(args.Nonat, false),
		InstallUser:                  defaultString(args.InstallUser, "ubuntu"),
		InstallUserPassword:          args.InstallUserPassword,
		InstallUserPasswordIsHash:    defaultBool(args.InstallUserPasswordIsHash, false),
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            defaultBool(args.DeleteDefaultUser, true),
		GuestNetworkInterface:        args.GuestNetworkInterface,
	}
}

func cloudInitOptionsFromGcp(args GcpPublisherArgs) cloudInitOptions {
	return cloudInitOptions{
		Bootstrap:                    defaultBool(args.Bootstrap, true),
		BootstrapURL:                 defaultString(args.BootstrapURL, defaultBootstrapURL),
		Nonat:                        defaultBool(args.Nonat, true),
		InstallUser:                  defaultString(args.InstallUser, "ubuntu"),
		InstallUserPassword:          args.InstallUserPassword,
		InstallUserPasswordIsHash:    defaultBool(args.InstallUserPasswordIsHash, false),
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            defaultBool(args.DeleteDefaultUser, true),
		GuestNetworkInterface:        args.GuestNetworkInterface,
	}
}

func renderUserData(publisherName string, registrationToken string, wizardPath *string) string {
	return renderUserDataWithOptions(publisherName, registrationToken, wizardPath, cloudInitOptions{
		Bootstrap:    true,
		BootstrapURL: defaultBootstrapURL,
		Nonat:        true,
	})
}

func renderUserDataWithOptions(publisherName string, registrationToken string, wizardPath *string, options cloudInitOptions) string {
	installUser := defaultString(&options.InstallUser, "ubuntu")
	path := defaultString(wizardPath, "/home/"+installUser+"/npa_publisher_wizard")
	if !options.Bootstrap && !options.Nonat && installUser == "ubuntu" && options.InstallUserPassword == nil && len(options.InstallUserSSHAuthorizedKeys) == 0 && options.GuestNetworkInterface == nil {
		return strings.Join([]string{
			"#cloud-config",
			"hostname: " + publisherName,
			"preserve_hostname: false",
			"runcmd:",
			fmt.Sprintf("  - [ %s, -token, \"%s\" ]", path, escapeDoubleQuoted(registrationToken)),
			"",
		}, "\n")
	}
	bootstrapURL := defaultString(&options.BootstrapURL, defaultBootstrapURL)
	lines := []string{
		"#cloud-config",
		"hostname: " + publisherName,
		"preserve_hostname: false",
		"",
		"system_info:",
		"  default_user:",
		"    name: " + installUser,
		"",
		"users:",
		"  - name: " + installUser,
		"    groups: [sudo]",
		"    sudo: \"ALL=(ALL) NOPASSWD:ALL\"",
		"    shell: /bin/bash",
		fmt.Sprintf("    lock_passwd: %t", options.InstallUserPassword == nil),
	}
	if len(options.InstallUserSSHAuthorizedKeys) > 0 {
		lines = append(lines, "    ssh_authorized_keys:")
		for _, key := range options.InstallUserSSHAuthorizedKeys {
			lines = append(lines, fmt.Sprintf("      - \"%s\"", escapeDoubleQuoted(key)))
		}
	}
	if options.InstallUserPassword != nil {
		lines = append(lines,
			"",
			"chpasswd:",
			"  expire: false",
			"  users:",
			"    - name: "+installUser,
			fmt.Sprintf("      password: \"%s\"", escapeDoubleQuoted(*options.InstallUserPassword)),
			fmt.Sprintf("      type: %s", map[bool]string{true: "hash", false: "text"}[options.InstallUserPasswordIsHash]),
			"ssh_pwauth: true",
		)
	}
	if options.GuestNetworkInterface != nil {
		lines = append(lines, "")
		lines = append(lines, renderNetplan(options.GuestNetworkInterface)...)
	}
	lines = append(lines, "", "runcmd:")
	if options.GuestNetworkInterface != nil {
		lines = append(lines,
			"  - chmod 0600 /etc/netplan/60-cloudinit-override.yaml",
			"  - netplan apply",
		)
	}
	if options.DeleteDefaultUser && installUser != "ubuntu" {
		lines = append(lines,
			"  - pkill -KILL -u ubuntu || true",
			"  - userdel -r ubuntu 2>/dev/null || true",
		)
	}
	if options.Bootstrap || options.Nonat {
		lines = append(lines, "  - chmod 1777 /tmp")
	}
	if options.Nonat {
		lines = append(lines,
			fmt.Sprintf("  - install -d -o %s -g %s -m 0755 /home/%s/resources", installUser, installUser, installUser),
			fmt.Sprintf("  - install -o %s -g %s -m 0644 /dev/null /home/%s/resources/.nonat", installUser, installUser, installUser),
		)
	}
	if options.Bootstrap {
		lines = append(lines, fmt.Sprintf("  - su - %s -c 'curl -fsSL %s | sudo bash'", installUser, bootstrapURL))
	}
	lines = append(lines, fmt.Sprintf("  - su - %s -c 'sudo %s -token \"%s\"'", installUser, path, escapeSingleQuoted(registrationToken)), "")
	return strings.Join(lines, "\n")
}

func renderUserDataBase64WithOptions(publisherName string, registrationToken string, wizardPath *string, options cloudInitOptions) string {
	return base64.StdEncoding.EncodeToString([]byte(renderUserDataWithOptions(publisherName, registrationToken, wizardPath, options)))
}

func renderUserDataOutputWithOptions(publisherName string, registrationToken pulumi.StringOutput, wizardPath *string, options cloudInitOptions) pulumi.StringOutput {
	return registrationToken.ApplyT(func(token string) string {
		return renderUserDataWithOptions(publisherName, token, wizardPath, options)
	}).(pulumi.StringOutput)
}

func renderUserDataBase64OutputWithOptions(publisherName string, registrationToken pulumi.StringOutput, wizardPath *string, options cloudInitOptions) pulumi.StringOutput {
	return registrationToken.ApplyT(func(token string) string {
		return renderUserDataBase64WithOptions(publisherName, token, wizardPath, options)
	}).(pulumi.StringOutput)
}

func renderMetadata(publisherName string) string {
	return "instance-id: " + publisherName + "\nlocal-hostname: " + publisherName + "\n"
}

func newKubernetesRelease(ctx *pulumi.Context, component pulumi.Resource, componentName string, releaseName string, namespaceName string, args KubernetesPublisherArgs, values pulumi.Map, dependsOn []pulumi.Resource) (*k8shelm.Release, error) {
	return k8shelm.NewRelease(ctx, componentName+"-"+releaseName, &k8shelm.ReleaseArgs{
		Name:            pulumi.StringPtr(releaseName),
		Namespace:       pulumi.StringPtr(namespaceName),
		Chart:           pulumi.String("kubernetes-netskope-publisher"),
		Version:         pulumi.StringPtr(defaultString(args.ChartVersion, "~> 1.4")),
		RepositoryOpts:  &k8shelm.RepositoryOptsArgs{Repo: pulumi.StringPtr(defaultString(args.ChartRepository, "oci://ghcr.io/johnneerdael/charts"))},
		CreateNamespace: pulumi.BoolPtr(false),
		Atomic:          pulumi.BoolPtr(true),
		SkipAwait:       pulumi.BoolPtr(false),
		Timeout:         pulumi.IntPtr(300),
		Values:          values,
	}, pulumi.Parent(component), pulumi.DependsOn(dependsOn))
}

func kubernetesValues(args KubernetesPublisherArgs, modeValues pulumi.Map) pulumi.Map {
	workloadType := defaultString(args.WorkloadType, "daemonset")
	values := pulumi.Map{
		"workload": pulumi.Map{
			"type": pulumi.String(workloadType),
		},
		"hpa": pulumi.Map{
			"enabled":     pulumi.Bool(defaultBool(args.HPAEnabled, false) && workloadType == "statefulset"),
			"minReplicas": pulumi.Int(defaultInt(args.HPAMinReplicas, 2)),
			"maxReplicas": pulumi.Int(defaultInt(args.HPAMaxReplicas, 6)),
		},
		"commonLabels": toStringMap(args.Tags),
	}
	image := pulumi.Map{}
	if args.ImageRepository != nil {
		image["repository"] = pulumi.String(*args.ImageRepository)
	}
	if args.ImageTag != nil {
		image["tag"] = pulumi.String(*args.ImageTag)
	}
	if len(image) > 0 {
		values["image"] = image
	}
	for key, value := range modeValues {
		values[key] = value
	}
	for key, value := range args.ChartValues {
		values[key] = toPulumiInput(value)
	}
	return values
}

func toPulumiInput(value interface{}) pulumi.Input {
	switch typed := value.(type) {
	case nil:
		return nil
	case pulumi.Input:
		return typed
	case string:
		return pulumi.String(typed)
	case bool:
		return pulumi.Bool(typed)
	case int:
		return pulumi.Int(typed)
	case int64:
		return pulumi.Int(int(typed))
	case float64:
		return pulumi.Float64(typed)
	case []interface{}:
		result := pulumi.Array{}
		for _, item := range typed {
			result = append(result, toPulumiInput(item))
		}
		return result
	case map[string]interface{}:
		result := pulumi.Map{}
		for key, item := range typed {
			result[key] = toPulumiInput(item)
		}
		return result
	default:
		return pulumi.String(fmt.Sprint(typed))
	}
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
	} else {
		marketplace := args.Marketplace
		if marketplace == nil && defaultBool(args.Bootstrap, false) {
			marketplace = &AzureMarketplaceImage{
				Publisher: "Canonical",
				Offer:     "0001-com-ubuntu-minimal-jammy",
				SKU:       "minimal-22_04-lts-gen2",
			}
		}
		if marketplace != nil {
			profile.ImageReference = &azurecompute.ImageReferenceArgs{
				Publisher: pulumi.StringPtr(marketplace.Publisher),
				Offer:     pulumi.StringPtr(marketplace.Offer),
				Sku:       pulumi.StringPtr(marketplace.SKU),
				Version:   pulumi.StringPtr(defaultString(marketplace.Version, "latest")),
			}
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

func renderNetplan(networkInterface *GuestNetworkInterface) []string {
	lines := []string{
		"write_files:",
		"  - path: /etc/netplan/60-cloudinit-override.yaml",
		"    owner: root:root",
		"    permissions: \"0600\"",
		"    content: |",
		"      network:",
		"        version: 2",
		"        ethernets:",
		"          " + networkInterface.Name + ":",
		fmt.Sprintf("            dhcp4: %t", defaultBool(networkInterface.DHCP4, false)),
	}
	if len(networkInterface.Addresses) > 0 {
		lines = append(lines, "            addresses:")
		for _, address := range networkInterface.Addresses {
			lines = append(lines, "              - "+address)
		}
	}
	if networkInterface.Gateway4 != nil {
		lines = append(lines, "            gateway4: "+*networkInterface.Gateway4)
	}
	if len(networkInterface.Nameservers) > 0 {
		lines = append(lines, "            nameservers:", "              addresses:")
		for _, nameserver := range networkInterface.Nameservers {
			lines = append(lines, "                - "+nameserver)
		}
	}
	if networkInterface.MTU != nil {
		lines = append(lines, fmt.Sprintf("            mtu: %d", *networkInterface.MTU))
	}
	return lines
}

func escapeDoubleQuoted(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `"`, `\"`)
}

func escapeSingleQuoted(value string) string {
	return strings.ReplaceAll(value, `'`, `'"'"'`)
}
