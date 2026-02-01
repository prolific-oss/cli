package collection

import (
	"errors"
	"fmt"
	"io"

	"github.com/pkg/browser"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/ui"
	collectionui "github.com/prolific-oss/cli/ui/collection"
	"github.com/spf13/cobra"
)

// PreviewOptions is the options for the preview collection command.
type PreviewOptions struct {
	Args []string
}

// NewPreviewCommand creates a new `collection preview` command to open a collection
// preview in the browser.
func NewPreviewCommand(c client.API, w io.Writer) *cobra.Command {
	var opts PreviewOptions

	cmd := &cobra.Command{
		Use:   "preview <collection-id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Preview a collection in the browser",
		Long: `Preview a collection in the browser

Opens the collection in your default web browser so you can preview
it before publishing.`,
		Example: `
Preview a collection in the browser

$ prolific collection preview 123456789
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 1 || opts.Args[0] == "" {
				return errors.New("please provide a collection ID")
			}

			collectionID := opts.Args[0]

			// Fetch collection to validate access
			_, err := c.GetCollection(collectionID)
			if err != nil {
				if shared.IsFeatureNotEnabledError(err) {
					ui.RenderFeatureAccessMessage(FeatureNameAITBCollection, FeatureContactURLAITBCollection)
					return nil
				}
				return fmt.Errorf("error: %s", err.Error())
			}

			// Open the collection preview in the browser
			previewURL := collectionui.GetCollectionPreviewURL(collectionID)
			if err := browser.OpenURL(previewURL); err != nil {
				return fmt.Errorf("failed to open browser: %s", err.Error())
			}

			fmt.Fprintln(w, "Opening collection preview in browser...")
			fmt.Fprintln(w)
			fmt.Fprintln(w, previewURL)

			return nil
		},
	}

	return cmd
}
