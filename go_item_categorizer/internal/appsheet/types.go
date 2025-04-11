package appsheet

import (
	"encoding/json"
	"fmt"
	"io"
)

// AppSheetAction represents the valid actions for the AppSheet API.
type AppSheetAction string
type AppSheetTable string

// Constants defining the supported AppSheet actions.
const (
	// ActionFind instructs the API to find records.
	ActionFind AppSheetAction = "Find"
	// ActionAdd instructs the API to add records.
	ActionAdd AppSheetAction = "Add"

	TableCategories AppSheetTable = "Kategóriák"
	TableExpenses   AppSheetTable = "Kiadások"
)

// AppSheetActionRequest defines the structure for the body of an AppSheet Action API call.
type AppSheetActionRequest struct {
	Action     AppSheetAction           `json:"Action"`
	Properties map[string]interface{}   `json:"Properties,omitempty"`
	Rows       []map[string]interface{} `json:"Rows,omitempty"` // Used for Add/Edit/Delete, often empty/nil for Find
}

type AppSheetTableData[T any] struct {
	Data []T
}

func (tableData *AppSheetTableData[T]) ReadTable(responseBody io.Reader) error {
	decoder := json.NewDecoder(responseBody)
	if err := decoder.Decode(&tableData.Data); err != nil {
		return fmt.Errorf("failed to decode response body for table read: %w", err)
	}

	return nil
}

type TableReader interface {
	ReadTable(io.Reader) error
}

type Category struct {
	RowNumber    string `json:"_RowNumber"`      // Matches "_RowNumber" key
	CategoryId   string `json:"Kategória ID"`    // Matches "Kategória ID" key
	CostGroup    string `json:"Költség csoport"` // Matches "Költség csoport" key
	MainCategory string `json:"Kategória"`       // Matches "Kategória" key
	SubCategory  string `json:"Alkategória"`     // Matches "Alkategória" key
}

// func (c *Category) ProcessBody(apiResponse []byte) error {
// 	json.Unmarshal(apiResponse, )
// }

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
