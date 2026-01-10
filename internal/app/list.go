package app

import (
	"context"
	"fmt"
	"log/slog"
)

func ListInputs(ctx context.Context, lister SourceLister, filter SourceFilter, writer SourceListWriter) error {
	slog.Info("Starting file listing")

	allSources, err := lister.ListSources(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sources: %w", err)
	}
	slog.Debug("Raw sources found", "count", len(allSources))

	filteredSources := filter.Filter(allSources)
	slog.Info("Filtered source list",
		"original_count", len(allSources),
		"final_count", len(filteredSources),
	)

	slog.Info("Outputting source list")
	err = writer.WriteSourceList(ctx, filteredSources)
	if err != nil {
		return fmt.Errorf("failed to write source list: %w", err)
	}

	return nil
}
