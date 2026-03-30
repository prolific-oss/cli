package template

import (
	"io"

	"github.com/spf13/cobra"
)

// NewTemplateCommand creates the parent `template` command with subcommands.
func NewTemplateCommand(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Browse and retrieve study and collection templates",
		Long: `Templates are bundled example files that demonstrate how to create studies
and collections via the CLI. Use these as starting points for your own
configurations.`,
	}

	cmd.AddCommand(
		NewListCommand(w),
		NewViewCommand(w),
	)

	return cmd
}
