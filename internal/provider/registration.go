package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"
)

type NetskopeRegistration struct{}

type NetskopeRegistrationArgs struct {
	PublisherNames []string `pulumi:"publisherNames"`
	TenantURL      string   `pulumi:"tenantUrl"`
	APIToken       string   `pulumi:"apiToken" provider:"secret"`
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
	netskopeClient := newNetskopeClient(args.TenantURL, args.APIToken, client)

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

type netskopeClient struct {
	apiBase  string
	apiToken string
	client   *http.Client
}

func newNetskopeClient(tenantURL string, apiToken string, client *http.Client) netskopeClient {
	return netskopeClient{
		apiBase:  strings.TrimRight(tenantURL, "/") + "/api/v2/infrastructure/publishers",
		apiToken: apiToken,
		client:   client,
	}
}

func (client netskopeClient) listPublishers(ctx context.Context) (map[string]int, error) {
	var response struct {
		Data struct {
			Publishers []struct {
				Name string      `json:"publisher_name"`
				ID   interface{} `json:"publisher_id"`
			} `json:"publishers"`
		} `json:"data"`
	}
	if err := client.request(ctx, "List publishers", http.MethodGet, client.apiBase, nil, &response); err != nil {
		return nil, err
	}

	publishers := make(map[string]int, len(response.Data.Publishers))
	for _, publisher := range response.Data.Publishers {
		id, err := parsePublisherID(publisher.ID)
		if err != nil {
			return nil, fmt.Errorf("List publishers returned invalid publisher ID for %s: %w", publisher.Name, err)
		}
		publishers[publisher.Name] = id
	}
	return publishers, nil
}

func (client netskopeClient) createPublisher(ctx context.Context, name string) (int, error) {
	var response struct {
		Data struct {
			ID interface{} `json:"id"`
		} `json:"data"`
	}
	body := map[string]string{"name": name}
	if err := client.request(ctx, "Create publisher "+name, http.MethodPost, client.apiBase, body, &response); err != nil {
		return 0, err
	}
	return parsePublisherID(response.Data.ID)
}

func (client netskopeClient) generateRegistrationToken(ctx context.Context, publisherID int) (string, error) {
	var response struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	url := fmt.Sprintf("%s/%d/registration_token", client.apiBase, publisherID)
	if err := client.request(ctx, fmt.Sprintf("Generate registration token for publisher %d", publisherID), http.MethodPost, url, nil, &response); err != nil {
		return "", err
	}
	return response.Data.Token, nil
}

func (client netskopeClient) request(
	ctx context.Context,
	operation string,
	method string,
	url string,
	body any,
	output any,
) error {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}

	request, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return err
	}
	request.Header.Set("Netskope-Api-Token", client.apiToken)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := client.client.Do(request)
	if err != nil {
		return fmt.Errorf("%s failed: %w", operation, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s failed (status=%d)", operation, response.StatusCode)
	}
	if output == nil {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return fmt.Errorf("%s returned invalid JSON: %w", operation, err)
	}
	return nil
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
