package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

type PrivateApp struct{}

type PrivateAppProtocol struct {
	Type  string `pulumi:"type"`
	Ports string `pulumi:"ports"`
}

type PrivateAppArgs struct {
	TenantURL            string               `pulumi:"tenantUrl"`
	APIToken             *string              `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken          *string              `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode             *string              `pulumi:"authMode,optional"`
	OAuth2               *NetskopeOAuth2Args  `pulumi:"oauth2,optional"`
	AppName              string               `pulumi:"appName"`
	AppType              *string              `pulumi:"appType,optional"`
	Host                 string               `pulumi:"host"`
	Hosts                []string             `pulumi:"hosts,optional"`
	Protocols            []PrivateAppProtocol `pulumi:"protocols"`
	ClientlessAccess     bool                 `pulumi:"clientlessAccess"`
	IsUserPortalApp      bool                 `pulumi:"isUserPortalApp"`
	UsePublisherDNS      bool                 `pulumi:"usePublisherDns"`
	TrustSelfSignedCerts bool                 `pulumi:"trustSelfSignedCerts"`
	Tags                 []string             `pulumi:"tags,optional"`
	AdoptExisting        *bool                `pulumi:"adoptExisting,optional"`
}

type PrivateAppOutputs struct {
	PrivateAppArgs
	AppID int `pulumi:"appId"`
}

func (*PrivateApp) Annotate(a infer.Annotator) {
	a.SetToken("index", "PrivateApp")
}

func (*PrivateApp) Create(ctx context.Context, req infer.CreateRequest[PrivateAppArgs]) (infer.CreateResponse[PrivateAppOutputs], error) {
	output := PrivateAppOutputs{PrivateAppArgs: req.Inputs}
	if req.DryRun {
		return infer.CreateResponse[PrivateAppOutputs]{ID: req.Inputs.AppName, Output: output}, nil
	}

	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	existing, err := client.findPrivateAppByName(ctx, req.Inputs.AppName)
	if err != nil {
		return infer.CreateResponse[PrivateAppOutputs]{}, err
	}

	payload := privateAppPayloadFromArgs(req.Inputs)
	if existing != nil {
		if !defaultBool(req.Inputs.AdoptExisting, false) {
			return infer.CreateResponse[PrivateAppOutputs]{}, fmt.Errorf("private app %q already exists; import it or set adoptExisting: true to manage it", req.Inputs.AppName)
		}
		updated, err := client.updatePrivateApp(ctx, existing.resourceID(), payload)
		if err != nil {
			return infer.CreateResponse[PrivateAppOutputs]{}, err
		}
		output.AppID = updated.resourceID()
		return infer.CreateResponse[PrivateAppOutputs]{ID: strconv.Itoa(output.AppID), Output: output}, nil
	}

	created, err := client.createPrivateApp(ctx, payload)
	if err != nil {
		return infer.CreateResponse[PrivateAppOutputs]{}, err
	}
	output.AppID = created.resourceID()
	return infer.CreateResponse[PrivateAppOutputs]{ID: strconv.Itoa(output.AppID), Output: output}, nil
}

func (*PrivateApp) Read(ctx context.Context, req infer.ReadRequest[PrivateAppArgs, PrivateAppOutputs]) (infer.ReadResponse[PrivateAppArgs, PrivateAppOutputs], error) {
	return infer.ReadResponse[PrivateAppArgs, PrivateAppOutputs]{ID: req.ID, Inputs: req.Inputs, State: req.State}, nil
}

func (*PrivateApp) Update(ctx context.Context, req infer.UpdateRequest[PrivateAppArgs, PrivateAppOutputs]) (infer.UpdateResponse[PrivateAppOutputs], error) {
	output := PrivateAppOutputs{PrivateAppArgs: req.Inputs}
	appID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.UpdateResponse[PrivateAppOutputs]{}, fmt.Errorf("invalid private app ID %q: %w", req.ID, err)
	}
	output.AppID = appID
	if req.DryRun {
		return infer.UpdateResponse[PrivateAppOutputs]{Output: output}, nil
	}

	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	updated, err := client.updatePrivateApp(ctx, appID, privateAppPayloadFromArgs(req.Inputs))
	if err != nil {
		return infer.UpdateResponse[PrivateAppOutputs]{}, err
	}
	output.AppID = updated.resourceID()
	return infer.UpdateResponse[PrivateAppOutputs]{Output: output}, nil
}

func (*PrivateApp) Delete(ctx context.Context, req infer.DeleteRequest[PrivateAppOutputs]) (infer.DeleteResponse, error) {
	appID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid private app ID %q: %w", req.ID, err)
	}
	client := newResourceClient(req.State.TenantURL, req.State.APIToken, req.State.BearerToken, req.State.AuthMode, req.State.OAuth2, http.DefaultClient)
	return infer.DeleteResponse{}, client.deletePrivateApp(ctx, appID)
}

func newResourceClient(tenantURL string, apiToken *string, bearerToken *string, authMode *string, oauth2 *NetskopeOAuth2Args, httpClient *http.Client) netskopeClient {
	return newNetskopeClient(netskopeClientConfig{
		TenantURL:   tenantURL,
		APIToken:    apiToken,
		BearerToken: stringValue(bearerToken),
		AuthMode:    defaultString(authMode, "token"),
		OAuth2:      oauth2,
		HTTPClient:  httpClient,
	})
}

func privateAppPayloadFromArgs(args PrivateAppArgs) privateAppPayload {
	host := any(args.Host)
	if len(args.Hosts) > 0 {
		host = args.Hosts
	}

	protocols := make([]privateAppProtocol, 0, len(args.Protocols))
	for _, protocol := range args.Protocols {
		protocols = append(protocols, privateAppProtocol{
			Type:  protocol.Type,
			Port:  protocol.Ports,
			Ports: protocol.Ports,
		})
	}

	tags := make([]privateAppTag, 0, len(args.Tags))
	for _, tag := range args.Tags {
		tags = append(tags, privateAppTag{TagName: tag})
	}

	return privateAppPayload{
		AppName:              args.AppName,
		AppType:              defaultString(args.AppType, "client"),
		Host:                 host,
		ClientlessAccess:     args.ClientlessAccess,
		IsUserPortalApp:      args.IsUserPortalApp,
		Protocols:            protocols,
		TrustSelfSignedCerts: args.TrustSelfSignedCerts,
		UsePublisherDNS:      args.UsePublisherDNS,
		PrivateAppTags:       tags,
		Tags:                 tags,
	}
}
