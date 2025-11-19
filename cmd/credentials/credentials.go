package credentials

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// getCredentials extracts credentials from either a file or command line arguments.
// It handles reading from a file (via filePath), trimming trailing newlines that editors add,
// or using a string argument at the specified index.
func getCredentials(filePath string, args []string, argIndex int) (string, error) {
	if filePath != "" {
		// Read from file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("unable to read file: %w", err)
		}
		// Trim trailing whitespace (including newlines) that editors add to files
		return strings.TrimRight(string(data), "\n\r"), nil
	}

	if argIndex < len(args) {
		// Use provided argument
		return args[argIndex], nil
	}

	return "", fmt.Errorf("credentials must be provided either as an argument or via -f flag")
}

// NewCredentialsCommand creates a new `credentials` command
func NewCredentialsCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credentials",
		Short: "Manage credential pools",
		Long:  `Create and manage credential pools for studies that require authentication credentials`,
	}

	cmd.AddCommand(
		NewCreateCommand(client, w),
		NewUpdateCommand(client, w),
	)

	return cmd
}
