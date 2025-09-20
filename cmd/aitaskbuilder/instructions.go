package aitaskbuilder

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewInstructionsCommand creates a new `instructions` command under aitaskbuilder
func NewInstructionsCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instructions",
		Short: "AI Task Builder instructions operations",
		Long:  "Manage AI Task Builder instructions - create instructions for batches",
	}

	cmd.AddCommand(
		NewInstructionsCreateCommand(client, w),
	)

	return cmd
}
