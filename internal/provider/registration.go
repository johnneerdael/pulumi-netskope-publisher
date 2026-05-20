package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"
)

type NetskopeRegistration struct{}

type NetskopeRegistrationArgs struct {
	PublisherNames []string            `pulumi:"publisherNames"`
	TenantURL      string              `pulumi:"tenantUrl"`
	APIToken       *string             `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken    *string             `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode       *string             `pulumi:"authMode,optional"`
	OAuth2         *NetskopeOAuth2Args `pulumi:"oauth2,optional"`
}

type RegistrationRecord struct {
	PublisherID       int    `pulumi:"publisherId"`
	RegistrationToken string `pulumi:"registrationToken" provider:"secret"`
	ExistedBefore     bool   `pulumi:"existedBefore"`
}

type NetskopeRegistrationOutputs struct {
	NetskopeRegistrationArgs
	Registrations map[string]RegistrationRecord `pulumi:"registrations"`
}

func (*NetskopeRegistration) Annotate(a infer.Annotator) {
	a.SetToken("index", "NetskopeRegistration")
}

func (*NetskopeRegistration) Create(
	ctx context.Context,
	req infer.CreateRequest[NetskopeRegistrationArgs],
) (infer.CreateResponse[NetskopeRegistrationOutputs], error) {
	output := NetskopeRegistrationOutputs{
		NetskopeRegistrationArgs: req.Inputs,
		Registrations:            emptyRegistrationRecords(req.Inputs.PublisherNames),
	}
	id := strings.Join(req.Inputs.PublisherNames, ",")

	if req.DryRun {
		return infer.CreateResponse[NetskopeRegistrationOutputs]{
			ID:     id,
			Output: output,
		}, nil
	}

	registrations, err := resolveNetskopeRegistrations(ctx, req.Inputs, http.DefaultClient)
	if err != nil {
		return infer.CreateResponse[NetskopeRegistrationOutputs]{}, err
	}
	output.Registrations = registrations

	return infer.CreateResponse[NetskopeRegistrationOutputs]{
		ID:     id,
		Output: output,
	}, nil
}

func emptyRegistrationRecords(names []string) map[string]RegistrationRecord {
	registrations := make(map[string]RegistrationRecord, len(names))
	for _, name := range names {
		registrations[name] = RegistrationRecord{}
	}
	return registrations
}

func resolveNetskopeRegistrations(
	ctx context.Context,
	args NetskopeRegistrationArgs,
	client *http.Client,
) (map[string]RegistrationRecord, error) {
	netskopeClient := newNetskopeClient(netskopeClientConfig{
		TenantURL:   args.TenantURL,
		APIToken:    args.APIToken,
		BearerToken: stringValue(args.BearerToken),
		AuthMode:    defaultString(args.AuthMode, "token"),
		OAuth2:      args.OAuth2,
		HTTPClient:  client,
	})

	existingByName, err := netskopeClient.listPublishers(ctx)
	if err != nil {
		return nil, err
	}

	registrations := make(map[string]RegistrationRecord, len(args.PublisherNames))
	for _, publisherName := range args.PublisherNames {
		publisherID, existedBefore := existingByName[publisherName]
		if !existedBefore {
			publisherID, err = netskopeClient.createPublisher(ctx, publisherName)
			if err != nil {
				return nil, err
			}
		}

		token, err := netskopeClient.generateRegistrationToken(ctx, publisherID)
		if err != nil {
			return nil, err
		}

		registrations[publisherName] = RegistrationRecord{
			PublisherID:       publisherID,
			RegistrationToken: token,
			ExistedBefore:     existedBefore,
		}
	}

	return registrations, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func parsePublisherID(value interface{}) (int, error) {
	switch value := value.(type) {
	case float64:
		return int(value), nil
	case string:
		return strconv.Atoi(value)
	default:
		return 0, fmt.Errorf("expected string or number, got %T", value)
	}
}
