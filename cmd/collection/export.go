package collection

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

const (
	exportPollInterval = 3 * time.Second
	exportTimeout      = 10 * time.Minute

	exportStatusGenerating = "generating"
	exportStatusComplete   = "complete"
	exportStatusFailed     = "failed"
)

// pollSleep is the sleep function used between export status polls.
// Replaced in tests via SetPollSleepForTesting to avoid real delays.
var pollSleep func(time.Duration) = time.Sleep

// downloadClient is the HTTP client used to fetch the export ZIP.
// Replaced in tests via SetDownloadClientForTesting to supply a TLS-aware client.
var downloadClient = http.DefaultClient

// ExportOptions is the options for the collection export command.
type ExportOptions struct {
	Args   []string
	Output string
}

// NewExportCommand creates a new `collection export` command to export a
// collection's responses as a ZIP archive.
func NewExportCommand(c client.API, w io.Writer) *cobra.Command {
	var opts ExportOptions

	cmd := &cobra.Command{
		Use:   "export <collection-id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Export a collection's responses as a ZIP archive",
		Long: `Export a collection's responses as a ZIP archive

This command requests an export of an AI Task Builder Collection and downloads
the resulting ZIP file. The archive contains:

  - responses.jsonl  — one JSON record per submission
  - collection.json  — collection metadata and instruction definitions
  - README.md        — usage guide with example code
  - files/           — participant-uploaded files (if any)

The export is generated asynchronously. This command will poll until the
archive is ready and then download it automatically.`,
		Example: `
Export a collection to the default filename (<collection-id>-export.zip):

$ prolific collection export 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a

Export to a custom output path:

$ prolific collection export 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a --output /tmp/my-export.zip
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 1 || opts.Args[0] == "" {
				return errors.New("please provide a collection ID")
			}

			if opts.Output == "" {
				opts.Output = fmt.Sprintf("%s-export.zip", opts.Args[0])
			}

			return exportCollection(c, opts, w)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path (default: <collection-id>-export.zip)")

	return cmd
}

func exportCollection(c client.API, opts ExportOptions, w io.Writer) error {
	collectionID := opts.Args[0]

	fmt.Fprintf(w, "Requesting export for collection %s...\n", collectionID)

	// Step 1: POST to initiate the export job.
	initResult, err := c.InitiateCollectionExport(collectionID)
	if err != nil {
		if shared.IsFeatureNotEnabledError(err) {
			ui.RenderFeatureAccessMessage(FeatureNameAITBCollection, FeatureContactURLAITBCollection)
			return nil
		}
		return fmt.Errorf("error requesting export: %s", err.Error())
	}

	// If already complete (cached result), download immediately.
	if initResult.Status == exportStatusComplete {
		return downloadExport(initResult.URL, opts.Output, w)
	}

	if initResult.Status != exportStatusGenerating {
		return fmt.Errorf("unexpected export status %q for collection %s", initResult.Status, collectionID)
	}

	exportID := initResult.ExportID

	// Step 2: Poll GET until complete or failed.
	deadline := time.Now().Add(exportTimeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("export timed out after 10 minutes for collection %s", collectionID)
		}

		fmt.Fprint(w, ".")
		pollSleep(exportPollInterval)

		pollResult, err := c.GetCollectionExportStatus(collectionID, exportID)
		if err != nil {
			return fmt.Errorf("error polling export status: %s", err.Error())
		}

		switch pollResult.Status {
		case exportStatusComplete:
			return downloadExport(pollResult.URL, opts.Output, w)
		case exportStatusFailed:
			return fmt.Errorf("export generation failed for collection %s", collectionID)
		case exportStatusGenerating:
			// continue polling
		default:
			return fmt.Errorf("unexpected export status %q for collection %s", pollResult.Status, collectionID)
		}
	}
}

func downloadExport(url, outputPath string, w io.Writer) error {
	fmt.Fprintf(w, "\nExport ready. Downloading to %s...\n", outputPath)
	if err := downloadFile(url, outputPath); err != nil {
		return fmt.Errorf("error downloading export: %s", err.Error())
	}
	fmt.Fprintf(w, "Export saved to %s\n", outputPath)
	return nil
}

func downloadFile(rawURL, outputPath string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid download URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("download URL must use HTTPS, got %q", parsed.Scheme)
	}

	ctx, cancel := context.WithTimeout(context.Background(), exportTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := downloadClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download request failed with status %d", resp.StatusCode)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
