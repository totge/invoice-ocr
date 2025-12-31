package appsheet

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// TestBuildRequestBody isolates the pure logic of creating the JSON payload.
func TestClient_buildRequestBody(t *testing.T) {
	// A dummy client is needed to call the method, but its fields won't be used.
	c := &Client{}

	type testRow struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	testCases := []struct {
		name        string
		action      appSheetAction
		data        any
		wantJSONMap map[string]interface{} // We'll compare unmarshalled maps for robustness
		expectError bool
	}{
		{
			name:   "Find action with nil data",
			action: actionFind,
			data:   nil,
			wantJSONMap: map[string]interface{}{
				"Action":     "Find",
				"Properties": map[string]interface{}{"Locale": "hu-HU"},
			},
			expectError: false,
		},
		{
			name:   "Add action with data",
			action: actionAdd,
			data:   []testRow{{ID: "1", Name: "Test"}},
			wantJSONMap: map[string]interface{}{
				"Action":     "Add",
				"Properties": map[string]interface{}{"Locale": "hu-HU"},
				"Rows":       []interface{}{map[string]interface{}{"id": "1", "name": "Test"}},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bodyBuffer, err := c.buildRequestBody(tc.action, tc.data)

			if (err != nil) != tc.expectError {
				t.Fatalf("buildRequestBody() error = %v, wantErr %v", err, tc.expectError)
			}
			if tc.expectError {
				return
			}

			// Unmarshal the actual result to compare against the wanted map
			var gotMap map[string]interface{}
			if err := json.Unmarshal(bodyBuffer.Bytes(), &gotMap); err != nil {
				t.Fatalf("Failed to unmarshal actual response: %v", err)
			}

			if !reflect.DeepEqual(gotMap, tc.wantJSONMap) {
				t.Errorf("buildRequestBody() mismatch:\ngot = %s\nwant= %v", gotMap, tc.wantJSONMap)
			}
		})
	}
}

func TestClient_ReadCategories_Success(t *testing.T) {
	// 1. Setup the Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert that the request your client sent is correct
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("ApplicationAccessKey") != "fake-api-key" {
			t.Errorf("Expected ApplicationAccessKey header to be 'fake-api-key'")
		}

		// Check the request body
		body, _ := io.ReadAll(r.Body)
		var reqBody appSheetActionRequest
		json.Unmarshal(body, &reqBody)
		if reqBody.Action != actionFind {
			t.Errorf("Expected Action to be 'Find', got '%s'", reqBody.Action)
		}

		// Send back a canned, successful response
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		// This is the fake data that our client should successfully parse
		json.NewEncoder(w).Encode([]Category{
			{CategoryId: "cat-123", CostGroup: "Groceries", MainCategory: "Dairy", SubCategory: "Milk"},
			{CategoryId: "cat-456", CostGroup: "Utilities", MainCategory: "Home", SubCategory: "Internet"},
		})
	}))
	// Close the server when the test is done
	defer server.Close()

	// 2. Create a Client instance pointing to the mock server
	// We use server.URL which provides the http://localhost:xxxx address of the fake server
	client, err := NewClient(server.URL, "fake-app-id", "fake-api-key")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 3. Call the method we want to test
	categories, err := client.ReadCategories(context.Background())

	// 4. Assert the results
	if err != nil {
		t.Fatalf("ReadCategories() returned an unexpected error: %v", err)
	}
	if len(categories) != 2 {
		t.Fatalf("Expected 2 categories, got %d", len(categories))
	}
	if categories[0].CategoryId != "cat-123" {
		t.Errorf("Expected first category ID to be 'cat-123', got '%s'", categories[0].CategoryId)
	}
	if categories[1].CostGroup != "Utilities" {
		t.Errorf("Expected second cost group to be 'Utilities', got '%s'", categories[1].CostGroup)
	}
}

func TestClient_ReadCategories_APIError(t *testing.T) {
	// 1. Setup a Mock Server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send back a 500 Internal Server Error status code
		w.WriteHeader(http.StatusInternalServerError)
		// Write an error message in the body, which our client should capture
		w.Write([]byte("Something went wrong on the server"))
	}))
	defer server.Close()

	// 2. Create the Client
	client, _ := NewClient(server.URL, "fake-app-id", "fake-api-key")

	// 3. Call the method
	_, err := client.ReadCategories(context.Background())

	// 4. Assert that an error WAS returned
	if err == nil {
		t.Fatal("ReadCategories() was expected to return an error, but it didn't")
	}

	// Optional: Check if the error message contains the expected details
	expectedErrorSubstring := "status code 500"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error message to contain '%s', but got: %v", expectedErrorSubstring, err)
	}
}
