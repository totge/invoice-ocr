package appsheet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
	defaultBaseURL = "https://appsheet.com"
)

type Client struct {
	httpClient *http.Client
	apiBaseUrl *url.URL
	appId      string
	apiKey     string
}

func (c *Client) SendRequest(ctx context.Context, table AppSheetTable, method string, body io.Reader) (*http.Response, error) {

	endpointUrl := c.apiBaseUrl.JoinPath(string(table)).JoinPath("Action")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointUrl.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// --- 1. Set Common Headers Expected by the AppSheet API---
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ApplicationAccessKey", c.apiKey)

	// --- 2. Execute Request ---
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}


func ReadRecords[TargetType any](c AppSheetClient, ctx context.Context, table AppSheetTable) ([]TargetType, error) {

	// --- 1. Construct Request Body (Consistent "Find" Action) ---
	requestBody := AppSheetActionRequest[TargetType]{
		Action: ActionFind,
		Properties: map[string]interface{}{
			// Add default properties or allow passing them via options if needed
			"Locale": "hu-HU",
		},
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// --- 2. Execute Request ---
	resp, err := c.SendRequest(ctx, table, http.MethodPost, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to execute request to %s: %w", table, err)
	}
	defer resp.Body.Close()

	// TODO: Check if api returns in-body errors
	// TODO: Is status checking unifor accross different types os api calls? if yes it should go to the do request function
	// --- 3. Handle Response Status Code ---
	if resp.StatusCode != http.StatusOK {
		// Attempt to read error body for more info
		bodyBytes, _ := io.ReadAll(resp.Body) // Use io.ReadAll
		return nil, fmt.Errorf("unexpected status code %d from AppSheet API for table '%s': %s", resp.StatusCode, string(table), string(bodyBytes))
	}

	// --- 4. Decode Successful Response using Generic Type ---
	// Use json.NewDecoder for efficiency
	parsedResponse := make([]TargetType, 0, 200)
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&parsedResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response body for table read: %w", err)
	}

	return parsedResponse, nil
}

func WriteRecords[TargetType any](c AppSheetClient, ctx context.Context, table AppSheetTable, records []TargetType) error {
	// --- 1. Construct the request body ---

	requestBody := AppSheetActionRequest[TargetType]{
		Action: ActionAdd,
		Properties: map[string]interface{}{
			// Add default properties or allow passing them via options if needed
			"Locale": "hu-HU",
		},
		Rows: records,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// --- 2. Execute Request ---
	resp, err := c.SendRequest(ctx, table, http.MethodPost, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to execute request to %s: %w", table, err)
	}
	defer resp.Body.Close()

	// TODO: Check if api returns in-body errors
	// --- 6. Handle Response Status Code ---
	if resp.StatusCode != http.StatusOK {
		// Attempt to read error body for more info
		bodyBytes, _ := io.ReadAll(resp.Body) // Use io.ReadAll
		return fmt.Errorf("unexpected status code %d from AppSheet API for table '%s': %s", resp.StatusCode, string(table), string(bodyBytes))
	}

	return nil
}

func NewClient(baseUrl string, appId string, apiKey string) (*Client, error) {
	if appId == "" || apiKey == "" {
		return nil, fmt.Errorf("appId and apiKey must not be empty")
	}

	if baseUrl == "" {
		baseUrl = defaultBaseURL
	}

	parsedBaseURL, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL %q: %w", baseUrl, err)
	}

	parsedBaseURL = parsedBaseURL.JoinPath("/api/v2/apps/", appId, "/tables/")

	// Create an http.Client with a timeout
	httpClient := &http.Client{
		Timeout: defaultTimeout,
		// You can customize Transport later if needed (e.g., for retries)
	}

	client := &Client{
		httpClient: httpClient,
		apiBaseUrl: parsedBaseURL,
		apiKey:     apiKey,
		appId:      appId,
	}
	return client, nil
}