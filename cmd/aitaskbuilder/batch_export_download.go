package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// BatchExportDownloadOptions holds the options for the batch export
// download command.
type BatchExportDownloadOptions struct {
	Args   []string
	Output string
}

// NewBatchExportDownloadCommand creates a new `aitaskbuilder batch export
// download` command to download an existing export job's ZIP archive
// without requesting a new export.
func NewBatchExportDownloadCommand(c client.API, w io.Writer) *cobra.Command {
	var opts BatchExportDownloadOptions

	cmd := &cobra.Command{
		Use:   "download <batch-id> <export-id>",
		Args:  cobra.ExactArgs(2),
		Short: "Download an existing batch export job",
		Long: `Download an existing batch export job

This command downloads the ZIP archive for an export job that was already
requested — for example via 'batch export' or one shown by
'batch export list' — without starting a new export.

If the job is still generating, this command polls until it completes and
then downloads it automatically, just like 'batch export' does for a
newly-requested job. If the job failed, request a fresh export with
'batch export <batch-id>' instead.`,
		Example: `
Download an existing, completed export job:

$ prolific aitaskbuilder batch export download 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a export-job-uuid-456

Download to a custom output path:

$ prolific aitaskbuilder batch export download 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a export-job-uuid-456 --output /tmp/my-export.zip
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 2 || opts.Args[0] == "" || opts.Args[1] == "" {
				return errors.New("please provide a batch ID and an export ID")
			}

			if opts.Output == "" {
				opts.Output = fmt.Sprintf("%s-export-%s.zip", opts.Args[0], opts.Args[1])
			}

			return downloadBatchExport(c, opts, w)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path (default: <batch-id>-export-<export-id>.zip)")

	return cmd
}

func downloadBatchExport(c client.API, opts BatchExportDownloadOptions, w io.Writer) error {
	batchID := opts.Args[0]
	exportID := opts.Args[1]

	fmt.Fprintf(w, "Checking export %s for batch %s...\n", exportID, batchID)

	status, err := c.GetBatchExportStatus(batchID, exportID)
	if err != nil {
		return fmt.Errorf("error checking export status: %s", err.Error())
	}

	switch status.Status {
	case batchExportStatusComplete:
		return batchDownloadExport(status.URL, opts.Output, w)
	case batchExportStatusFailed:
		return fmt.Errorf("export %s for batch %s failed to generate; request a new export with 'aitaskbuilder batch export %s'", exportID, batchID, batchID)
	case batchExportStatusGenerating:
		url, err := pollBatchExportUntilDone(c, batchID, exportID, w)
		if err != nil {
			return err
		}
		return batchDownloadExport(url, opts.Output, w)
	default:
		return fmt.Errorf("unexpected export status %q for batch %s", status.Status, batchID)
	}
}
