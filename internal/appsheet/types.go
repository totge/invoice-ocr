package appsheet

import (
	"context"
)

// appSheetAction represents the valid actions for the AppSheet API.
type appSheetAction string
type appSheetTable string

type AppSheetClient interface {
	ReadExpenses(context.Context) ([]Expense, error)
	ReadCategories(context.Context) ([]Category, error)
	WriteExpenseStage(context.Context, []ExpenseStage) error
}

// Constants defining the supported AppSheet actions.
const (
	// actionFind instructs the API to find records.
	actionFind appSheetAction = "Find"
	// actionAdd instructs the API to add records.
	actionAdd appSheetAction = "Add"

	tableCategories   appSheetTable = "Kategóriák"
	tableExpenses     appSheetTable = "Kiadások"
	tableExpenseStage appSheetTable = "expense_stage"
)

// appSheetActionRequest defines the structure for the body of an AppSheet Action API call.
type appSheetActionRequest struct {
	Action     appSheetAction         `json:"Action"`
	Properties map[string]interface{} `json:"Properties,omitempty"`
	Rows       any                    `json:"Rows,omitempty"` // Used for Add/Edit/Delete, often empty/nil for Find
}

type Category struct {
	RowNumber    string `json:"_RowNumber,omitempty"`
	CategoryId   string `json:"Kategória ID"`
	CostGroup    string `json:"Költség csoport"`
	MainCategory string `json:"Kategória"`
	SubCategory  string `json:"Alkategória"`
}

type Expense struct {
	RowNumber     string `json:"_RowNumber,omitempty"`
	Id            string `json:"ID,omitempty"`
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
	RowNumber    string `json:"_RowNumber,omitempty"`
	Id           string `json:"ID,omitempty"`
	ReceiptId    string `json:"Vásárlás ID"`
	ExpenseDate  string `json:"Vásárlás dátum"`
	CategoryId   string `json:"Kategória ID,omitempty"`
	CostGroup    string `json:"Költség csoport"`
	MainCategory string `json:"Kategória"`
	SubCategory  string `json:"Alkategória"`
	Name         string `json:"Megnevezés"`
	Amount       int    `json:"Összeg"`
	OriginalName string `json:"Eredeti név"`
	Approved     bool   `json:"Jóváhagyva"`
}
