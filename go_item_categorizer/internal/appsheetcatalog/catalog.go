package appsheetcatalog

import (
	"context"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheet"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type Catalog struct {
	Client      *appsheet.Client
	productList []domain.ProductClassification
}

func (c *Catalog) fetchCategories(ctx context.Context) (map[string]appsheet.Category, error) {
	categories, err := appsheet.ReadRecords[appsheet.Category](c.Client, ctx, appsheet.TableCategories)
	if err != nil {
		// TODO: add context to the error
		return nil, err
	}

	categoryMap := make(map[string]appsheet.Category)

	for _, cat := range categories {
		categoryMap[cat.CategoryId] = cat
	}
	return categoryMap, nil
}

// TODO: should it return all the expenses with all the expense related data? Is the naming ok?
func (c *Catalog) fetchExpenses(ctx context.Context) ([]appsheet.Expense, error) {
	expenses, err := appsheet.ReadRecords[appsheet.Expense](c.Client, ctx, appsheet.TableExpenses)
	if err != nil {
		// TODO: add context to the error
		return nil, err
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

func (c *Catalog) ListProducts() ([]domain.ProductClassification, error) {

	// if product list is cached, just return it
	if c.productList != nil {
		return c.productList, nil
	}

	// create product list, when cache is empty
	ctx := context.Background()

	categories, err := c.fetchCategories(ctx)
	if err != nil {
		// TODO: add context to the error
		return nil, err
	}

	expenseList, err := c.fetchExpenses(ctx)
	if err != nil {
		// TODO: add context to the error
		return nil, err
	}

	productList := c.buildProductList(categories, expenseList)

	return productList, nil
}
