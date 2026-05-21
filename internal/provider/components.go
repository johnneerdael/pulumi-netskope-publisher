package provider

import (
	"encoding/base64"
	"fmt"
	"reflect"
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
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

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
	if err := validateProviderCatalogArgs("AwsPublisher", args); err != nil {
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

		outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), instance.PrivateIp, instance.PublicIp, args.PlacementLabels)
	}

	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type AzurePublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

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
	if err := validateProviderCatalogArgs("AzurePublisher", args); err != nil {
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
		outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), publicIPOutput, args.PlacementLabels)
	}

	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type GcpPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

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
	if err := validateProviderCatalogArgs("GcpPublisher", args); err != nil {
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

		outputs[publisherName] = publisherOutput(registration, instance.InstanceId, pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput(), args.PlacementLabels)
	}

	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type KubernetesPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

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
		if args.TenantURL == nil || *args.TenantURL == "" || !hasManagedRegistrationAuth(args.common()) {
			return nil, fmt.Errorf("tenantUrl and a bearer token or oauth2 credentials are required in api enrollment mode")
		}
		apiAuthMode := defaultString(args.AuthMode, "token")
		apiSecret, apiAuthValues, err := kubernetesAPIAuth(ctx, component, name, namespaceName, args, namespace)
		if err != nil {
			return nil, err
		}
		apiValues := pulumi.Map{
			"baseUrl":         pulumi.String(*args.TenantURL),
			"authMode":        pulumi.String(apiAuthMode),
			"cleanupOnDelete": pulumi.Bool(false),
		}
		for key, value := range apiAuthValues {
			apiValues[key] = value
		}

		releaseName := "npa-publisher"
		release, err := newKubernetesRelease(ctx, component, name, releaseName, namespaceName, args, kubernetesValues(args, pulumi.Map{
			"enrollment": pulumi.Map{
				"mode": pulumi.String("api"),
				"api":  apiValues,
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

func kubernetesAPIAuth(ctx *pulumi.Context, component pulumi.Resource, name string, namespaceName string, args KubernetesPublisherArgs, namespace pulumi.Resource) (*k8score.Secret, pulumi.Map, error) {
	if defaultString(args.AuthMode, "token") == "oauth2" {
		secret, err := k8score.NewSecret(ctx, name+"-api-oauth", &k8score.SecretArgs{
			Metadata: &k8smeta.ObjectMetaArgs{
				Name:      pulumi.StringPtr("npa-api-oauth"),
				Namespace: pulumi.StringPtr(namespaceName),
			},
			StringData: pulumi.StringMap{
				"client-id":     pulumi.String(args.OAuth2.ClientID),
				"client-secret": pulumi.ToSecret(pulumi.String(args.OAuth2.ClientSecret)).(pulumi.StringOutput),
			},
			Type: pulumi.StringPtr("Opaque"),
		}, pulumi.Parent(component), pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return nil, nil, err
		}
		oauth2 := pulumi.Map{
			"tokenUrl":        pulumi.String(args.OAuth2.TokenURL),
			"existingSecret":  pulumi.String("npa-api-oauth"),
			"clientIdKey":     pulumi.String("client-id"),
			"clientSecretKey": pulumi.String("client-secret"),
			"scope":           pulumi.String(defaultString(args.OAuth2.Scope, "")),
		}
		return secret, pulumi.Map{"oauth2": oauth2}, nil
	}

	token := stringValue(args.BearerToken)
	if token == "" {
		token = stringValue(args.APIToken)
	}
	secret, err := k8score.NewSecret(ctx, name+"-api-token", &k8score.SecretArgs{
		Metadata: &k8smeta.ObjectMetaArgs{
			Name:      pulumi.StringPtr("npa-api-token"),
			Namespace: pulumi.StringPtr(namespaceName),
		},
		StringData: pulumi.StringMap{
			"api-token": pulumi.ToSecret(pulumi.String(token)).(pulumi.StringOutput),
		},
		Type: pulumi.StringPtr("Opaque"),
	}, pulumi.Parent(component), pulumi.DependsOn([]pulumi.Resource{namespace}))
	if err != nil {
		return nil, nil, err
	}
	return secret, pulumi.Map{
		"existingSecret": pulumi.String("npa-api-token"),
		"tokenKey":       pulumi.String("api-token"),
	}, nil
}

type VspherePublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

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
	if err := validateProviderCatalogArgs("VspherePublisher", args); err != nil {
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

		outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), vm.DefaultIpAddress, pulumi.String("").ToStringOutput(), args.PlacementLabels)
	}

	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{
		"publisherNames": component.PublisherNames,
		"publishers":     component.Publishers,
	})
}

type HypervPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

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
	if err := validateProviderCatalogArgs("HypervPublisher", args); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("Hyper-V support requires a stable Pulumi Hyper-V provider SDK")
}

type EsxiPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	DiskStore                    string                 `pulumi:"diskStore"`
	VirtualNetwork               string                 `pulumi:"virtualNetwork"`
	OS                           *string                `pulumi:"os,optional"`
	Memory                       *int                   `pulumi:"memory,optional"`
	NumVCpus                     *int                   `pulumi:"numVCpus,optional"`
	DiskSize                     *int                   `pulumi:"diskSize,optional"`
}

type EsxiPublisher struct {
	pulumi.ResourceState
	EsxiPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*EsxiPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "EsxiPublisher") }

type HcloudPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	ServerType                   *string                `pulumi:"serverType,optional"`
	Image                        *string                `pulumi:"image,optional"`
	Location                     *string                `pulumi:"location,optional"`
	Datacenter                   *string                `pulumi:"datacenter,optional"`
	SSHKeys                      []string               `pulumi:"sshKeys,optional"`
	FirewallIDs                  []int                  `pulumi:"firewallIds,optional"`
	NetworkID                    *int                   `pulumi:"networkId,optional"`
	AssignPublicIP               *bool                  `pulumi:"assignPublicIp,optional"`
}

type HcloudPublisher struct {
	pulumi.ResourceState
	HcloudPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*HcloudPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "HcloudPublisher") }

type NutanixPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	ClusterUUID                  string                 `pulumi:"clusterUuid"`
	ImageUUID                    *string                `pulumi:"imageUuid,optional"`
	SubnetUUID                   *string                `pulumi:"subnetUuid,optional"`
	NumVCpus                     *int                   `pulumi:"numVCpus,optional"`
	NumCoresPerVcpu              *int                   `pulumi:"numCoresPerVcpu,optional"`
	MemorySizeMib                *int                   `pulumi:"memorySizeMib,optional"`
}

type NutanixPublisher struct {
	pulumi.ResourceState
	NutanixPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*NutanixPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "NutanixPublisher") }

type OpenstackPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	ImageName                    string                 `pulumi:"imageName"`
	FlavorName                   string                 `pulumi:"flavorName"`
	NetworkName                  string                 `pulumi:"networkName"`
	KeyPair                      *string                `pulumi:"keyPair,optional"`
	SecurityGroups               []string               `pulumi:"securityGroups,optional"`
	AvailabilityZone             *string                `pulumi:"availabilityZone,optional"`
	AssignFloatingIP             *bool                  `pulumi:"assignFloatingIp,optional"`
	FloatingIPPool               *string                `pulumi:"floatingIpPool,optional"`
}

type OpenstackPublisher struct {
	pulumi.ResourceState
	OpenstackPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*OpenstackPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "OpenstackPublisher") }

type OvhPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	ServiceName                  string                 `pulumi:"serviceName"`
	Region                       string                 `pulumi:"region"`
	ImageID                      string                 `pulumi:"imageId"`
	FlavorID                     string                 `pulumi:"flavorId"`
	SSHKeyName                   *string                `pulumi:"sshKeyName,optional"`
	NetworkID                    *string                `pulumi:"networkId,optional"`
}

type OvhPublisher struct {
	pulumi.ResourceState
	OvhPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*OvhPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "OvhPublisher") }

type ScalewayPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	Type                         *string                `pulumi:"type,optional"`
	Image                        *string                `pulumi:"image,optional"`
	Zone                         *string                `pulumi:"zone,optional"`
	SecurityGroupID              *string                `pulumi:"securityGroupId,optional"`
	EnableDynamicIP              *bool                  `pulumi:"enableDynamicIp,optional"`
}

type ScalewayPublisher struct {
	pulumi.ResourceState
	ScalewayPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*ScalewayPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "ScalewayPublisher") }

type OciPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	CompartmentID                string                 `pulumi:"compartmentId"`
	AvailabilityDomain           string                 `pulumi:"availabilityDomain"`
	Shape                        *string                `pulumi:"shape,optional"`
	SubnetID                     string                 `pulumi:"subnetId"`
	ImageID                      string                 `pulumi:"imageId"`
	SSHPublicKey                 *string                `pulumi:"sshPublicKey,optional"`
	AssignPublicIP               *bool                  `pulumi:"assignPublicIp,optional"`
}

