package aitaskbuilder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

const (
	batchExportPollInterval = 3 * time.Second
	batchExportTimeout      = 10 * time.Minute

	batchExportStatusGenerating = "generating"
	batchExportStatusComplete   = "complete"
	batchExportStatusFailed     = "failed"
)

// batchExportPollSleep is the sleep function used between export status polls.
// Replaced in tests via SetBatchExportPollSleepForTesting to avoid real delays.
var batchExportPollSleep func(time.Duration) = time.Sleep

// batchExportDownloadClient is the HTTP client used to fetch the export ZIP.
// Replaced in tests via SetBatchExportDownloadClientForTesting to supply a TLS-aware client.
var batchExportDownloadClient = http.DefaultClient

// BatchExportOptions holds the options for the batch export command.
type BatchExportOptions struct {
	Args   []string
	Output string
}

// NewBatchExportCommand creates a new `aitaskbuilder batch export` command to
// export a batch's responses as a ZIP archive.
func NewBatchExportCommand(c client.API, w io.Writer) *cobra.Command {
	var opts BatchExportOptions

	cmd := &cobra.Command{
		Use:   "export <batch-id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Export a batch's responses as a ZIP archive",
		Long: `Export a batch's responses as a ZIP archive

This command requests an export of an AI Task Builder Batch and downloads
the resulting ZIP file. The archive contains:

  - responses-by-submission.jsonl  — one JSON record per participant submission
  - responses-by-task.jsonl        — one JSON record per task (for inter-rater comparison)
  - batch.json                     — batch metadata and instruction definitions
  - README.md                      — usage guide with example code
  - files/                         — participant-uploaded files (if any)

The export is generated asynchronously. This command will poll until the
archive is ready and then download it automatically.`,
		Example: `
Export a batch to the default filename (<batch-id>-export-<timestamp>.zip):

$ prolific aitaskbuilder batch export 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a

Export to a custom output path:

$ prolific aitaskbuilder batch export 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a --output /tmp/my-export.zip
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 1 || opts.Args[0] == "" {
				return errors.New("please provide a batch ID")
			}

			if opts.Output == "" {
				opts.Output = fmt.Sprintf("%s-export-%s.zip", opts.Args[0], time.Now().Format("20060102-150405"))
			}

			return exportBatch(c, opts, w)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path (default: <batch-id>-export-<timestamp>.zip)")

	return cmd
}

func exportBatch(c client.API, opts BatchExportOptions, w io.Writer) error {
	batchID := opts.Args[0]

	fmt.Fprintf(w, "Requesting export for batch %s...\n", batchID)

	// Step 1: POST to initiate the export job.
	initResult, err := c.InitiateBatchExport(batchID)
	if err != nil {
		return fmt.Errorf("error requesting export: %s", err.Error())
	}

	// If already complete (cached result), download immediately.
	if initResult.Status == batchExportStatusComplete {
		return batchDownloadExport(initResult.URL, opts.Output, w)
	}

	if initResult.Status != batchExportStatusGenerating {
		return fmt.Errorf("unexpected export status %q for batch %s", initResult.Status, batchID)
	}

	exportID := initResult.ExportID

	// Step 2: Poll GET until complete or failed.
	deadline := time.Now().Add(batchExportTimeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("export timed out after 10 minutes for batch %s", batchID)
		}

		fmt.Fprint(w, ".")
		batchExportPollSleep(batchExportPollInterval)

		pollResult, err := c.GetBatchExportStatus(batchID, exportID)
		if err != nil {
			return fmt.Errorf("error polling export status: %s", err.Error())
		}

		switch pollResult.Status {
		case batchExportStatusComplete:
			return batchDownloadExport(pollResult.URL, opts.Output, w)
		case batchExportStatusFailed:
			return fmt.Errorf("export generation failed for batch %s", batchID)
		case batchExportStatusGenerating:
			// continue polling
		default:
			return fmt.Errorf("unexpected export status %q for batch %s", pollResult.Status, batchID)
		}
	}
}

func batchDownloadExport(rawURL, outputPath string, w io.Writer) error {
	fmt.Fprintf(w, "\nExport ready. Downloading to %s...\n", outputPath)
	if err := batchDownloadFile(rawURL, outputPath); err != nil {
		return fmt.Errorf("error downloading export: %s", err.Error())
	}
	fmt.Fprintf(w, "Export saved to %s\n", outputPath)
	return nil
}

func batchDownloadFile(rawURL, outputPath string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid download URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("download URL must use HTTPS, got %q", parsed.Scheme)
	}

	ctx, cancel := context.WithTimeout(context.Background(), batchExportTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := batchExportDownloadClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download request failed with status %d", resp.StatusCode)
	}

	// Write to a temp file in the same directory so the final rename is
	// guaranteed to be on the same filesystem (cross-device renames fail).
	// The temp file is removed on any error to avoid leaving partial ZIPs.
	tmp, err := os.CreateTemp(filepath.Dir(outputPath), ".export-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	_, copyErr := io.Copy(tmp, resp.Body)
	closeErr := tmp.Close()

	if copyErr != nil || closeErr != nil {
		_ = os.Remove(tmpPath)
		if copyErr != nil {
			return fmt.Errorf("failed to write download: %w", copyErr)
		}
		return fmt.Errorf("failed to close temp file: %w", closeErr)
	}

	if err := os.Rename(tmpPath, outputPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to move download into place: %w", err)
	}

	return nil
}
