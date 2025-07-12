package appsheet

import (
	"context"
)

// AppSheetAction represents the valid actions for the AppSheet API.
type AppSheetAction string
type AppSheetTable string

type AppSheetClient interface {
	ReadExpenses(context.Context) ([]Expense, error)
	ReadCategories(context.Context) ([]Category, error)
	WriteExpenseStage(context.Context, []ExpenseStage) error
}

// Constants defining the supported AppSheet actions.
const (
	// ActionFind instructs the API to find records.
	ActionFind AppSheetAction = "Find"
	// ActionAdd instructs the API to add records.
	ActionAdd AppSheetAction = "Add"

	TableCategories   AppSheetTable = "Kategóriák"
	TableExpenses     AppSheetTable = "Kiadások"
	TableExpenseStage AppSheetTable = "expense_stage"
)

// // AppSheetActionRequest defines the structure for the body of an AppSheet Action API call.
// type AppSheetActionRequest[T any] struct {
// 	Action     AppSheetAction         `json:"Action"`
// 	Properties map[string]interface{} `json:"Properties,omitempty"`
// 	Rows       []T                    `json:"Rows,omitempty"` // Used for Add/Edit/Delete, often empty/nil for Find
// }

// AppSheetActionRequest defines the structure for the body of an AppSheet Action API call.
type AppSheetActionRequest struct {
	Action     AppSheetAction         `json:"Action"`
	Properties map[string]interface{} `json:"Properties,omitempty"`
	Rows       any                    `json:"Rows,omitempty"` // Used for Add/Edit/Delete, often empty/nil for Find
}

type Category struct {
	RowNumber    string `json:"_RowNumber"`      // Matches "_RowNumber" key
	CategoryId   string `json:"Kategória ID"`    // Matches "Kategória ID" key
	CostGroup    string `json:"Költség csoport"` // Matches "Költség csoport" key
	MainCategory string `json:"Kategória"`       // Matches "Kategória" key
	SubCategory  string `json:"Alkategória"`     // Matches "Alkategória" key
}

type Expense struct {
	RowNumber     string `json:"_RowNumber"`
	Id            string `json:"ID"`
	EntryDate     string `json:"Rögzítés dátum"`
	SameDayEntry  string `json:"Mai rögzítés"`
	ExpenseDate   string `json:"Kiadás dátum"`
	IsPlanned     string `json:"Tervezett"`
	Regularity    string `json:"Rendszeresség"`
	CategoryId    string `json:"Kategória ID"`
	CostGroup     string `json:"Költség csoport"`
	MainCategory  string `json:"Kategória"`
	SubCategory   string `json:"Alkategória"`
	Name          string `json:"Megnevezés"`
	Amount        string `json:"Összeg"`
	PaymentMethod string `json:"Fizetési mód"`
}

type ExpenseStage struct {
	RowNumber    string `json:"_RowNumber"`
	Id           string `json:"ID"`
	ReceiptId    string `json:"Vásárlás ID"`
	ExpenseDate  string `json:"Vásárlás dátum"`
	CategoryId   string `json:"Kategória ID"`
	CostGroup    string `json:"Költség csoport"`
	MainCategory string `json:"Kategória"`
	SubCategory  string `json:"Alkategória"`
	Name         string `json:"Megnevezés"`
	Amount       int    `json:"Összeg"`
	OriginalName string `json:"Eredeti név"`
	Approved     bool   `json:"Jóváhagyva"`
}
