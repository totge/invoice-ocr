package appsheet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// AppSheetActionRequest defines the structure for the body of an AppSheet Action API call.
type AppSheetActionRequestv2 struct {
	Action     AppSheetAction         `json:"Action"`
	Properties map[string]interface{} `json:"Properties,omitempty"`
	Rows       any                    `json:"Rows,omitempty"` // Used for Add/Edit/Delete, often empty/nil for Find
}

type Clientv2 struct {
	httpClient *http.Client
	apiBaseUrl *url.URL
	appId      string
	apiKey     string
}

// Handles sending a request and processing the response from the Appsheet API
// 'rows' represent a slice of structs that can be marshalled to the targeted table schema for write requests
// 'decodeTarget' represents a slice structs of the expected type, mapping to the table schema for read requests
func (c *Clientv2) doRequest(ctx context.Context, table AppSheetTable, action AppSheetAction, rows any, decodeTarget any) error {

	// --- 1. Build the url for the request
	endpointUrl := c.apiBaseUrl.JoinPath(string(table)).JoinPath("Action")
	// --- 2. Build the request body
	body, err := c.buildRequestBody(action, rows)
	if err != nil {
		return fmt.Errorf("failed to build request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointUrl.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	// --- 3. Set Headers Expected by the AppSheet API
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ApplicationAccessKey", c.apiKey)

	// --- 4. Execute Request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// --- 5. Handle unsucessful request response
	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		// Attempt to read error body for more info
		return fmt.Errorf("unexpected status code %d from AppSheet API for table '%s': %s", resp.StatusCode, string(table), string(respBody))
	}

	// --- 6. Decode response to the target
	if decodeTarget != nil {
		if err := json.NewDecoder(resp.Body).Decode(decodeTarget); err != nil {
			return fmt.Errorf("failed to decode successful response body: %w", err)
		}
	}
	return nil
}

func (c *Clientv2) buildRequestBody(action AppSheetAction, data any) (*bytes.Buffer, error) {
	requestBody := AppSheetActionRequestv2{
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

func (c *Clientv2) ReadExpenses(ctx context.Context) ([]Expense, error) {
	var expenses []Expense

	err := c.doRequest(ctx, TableExpenses, ActionFind, nil, &expenses)
	if err != nil {
		return nil, err
	}

	return expenses, nil
}

func (c *Clientv2) ReadCategories(ctx context.Context) ([]Category, error) {
	var categories []Category
	err := c.doRequest(ctx, TableCategories, ActionFind, nil, &categories)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (c *Clientv2) WriteExpenseStage(ctx context.Context, rows []ExpenseStage) error {
	err := c.doRequest(ctx, TableExpenseStage, ActionAdd, rows, nil)
	if err != nil {
		return err
	}

	return nil
}

func NewClientv2(baseUrl string, appId string, apiKey string) (*Clientv2, error) {
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

	client := &Clientv2{
		httpClient: httpClient,
		apiBaseUrl: parsedBaseURL,
		apiKey:     apiKey,
		appId:      appId,
	}
	return client, nil
}
