package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type netskopeClientConfig struct {
	TenantURL   string
	APIToken    *string
	BearerToken string
	AuthMode    string
	OAuth2      *NetskopeOAuth2Args
	HTTPClient  *http.Client
}

type netskopeClient struct {
	tenantURL   string
	bearerToken string
	authMode    string
	oauth2      *NetskopeOAuth2Args
	accessToken string
	httpClient  *http.Client
}

func newNetskopeClient(config netskopeClientConfig) netskopeClient {
	token := config.BearerToken
	if token == "" {
		token = stringValue(config.APIToken)
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return netskopeClient{
		tenantURL:   strings.TrimRight(config.TenantURL, "/"),
		bearerToken: token,
		authMode:    defaultString(&config.AuthMode, "token"),
		oauth2:      config.OAuth2,
		httpClient:  httpClient,
	}
}

func (client *netskopeClient) endpoint(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return client.tenantURL + path
}

func (client *netskopeClient) request(
	ctx context.Context,
	operation string,
	method string,
	path string,
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

	request, err := http.NewRequestWithContext(ctx, method, client.endpoint(path), reader)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("%s failed: %w", operation, err)
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("%s failed reading response: %w", operation, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s failed (status=%d): %s", operation, response.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	if output == nil || len(bodyBytes) == 0 {
		return nil
	}
	if err := json.Unmarshal(bodyBytes, output); err != nil {
		return fmt.Errorf("%s returned invalid JSON: %w", operation, err)
	}
	return nil
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
	if err := client.request(ctx, "List publishers", http.MethodGet, "/api/v2/infrastructure/publishers", nil, &response); err != nil {
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
	if err := client.request(ctx, "Create publisher "+name, http.MethodPost, "/api/v2/infrastructure/publishers", map[string]string{"name": name}, &response); err != nil {
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
	path := fmt.Sprintf("/api/v2/infrastructure/publishers/%d/registration_token", publisherID)
	if err := client.request(ctx, fmt.Sprintf("Generate registration token for publisher %d", publisherID), http.MethodPost, path, nil, &response); err != nil {
		return "", err
	}
	return response.Data.Token, nil
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

	response, err := client.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("Fetch OAuth2 access token failed: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Fetch OAuth2 access token returned unreadable body: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Fetch OAuth2 access token failed (status=%d): %s", response.StatusCode, strings.TrimSpace(string(body)))
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
