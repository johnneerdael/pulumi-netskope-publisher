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

var errNetskopeNotFound = fmt.Errorf("netskope resource not found")

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
	if response.StatusCode == http.StatusNotFound {
		return errNetskopeNotFound
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

type privateAppProtocol struct {
	Type string `json:"type"`
	Port string `json:"port,omitempty"`
}

type privateAppTag struct {
	TagID   int    `json:"tag_id,omitempty"`
	TagName string `json:"tag_name"`
}

type privateAppPublisher struct {
	PublisherID   string `json:"publisher_id"`
	PublisherName string `json:"publisher_name,omitempty"`
}

type privateAppPayload struct {
	AppName              string                `json:"app_name"`
	AppType              string                `json:"app_type,omitempty"`
	Host                 any                   `json:"host"`
	ClientlessAccess     bool                  `json:"clientless_access"`
	IsUserPortalApp      bool                  `json:"is_user_portal_app"`
	Protocols            []privateAppProtocol  `json:"protocols"`
	TrustSelfSignedCerts bool                  `json:"trust_self_signed_certs"`
	UsePublisherDNS      bool                  `json:"use_publisher_dns"`
	PrivateAppTags       []privateAppTag       `json:"private_app_tags,omitempty"`
	Tags                 []privateAppTag       `json:"tags,omitempty"`
	Publishers           []privateAppPublisher `json:"publishers,omitempty"`
}

type privateAppRecord struct {
	AppID   int             `json:"app_id"`
	ID      int             `json:"id"`
	AppName string          `json:"app_name"`
	Name    string          `json:"name"`
	Host    any             `json:"host"`
	Tags    []privateAppTag `json:"tags"`
}

func (app privateAppRecord) resourceID() int {
	if app.AppID != 0 {
		return app.AppID
	}
	return app.ID
}

type privateAppsListResponse struct {
	Status string `json:"status"`
	Data   struct {
		PrivateApps []privateAppRecord `json:"private_apps"`
	} `json:"data"`
}

func (client *netskopeClient) listPrivateApps(ctx context.Context) ([]privateAppRecord, error) {
	var response privateAppsListResponse
	if err := client.request(ctx, "List private apps", http.MethodGet, "/api/v2/steering/apps/private", nil, &response); err != nil {
		return nil, err
	}
	return response.Data.PrivateApps, nil
}

func (client *netskopeClient) findPrivateAppByName(ctx context.Context, name string) (*privateAppRecord, error) {
	apps, err := client.listPrivateApps(ctx)
	if err != nil {
		return nil, err
	}
	for _, app := range apps {
		if app.AppName == name || app.Name == name {
			return &app, nil
		}
	}
	return nil, nil
}

func (client *netskopeClient) createPrivateApp(ctx context.Context, payload privateAppPayload) (privateAppRecord, error) {
	var response struct {
		Status string           `json:"status"`
		Data   privateAppRecord `json:"data"`
	}
	if err := client.request(ctx, "Create private app "+payload.AppName, http.MethodPost, "/api/v2/steering/apps/private", payload, &response); err != nil {
		return privateAppRecord{}, err
	}
	return response.Data, nil
}

func (client *netskopeClient) getPrivateApp(ctx context.Context, id int) (privateAppRecord, error) {
	var response struct {
		Status string           `json:"status"`
		Data   privateAppRecord `json:"data"`
	}
	path := fmt.Sprintf("/api/v2/steering/apps/private/%d", id)
	if err := client.request(ctx, fmt.Sprintf("Get private app %d", id), http.MethodGet, path, nil, &response); err != nil {
		return privateAppRecord{}, err
	}
	return response.Data, nil
}

func (client *netskopeClient) updatePrivateApp(ctx context.Context, id int, payload privateAppPayload) (privateAppRecord, error) {
	var response struct {
		Status string           `json:"status"`
		Data   privateAppRecord `json:"data"`
	}
	path := fmt.Sprintf("/api/v2/steering/apps/private/%d", id)
	if err := client.request(ctx, "Update private app "+payload.AppName, http.MethodPut, path, payload, &response); err != nil {
		return privateAppRecord{}, err
	}
	return response.Data, nil
}

func (client *netskopeClient) deletePrivateApp(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v2/steering/apps/private/%d", id)
	return client.request(ctx, fmt.Sprintf("Delete private app %d", id), http.MethodDelete, path, nil, nil)
}

type privateAppPublisherAssignment struct {
	PublisherID int `json:"publisher_id"`
}

type privateAppRecordWithPublishers struct {
	AppID                       int                             `json:"app_id"`
	ID                          int                             `json:"id"`
	AppName                     string                          `json:"app_name"`
	Name                        string                          `json:"name"`
	Tags                        []privateAppTag                 `json:"tags"`
	ServicePublisherAssignments []privateAppPublisherAssignment `json:"service_publisher_assignments"`
}

func (app privateAppRecordWithPublishers) resourceID() int {
	if app.AppID != 0 {
		return app.AppID
	}
	return app.ID
}

type privateAppsWithPublishersListResponse struct {
	Status string `json:"status"`
	Data   struct {
		PrivateApps []privateAppRecordWithPublishers `json:"private_apps"`
	} `json:"data"`
}

func (client *netskopeClient) listPrivateAppsWithPublishers(ctx context.Context) ([]privateAppRecordWithPublishers, error) {
	var response privateAppsWithPublishersListResponse
	if err := client.request(ctx, "List private apps", http.MethodGet, "/api/v2/steering/apps/private", nil, &response); err != nil {
		return nil, err
	}
	return response.Data.PrivateApps, nil
}

func (client *netskopeClient) replacePrivateAppPublishers(ctx context.Context, appIDs []int, publisherIDs []int) error {
	privateAppIDs := make([]string, 0, len(appIDs))
	for _, id := range appIDs {
		privateAppIDs = append(privateAppIDs, strconv.Itoa(id))
	}
	publisherIDValues := make([]string, 0, len(publisherIDs))
	for _, id := range publisherIDs {
		publisherIDValues = append(publisherIDValues, strconv.Itoa(id))
	}
	body := map[string]any{
		"private_app_ids": privateAppIDs,
		"publisher_ids":   publisherIDValues,
	}
	return client.request(ctx, "Replace private app publishers", http.MethodPut, "/api/v2/steering/apps/private/publishers", body, nil)
}

func (client *netskopeClient) deletePrivateAppPublishers(ctx context.Context, appIDs []int, publisherIDs []int) error {
	privateAppIDs := make([]string, 0, len(appIDs))
	for _, id := range appIDs {
		privateAppIDs = append(privateAppIDs, strconv.Itoa(id))
	}
	publisherIDValues := make([]string, 0, len(publisherIDs))
	for _, id := range publisherIDs {
		publisherIDValues = append(publisherIDValues, strconv.Itoa(id))
	}
	body := map[string]any{
		"private_app_ids": privateAppIDs,
		"publisher_ids":   publisherIDValues,
	}
	return client.request(ctx, "Delete private app publishers", http.MethodDelete, "/api/v2/steering/apps/private/publishers", body, nil)
}

type policyGroupRecord struct {
	ID   int    `json:"group_id"`
	Name string `json:"group_name"`
}

type realtimePolicyAction struct {
	ActionName string `json:"action_name"`
}

type realtimePolicyRuleData struct {
	PrivateApps         []string             `json:"privateApps,omitempty"`
	PrivateAppTags      []string             `json:"privateAppTags,omitempty"`
	Users               []string             `json:"users,omitempty"`
	UserGroups          []string             `json:"userGroups,omitempty"`
	MatchCriteriaAction realtimePolicyAction `json:"match_criteria_action"`
}

type realtimePolicyPayload struct {
	RuleName  string                 `json:"rule_name"`
	GroupID   string                 `json:"group_id,omitempty"`
	GroupName string                 `json:"group_name,omitempty"`
	RuleData  realtimePolicyRuleData `json:"rule_data"`
	Enabled   string                 `json:"enabled"`
}

type realtimePolicyRecord struct {
	RuleID   int    `json:"rule_id"`
	RuleName string `json:"rule_name"`
}

type realtimePolicyEnvelope struct {
	Status string               `json:"status"`
	Data   realtimePolicyRecord `json:"data"`
}

func (client *netskopeClient) findPolicyGroupByName(ctx context.Context, name string) (*policyGroupRecord, error) {
	var response []policyGroupRecord
	if err := client.request(ctx, "List policy groups", http.MethodGet, "/api/v2/policy/npa/policygroups", nil, &response); err != nil {
		return nil, err
	}
	for _, group := range response {
		if group.Name == name {
			return &group, nil
		}
	}
	return nil, nil
}

func (client *netskopeClient) createRealtimePolicy(ctx context.Context, payload realtimePolicyPayload) (realtimePolicyRecord, error) {
	var response realtimePolicyRecord
	if err := client.request(ctx, "Create realtime protection policy "+payload.RuleName, http.MethodPost, "/api/v2/policy/npa/rules", payload, &response); err != nil {
		return realtimePolicyRecord{}, err
	}
	return response, nil
}

func (client *netskopeClient) updateRealtimePolicy(ctx context.Context, id int, payload realtimePolicyPayload) (realtimePolicyRecord, error) {
	var response realtimePolicyEnvelope
	path := fmt.Sprintf("/api/v2/policy/npa/rules/%d", id)
	if err := client.request(ctx, "Update realtime protection policy "+payload.RuleName, http.MethodPatch, path, payload, &response); err != nil {
		return realtimePolicyRecord{}, err
	}
	return response.Data, nil
}

func (client *netskopeClient) getRealtimePolicy(ctx context.Context, id int) (realtimePolicyRecord, error) {
	var response realtimePolicyEnvelope
	path := fmt.Sprintf("/api/v2/policy/npa/rules/%d", id)
	if err := client.request(ctx, fmt.Sprintf("Get realtime protection policy %d", id), http.MethodGet, path, nil, &response); err != nil {
		return realtimePolicyRecord{}, err
	}
	return response.Data, nil
}

func (client *netskopeClient) deleteRealtimePolicy(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v2/policy/npa/rules/%d", id)
	return client.request(ctx, fmt.Sprintf("Delete realtime protection policy %d", id), http.MethodDelete, path, nil, nil)
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
