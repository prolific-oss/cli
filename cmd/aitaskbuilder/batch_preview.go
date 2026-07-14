package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/pkg/browser"
	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// BrowserOpener is a function type for opening URLs in a browser.
// This allows for dependency injection in tests.
type BrowserOpener func(url string) error

// DefaultBrowserOpener uses the system browser to open URLs.
var DefaultBrowserOpener BrowserOpener = browser.OpenURL

// BatchPreviewOptions is the options for the batch preview command.
type BatchPreviewOptions struct {
	Args          []string
	BatchID       string
	BrowserOpener BrowserOpener
}

// NewBatchPreviewCommand creates a new `batch preview` command to open a batch
// preview in the browser.
func NewBatchPreviewCommand(c client.API, w io.Writer) *cobra.Command {
	return NewBatchPreviewCommandWithOpener(c, w, DefaultBrowserOpener)
}

// NewBatchPreviewCommandWithOpener creates a new `batch preview` command with a
// custom browser opener. This is useful for testing to avoid opening actual
// browser windows.
func NewBatchPreviewCommandWithOpener(c client.API, w io.Writer, browserOpener BrowserOpener) *cobra.Command {
	var opts BatchPreviewOptions
	opts.BrowserOpener = browserOpener

	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview a batch in the browser",
		Long: `Preview a batch in the browser

Opens the batch's first task group in your default web browser so you can
preview it before launch.`,
		Example: `
Preview an AI Task Builder batch in the browser:
$ prolific aitaskbuilder batch preview -b <batch_id>
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderAITaskBuilderBatchPreview(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.BatchID, "batch-id", "b", "", "Batch ID (required) - The ID of the batch to preview.")

	_ = cmd.MarkFlagRequired("batch-id")

	return cmd
}

// renderAITaskBuilderBatchPreview validates access to the batch, resolves a task
// group, builds the preview URL, prints it, and attempts to open it in a browser.
func renderAITaskBuilderBatchPreview(c client.API, opts BatchPreviewOptions, w io.Writer) error {
	if opts.BatchID == "" {
		return errors.New(ErrBatchIDRequired)
	}

	// Fetch batch to validate access
	_, err := c.GetAITaskBuilderBatch(opts.BatchID)
	if err != nil {
		return fmt.Errorf("failed to get batch: %s", err.Error())
	}

	taskGroups, err := c.GetAITaskBuilderTaskGroups(opts.BatchID)
	if err != nil {
		return fmt.Errorf("failed to get task groups: %s", err.Error())
	}
	if len(*taskGroups) == 0 {
		return fmt.Errorf("%s %s", ErrNoTaskGroupsFound, opts.BatchID)
	}

	taskGroupID := (*taskGroups)[0]

	// Build the preview URL and display it
	previewURL := GetBatchPreviewURL(opts.BatchID, taskGroupID)
	fmt.Fprintln(w, "Opening batch preview in browser...")
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
}
