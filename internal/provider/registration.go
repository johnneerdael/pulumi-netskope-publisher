package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	netskopeClient := newNetskopeClient(args.TenantURL, args.APIToken, args.BearerToken, args.AuthMode, args.OAuth2, client)

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
	apiBase     string
	bearerToken string
	authMode    string
	oauth2      *NetskopeOAuth2Args
	accessToken string
	client      *http.Client
}

func newNetskopeClient(tenantURL string, apiToken *string, bearerToken *string, authMode *string, oauth2 *NetskopeOAuth2Args, client *http.Client) netskopeClient {
	token := stringValue(bearerToken)
	if token == "" {
		token = stringValue(apiToken)
	}
	return netskopeClient{
		apiBase:     strings.TrimRight(tenantURL, "/") + "/api/v2/infrastructure/publishers",
		bearerToken: token,
		authMode:    defaultString(authMode, "token"),
		oauth2:      oauth2,
		client:      client,
	}
}

func (client *netskopeClient) listPublishers(ctx context.Context) (map[string]int, error) {
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

func (client *netskopeClient) createPublisher(ctx context.Context, name string) (int, error) {
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

func (client *netskopeClient) generateRegistrationToken(ctx context.Context, publisherID int) (string, error) {
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

func (client *netskopeClient) request(
	ctx context.Context,
	operation string,
	method string,
	url string,
	body any,
	output any,
) error {
	token, err := client.resolveAccessToken(ctx)
	if err != nil {
		return err
	}

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
	request.Header.Set("Authorization", "Bearer "+token)
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

func (client *netskopeClient) resolveAccessToken(ctx context.Context) (string, error) {
	switch client.authMode {
	case "", "token":
		if client.bearerToken == "" {
			return "", fmt.Errorf("bearerToken or apiToken is required for token authentication")
		}
		return client.bearerToken, nil
	case "oauth2":
		if client.accessToken != "" {
			return client.accessToken, nil
		}
		token, err := client.fetchOAuth2AccessToken(ctx)
		if err != nil {
			return "", err
		}
		client.accessToken = token
		return token, nil
	default:
		return "", fmt.Errorf("unsupported authMode %q", client.authMode)
	}
}

func (client *netskopeClient) fetchOAuth2AccessToken(ctx context.Context) (string, error) {
	if client.oauth2 == nil || client.oauth2.TokenURL == "" || client.oauth2.ClientID == "" || client.oauth2.ClientSecret == "" {
		return "", fmt.Errorf("oauth2.tokenUrl, clientId, and clientSecret are required for OAuth2 authentication")
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", client.oauth2.ClientID)
	form.Set("client_secret", client.oauth2.ClientSecret)
	if client.oauth2.Scope != nil && *client.oauth2.Scope != "" {
		form.Set("scope", *client.oauth2.Scope)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.oauth2.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("Fetch OAuth2 access token failed: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Fetch OAuth2 access token returned unreadable body: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Fetch OAuth2 access token failed (status=%d)", response.StatusCode)
	}
	var parsed struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("Fetch OAuth2 access token returned invalid JSON: %w", err)
	}
	if parsed.AccessToken == "" {
		return "", fmt.Errorf("Fetch OAuth2 access token returned no access_token")
	}
	return parsed.AccessToken, nil
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
