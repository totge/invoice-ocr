package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/googledrive"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/localfilelister"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/sourcefilter"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/terminalwriter"
)

// NewListCmd creates the "list" subcommand for listing source files.
func NewListCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available receipt files from a source",
		Long:  `List discovers receipt files from a local directory or Google Drive and displays them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, v)
		},
	}

	cmd.Flags().String("path", "", "path to source files")
	cmd.Flags().String("source", "local", "source system: 'local' or 'gdrive'")
	cmd.Flags().String("target", "terminal", "output target")
	cmd.Flags().Int("limit", 10, "max number of files to list (0 for unlimited)")

	cmd.MarkFlagRequired("path")

	return cmd
}

func runList(cmd *cobra.Command, v *viper.Viper) error {
	// Read flags
	path, _ := cmd.Flags().GetString("path")
	source, _ := cmd.Flags().GetString("source")
	target, _ := cmd.Flags().GetString("target")
	limit, _ := cmd.Flags().GetInt("limit")

	// Validate source
	switch source {
	case "local":
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return fmt.Errorf("input directory does not exist: %s", path)
		} else if err != nil {
			return fmt.Errorf("failed to check input path %s: %w", path, err)
		} else if !info.IsDir() {
			return fmt.Errorf("input path is not a directory: %s", path)
		}
	case "gdrive":
		// no preemptive validation
	default:
		return fmt.Errorf("invalid --source: %s (must be 'local' or 'gdrive')", source)
	}

	// Validate target
	switch target {
	case "terminal":
		// valid
	default:
		return fmt.Errorf("invalid --target: %s (must be 'terminal')", target)
	}

	if limit < 0 {
		slog.Warn("--limit should be >= 0 (0 means unlimited)", "limit", limit)
	}

	gdriveKeyPath := v.GetString("google-drive-key-path")

	slog.Debug("List command configuration",
		"path", path,
		"source", source,
		"target", target,
		"limit", limit,
	)

	ctx, cancel := context.WithTimeout(cmd.Context(), 1*time.Minute)
	defer cancel()

	// Construct dependencies

	// Source lister
	var lister app.SourceLister
	switch source {
	case "local":
		lister = localfilelister.New(path)
	case "gdrive":
		var err error
		lister, err = googledrive.New(ctx, gdriveKeyPath, path)
		if err != nil {
			return fmt.Errorf("failed to initialize Google Drive client: %w", err)
		}
	}

	// Filter
	filter := sourcefilter.New(limit)

	// Writer
	var writer app.SourceListWriter
	switch target {
	case "terminal":
		writer = terminalwriter.New()
	}

	return app.ListInputs(ctx, lister, filter, writer)
}
