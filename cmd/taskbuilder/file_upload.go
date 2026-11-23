package taskbuilder

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

const supportedHostedFileExtensions = ".mp4, .webm"

var validHostedFileExtensions = map[string]bool{
	".mp4":  true,
	".webm": true,
}

// FileUploadOptions are the options for uploading a hosted file.
type FileUploadOptions struct {
	WorkspaceID string
	FilePath    string
}

// NewFileUploadCommand creates a new command for uploading a hosted file.
func NewFileUploadCommand(client client.API, w io.Writer) *cobra.Command {
	var opts FileUploadOptions

	cmd := &cobra.Command{
		Use:           "upload",
		Short:         "Upload a video and print its public URL",
		SilenceErrors: true,
		Long: fmt.Sprintf(`Upload a video file and print the public URL it is served from.

Supported file types: %s.

The URL can be used in custom HTML, for example as the source of a video element.
Use "prolific taskbuilder file delete" to remove the file again.`, supportedHostedFileExtensions),
		Example: `
Upload a video to a workspace:
$ prolific taskbuilder file upload -w <workspace_id> -f ./intro.mp4
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := uploadHostedFile(client, opts, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace-id", "w", "", "The ID of the workspace to upload to (required)")
	flags.StringVarP(&opts.FilePath, "file", "f", "", "The path to the video file to upload (required)")

	_ = cmd.MarkFlagRequired("workspace-id")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func uploadHostedFile(client client.API, opts FileUploadOptions, w io.Writer) error {
	if opts.WorkspaceID == "" {
		return errors.New("workspace ID is required")
	}

	if opts.FilePath == "" {
		return errors.New(ErrFilePathRequired)
	}

	fileName := filepath.Base(opts.FilePath)
	if !validHostedFileExtensions[strings.ToLower(filepath.Ext(fileName))] {
		return fmt.Errorf("unsupported file type %q; supported types are %s", filepath.Ext(fileName), supportedHostedFileExtensions)
	}

	info, err := os.Stat(opts.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", opts.FilePath)
		}

		return fmt.Errorf("failed to inspect file %s: %w", opts.FilePath, err)
	}

	if info.IsDir() {
		return fmt.Errorf("file path is a directory: %s", opts.FilePath)
	}

	if info.Size() == 0 {
		return fmt.Errorf("file is empty: %s", opts.FilePath)
	}

	response, err := client.GetFileUploadURL(opts.WorkspaceID, fileName)
	if err != nil {
		return fmt.Errorf("failed to get upload URL: %w", err)
	}

	if response.PublicURL == "" {
		return errors.New("public_url is missing in response")
	}

	if err := uploadFileToPresignedURL(opts.FilePath, response.UploadURL, "PUT", response.ContentType); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	fmt.Fprintln(w, response.PublicURL)

	return nil
}
