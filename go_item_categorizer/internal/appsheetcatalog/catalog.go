package appsheetcatalog

import (
	"context"
	"fmt"
	"sync"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheet"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type DataSource interface {
	ReadCategories(ctx context.Context) ([]appsheet.Category, error)
	ReadExpenses(ctx context.Context) ([]appsheet.Expense, error)
}

// Catalog builds and caches a list of product classifications from a data source.
type Catalog struct {
	source      DataSource
	productList []domain.ProductClassification
	initOnce    sync.Once
	initErr     error
}

var _ app.ProductLister = (*Catalog)(nil)

func (c *Catalog) fetchCategories(ctx context.Context) (map[string]appsheet.Category, error) {
	categories, err := c.source.ReadCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch categories for catalog build: %w", err)
	}

	categoryMap := make(map[string]appsheet.Category)

	for _, cat := range categories {
		categoryMap[cat.CategoryId] = cat
	}
	return categoryMap, nil
}

func (c *Catalog) fetchExpenses(ctx context.Context) ([]appsheet.Expense, error) {
	expenses, err := c.source.ReadExpenses(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch expenses for catalog build: %w", err)
	}

	return expenses, nil
}

func (c *Catalog) buildProductList(categoryMap map[string]appsheet.Category, expenses []appsheet.Expense) []domain.ProductClassification {
	// Define a local struct for the composite key
	type expenseKey struct {
		ProductName string
		CategoryId  string
	}

	// Expense data needs to be deduplicated
	uniqueProducts := make(map[expenseKey]domain.ProductClassification)

	// for each expense create a ProductClassification struct
	for _, expense := range expenses {
		key := expenseKey{expense.Name, expense.CategoryId}

		if _, ok := uniqueProducts[key]; !ok {
			// getting corresponding category from category map
			catInfo := categoryMap[expense.CategoryId]

			uniqueProducts[key] = domain.ProductClassification{ProductName: expense.Name, CostGroup: catInfo.CostGroup, MainCategory: catInfo.MainCategory, SubCategory: catInfo.SubCategory, CategoryID: catInfo.CategoryId}
		}
	}

	// allocate and populate the list to return
	productList := make([]domain.ProductClassification, 0, len(uniqueProducts))

	for _, v := range uniqueProducts {
		productList = append(productList, v)
	}

	return productList
}
func (c *Catalog) init(ctx context.Context) {
	categories, err := c.fetchCategories(ctx)
	if err != nil {
		c.initErr = err
	}

	expenseList, err := c.fetchExpenses(ctx)
	if err != nil {
		c.initErr = err
	}

	productList := c.buildProductList(categories, expenseList)

	c.productList = productList
}

func (c *Catalog) ListProducts(ctx context.Context) ([]domain.ProductClassification, error) {
	c.initOnce.Do(func() {
		c.init(ctx)
	})

	return c.productList, c.initErr
}

// New is the constructor for the Catalog.
// It takes any type that satisfies the DataSource interface and returns a ready-to-use Catalog.
func New(source DataSource) *Catalog {
	return &Catalog{
		source: source,
		// The other fields (initOnce, productList, initErr) are automatically
		// set to their correct zero values.
	}
}
