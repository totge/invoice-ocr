package main

import (
	"fmt"
	"os"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/cli"
)

func main() {

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1] // The first argument after the program name is the command

	// os.Args[2:] will be the arguments for the specific command
	args := os.Args[2:]

	var err error
	switch command {
	case "process":
		err = cli.RunProcessCommand(args)
	case "extract":
		err = cli.RunExtractCommand(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}

func printUsage() {
	fmt.Println("Receip processor usage:")
	fmt.Println("Usage: categorizer <command> [flags]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("\tprocess\tProcess an input json file and save the result to a file or AppSheet")
	fmt.Println("\textract\tExtract data from receipt image in astructured format and save the result to a file")

	fmt.Println("\thelp\tShow this help message")
}