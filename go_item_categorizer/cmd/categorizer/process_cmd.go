package main


// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"os"
// 	"strconv"
// 	"time"

// 	"github.com/joho/godotenv"
// 	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheet"
// 	"github.com/totge/invoice-oc/go_item_categorizer/internal/catalog"
// 	"github.com/totge/invoice-oc/go_item_categorizer/internal/categorizer"
// 	"github.com/totge/invoice-oc/go_item_categorizer/internal/llm"
// 	"github.com/totge/invoice-oc/go_item_categorizer/internal/receipt"
// )

// // Config struct to hold application configuration
// type Config struct {
// 	GeminiApiKey    string
// 	AppSheetApiKey  string
// 	AppSheetAppId   string
// 	AppSheetBaseUrl string // Optional, might have a default
// }


// // main is the entry point of the application
// func Process() {


// }

// func loadConfig() (Config, error) {
// 	allEnvVarsFound := true
// 	var c Config

// 	geminiApiKey := os.Getenv("GEMINI_API_KEY")
// 	if geminiApiKey == "" {
// 		log.Fatal("FATAL: GEMINI_API_KEY environment variable not set.")
// 		allEnvVarsFound = false
// 	}

// 	appSheetApiKey := os.Getenv("APPSHEET_API_KEY")
// 	if appSheetApiKey == "" {
// 		fmt.Println("FATAL: APPSHEET_API_KEY environment variable not set.")
// 		allEnvVarsFound = false
// 	}

// 	appSheetAppId := os.Getenv("APPSHEET_APP_ID")
// 	if appSheetAppId == "" {
// 		fmt.Println("FATAL: APPSHEET_APP_ID environment variable not set.")
// 		allEnvVarsFound = false
// 	}

// 	appSheetBaseUrl := os.Getenv("APPSHEET_BASE_URL")
// 	if appSheetBaseUrl == "" {
// 		fmt.Println("FATAL: APPSHEET_BASE_URL environment variable not set.")
// 		allEnvVarsFound = false
// 	}

// 	if !allEnvVarsFound {
// 		return c, errors.New("missing config from environment")
// 	}

// 	c.AppSheetApiKey = appSheetApiKey
// 	c.AppSheetAppId = appSheetAppId
// 	c.AppSheetBaseUrl = appSheetBaseUrl
// 	c.GeminiApiKey = geminiApiKey

// 	return c, nil
// }
