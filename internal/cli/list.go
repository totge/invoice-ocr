package cli

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

// ListOptions holds all inputs for the list command.
type ListOptions struct {
	Source        string
	Path          string
	Target        string
	Limit         int
	GDriveKeyPath string
}

// Validate checks that all options are valid.
func (o *ListOptions) Validate() error {
	switch o.Source {
	case "local":
		info, err := os.Stat(o.Path)
		if os.IsNotExist(err) {
			return fmt.Errorf("input directory does not exist: %s", o.Path)
		} else if err != nil {
			return fmt.Errorf("failed to check input path %s: %w", o.Path, err)
		} else if !info.IsDir() {
			return fmt.Errorf("input path is not a directory: %s", o.Path)
		}
	case "gdrive":
		// no preemptive validation
	default:
		return fmt.Errorf("invalid --source: %s (must be 'local' or 'gdrive')", o.Source)
	}

	switch o.Target {
	case "terminal":
		// valid
	default:
		return fmt.Errorf("invalid --target: %s (must be 'terminal')", o.Target)
	}

	if o.Limit < 0 {
		return fmt.Errorf("--limit must be >= 0 (0 means unlimited), got %d", o.Limit)
	}

	return nil
}

func newListOptions(cmd *cobra.Command, v *viper.Viper) ListOptions {
	source, _ := cmd.Flags().GetString("source")
	path, _ := cmd.Flags().GetString("path")
	target, _ := cmd.Flags().GetString("target")
	limit, _ := cmd.Flags().GetInt("limit")

	return ListOptions{
		Source:        source,
		Path:          path,
		Target:        target,
		Limit:         limit,
		GDriveKeyPath: v.GetString("google-drive-key-path"),
	}
}

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

	cmd.Flags().StringP("source", "s", "local", "input source: 'local' or 'gdrive'")
	cmd.Flags().StringP("path", "p", "", "path to source directory")
	cmd.Flags().StringP("target", "t", "terminal", "output target: 'terminal'")
	cmd.Flags().IntP("limit", "l", 10, "max number of files to list (0 for unlimited)")

	cmd.MarkFlagRequired("path")

	return cmd
}

func runList(cmd *cobra.Command, v *viper.Viper) error {
	opts := newListOptions(cmd, v)
	if err := opts.Validate(); err != nil {
		return err
	}

	slog.Debug("List command configuration",
		"path", opts.Path,
		"source", opts.Source,
		"target", opts.Target,
		"limit", opts.Limit,
	)

	ctx, cancel := context.WithTimeout(cmd.Context(), 1*time.Minute)
	defer cancel()

	// Source lister
	var lister app.SourceLister
	switch opts.Source {
	case "local":
		lister = localfilelister.New(opts.Path)
	case "gdrive":
		var err error
		lister, err = googledrive.New(ctx, opts.GDriveKeyPath, opts.Path)
		if err != nil {
			return fmt.Errorf("failed to initialize Google Drive client: %w", err)
		}
	}

	filter := sourcefilter.New(opts.Limit)

	var writer app.SourceListWriter
	switch opts.Target {
	case "terminal":
		writer = terminalwriter.New()
	}

	return app.ListInputs(ctx, lister, filter, writer)
}
