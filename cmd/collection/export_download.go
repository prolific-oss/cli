package collection

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// ExportDownloadOptions is the options for the collection export download
// command.
type ExportDownloadOptions struct {
	Args   []string
	Output string
}

// NewExportDownloadCommand creates a new `collection export download`
// command to download an existing export job's ZIP archive without
// requesting a new export.
func NewExportDownloadCommand(c client.API, w io.Writer) *cobra.Command {
	var opts ExportDownloadOptions

	cmd := &cobra.Command{
		Use:   "download <collection-id> <export-id>",
		Args:  cobra.ExactArgs(2),
		Short: "Download an existing collection export job",
		Long: `Download an existing collection export job

This command downloads the ZIP archive for an export job that was already
requested — for example via 'collection export' or one shown by
'collection export list' — without starting a new export.

If the job is still generating, this command polls until it completes and
then downloads it automatically, just like 'collection export' does for a
newly-requested job. If the job failed, request a fresh export with
'collection export <collection-id>' instead.`,
		Example: `
Download an existing, completed export job:

$ prolific collection export download 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a export-job-uuid-456

Download to a custom output path:

$ prolific collection export download 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a export-job-uuid-456 --output /tmp/my-export.zip
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 2 || opts.Args[0] == "" || opts.Args[1] == "" {
				return errors.New("please provide a collection ID and an export ID")
			}

			if opts.Output == "" {
				opts.Output = fmt.Sprintf("%s-export-%s.zip", opts.Args[0], opts.Args[1])
			}

			return downloadCollectionExport(c, opts, w)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path (default: <collection-id>-export-<export-id>.zip)")

	return cmd
}

func downloadCollectionExport(c client.API, opts ExportDownloadOptions, w io.Writer) error {
	collectionID := opts.Args[0]
	exportID := opts.Args[1]

	fmt.Fprintf(w, "Checking export %s for collection %s...\n", exportID, collectionID)

	status, err := c.GetCollectionExportStatus(collectionID, exportID)
	if err != nil {
		if shared.IsFeatureNotEnabledError(err) {
			ui.RenderFeatureAccessMessage(FeatureNameAITBCollection, FeatureContactURLAITBCollection)
			return nil
		}
		return fmt.Errorf("error checking export status: %s", err.Error())
	}

	switch status.Status {
	case exportStatusComplete:
		return downloadExport(status.URL, opts.Output, w)
	case exportStatusFailed:
		return fmt.Errorf("export %s for collection %s failed to generate; request a new export with 'collection export %s'", exportID, collectionID, collectionID)
	case exportStatusGenerating:
		url, err := pollCollectionExportUntilDone(c, collectionID, exportID, w)
		if err != nil {
			return err
		}
		return downloadExport(url, opts.Output, w)
	default:
		return fmt.Errorf("unexpected export status %q for collection %s", status.Status, collectionID)
	}
}
