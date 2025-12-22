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

// Handles sending a request and processing the response from the Appsheet API
// 'rows' represent a slice of structs that can be marshalled to the targeted table schema for write requests
// 'decodeTarget' represents a slice structs of the expected type, mapping to the table schema for read requests
func (c *Client) doRequest(ctx context.Context, table appSheetTable, action appSheetAction, rows any, decodeTarget any) error {

	// --- 1. Build the url for the request
	endpointUrl := c.apiBaseUrl.JoinPath(string(table)).JoinPath("Action")
	// --- 2. Build the request body
	body, err := c.buildRequestBody(action, rows)
	if err != nil {
		return fmt.Errorf("failed to build apsheet request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointUrl.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create http request object: %w", err)
	}
	// --- 3. Set Headers Expected by the AppSheet API
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ApplicationAccessKey", c.apiKey)

	// --- 4. Execute Request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http client failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// --- 5. Handle unsucessful request response
	if resp.StatusCode != http.StatusOK {
		// ignore error, main information in the status code
		respBody, _ := io.ReadAll(resp.Body)

		// Attempt to read error body for more info
		return fmt.Errorf("appsheet API returned non-OK status code %d from AppSheet API for table '%s': %s", resp.StatusCode, string(table), string(respBody))
	}

	// --- 6. Decode response to the target
	if decodeTarget != nil {
		if err := json.NewDecoder(resp.Body).Decode(decodeTarget); err != nil {
			return fmt.Errorf("failed to decode successful appsheet response body: %w", err)
		}
	}
	return nil
}

func (c *Client) buildRequestBody(action appSheetAction, data any) (*bytes.Buffer, error) {
	requestBody := appSheetActionRequest{
		Action: action,
		Properties: map[string]interface{}{
			// Add default properties or allow passing them via options if needed
			"Locale": "hu-HU",
		},
		Rows: data,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	return bytes.NewBuffer(requestBodyBytes), nil
}

func (c *Client) ReadExpenses(ctx context.Context) ([]Expense, error) {
	var expenses []Expense

	err := c.doRequest(ctx, TableExpenses, actionFind, nil, &expenses)
	if err != nil {
		return nil, fmt.Errorf("could not read expenses from appsheet: %w", err)
	}

	return expenses, nil
}

func (c *Client) ReadCategories(ctx context.Context) ([]Category, error) {
	var categories []Category
	err := c.doRequest(ctx, TableCategories, actionFind, nil, &categories)
	if err != nil {
		return nil, fmt.Errorf("could not read categories from appsheet: %w", err)
	}

	return categories, nil
}

func (c *Client) WriteExpenseStage(ctx context.Context, rows []ExpenseStage) error {
	err := c.doRequest(ctx, TableExpenseStage, actionAdd, rows, nil)
	if err != nil {
		return fmt.Errorf("could not write expenses to the stage in appsheet: %w", err)
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
