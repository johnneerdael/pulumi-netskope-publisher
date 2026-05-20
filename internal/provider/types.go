package provider

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

type CommonPublisherArgs struct {
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
	Bootstrap       *bool                                 `pulumi:"bootstrap,optional"`
	BootstrapURL    *string                               `pulumi:"bootstrapUrl,optional"`
	Nonat           *bool                                 `pulumi:"nonat,optional"`

	InstallUser                  *string                `pulumi:"installUser,optional"`
	InstallUserPassword          *string                `pulumi:"installUserPassword,optional" provider:"secret"`
	InstallUserPasswordIsHash    *bool                  `pulumi:"installUserPasswordIsHash,optional"`
	InstallUserSSHAuthorizedKeys []string               `pulumi:"installUserSshAuthorizedKeys,optional"`
	DeleteDefaultUser            *bool                  `pulumi:"deleteDefaultUser,optional"`
	GuestNetworkInterface        *GuestNetworkInterface `pulumi:"guestNetworkInterface,optional"`
}

type NetskopeOAuth2Args struct {
	TokenURL     string  `pulumi:"tokenUrl"`
	ClientID     string  `pulumi:"clientId"`
	ClientSecret string  `pulumi:"clientSecret" provider:"secret"`
	Scope        *string `pulumi:"scope,optional"`
}

type PublisherRegistrationInput struct {
	PublisherID       int    `pulumi:"publisherId"`
	RegistrationToken string `pulumi:"registrationToken" provider:"secret"`
	ExistedBefore     bool   `pulumi:"existedBefore,optional"`
}

type PublisherOutput struct {
	PublisherID       pulumi.IntOutput         `pulumi:"publisherId"`
	RegistrationToken pulumi.StringOutput      `pulumi:"registrationToken" provider:"secret"`
	VMID              pulumi.StringOutput      `pulumi:"vmId"`
	PrivateIP         pulumi.StringOutput      `pulumi:"privateIp"`
	PublicIP          pulumi.StringOutput      `pulumi:"publicIp"`
	PlacementLabels   pulumi.StringArrayOutput `pulumi:"placementLabels"`
}

type PublisherComponent struct {
	pulumi.ResourceState

	PublisherNames pulumi.StringArrayOutput `pulumi:"publisherNames"`
	Publishers     pulumi.MapOutput         `pulumi:"publishers" provider:"secret"`
}

type MetadataOptions struct {
	HTTPEndpoint *string `pulumi:"httpEndpoint,optional"`
	HTTPTokens   *string `pulumi:"httpTokens,optional"`
}

type GuestNetworkInterface struct {
	Name        string   `pulumi:"name"`
	DHCP4       *bool    `pulumi:"dhcp4,optional"`
	Addresses   []string `pulumi:"addresses,optional"`
	Gateway4    *string  `pulumi:"gateway4,optional"`
	Nameservers []string `pulumi:"nameservers,optional"`
	MTU         *int     `pulumi:"mtu,optional"`
}

type AzureMarketplaceImage struct {
	Publisher string  `pulumi:"publisher"`
	Offer     string  `pulumi:"offer"`
	SKU       string  `pulumi:"sku"`
	Version   *string `pulumi:"version,optional"`
}

type AzureOsDisk struct {
	Type   *string `pulumi:"type,optional"`
	SizeGB *int    `pulumi:"sizeGb,optional"`
}

type GcpServiceAccount struct {
	Email  string   `pulumi:"email"`
	Scopes []string `pulumi:"scopes,optional"`
}

type HypervHardDrive struct {
	Path               string  `pulumi:"path"`
	ControllerType     *string `pulumi:"controllerType,optional"`
	ControllerNumber   *int    `pulumi:"controllerNumber,optional"`
	ControllerLocation *int    `pulumi:"controllerLocation,optional"`
}
