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
	defaultBaseURL = "https://appsheet.com" // Verify this URL
)

type Client struct {
	httpClient *http.Client
	apiBaseUrl *url.URL
	appId      string
	apiKey     string
}

func ReadRecords[TargetType any](c *Client, ctx context.Context, table AppSheetTable) ([]TargetType, error) {

	// --- 1. Construct URL for the endpoint ---
	// TODO: Make it nicer
	endpointUrl := c.apiBaseUrl.JoinPath(string(table)).JoinPath("Action")
	// --- 2. Construct Request Body (Consistent "Find" Action) ---
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

	// --- 3. Create Request ---
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointUrl.String(), bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// --- 4. Set Headers ---
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ApplicationAccessKey", c.apiKey)

	// --- 5. Execute Request ---
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request to %s: %w", endpointUrl.Redacted(), err)
	}
	defer resp.Body.Close()

	// TODO: Check if api returns in-body errors
	// --- 6. Handle Response Status Code ---
	if resp.StatusCode != http.StatusOK {
		// Attempt to read error body for more info
		bodyBytes, _ := io.ReadAll(resp.Body) // Use io.ReadAll
		return nil, fmt.Errorf("unexpected status code %d from AppSheet API for table '%s': %s", resp.StatusCode, string(table), string(bodyBytes))
	}

	// --- 7. Decode Successful Response using Generic Type ---
	// Use json.NewDecoder for efficiency
	parsedResponse := make([]TargetType, 0, 100)
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&parsedResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response body for table read: %w", err)
	}

	return parsedResponse, nil
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
