package googledrive

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Source implements both SourceLister and ReceiptImageReader for Google Drive.
type Source struct {
	service  *drive.Service
	targetID string // The File ID or Folder ID this source is bound to
}

func New(ctx context.Context, credentialsPath string, uri string) (*Source, error) {
	slog.Debug("Initializing Google Drive source", "uri", uri)

	// 1. Parse the URI to get the ID
	id, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	// 2. Initialize the Drive Service
	// We use the DriveReadOnlyScope because we only need to list and download.
	srv, err := drive.NewService(ctx,
		option.WithCredentialsFile(credentialsPath),
		option.WithScopes(drive.DriveReadonlyScope),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create drive client: %w", err)
	}

	return &Source{
		service:  srv,
		targetID: id,
	}, nil
}

func (s *Source) ListSources(ctx context.Context) ([]domain.SourceInfo, error) {
	slog.Debug("Listing Drive files", "folder_id", s.targetID)

	// Query: Parents contains ID AND Not Trashed
	query := fmt.Sprintf("'%s' in parents and trashed = false", s.targetID)

	call := s.service.Files.List().
		Q(query).
		Fields("files(id, name, size, modifiedTime, mimeType)").
		Context(ctx)

	fileList, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("drive list call failed: %w", err)
	}

	var sources []domain.SourceInfo
	for _, f := range fileList.Files {
		// Skip folders, we only want processable files
		if f.MimeType == "application/vnd.google-apps.folder" {
			continue
		}

		// Parse the modification time (RFC3339)
		modTime, _ := time.Parse(time.RFC3339, f.ModifiedTime)

		sources = append(sources, domain.SourceInfo{
			Name: f.Name,
			// The reference is the full URI, so it can be passed back to --input later
			Reference: fmt.Sprintf("gdrive://%s", f.Id),
			Size:      f.Size,
			ModTime:   modTime,
			// We can get the extension from the name, or map the MimeType
			Extension: filepath.Ext(f.Name),
		})
	}

	slog.Debug("Drive listing complete", "files_found", len(sources))
	return sources, nil
}

// ParseURI extracts the ID from a gdrive://ID string.
func ParseURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, "gdrive://") {
		return "", fmt.Errorf("invalid gdrive uri: %s", uri)
	}
	// Trim the prefix
	id := strings.TrimPrefix(uri, "gdrive://")
	if id == "" {
		return "", fmt.Errorf("gdrive uri is missing ID")
	}
	return id, nil
}
