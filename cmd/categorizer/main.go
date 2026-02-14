package main

import (
	"fmt"
	"os"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
