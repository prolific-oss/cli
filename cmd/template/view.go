package template

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/docs/examples"
	"github.com/spf13/cobra"
)

// NewViewCommand creates the `template <id>` subcommand.
func NewViewCommand(w io.Writer) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "view <template-id>",
		Short: "Display the contents of a template",
		Long:  `Output the raw JSON or YAML content of a template by its ID.`,
		Example: `
View a template (defaults to JSON when both formats exist)
$ prolific template view standard-sample

View the YAML variant
$ prolific template view standard-sample --format yaml

Use the output to create a study
$ prolific template view standard-sample > my-study.json
$ prolific study create -t my-study.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			templates := listTemplates()

			var match *Template
			for i, t := range templates {
				if t.ID != id {
					continue
				}
				if format != "" && t.Format == format {
					match = &templates[i]
					break
				}
				if match == nil {
					match = &templates[i]
				}
			}

			if match == nil {
				return fmt.Errorf("template %q not found, run 'prolific template list' to see available templates", id)
			}

			data, err := examples.FS.ReadFile(match.Filename)
			if err != nil {
				return fmt.Errorf("reading template %s: %w", id, err)
			}
			fmt.Fprint(w, string(data))
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&format, "format", "f", "", "Preferred format when both exist (json or yaml)")

	return cmd
}
