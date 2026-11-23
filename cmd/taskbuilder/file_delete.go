package taskbuilder

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewFileDeleteCommand creates a new command for deleting a hosted file.
func NewFileDeleteCommand(c client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <public-url>",
		Args:  cobra.ExactArgs(1),
		Short: "Delete an uploaded video",
		Long: `Delete an uploaded video using the public URL printed by "file upload".

Deleting a file that has already been deleted is not an error. The API does not say
whether the file existed, so the command reports that the file is no longer hosted.`,
		Example: `
$ prolific taskbuilder file delete https://example.com/files/<workspace_id>/<file_id>/intro.mp4
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteHostedFile(c, args[0], w)
		},
	}

	return cmd
}

// parseHostedFileURL extracts the workspace ID, file ID and file name from a public file URL.
func parseHostedFileURL(rawURL string) (workspaceID, fileID, fileName string, err error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Host == "" {
		return "", "", "", fmt.Errorf("%q is not a valid file URL", rawURL)
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "files" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", "", "", fmt.Errorf("%q is not a file URL; expected https://<host>/files/<workspace_id>/<file_id>/<file_name>", rawURL)
	}

	return parts[1], parts[2], parts[3], nil
}

func deleteHostedFile(c client.API, rawURL string, w io.Writer) error {
	if strings.TrimSpace(rawURL) == "" {
		return errors.New("please provide the public URL of the file")
	}

	workspaceID, fileID, fileName, err := parseHostedFileURL(rawURL)
	if err != nil {
		return err
	}

	if err := c.DeleteFile(workspaceID, fileID, fileName); err != nil {
		return fmt.Errorf("error deleting file: %s", err.Error())
	}

	// The API returns 204 whether or not the file existed, so do not claim it was removed by this call.
	fmt.Fprintf(w, "%s is no longer hosted (deleted now, or it was already gone).\n", fileName)

	return nil
}
