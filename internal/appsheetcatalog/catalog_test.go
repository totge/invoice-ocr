package appsheetcatalog

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheet"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

// mockDataSource is a test implementation of the DataSource interface.
type mockDataSource struct {
	// Fields to control the mock's behavior
	CategoriesToReturn []appsheet.Category
	ExpensesToReturn   []appsheet.Expense
	ErrToReturn        error

	// Fields to track how many times methods were called
	ReadCategoriesCalls int
	ReadExpensesCalls   int
}

// ReadCategories implements the DataSource interface for the mock.
func (m *mockDataSource) ReadCategories(ctx context.Context) ([]appsheet.Category, error) {
	m.ReadCategoriesCalls++ // Increment the call counter
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	return m.CategoriesToReturn, nil
}

// ReadExpenses implements the DataSource interface for the mock.
func (m *mockDataSource) ReadExpenses(ctx context.Context) ([]appsheet.Expense, error) {
	m.ReadExpensesCalls++ // Increment the call counter
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	return m.ExpensesToReturn, nil
}

func TestBuildProductList(t *testing.T) {
	// --- Setup ---
	// Dummy catalog instance needed to call the method. Its fields don't matter here.
	c := &Catalog{}

	categories := map[string]appsheet.Category{
		"cat-1": {CategoryId: "cat-1", CostGroup: "Groceries", MainCategory: "Dairy"},
		"cat-2": {CategoryId: "cat-2", CostGroup: "Hardware", MainCategory: "Tools"},
	}

	expenses := []appsheet.Expense{
		{Name: "Milk", CategoryId: "cat-1"},
		{Name: "Milk", CategoryId: "cat-1"}, // Duplicate to test deduplication
		{Name: "Hammer", CategoryId: "cat-2"},
		{Name: "Screwdriver", CategoryId: "cat-2"},
		// {Name: "  Milk  ", CategoryId: "cat-1"}, // Test trimming of whitespace
		// {Name: "Unknown", CategoryId: "cat-3"},  // Test item with no matching category
	}

	// --- Act ---
	productList := c.buildProductList(categories, expenses)

	// --- Assert ---
	if len(productList) != 3 {
		t.Fatalf("Expected 3 unique products, but got %d", len(productList))
	}

	// To make the test robust, we'll check for the presence of items
	// instead of relying on a specific order.
	expectedProducts := map[string]domain.ProductClassification{
		"Milk":        {ProductName: "Milk", CategoryID: "cat-1", CostGroup: "Groceries"},
		"Hammer":      {ProductName: "Hammer", CategoryID: "cat-2", CostGroup: "Hardware"},
		"Screwdriver": {ProductName: "Screwdriver", CategoryID: "cat-2", CostGroup: "Hardware"},
	}

	for _, p := range productList {
		expected, ok := expectedProducts[p.ProductName]
		if !ok {
			t.Errorf("Got unexpected product in list: %s", p.ProductName)
			continue
		}
		// Basic check, you can use reflect.DeepEqual for a full struct check
		if p.CostGroup != expected.CostGroup {
			t.Errorf("For product %s, expected cost group %s, got %s", p.ProductName, expected.CostGroup, p.CostGroup)
		}
		delete(expectedProducts, p.ProductName) // Mark as found
	}

	if len(expectedProducts) > 0 {
		t.Errorf("Did not find all expected products. Missing: %v", reflect.ValueOf(expectedProducts).MapKeys())
	}
}

func TestCatalog_ListProducts(t *testing.T) {
	ctx := context.Background()

	// --- Sub-test 1: Happy Path and Caching ---
	t.Run("success and caching", func(t *testing.T) {
		// Setup the mock with successful data
		mock := &mockDataSource{
			CategoriesToReturn: []appsheet.Category{{CategoryId: "cat-1", CostGroup: "TestGroup"}},
			ExpensesToReturn:   []appsheet.Expense{{Name: "TestItem", CategoryId: "cat-1"}},
		}
		catalog := New(mock)

		// First call - should hit the data source
		products, err := catalog.ListProducts(ctx)

		// Assertions for first call
		if err != nil {
			t.Fatalf("First call to ListProducts failed: %v", err)
		}
		if len(products) != 1 {
			t.Fatalf("Expected 1 product, got %d", len(products))
		}
		if mock.ReadCategoriesCalls != 1 || mock.ReadExpensesCalls != 1 {
			t.Errorf("Expected data source to be called once, got Cats:%d, Exps:%d", mock.ReadCategoriesCalls, mock.ReadExpensesCalls)
		}

		// Second call - should serve from cache
		products2, err2 := catalog.ListProducts(ctx)

		// Assertions for second call
		if err2 != nil {
			t.Fatalf("Second call to ListProducts failed: %v", err2)
		}
		if !reflect.DeepEqual(products, products2) {
			t.Error("Data from second call does not match first call")
		}
		// CRITICAL: Call counts should NOT have increased
		if mock.ReadCategoriesCalls != 1 || mock.ReadExpensesCalls != 1 {
			t.Errorf("Expected data source NOT to be called again, but got Cats:%d, Exps:%d", mock.ReadCategoriesCalls, mock.ReadExpensesCalls)
		}
	})

	// --- Sub-test 2: Data Source Returns an Error ---
	t.Run("data source error", func(t *testing.T) {
		// Setup the mock to return an error
		expectedErr := errors.New("API is down")
		mock := &mockDataSource{
			ErrToReturn: expectedErr,
		}
		catalog := New(mock)

		// First call - should fail and cache the error
		_, err := catalog.ListProducts(ctx)

		// Assertions for first call
		if err == nil {
			t.Fatal("Expected an error from ListProducts, but got nil")
		}
		// Check if the original error is preserved in the chain
		if !errors.Is(err, expectedErr) {
			t.Fatalf("Expected error to wrap '%v', but it didn't. Got: %v", expectedErr, err)
		}

		// Second call - should return the cached error without calling the source again
		mock.ReadCategoriesCalls = 0 // Reset counter to be sure
		_, err2 := catalog.ListProducts(ctx)

		if mock.ReadCategoriesCalls != 0 {
			t.Error("Expected data source not to be called on second attempt after a permanent error")
		}
		if !errors.Is(err2, expectedErr) {
			t.Fatalf("Expected second call to return the same cached error, but it didn't. Got: %v", err2)
		}
	})
}
