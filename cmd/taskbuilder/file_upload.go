package taskbuilder

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
)

const supportedHostedFileExtensions = ".mp4, .webm"

// hostedFileMaxBytes is the largest object S3 accepts in a single PUT (5 GiB).
const hostedFileMaxBytes int64 = 5 << 30

const hostedFileScanNotice = "Note: the URL returns 403 until the file has been virus scanned, which can take a minute or two.\n"

// validHostedFileExtensions is deliberately narrower than validVideoURLFileExtensions: the file
// endpoint does not accept .mov.
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
The file is virus scanned after upload, so the URL returns 403 for a minute or two before it
serves the video. The limit is 5 GiB per file.
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

			// stderr, so that stdout stays the bare URL for scripts.
			fmt.Fprint(cmd.ErrOrStderr(), hostedFileScanNotice)

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

	if info.Size() > hostedFileMaxBytes {
		return fmt.Errorf("file is %d bytes, the limit is %d bytes (5 GiB)", info.Size(), hostedFileMaxBytes)
	}

	response, err := client.GetFileUploadURL(opts.WorkspaceID, fileName)
	if err != nil {
		if shared.IsFileHostingNotEnabledError(err) {
			return errors.New(shared.ErrFileHostingNotEnabled)
		}

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
