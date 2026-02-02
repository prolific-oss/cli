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

// BrowserOpener is a function type for opening URLs in a browser.
// This allows for dependency injection in tests.
type BrowserOpener func(url string) error

// DefaultBrowserOpener uses the system browser to open URLs.
var DefaultBrowserOpener BrowserOpener = browser.OpenURL

// PreviewOptions is the options for the preview collection command.
type PreviewOptions struct {
	Args          []string
	BrowserOpener BrowserOpener
}

// NewPreviewCommand creates a new `collection preview` command to open a collection
// preview in the browser.
func NewPreviewCommand(c client.API, w io.Writer) *cobra.Command {
	return NewPreviewCommandWithOpener(c, w, DefaultBrowserOpener)
}

// NewPreviewCommandWithOpener creates a new `collection preview` command with a custom browser opener.
// This is useful for testing to avoid opening actual browser windows.
func NewPreviewCommandWithOpener(c client.API, w io.Writer, browserOpener BrowserOpener) *cobra.Command {
	var opts PreviewOptions
	opts.BrowserOpener = browserOpener

	cmd := &cobra.Command{
		Use: "preview <collection-id>",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || args[0] == "" {
				return errors.New("please provide a collection ID")
			}
			return nil
		},
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
			collectionID := opts.Args[0]

			// Fetch collection to validate access
			_, err := c.GetCollection(collectionID)
			if err != nil {
				if shared.IsFeatureNotEnabledError(err) {
					ui.RenderFeatureAccessMessage(FeatureNameAITBCollection, FeatureContactURLAITBCollection)
					return nil
				}
				return fmt.Errorf("failed to get collection: %s", err.Error())
			}

			// Build the preview URL and display it
			previewURL := collectionui.GetCollectionPreviewURL(collectionID)
			fmt.Fprintln(w, "Opening collection preview in browser...")
			fmt.Fprintln(w)
			fmt.Fprintln(w, previewURL)

			// Attempt to open the browser - don't fail if it doesn't work
			// (e.g., in headless/CI environments)
			if opts.BrowserOpener != nil {
				if err := opts.BrowserOpener(previewURL); err != nil {
					fmt.Fprintln(w, "(Browser did not open automatically - use the URL above)")
				}
			}

			return nil
		},
	}

	return cmd
}
