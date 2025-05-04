package receipt

import (
	"io"
	"reflect" // Needed for deep comparison of structs/slices
	"strings"
	"testing" // The core testing package
)

func TestParseReceipt(t *testing.T) { // Test function signature
	// Define test cases using a struct slice
	testCases := []struct {
		name        string    // Name for the test case (shows up in verbose output)
		inputJSON   io.Reader // The input JSON string
		wantReceipt *Receipt  // The expected output struct (nil if error is expected)
		wantErr     bool      // Whether an error is expected
	}{
		// --- Test Case 1: Valid JSON ---
		{
			name: "valid receipt json",
			inputJSON: strings.NewReader(`{
				"datetime": "2024-05-05T10:00:00Z",
				"items": [
					{"name": "Milk", "price": 100, "discount": 0},
					{"name": "Bread", "price": 150, "discount": 10}
				],
				"parsed_total": 250
			}`),
			wantReceipt: &Receipt{
				Timestamp: "2024-05-05T10:00:00Z",
				Items: []Item{
					{Name: "Milk", Price: 100, Discount: 0},
					{Name: "Bread", Price: 150, Discount: 10},
				},
				ParsedTotal: 250,
			},
			wantErr: false, // No error expected
		},
		// --- Test Case 2: Invalid JSON ---
		{
			name:        "invalid json structure",
			inputJSON:   strings.NewReader(`{"datetime": "2024-05-05T10:00:00Z", "items": "not an array"}`),
			wantReceipt: nil,  // Expecting nil receipt on error
			wantErr:     true, // Expecting an unmarshalling error
		},
		// --- Test Case 3: Empty JSON ---
		{
			name:        "empty json string",
			inputJSON:   strings.NewReader(``),
			wantReceipt: nil,
			wantErr:     true, // Expecting EOF error
		},
		// --- Test Case 4: Structurally valid but empty values ---
		{
			name:      "valid structure empty values",
			inputJSON: strings.NewReader(`{}`), // Valid JSON, but fields will be zero-valued
			wantReceipt: &Receipt{ // Expecting zero values for string/int/slice
				Timestamp:   "",
				Items:       nil, // Note: Unmarshalling into empty slice results in nil, not empty slice []Item{}
				ParsedTotal: 0,
			},
			wantErr: false,
		},
		// --- Add more test cases! ---
		//  - Missing fields
		//  - Extra fields (should usually be ignored by json.Unmarshal)
		//  - Incorrect data types (e.g., price as string)
	}

	// Loop through the test cases
	for _, tc := range testCases {
		// t.Run creates a subtest, making output easier to read
		t.Run(tc.name, func(t *testing.T) {
			// Convert input string to byte slice for the function
			jsonData := tc.inputJSON

			// Call the function under test
			gotReceipt, err := ParseReceipt(jsonData)

			// --- Assertions ---

			// 1. Check if error status matches expectation
			if (err != nil) != tc.wantErr {
				// If we got an error but didn't expect one, OR
				// if we didn't get an error but expected one
				t.Fatalf("ParseReceipt() error = %v, wantErr %v", err, tc.wantErr)
				// Use Fatalf because if error status is wrong, no point comparing receipts
			}

			// 2. If no error was expected, compare the resulting structs
			//    Use reflect.DeepEqual for comparing complex types like structs/slices/maps
			if !tc.wantErr && !reflect.DeepEqual(gotReceipt, tc.wantReceipt) {
				t.Errorf("ParseReceipt() got = %v, want %v", gotReceipt, tc.wantReceipt)
				// Use Errorf to report mismatch but allow other tests to run
			}

			// Optional: If an error IS expected, you might want to check *which* error
			// if tc.wantErr && !errors.Is(err, specificExpectedError) { // Example
			//     t.Errorf("ParseReceipt() expected error type %T, but got %T", specificExpectedError, err)
			// }
		})
	}
}
