package appsheet

import (
	"fmt"
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

	parsedBaseURL.JoinPath("/api/v2/apps/", appId, "/tables/")

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