type OciPublisher struct {
	pulumi.ResourceState
	OciPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*OciPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "OciPublisher") }

type AlicloudPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	InstanceType                 *string                `pulumi:"instanceType,optional"`
	ImageID                      string                 `pulumi:"imageId"`
	VswitchID                    string                 `pulumi:"vswitchId"`
	SecurityGroupIDs             []string               `pulumi:"securityGroupIds"`
	KeyName                      *string                `pulumi:"keyName,optional"`
	AllocatePublicIP             *bool                  `pulumi:"allocatePublicIp,optional"`
}

type AlicloudPublisher struct {
	pulumi.ResourceState
	AlicloudPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*AlicloudPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "AlicloudPublisher") }

type ProxmoxvePublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	NodeName                     string                 `pulumi:"nodeName"`
	DatastoreID                  string                 `pulumi:"datastoreId"`
	TemplateVMID                 int                    `pulumi:"templateVmId"`
	CloneNodeName                *string                `pulumi:"cloneNodeName,optional"`
	VMID                         *int                   `pulumi:"vmId,optional"`
	PoolID                       *string                `pulumi:"poolId,optional"`
	CPUCores                     *int                   `pulumi:"cpuCores,optional"`
	Memory                       *int                   `pulumi:"memory,optional"`
	DiskSize                     *int                   `pulumi:"diskSize,optional"`
	NetworkBridge                *string                `pulumi:"networkBridge,optional"`
	NetworkModel                 *string                `pulumi:"networkModel,optional"`
	VlanID                       *int                   `pulumi:"vlanId,optional"`
	Started                      *bool                  `pulumi:"started,optional"`
	OnBoot                       *bool                  `pulumi:"onBoot,optional"`
	FullClone                    *bool                  `pulumi:"fullClone,optional"`
	IPAddress                    *string                `pulumi:"ipAddress,optional"`
	Gateway                      *string                `pulumi:"gateway,optional"`
	Nameservers                  []string               `pulumi:"nameservers,optional"`
}

type ProxmoxvePublisher struct {
	pulumi.ResourceState
	ProxmoxvePublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*ProxmoxvePublisher) Annotate(a infer.Annotator) { a.SetToken("index", "ProxmoxvePublisher") }

type DigitaloceanPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	Region                       string                 `pulumi:"region"`
	Size                         *string                `pulumi:"size,optional"`
	Image                        *string                `pulumi:"image,optional"`
	SSHKeys                      []string               `pulumi:"sshKeys,optional"`
	VpcUUID                      *string                `pulumi:"vpcUuid,optional"`
	Monitoring                   *bool                  `pulumi:"monitoring,optional"`
	Ipv6                         *bool                  `pulumi:"ipv6,optional"`
}

type DigitaloceanPublisher struct {
	pulumi.ResourceState
	DigitaloceanPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*DigitaloceanPublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "DigitaloceanPublisher")
}

type VultrPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	Region                       string                 `pulumi:"region"`
	Plan                         string                 `pulumi:"plan"`
	OSID                         *int                   `pulumi:"osId,optional"`
	ImageID                      *string                `pulumi:"imageId,optional"`
	SSHKeyIDs                    []string               `pulumi:"sshKeyIds,optional"`
	Vpc2IDs                      []string               `pulumi:"vpc2Ids,optional"`
	EnableIpv6                   *bool                  `pulumi:"enableIpv6,optional"`
	FirewallGroupID              *string                `pulumi:"firewallGroupId,optional"`
}

type VultrPublisher struct {
	pulumi.ResourceState
	VultrPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*VultrPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "VultrPublisher") }

type ExoscalePublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                    `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                  `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                    `pulumi:"nonat,optional"`
	InstallUser                  *string                  `pulumi:"installUser,optional"`
	InstallUserPassword          *string                  `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                    `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string                 `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                    `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface   `pulumi:"guestNetworkInterface,optional"`
	Zone                         string                   `pulumi:"zone"`
	Type                         string                   `pulumi:"type"`
	TemplateID                   string                   `pulumi:"templateId"`
	DiskSize                     int                      `pulumi:"diskSize"`
	SSHKeys                      []string                 `pulumi:"sshKeys,optional"`
	SecurityGroupIDs             []string                 `pulumi:"securityGroupIds,optional"`
	NetworkInterfaces            []map[string]interface{} `pulumi:"networkInterfaces,optional"`
}

type ExoscalePublisher struct {
	pulumi.ResourceState
	ExoscalePublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*ExoscalePublisher) Annotate(a infer.Annotator) { a.SetToken("index", "ExoscalePublisher") }

type UpcloudPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                    `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                  `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                    `pulumi:"nonat,optional"`
	InstallUser                  *string                  `pulumi:"installUser,optional"`
	InstallUserPassword          *string                  `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                    `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string                 `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                    `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface   `pulumi:"guestNetworkInterface,optional"`
	Zone                         string                   `pulumi:"zone"`
	Hostname                     *string                  `pulumi:"hostname,optional"`
	Plan                         *string                  `pulumi:"plan,optional"`
	Template                     *string                  `pulumi:"template,optional"`
	NetworkInterfaces            []map[string]interface{} `pulumi:"networkInterfaces,optional"`
}

type UpcloudPublisher struct {
	pulumi.ResourceState
	UpcloudPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*UpcloudPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "UpcloudPublisher") }

type StackitPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                    `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                  `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                    `pulumi:"nonat,optional"`
	InstallUser                  *string                  `pulumi:"installUser,optional"`
	InstallUserPassword          *string                  `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                    `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string                 `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                    `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface   `pulumi:"guestNetworkInterface,optional"`
	ProjectID                    string                   `pulumi:"projectId"`
	MachineType                  string                   `pulumi:"machineType"`
	ImageID                      string                   `pulumi:"imageId"`
	AvailabilityZone             *string                  `pulumi:"availabilityZone,optional"`
	KeypairName                  *string                  `pulumi:"keypairName,optional"`
	NetworkInterfaces            []map[string]interface{} `pulumi:"networkInterfaces,optional"`
}

type StackitPublisher struct {
	pulumi.ResourceState
	StackitPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*StackitPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "StackitPublisher") }

type EquinixPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	ProjectID                    string                 `pulumi:"projectId"`
	Metro                        string                 `pulumi:"metro"`
	Plan                         string                 `pulumi:"plan"`
	OperatingSystem              *string                `pulumi:"operatingSystem,optional"`
	BillingCycle                 *string                `pulumi:"billingCycle,optional"`
	ProjectSSHKeyIDs             []string               `pulumi:"projectSshKeyIds,optional"`
	UserSSHKeyIDs                []string               `pulumi:"userSshKeyIds,optional"`
}

type EquinixPublisher struct {
	pulumi.ResourceState
	EquinixPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*EquinixPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "EquinixPublisher") }

type OutscalePublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	ImageID                      string                 `pulumi:"imageId"`
	VMType                       *string                `pulumi:"vmType,optional"`
	SubnetID                     *string                `pulumi:"subnetId,optional"`
	KeypairName                  *string                `pulumi:"keypairName,optional"`
	SecurityGroupIDs             []string               `pulumi:"securityGroupIds,optional"`
	PlacementSubregionName       *string                `pulumi:"placementSubregionName,optional"`
}

type OutscalePublisher struct {
	pulumi.ResourceState
	OutscalePublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*OutscalePublisher) Annotate(a infer.Annotator) { a.SetToken("index", "OutscalePublisher") }

type OpentelekomcloudPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                    `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                  `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                    `pulumi:"nonat,optional"`
	InstallUser                  *string                  `pulumi:"installUser,optional"`
	InstallUserPassword          *string                  `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                    `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string                 `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                    `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface   `pulumi:"guestNetworkInterface,optional"`
	ImageName                    *string                  `pulumi:"imageName,optional"`
	ImageID                      *string                  `pulumi:"imageId,optional"`
	FlavorName                   *string                  `pulumi:"flavorName,optional"`
	FlavorID                     *string                  `pulumi:"flavorId,optional"`
	Networks                     []map[string]interface{} `pulumi:"networks"`
	KeyPair                      *string                  `pulumi:"keyPair,optional"`
	AvailabilityZone             *string                  `pulumi:"availabilityZone,optional"`
	SecurityGroups               []string                 `pulumi:"securityGroups,optional"`
}

type OpentelekomcloudPublisher struct {
	pulumi.ResourceState
	OpentelekomcloudPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*OpentelekomcloudPublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "OpentelekomcloudPublisher")
}

type TencentcloudPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	AvailabilityZone             string                 `pulumi:"availabilityZone"`
	ImageID                      string                 `pulumi:"imageId"`
	InstanceType                 *string                `pulumi:"instanceType,optional"`
	SubnetID                     *string                `pulumi:"subnetId,optional"`
	VpcID                        *string                `pulumi:"vpcId,optional"`
	KeyName                      *string                `pulumi:"keyName,optional"`
	SecurityGroups               []string               `pulumi:"securityGroups,optional"`
	SystemDiskType               *string                `pulumi:"systemDiskType,optional"`
	SystemDiskSize               *int                   `pulumi:"systemDiskSize,optional"`
}

type TencentcloudPublisher struct {
	pulumi.ResourceState
	TencentcloudPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*TencentcloudPublisher) Annotate(a infer.Annotator) {
	a.SetToken("index", "TencentcloudPublisher")
}

type YandexPublisherArgs struct {
	NamePrefix      *string                               `pulumi:"namePrefix,optional"`
	Names           []string                              `pulumi:"names,optional"`
	Replicas        *int                                  `pulumi:"replicas,optional"`
	PlacementLabels []string                              `pulumi:"placementLabels,optional"`
	TenantURL       *string                               `pulumi:"tenantUrl,optional"`
	APIToken        *string                               `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string                               `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string                               `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args                   `pulumi:"oauth2,optional"`
	WizardPath      *string                               `pulumi:"wizardPath,optional"`
	Tags            map[string]string                     `pulumi:"tags,optional"`
	Registrations   map[string]PublisherRegistrationInput `pulumi:"registrations,optional"`

	Bootstrap                    *bool                  `pulumi:"bootstrap,optional"`
	BootstrapURL                 *string                `pulumi:"bootstrapUrl,optional"`
	Nonat                        *bool                  `pulumi:"nonat,optional"`
	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
	Zone                         *string                `pulumi:"zone,optional"`
	PlatformID                   *string                `pulumi:"platformId,optional"`
	ImageID                      string                 `pulumi:"imageId"`
	SubnetID                     string                 `pulumi:"subnetId"`
	Cores                        *int                   `pulumi:"cores,optional"`
	Memory                       *int                   `pulumi:"memory,optional"`
	CoreFraction                 *int                   `pulumi:"coreFraction,optional"`
	Nat                          *bool                  `pulumi:"nat,optional"`
	SSHKeys                      []string               `pulumi:"sshKeys,optional"`
}

type YandexPublisher struct {
	pulumi.ResourceState
	YandexPublisherArgs
	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

func (*YandexPublisher) Annotate(a infer.Annotator) { a.SetToken("index", "YandexPublisher") }

type rawVMResource struct {
	pulumi.CustomResourceState

	AccessIPV4         pulumi.StringOutput      `pulumi:"accessIpV4"`
	AccessPrivateIpv4  pulumi.StringOutput      `pulumi:"accessPrivateIpv4"`
	AccessPublicIpv4   pulumi.StringOutput      `pulumi:"accessPublicIpv4"`
	Address            pulumi.StringOutput      `pulumi:"address"`
	Addresses          pulumi.ArrayOutput       `pulumi:"addresses"`
	InternalIP         pulumi.StringOutput      `pulumi:"internalIp"`
	IPAddresses        pulumi.StringArrayOutput `pulumi:"ipAddresses"`
	Ipv4Address        pulumi.StringOutput      `pulumi:"ipv4Address"`
	Ipv4AddressPrivate pulumi.StringOutput      `pulumi:"ipv4AddressPrivate"`
	Ipv4Addresses      pulumi.ArrayOutput       `pulumi:"ipv4Addresses"`
	MainIP             pulumi.StringOutput      `pulumi:"mainIp"`
	Networks           pulumi.ArrayOutput       `pulumi:"networks"`
	NicListStatuses    pulumi.AnyOutput         `pulumi:"nicListStatuses"`
	PrimaryIPAddress   pulumi.StringOutput      `pulumi:"primaryIpAddress"`
	PrivateIP          pulumi.StringOutput      `pulumi:"privateIp"`
	PrivateIPs         pulumi.ArrayOutput       `pulumi:"privateIps"`
	PublicIP           pulumi.StringOutput      `pulumi:"publicIp"`
	PublicIPAddress    pulumi.StringOutput      `pulumi:"publicIpAddress"`
	PublicIPs          pulumi.ArrayOutput       `pulumi:"publicIps"`
}

func NewEsxiPublisher(ctx *pulumi.Context, name string, args EsxiPublisherArgs, opts ...pulumi.ResourceOption) (*EsxiPublisher, error) {
	component := &EsxiPublisher{EsxiPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("EsxiPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}
	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		vm := &rawVMResource{}
		err := ctx.RegisterResource("esxi-native:index:VirtualMachine", name+"-"+publisherName, pulumi.Map{
			"name":         pulumi.String(publisherName),
			"diskStore":    pulumi.String(args.DiskStore),
			"os":           pulumi.String(defaultString(args.OS, "ubuntu-64")),
			"memSize":      pulumi.Int(defaultInt(args.Memory, 4096)),
			"numVCpus":     pulumi.Int(defaultInt(args.NumVCpus, 2)),
			"bootDiskSize": pulumi.Int(defaultInt(args.DiskSize, 64)),
			"networkInterfaces": pulumi.Array{pulumi.Map{
				"virtualNetwork": pulumi.String(args.VirtualNetwork),
			}},
			"info": pulumi.Array{pulumi.Map{
				"key":   pulumi.String("guestinfo.userdata"),
				"value": renderUserDataBase64OutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromCommon(args.common(), true)),
			}, pulumi.Map{
				"key":   pulumi.String("guestinfo.userdata.encoding"),
				"value": pulumi.String("base64"),
			}},
			"power": pulumi.String("on"),
		}, vm, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}
		outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput(), args.PlacementLabels)
	}
	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewHcloudPublisher(ctx *pulumi.Context, name string, args HcloudPublisherArgs, opts ...pulumi.ResourceOption) (*HcloudPublisher, error) {
	component := &HcloudPublisher{HcloudPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("HcloudPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}
	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		userData := renderUserDataOutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromCommon(args.common(), true))
		inputs := pulumi.Map{
			"name":        pulumi.String(publisherName),
			"serverType":  pulumi.String(defaultString(args.ServerType, "cx22")),
			"image":       pulumi.String(defaultString(args.Image, "ubuntu-22.04")),
			"location":    stringPtrInput(args.Location),
			"datacenter":  stringPtrInput(args.Datacenter),
			"sshKeys":     toStringArray(args.SSHKeys),
			"firewallIds": toIntArray(args.FirewallIDs),
			"publicNets": pulumi.Array{pulumi.Map{
				"ipv4Enabled": pulumi.Bool(defaultBool(args.AssignPublicIP, true)),
				"ipv6Enabled": pulumi.Bool(defaultBool(args.AssignPublicIP, true)),
			}},
			"userData": plainUserData(userData),
			"labels":   toStringMap(args.Tags),
		}
		if args.NetworkID != nil {
			inputs["networks"] = pulumi.Array{pulumi.Map{"networkId": pulumi.Int(*args.NetworkID)}}
		}
		server := &rawVMResource{}
		if err := ctx.RegisterResource("hcloud:index/server:Server", name+"-"+publisherName, inputs, server, pulumi.Parent(component)); err != nil {
			return nil, err
		}
		outputs[publisherName] = publisherOutput(registration, server.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), server.Ipv4Address, args.PlacementLabels)
	}
	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewNutanixPublisher(ctx *pulumi.Context, name string, args NutanixPublisherArgs, opts ...pulumi.ResourceOption) (*NutanixPublisher, error) {
	component := &NutanixPublisher{NutanixPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("NutanixPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}
	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		userData := renderUserDataOutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromCommon(args.common(), true))
		inputs := pulumi.Map{
			"name":                                pulumi.String(publisherName),
			"clusterUuid":                         pulumi.String(args.ClusterUUID),
			"numSockets":                          pulumi.Int(defaultInt(args.NumVCpus, 2)),
			"numVcpusPerSocket":                   pulumi.Int(defaultInt(args.NumCoresPerVcpu, 1)),
			"memorySizeMib":                       pulumi.Int(defaultInt(args.MemorySizeMib, 4096)),
			"guestCustomizationCloudInitUserData": base64UserData(userData),
		}
		if args.ImageUUID != nil {
			inputs["diskLists"] = pulumi.Array{pulumi.Map{"dataSourceReference": pulumi.Map{"kind": pulumi.String("image"), "uuid": pulumi.String(*args.ImageUUID)}}}
		}
		if args.SubnetUUID != nil {
			inputs["nicLists"] = pulumi.Array{pulumi.Map{"subnetUuid": pulumi.String(*args.SubnetUUID), "nicType": pulumi.String("NORMAL_NIC"), "model": pulumi.String("VIRTIO")}}
		}
		vm := &rawVMResource{}
		if err := ctx.RegisterResource("nutanix:index/virtualMachine:VirtualMachine", name+"-"+publisherName, inputs, vm, pulumi.Parent(component)); err != nil {
			return nil, err
		}
		outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), firstNutanixPrivateIP(vm.NicListStatuses), pulumi.String("").ToStringOutput(), args.PlacementLabels)
	}
	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewOpenstackPublisher(ctx *pulumi.Context, name string, args OpenstackPublisherArgs, opts ...pulumi.ResourceOption) (*OpenstackPublisher, error) {
	component := &OpenstackPublisher{OpenstackPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("OpenstackPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}
	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		userData := renderUserDataOutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromCommon(args.common(), true))
		instance := &rawVMResource{}
		err := ctx.RegisterResource("openstack:compute/instance:Instance", name+"-"+publisherName, pulumi.Map{
			"name":             pulumi.String(publisherName),
			"imageName":        pulumi.String(args.ImageName),
			"flavorName":       pulumi.String(args.FlavorName),
			"networks":         pulumi.Array{pulumi.Map{"name": pulumi.String(args.NetworkName)}},
			"keyPair":          stringPtrInput(args.KeyPair),
			"securityGroups":   toStringArray(args.SecurityGroups),
			"availabilityZone": stringPtrInput(args.AvailabilityZone),
			"userData":         plainUserData(userData),
		}, instance, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}
		publicIP := instance.AccessIPV4
		if defaultBool(args.AssignFloatingIP, false) {
			floatingIP := &rawVMResource{}
			if err := ctx.RegisterResource("openstack:networking/floatingIp:FloatingIp", name+"-"+publisherName+"-fip", pulumi.Map{
				"pool": stringPtrInput(args.FloatingIPPool),
			}, floatingIP, pulumi.Parent(component)); err != nil {
				return nil, err
			}
			association := &rawVMResource{}
			if err := ctx.RegisterResource("openstack:networking/floatingIpAssociate:FloatingIpAssociate", name+"-"+publisherName+"-fip-association", pulumi.Map{
				"floatingIp": floatingIP.Address,
				"portId":     firstOpenstackNetworkPort(instance.Networks),
			}, association, pulumi.Parent(component)); err != nil {
				return nil, err
			}
			publicIP = floatingIP.Address
		}
		outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), publicIP, args.PlacementLabels)
	}
	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func firstOpenstackNetworkPort(networks pulumi.ArrayOutput) pulumi.StringOutput {
	return networks.ApplyT(func(values []interface{}) string {
		if len(values) == 0 {
			return ""
		}
		network, ok := values[0].(map[string]interface{})
		if !ok {
			return ""
		}
		port, ok := network["port"]
		if !ok || port == nil {
			return ""
		}
		return fmt.Sprint(port)
	}).(pulumi.StringOutput)
}

func NewOvhPublisher(ctx *pulumi.Context, name string, args OvhPublisherArgs, opts ...pulumi.ResourceOption) (*OvhPublisher, error) {
	component := &OvhPublisher{OvhPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("OvhPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}
	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		userData := renderUserDataOutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromCommon(args.common(), true))
		network := pulumi.Map{"public": pulumi.Bool(true)}
		if args.NetworkID != nil {
			network["private"] = pulumi.Map{"network": pulumi.Map{"id": pulumi.String(*args.NetworkID)}}
		}
		inputs := pulumi.Map{
			"serviceName":   pulumi.String(args.ServiceName),
			"name":          pulumi.String(publisherName),
			"region":        pulumi.String(args.Region),
			"billingPeriod": pulumi.String("hourly"),
			"bootFrom":      pulumi.Map{"imageId": pulumi.String(args.ImageID)},
			"flavor":        pulumi.Map{"flavorId": pulumi.String(args.FlavorID)},
			"network":       network,
			"userData":      plainUserData(userData),
		}
		if args.SSHKeyName != nil {
			inputs["sshKey"] = pulumi.Map{"name": pulumi.String(*args.SSHKeyName)}
		}
		instance := &rawVMResource{}
		if err := ctx.RegisterResource("ovh:CloudProject/instance:Instance", name+"-"+publisherName, inputs, instance, pulumi.Parent(component)); err != nil {
			return nil, err
		}
		outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), firstMapFieldOutput(instance.Addresses, "ip"), args.PlacementLabels)
	}
	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewScalewayPublisher(ctx *pulumi.Context, name string, args ScalewayPublisherArgs, opts ...pulumi.ResourceOption) (*ScalewayPublisher, error) {
	component := &ScalewayPublisher{ScalewayPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("ScalewayPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}
	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		userData := renderUserDataOutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromCommon(args.common(), true))
		userDataPlacement := scalewayUserData(userData)
		server := &rawVMResource{}
		err := ctx.RegisterResource("scaleway:instance/server:Server", name+"-"+publisherName, pulumi.Map{
			"name":            pulumi.String(publisherName),
			"type":            pulumi.String(defaultString(args.Type, "DEV1-M")),
			"image":           pulumi.String(defaultString(args.Image, "ubuntu_jammy")),
			"zone":            stringPtrInput(args.Zone),
			"securityGroupId": stringPtrInput(args.SecurityGroupID),
			"enableDynamicIp": pulumi.Bool(defaultBool(args.EnableDynamicIP, true)),
			"cloudInit":       userDataPlacement["cloudInit"],
			"userData":        userDataPlacement["userData"],
			"tags":            stringMapToTagArray(args.Tags),
		}, server, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}
		outputs[publisherName] = publisherOutput(registration, server.ID().ToStringOutput(), firstMapFieldOutput(server.PrivateIPs, "address"), firstMapFieldOutput(server.PublicIPs, "address"), args.PlacementLabels)
	}
	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewOciPublisher(ctx *pulumi.Context, name string, args OciPublisherArgs, opts ...pulumi.ResourceOption) (*OciPublisher, error) {
	component := &OciPublisher{OciPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("OciPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}
	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		userData := renderUserDataOutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromCommon(args.common(), true))
		metadata := base64MetadataUserData(userData, "userData")
		if args.SSHPublicKey != nil {
			metadata["ssh_authorized_keys"] = pulumi.String(*args.SSHPublicKey)
		}
		instance := &rawVMResource{}
		err := ctx.RegisterResource("oci:Core/instance:Instance", name+"-"+publisherName, pulumi.Map{
			"displayName":        pulumi.String(publisherName),
			"compartmentId":      pulumi.String(args.CompartmentID),
			"availabilityDomain": pulumi.String(args.AvailabilityDomain),
			"shape":              pulumi.String(defaultString(args.Shape, "VM.Standard.E4.Flex")),
			"createVnicDetails": pulumi.Map{
				"subnetId":       pulumi.String(args.SubnetID),
				"assignPublicIp": pulumi.String(fmt.Sprint(defaultBool(args.AssignPublicIP, false))),
				"displayName":    pulumi.String(publisherName + "-vnic"),
			},
			"sourceDetails": pulumi.Map{"sourceType": pulumi.String("image"), "sourceId": pulumi.String(args.ImageID)},
			"metadata":      metadata,
			"freeformTags":  toStringMap(args.Tags),
		}, instance, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}
		outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), instance.PrivateIP, instance.PublicIP, args.PlacementLabels)
	}
	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewAlicloudPublisher(ctx *pulumi.Context, name string, args AlicloudPublisherArgs, opts ...pulumi.ResourceOption) (*AlicloudPublisher, error) {
	component := &AlicloudPublisher{AlicloudPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("AlicloudPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}
	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		userData := renderUserDataOutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromCommon(args.common(), true))
		instance := &rawVMResource{}
		err := ctx.RegisterResource("alicloud:ecs/instance:Instance", name+"-"+publisherName, pulumi.Map{
			"instanceName":            pulumi.String(publisherName),
			"instanceType":            pulumi.String(defaultString(args.InstanceType, "ecs.t6-c1m2.large")),
			"imageId":                 pulumi.String(args.ImageID),
			"vswitchId":               pulumi.String(args.VswitchID),
			"securityGroups":          toStringArray(args.SecurityGroupIDs),
			"keyName":                 stringPtrInput(args.KeyName),
			"internetMaxBandwidthOut": pulumi.Int(publicBandwidth(args.AllocatePublicIP)),
			"userData":                base64UserData(userData),
			"tags":                    toStringMap(args.Tags),
		}, instance, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}
		outputs[publisherName] = publisherOutput(registration, instance.ID().ToStringOutput(), instance.PrimaryIPAddress, instance.PublicIP, args.PlacementLabels)
	}
	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewProxmoxvePublisher(ctx *pulumi.Context, name string, args ProxmoxvePublisherArgs, opts ...pulumi.ResourceOption) (*ProxmoxvePublisher, error) {
	component := &ProxmoxvePublisher{ProxmoxvePublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("ProxmoxvePublisher", args); err != nil {
		return nil, err
	}
	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, name, args.common())
	if err != nil {
		return nil, err
	}
	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		userData := renderUserDataOutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromCommon(args.common(), true))
		userDataFile := &rawVMResource{}
		err := ctx.RegisterResource("proxmoxve:index/fileLegacy:FileLegacy", name+"-"+publisherName+"-user-data", pulumi.Map{
			"contentType": pulumi.String("snippets"),
			"datastoreId": pulumi.String(args.DatastoreID),
			"nodeName":    pulumi.String(args.NodeName),
			"sourceRaw": pulumi.Map{
				"data":     plainUserData(userData),
				"fileName": pulumi.String(publisherName + "-user-data.yaml"),
			},
		}, userDataFile, pulumi.Parent(component))
		if err != nil {
			return nil, err
		}

		vmInputs := pulumi.Map{
			"name":     pulumi.String(publisherName),
			"nodeName": pulumi.String(args.NodeName),
			"clone": pulumi.Map{
				"vmId":        pulumi.Int(args.TemplateVMID),
				"nodeName":    stringPtrInput(args.CloneNodeName),
				"datastoreId": pulumi.String(args.DatastoreID),
				"full":        pulumi.Bool(defaultBool(args.FullClone, true)),
			},
			"agent": pulumi.Map{
				"enabled": pulumi.Bool(true),
			},
			"cpu": pulumi.Map{
				"cores": pulumi.Int(defaultInt(args.CPUCores, 2)),
			},
			"memory": pulumi.Map{
				"dedicated": pulumi.Int(defaultInt(args.Memory, 4096)),
			},
			"networkDevices": pulumi.Array{pulumi.Map{
				"bridge": pulumi.String(defaultString(args.NetworkBridge, "vmbr0")),
				"model":  pulumi.String(defaultString(args.NetworkModel, "virtio")),
				"vlanId": intPtrInput(args.VlanID),
			}},
			"initialization": pulumi.Map{
				"datastoreId":    pulumi.String(args.DatastoreID),
				"userDataFileId": userDataFile.ID().ToStringOutput(),
				"ipConfigs": pulumi.Array{pulumi.Map{
					"ipv4": pulumi.Map{
						"address": pulumi.String(defaultString(args.IPAddress, "dhcp")),
						"gateway": stringPtrInput(args.Gateway),
					},
				}},
				"dns": pulumi.Map{
					"servers": toStringArray(args.Nameservers),
				},
			},
			"onBoot":          pulumi.Bool(defaultBool(args.OnBoot, true)),
			"operatingSystem": pulumi.Map{"type": pulumi.String("l26")},
			"poolId":          stringPtrInput(args.PoolID),
			"started":         pulumi.Bool(defaultBool(args.Started, true)),
			"tags":            stringMapToTagArray(args.Tags),
		}
		if args.VMID != nil {
			vmInputs["vmId"] = pulumi.Int(*args.VMID)
		}
		if args.DiskSize != nil {
			vmInputs["disks"] = pulumi.Array{pulumi.Map{
				"datastoreId": pulumi.String(args.DatastoreID),
				"interface":   pulumi.String("scsi0"),
				"size":        pulumi.Int(*args.DiskSize),
			}}
		}

		vm := &rawVMResource{}
		if err := ctx.RegisterResource("proxmoxve:index/vmLegacy:VmLegacy", name+"-"+publisherName, vmInputs, vm, pulumi.Parent(component)); err != nil {
			return nil, err
		}
		outputs[publisherName] = publisherOutput(registration, vm.ID().ToStringOutput(), firstNestedString(vm.Ipv4Addresses), pulumi.String("").ToStringOutput(), args.PlacementLabels)
	}
	component.PublisherNames = toStringArray(publisherNames).ToStringArrayOutput()
	component.Publishers = pulumi.ToSecret(outputs).(pulumi.MapOutput)
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

type rawBootstrapBuildResult struct {
	VMID      pulumi.StringOutput
	PrivateIP pulumi.StringOutput
	PublicIP  pulumi.StringOutput
}

func createRawBootstrapPublishers(
	ctx *pulumi.Context,
	component pulumi.Resource,
	componentName string,
	args CommonPublisherArgs,
	build func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error),
) (pulumi.StringArrayOutput, pulumi.MapOutput, error) {
	publisherNames, registrations, err := resolvePublisherInputs(ctx, component, componentName, args)
	if err != nil {
		return pulumi.StringArrayOutput{}, pulumi.MapOutput{}, err
	}
	outputs := pulumi.Map{}
	for _, publisherName := range publisherNames {
		registration := registrations[publisherName]
		userData := renderUserDataOutputWithOptions(publisherName, registration.RegistrationToken, args.WizardPath, cloudInitOptionsFromCommon(args, true))
		result, err := build(publisherName, userData)
		if err != nil {
			return pulumi.StringArrayOutput{}, pulumi.MapOutput{}, err
		}
		outputs[publisherName] = publisherOutput(registration, result.VMID, result.PrivateIP, result.PublicIP, args.PlacementLabels)
	}
	return toStringArray(publisherNames).ToStringArrayOutput(), pulumi.ToSecret(outputs).(pulumi.MapOutput), nil
}

func NewDigitaloceanPublisher(ctx *pulumi.Context, name string, args DigitaloceanPublisherArgs, opts ...pulumi.ResourceOption) (*DigitaloceanPublisher, error) {
	component := &DigitaloceanPublisher{DigitaloceanPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("DigitaloceanPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, publishers, err := createRawBootstrapPublishers(ctx, component, name, args.common(), func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error) {
		droplet := &rawVMResource{}
		err := ctx.RegisterResource("digitalocean:index/droplet:Droplet", name+"-"+publisherName, pulumi.Map{
			"name":       pulumi.String(publisherName),
			"region":     pulumi.String(args.Region),
			"size":       pulumi.String(defaultString(args.Size, "s-2vcpu-4gb")),
			"image":      pulumi.String(defaultString(args.Image, "ubuntu-22-04-x64")),
			"sshKeys":    toStringArray(args.SSHKeys),
			"vpcUuid":    stringPtrInput(args.VpcUUID),
			"monitoring": boolPtrInput(args.Monitoring),
			"ipv6":       boolPtrInput(args.Ipv6),
			"userData":   plainUserData(userData),
			"tags":       stringMapToColonTagArray(args.Tags),
		}, droplet, pulumi.Parent(component))
		return rawBootstrapBuildResult{droplet.ID().ToStringOutput(), droplet.Ipv4AddressPrivate, droplet.Ipv4Address}, err
	})
	if err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNames
	component.Publishers = publishers
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewVultrPublisher(ctx *pulumi.Context, name string, args VultrPublisherArgs, opts ...pulumi.ResourceOption) (*VultrPublisher, error) {
	component := &VultrPublisher{VultrPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("VultrPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, publishers, err := createRawBootstrapPublishers(ctx, component, name, args.common(), func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error) {
		instance := &rawVMResource{}
		err := ctx.RegisterResource("vultr:index/instance:Instance", name+"-"+publisherName, pulumi.Map{
			"label":           pulumi.String(publisherName),
			"hostname":        pulumi.String(publisherName),
			"region":          pulumi.String(args.Region),
			"plan":            pulumi.String(args.Plan),
			"osId":            intPtrInput(args.OSID),
			"imageId":         stringPtrInput(args.ImageID),
			"sshKeyIds":       toStringArray(args.SSHKeyIDs),
			"vpc2Ids":         toStringArray(args.Vpc2IDs),
			"enableIpv6":      boolPtrInput(args.EnableIpv6),
			"firewallGroupId": stringPtrInput(args.FirewallGroupID),
			"userData":        plainUserData(userData),
			"tags":            stringMapToColonTagArray(args.Tags),
		}, instance, pulumi.Parent(component))
		return rawBootstrapBuildResult{instance.ID().ToStringOutput(), instance.InternalIP, instance.MainIP}, err
	})
	if err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNames
	component.Publishers = publishers
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewExoscalePublisher(ctx *pulumi.Context, name string, args ExoscalePublisherArgs, opts ...pulumi.ResourceOption) (*ExoscalePublisher, error) {
	component := &ExoscalePublisher{ExoscalePublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("ExoscalePublisher", args); err != nil {
		return nil, err
	}
	publisherNames, publishers, err := createRawBootstrapPublishers(ctx, component, name, args.common(), func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error) {
		instance := &rawVMResource{}
		err := ctx.RegisterResource("exoscale:index/computeInstance:ComputeInstance", name+"-"+publisherName, pulumi.Map{
			"name":              pulumi.String(publisherName),
			"zone":              pulumi.String(args.Zone),
			"type":              pulumi.String(args.Type),
			"templateId":        pulumi.String(args.TemplateID),
			"diskSize":          pulumi.Int(args.DiskSize),
			"sshKeys":           toStringArray(args.SSHKeys),
			"securityGroupIds":  toStringArray(args.SecurityGroupIDs),
			"networkInterfaces": toMapArray(args.NetworkInterfaces),
			"userData":          plainUserData(userData),
			"labels":            toStringMap(args.Tags),
		}, instance, pulumi.Parent(component))
		return rawBootstrapBuildResult{instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), instance.PublicIPAddress}, err
	})
	if err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNames
	component.Publishers = publishers
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewUpcloudPublisher(ctx *pulumi.Context, name string, args UpcloudPublisherArgs, opts ...pulumi.ResourceOption) (*UpcloudPublisher, error) {
	component := &UpcloudPublisher{UpcloudPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("UpcloudPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, publishers, err := createRawBootstrapPublishers(ctx, component, name, args.common(), func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error) {
		server := &rawVMResource{}
		err := ctx.RegisterResource("upcloud:index/server:Server", name+"-"+publisherName, pulumi.Map{
			"hostname":          pulumi.String(defaultString(args.Hostname, publisherName)),
			"title":             pulumi.String(publisherName),
			"zone":              pulumi.String(args.Zone),
			"plan":              pulumi.String(defaultString(args.Plan, "2xCPU-4GB")),
			"template":          pulumi.String(defaultString(args.Template, "01000000-0000-4000-8000-000030220200")),
			"networkInterfaces": toMapArray(args.NetworkInterfaces),
			"metadata":          pulumi.Bool(true),
			"userData":          plainUserData(userData),
			"labels":            toStringMap(args.Tags),
		}, server, pulumi.Parent(component))
		return rawBootstrapBuildResult{server.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput()}, err
	})
	if err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNames
	component.Publishers = publishers
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewStackitPublisher(ctx *pulumi.Context, name string, args StackitPublisherArgs, opts ...pulumi.ResourceOption) (*StackitPublisher, error) {
	component := &StackitPublisher{StackitPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("StackitPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, publishers, err := createRawBootstrapPublishers(ctx, component, name, args.common(), func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error) {
		server := &rawVMResource{}
		err := ctx.RegisterResource("stackit:index/server:Server", name+"-"+publisherName, pulumi.Map{
			"name":              pulumi.String(publisherName),
			"projectId":         pulumi.String(args.ProjectID),
			"machineType":       pulumi.String(args.MachineType),
			"imageId":           pulumi.String(args.ImageID),
			"availabilityZone":  stringPtrInput(args.AvailabilityZone),
			"keypairName":       stringPtrInput(args.KeypairName),
			"networkInterfaces": toMapArray(args.NetworkInterfaces),
			"userData":          plainUserData(userData),
			"labels":            toStringMap(args.Tags),
		}, server, pulumi.Parent(component))
		return rawBootstrapBuildResult{server.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput()}, err
	})
	if err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNames
	component.Publishers = publishers
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewEquinixPublisher(ctx *pulumi.Context, name string, args EquinixPublisherArgs, opts ...pulumi.ResourceOption) (*EquinixPublisher, error) {
	component := &EquinixPublisher{EquinixPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("EquinixPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, publishers, err := createRawBootstrapPublishers(ctx, component, name, args.common(), func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error) {
		device := &rawVMResource{}
		err := ctx.RegisterResource("equinix:metal/device:Device", name+"-"+publisherName, pulumi.Map{
			"hostname":         pulumi.String(publisherName),
			"projectId":        pulumi.String(args.ProjectID),
			"metro":            pulumi.String(args.Metro),
			"plan":             pulumi.String(args.Plan),
			"operatingSystem":  pulumi.String(defaultString(args.OperatingSystem, "ubuntu_22_04")),
			"billingCycle":     pulumi.String(defaultString(args.BillingCycle, "hourly")),
			"projectSshKeyIds": toStringArray(args.ProjectSSHKeyIDs),
			"userSshKeyIds":    toStringArray(args.UserSSHKeyIDs),
			"userData":         plainUserData(userData),
			"tags":             stringMapToColonTagArray(args.Tags),
		}, device, pulumi.Parent(component))
		return rawBootstrapBuildResult{device.ID().ToStringOutput(), device.AccessPrivateIpv4, device.AccessPublicIpv4}, err
	})
	if err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNames
	component.Publishers = publishers
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewOutscalePublisher(ctx *pulumi.Context, name string, args OutscalePublisherArgs, opts ...pulumi.ResourceOption) (*OutscalePublisher, error) {
	component := &OutscalePublisher{OutscalePublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("OutscalePublisher", args); err != nil {
		return nil, err
	}
	publisherNames, publishers, err := createRawBootstrapPublishers(ctx, component, name, args.common(), func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error) {
		vm := &rawVMResource{}
		err := ctx.RegisterResource("outscale:index/vm:Vm", name+"-"+publisherName, pulumi.Map{
			"imageId":                pulumi.String(args.ImageID),
			"vmType":                 pulumi.String(defaultString(args.VMType, "tinav5.c2r4p1")),
			"subnetId":               stringPtrInput(args.SubnetID),
			"keypairName":            stringPtrInput(args.KeypairName),
			"securityGroupIds":       toStringArray(args.SecurityGroupIDs),
			"placementSubregionName": stringPtrInput(args.PlacementSubregionName),
			"userData":               plainUserData(userData),
		}, vm, pulumi.Parent(component))
		return rawBootstrapBuildResult{vm.ID().ToStringOutput(), vm.PrivateIP, vm.PublicIP}, err
	})
	if err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNames
	component.Publishers = publishers
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewOpentelekomcloudPublisher(ctx *pulumi.Context, name string, args OpentelekomcloudPublisherArgs, opts ...pulumi.ResourceOption) (*OpentelekomcloudPublisher, error) {
	component := &OpentelekomcloudPublisher{OpentelekomcloudPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("OpentelekomcloudPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, publishers, err := createRawBootstrapPublishers(ctx, component, name, args.common(), func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error) {
		instance := &rawVMResource{}
		err := ctx.RegisterResource("opentelekomcloud:index/computeInstanceV2:ComputeInstanceV2", name+"-"+publisherName, pulumi.Map{
			"name":             pulumi.String(publisherName),
			"imageName":        pulumi.String(defaultString(args.ImageName, "Ubuntu 22.04")),
			"imageId":          stringPtrInput(args.ImageID),
			"flavorName":       pulumi.String(defaultString(args.FlavorName, "s3.medium.2")),
			"flavorId":         stringPtrInput(args.FlavorID),
			"networks":         toMapArray(args.Networks),
			"keyPair":          stringPtrInput(args.KeyPair),
			"availabilityZone": stringPtrInput(args.AvailabilityZone),
			"securityGroups":   toStringArray(args.SecurityGroups),
			"userData":         plainUserData(userData),
		}, instance, pulumi.Parent(component))
		return rawBootstrapBuildResult{instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), instance.AccessIPV4}, err
	})
	if err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNames
	component.Publishers = publishers
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewTencentcloudPublisher(ctx *pulumi.Context, name string, args TencentcloudPublisherArgs, opts ...pulumi.ResourceOption) (*TencentcloudPublisher, error) {
	component := &TencentcloudPublisher{TencentcloudPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("TencentcloudPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, publishers, err := createRawBootstrapPublishers(ctx, component, name, args.common(), func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error) {
		instance := &rawVMResource{}
		err := ctx.RegisterResource("tencentcloud:index/instance:Instance", name+"-"+publisherName, pulumi.Map{
			"instanceName":            pulumi.String(publisherName),
			"hostname":                pulumi.String(publisherName),
			"availabilityZone":        pulumi.String(args.AvailabilityZone),
			"imageId":                 pulumi.String(args.ImageID),
			"instanceType":            pulumi.String(defaultString(args.InstanceType, "S5.MEDIUM4")),
			"subnetId":                stringPtrInput(args.SubnetID),
			"vpcId":                   stringPtrInput(args.VpcID),
			"keyName":                 stringPtrInput(args.KeyName),
			"securityGroups":          toStringArray(args.SecurityGroups),
			"systemDiskType":          stringPtrInput(args.SystemDiskType),
			"systemDiskSize":          intPtrInput(args.SystemDiskSize),
			"userDataRaw":             plainUserData(userData),
			"userDataReplaceOnChange": pulumi.Bool(true),
			"tags":                    toStringMap(args.Tags),
		}, instance, pulumi.Parent(component))
		return rawBootstrapBuildResult{instance.ID().ToStringOutput(), instance.PrivateIP, instance.PublicIP}, err
	})
	if err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNames
	component.Publishers = publishers
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func NewYandexPublisher(ctx *pulumi.Context, name string, args YandexPublisherArgs, opts ...pulumi.ResourceOption) (*YandexPublisher, error) {
	component := &YandexPublisher{YandexPublisherArgs: args}
	if err := ctx.RegisterComponentResource(p.GetTypeToken(ctx), name, component, opts...); err != nil {
		return nil, err
	}
	if err := validateProviderCatalogArgs("YandexPublisher", args); err != nil {
		return nil, err
	}
	publisherNames, publishers, err := createRawBootstrapPublishers(ctx, component, name, args.common(), func(publisherName string, userData pulumi.StringOutput) (rawBootstrapBuildResult, error) {
		metadata := metadataUserData(userData, "user-data")
		if len(args.SSHKeys) > 0 {
			metadata["ssh-keys"] = pulumi.String(strings.Join(args.SSHKeys, "\n"))
		}
		instance := &rawVMResource{}
		err := ctx.RegisterResource("yandex:index/computeInstance:ComputeInstance", name+"-"+publisherName, pulumi.Map{
			"name":       pulumi.String(publisherName),
			"hostname":   pulumi.String(publisherName),
			"zone":       stringPtrInput(args.Zone),
			"platformId": pulumi.String(defaultString(args.PlatformID, "standard-v3")),
			"bootDisk": pulumi.Map{"initializeParams": pulumi.Map{
				"imageId": pulumi.String(args.ImageID),
			}},
			"resources": pulumi.Map{
				"cores":        pulumi.Int(defaultInt(args.Cores, 2)),
				"memory":       pulumi.Int(defaultInt(args.Memory, 4)),
				"coreFraction": intPtrInput(args.CoreFraction),
			},
			"networkInterfaces": pulumi.Array{pulumi.Map{
				"subnetId": pulumi.String(args.SubnetID),
				"nat":      pulumi.Bool(defaultBool(args.Nat, false)),
			}},
			"metadata": metadata,
			"labels":   toStringMap(args.Tags),
		}, instance, pulumi.Parent(component))
		return rawBootstrapBuildResult{instance.ID().ToStringOutput(), pulumi.String("").ToStringOutput(), pulumi.String("").ToStringOutput()}, err
	})
	if err != nil {
		return nil, err
	}
	component.PublisherNames = publisherNames
	component.Publishers = publishers
	return component, ctx.RegisterResourceOutputs(component, pulumi.Map{"publisherNames": component.PublisherNames, "publishers": component.Publishers})
}

func (args AwsPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
	}
}

func (args AzurePublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
	}
}

func (args GcpPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
	}
}

func (args KubernetesPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
	}
}

func (args VspherePublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
	}
}

func (args EsxiPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
		Bootstrap: args.Bootstrap, BootstrapURL: args.BootstrapURL, Nonat: args.Nonat,
		InstallUser: args.InstallUser, InstallUserPassword: args.InstallUserPassword,
		InstallUserPasswordIsHash:    args.InstallUserPasswordIsHash,
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            args.DeleteDefaultUser, GuestNetworkInterface: args.GuestNetworkInterface,
	}
}

func (args HcloudPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
		Bootstrap: args.Bootstrap, BootstrapURL: args.BootstrapURL, Nonat: args.Nonat,
		InstallUser: args.InstallUser, InstallUserPassword: args.InstallUserPassword,
		InstallUserPasswordIsHash:    args.InstallUserPasswordIsHash,
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            args.DeleteDefaultUser, GuestNetworkInterface: args.GuestNetworkInterface,
	}
}

func (args NutanixPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
		Bootstrap: args.Bootstrap, BootstrapURL: args.BootstrapURL, Nonat: args.Nonat,
		InstallUser: args.InstallUser, InstallUserPassword: args.InstallUserPassword,
		InstallUserPasswordIsHash:    args.InstallUserPasswordIsHash,
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            args.DeleteDefaultUser, GuestNetworkInterface: args.GuestNetworkInterface,
	}
}

func (args OpenstackPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
		Bootstrap: args.Bootstrap, BootstrapURL: args.BootstrapURL, Nonat: args.Nonat,
		InstallUser: args.InstallUser, InstallUserPassword: args.InstallUserPassword,
		InstallUserPasswordIsHash:    args.InstallUserPasswordIsHash,
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            args.DeleteDefaultUser, GuestNetworkInterface: args.GuestNetworkInterface,
	}
}

func (args OvhPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
		Bootstrap: args.Bootstrap, BootstrapURL: args.BootstrapURL, Nonat: args.Nonat,
		InstallUser: args.InstallUser, InstallUserPassword: args.InstallUserPassword,
		InstallUserPasswordIsHash:    args.InstallUserPasswordIsHash,
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            args.DeleteDefaultUser, GuestNetworkInterface: args.GuestNetworkInterface,
	}
}

func (args ScalewayPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
		Bootstrap: args.Bootstrap, BootstrapURL: args.BootstrapURL, Nonat: args.Nonat,
		InstallUser: args.InstallUser, InstallUserPassword: args.InstallUserPassword,
		InstallUserPasswordIsHash:    args.InstallUserPasswordIsHash,
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            args.DeleteDefaultUser, GuestNetworkInterface: args.GuestNetworkInterface,
	}
}

func (args OciPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
		Bootstrap: args.Bootstrap, BootstrapURL: args.BootstrapURL, Nonat: args.Nonat,
		InstallUser: args.InstallUser, InstallUserPassword: args.InstallUserPassword,
		InstallUserPasswordIsHash:    args.InstallUserPasswordIsHash,
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            args.DeleteDefaultUser, GuestNetworkInterface: args.GuestNetworkInterface,
	}
}

func (args AlicloudPublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
		Bootstrap: args.Bootstrap, BootstrapURL: args.BootstrapURL, Nonat: args.Nonat,
		InstallUser: args.InstallUser, InstallUserPassword: args.InstallUserPassword,
		InstallUserPasswordIsHash:    args.InstallUserPasswordIsHash,
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            args.DeleteDefaultUser, GuestNetworkInterface: args.GuestNetworkInterface,
	}
}

func (args ProxmoxvePublisherArgs) common() CommonPublisherArgs {
	return CommonPublisherArgs{
		NamePrefix: args.NamePrefix, Names: args.Names, Replicas: args.Replicas,
		PlacementLabels: args.PlacementLabels,
		TenantURL:       args.TenantURL, APIToken: args.APIToken, BearerToken: args.BearerToken,
		AuthMode: args.AuthMode, OAuth2: args.OAuth2, WizardPath: args.WizardPath,
		Tags: args.Tags, Registrations: args.Registrations,
		Bootstrap: args.Bootstrap, BootstrapURL: args.BootstrapURL, Nonat: args.Nonat,
		InstallUser: args.InstallUser, InstallUserPassword: args.InstallUserPassword,
		InstallUserPasswordIsHash:    args.InstallUserPasswordIsHash,
		InstallUserSSHAuthorizedKeys: args.InstallUserSSHAuthorizedKeys,
		DeleteDefaultUser:            args.DeleteDefaultUser, GuestNetworkInterface: args.GuestNetworkInterface,
	}
}

func (args DigitaloceanPublisherArgs) common() CommonPublisherArgs {
	return commonFromExpandedArgs(args)
}
func (args VultrPublisherArgs) common() CommonPublisherArgs    { return commonFromExpandedArgs(args) }
func (args ExoscalePublisherArgs) common() CommonPublisherArgs { return commonFromExpandedArgs(args) }
func (args UpcloudPublisherArgs) common() CommonPublisherArgs  { return commonFromExpandedArgs(args) }
func (args StackitPublisherArgs) common() CommonPublisherArgs  { return commonFromExpandedArgs(args) }
func (args EquinixPublisherArgs) common() CommonPublisherArgs  { return commonFromExpandedArgs(args) }
func (args OutscalePublisherArgs) common() CommonPublisherArgs { return commonFromExpandedArgs(args) }
func (args OpentelekomcloudPublisherArgs) common() CommonPublisherArgs {
	return commonFromExpandedArgs(args)
}
func (args TencentcloudPublisherArgs) common() CommonPublisherArgs {
	return commonFromExpandedArgs(args)
}
func (args YandexPublisherArgs) common() CommonPublisherArgs { return commonFromExpandedArgs(args) }

func commonFromExpandedArgs(args interface{}) CommonPublisherArgs {
	value := reflect.ValueOf(args)
	return CommonPublisherArgs{
		NamePrefix: fieldValue[*string](value, "NamePrefix"), Names: fieldValue[[]string](value, "Names"), Replicas: fieldValue[*int](value, "Replicas"),
		PlacementLabels: fieldValue[[]string](value, "PlacementLabels"),
		TenantURL:       fieldValue[*string](value, "TenantURL"), APIToken: fieldValue[*string](value, "APIToken"), BearerToken: fieldValue[*string](value, "BearerToken"),
		AuthMode: fieldValue[*string](value, "AuthMode"), OAuth2: fieldValue[*NetskopeOAuth2Args](value, "OAuth2"), WizardPath: fieldValue[*string](value, "WizardPath"),
		Tags: fieldValue[map[string]string](value, "Tags"), Registrations: fieldValue[map[string]PublisherRegistrationInput](value, "Registrations"),
		Bootstrap: fieldValue[*bool](value, "Bootstrap"), BootstrapURL: fieldValue[*string](value, "BootstrapURL"), Nonat: fieldValue[*bool](value, "Nonat"),
		InstallUser: fieldValue[*string](value, "InstallUser"), InstallUserPassword: fieldValue[*string](value, "InstallUserPassword"),
		InstallUserPasswordIsHash:    fieldValue[*bool](value, "InstallUserPasswordIsHash"),
		InstallUserSSHAuthorizedKeys: fieldValue[[]string](value, "InstallUserSSHAuthorizedKeys"),
		DeleteDefaultUser:            fieldValue[*bool](value, "DeleteDefaultUser"),
		GuestNetworkInterface:        fieldValue[*GuestNetworkInterface](value, "GuestNetworkInterface"),
	}
}

func fieldValue[T any](value reflect.Value, name string) T {
	field := value.FieldByName(name)
	if !field.IsValid() || field.IsZero() {
		var zero T
		return zero
	}
	return field.Interface().(T)
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

	if args.TenantURL == nil || *args.TenantURL == "" || !hasManagedRegistrationAuth(args) {
		return nil, nil, fmt.Errorf("tenantUrl and a bearer token or oauth2 credentials are required when registrations are not provided")
	}

	registrationInputs := registrationAuthInputs(args)
	registrationInputs["publisherNames"] = toStringArray(names)
	registrationInputs["tenantUrl"] = pulumi.String(*args.TenantURL)

	registrationResource := &NetskopeRegistrationResource{}
	err = ctx.RegisterResource("netskope-publisher:index:NetskopeRegistration", componentName+"-registration", registrationInputs, registrationResource, pulumi.Parent(parent))
	if err != nil {
		return nil, nil, err
	}

	registrations := make(map[string]publisherRegistrationOutput, len(names))
	for _, name := range names {
		registrations[name] = registrationOutputFromMap(registrationResource.Registrations, name)
	}
	return names, registrations, nil
}

func hasManagedRegistrationAuth(args CommonPublisherArgs) bool {
	authMode := defaultString(args.AuthMode, "token")
	if authMode == "oauth2" {
		return args.OAuth2 != nil && args.OAuth2.TokenURL != "" && args.OAuth2.ClientID != "" && args.OAuth2.ClientSecret != ""
	}
	return stringValue(args.BearerToken) != "" || stringValue(args.APIToken) != ""
}

func registrationAuthInputs(args CommonPublisherArgs) pulumi.Map {
	inputs := pulumi.Map{}
	if args.AuthMode != nil {
		inputs["authMode"] = pulumi.String(*args.AuthMode)
	}
	if args.BearerToken != nil {
		inputs["bearerToken"] = pulumi.ToSecret(pulumi.String(*args.BearerToken))
	}
	if args.APIToken != nil {
		inputs["apiToken"] = pulumi.ToSecret(pulumi.String(*args.APIToken))
	}
	if args.OAuth2 != nil {
		oauth2 := pulumi.Map{
			"tokenUrl":     pulumi.String(args.OAuth2.TokenURL),
			"clientId":     pulumi.String(args.OAuth2.ClientID),
			"clientSecret": pulumi.ToSecret(pulumi.String(args.OAuth2.ClientSecret)),
		}
		if args.OAuth2.Scope != nil {
			oauth2["scope"] = pulumi.String(*args.OAuth2.Scope)
		}
		inputs["oauth2"] = oauth2
	}
	return inputs
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

func publisherOutput(registration publisherRegistrationOutput, vmID pulumi.StringOutput, privateIP pulumi.StringOutput, publicIP pulumi.StringOutput, placementLabels []string) pulumi.MapOutput {
	return pulumi.Map{
		"publisherId":       registration.PublisherID,
		"registrationToken": registration.RegistrationToken,
		"vmId":              vmID,
		"privateIp":         privateIP,
		"publicIp":          publicIP,
		"placementLabels":   toStringArray(placementLabels),
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

func cloudInitOptionsFromCommon(args CommonPublisherArgs, forceBootstrap bool) cloudInitOptions {
	bootstrap := defaultBool(args.Bootstrap, false)
	if forceBootstrap {
		bootstrap = true
	}
	return cloudInitOptions{
		Bootstrap:                    bootstrap,
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

func intPtrInput(value *int) pulumi.IntPtrInput {
	if value == nil {
		return nil
	}
	return pulumi.IntPtr(*value)
}

func boolPtrInput(value *bool) pulumi.BoolPtrInput {
	if value == nil {
		return nil
	}
	return pulumi.BoolPtr(*value)
}

func decodeBase64String(value string) string {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return value
	}
	return string(decoded)
}

func publicBandwidth(value *bool) int {
	if defaultBool(value, false) {
		return 10
	}
	return 0
}

func toStringArray(values []string) pulumi.StringArray {
	result := make(pulumi.StringArray, len(values))
	for i, value := range values {
		result[i] = pulumi.String(value)
	}
	return result
}

func toIntArray(values []int) pulumi.IntArray {
	result := make(pulumi.IntArray, len(values))
	for i, value := range values {
		result[i] = pulumi.Int(value)
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

func stringMapToTagArray(values map[string]string) pulumi.StringArray {
	result := pulumi.StringArray{}
	for key, value := range values {
		result = append(result, pulumi.String(key+"="+value))
	}
	return result
}

func stringMapToColonTagArray(values map[string]string) pulumi.StringArray {
	result := pulumi.StringArray{}
	for key, value := range values {
		result = append(result, pulumi.String(key+":"+value))
	}
	return result
}

func toMapArray(values []map[string]interface{}) pulumi.Array {
	result := pulumi.Array{}
	for _, value := range values {
		result = append(result, toPulumiMap(value))
	}
	return result
}

func toPulumiMap(values map[string]interface{}) pulumi.Map {
	result := pulumi.Map{}
	for key, value := range values {
		switch typed := value.(type) {
		case string:
			result[key] = pulumi.String(typed)
		case int:
			result[key] = pulumi.Int(typed)
		case float64:
			result[key] = pulumi.Float64(typed)
		case bool:
			result[key] = pulumi.Bool(typed)
		case map[string]interface{}:
			result[key] = toPulumiMap(typed)
		default:
			if value != nil {
				result[key] = pulumi.Any(value)
			}
		}
	}
	return result
}

func firstNestedString(values pulumi.ArrayOutput) pulumi.StringOutput {
	return values.ApplyT(func(values []interface{}) string {
		if len(values) == 0 {
			return ""
		}
		switch first := values[0].(type) {
		case []interface{}:
			if len(first) == 0 {
				return ""
			}
			return fmt.Sprint(first[0])
		case []string:
			if len(first) == 0 {
				return ""
			}
			return first[0]
		default:
			return fmt.Sprint(first)
		}
	}).(pulumi.StringOutput)
}

func firstStringOutput(values pulumi.StringArrayOutput) pulumi.StringOutput {
	return values.ApplyT(func(items []string) string {
		if len(items) == 0 {
			return ""
		}
		return items[0]
	}).(pulumi.StringOutput)
}

func firstMapFieldOutput(values pulumi.ArrayOutput, field string) pulumi.StringOutput {
	return values.ApplyT(func(items []interface{}) string {
		if len(items) == 0 {
			return ""
		}
		item, ok := items[0].(map[string]interface{})
		if !ok {
			return ""
		}
		value, ok := item[field]
		if !ok || value == nil {
			return ""
		}
		return fmt.Sprint(value)
	}).(pulumi.StringOutput)
}

func firstNutanixPrivateIP(values pulumi.AnyOutput) pulumi.StringOutput {
	return values.ApplyT(func(value interface{}) string {
		items, ok := value.([]interface{})
		if !ok {
			return ""
		}
		if len(items) == 0 {
			return ""
		}
		status, ok := items[0].(map[string]interface{})
		if !ok {
			return ""
		}
		endpoints, ok := status["ipEndpointLists"].([]interface{})
		if !ok || len(endpoints) == 0 {
			return ""
		}
		endpoint, ok := endpoints[0].(map[string]interface{})
		if !ok {
			return ""
		}
		ip, ok := endpoint["ip"]
		if !ok || ip == nil {
			return ""
		}
		return fmt.Sprint(ip)
	}).(pulumi.StringOutput)
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
