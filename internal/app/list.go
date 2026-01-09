package app

import (
	"context"
	"fmt"
)

func ListInputs(ctx context.Context, lister SourceLister, filter SourceFilter, writer SourceListWriter) error {
	allSources, err := lister.ListSources(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sources: %w", err)
	}

	filteredSources := filter.Filter(allSources)

	err = writer.WriteSourceList(ctx, filteredSources)
	if err != nil {
		return fmt.Errorf("failed to write source list: %w", err)
	}
	
	return nil
}
