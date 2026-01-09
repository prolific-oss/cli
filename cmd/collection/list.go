package collection

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/ui"
	"github.com/prolific-oss/cli/ui/collection"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing collections command.
type ListOptions struct {
	Args           []string
	Csv            bool
	Json           bool
	Fields         string
	NonInteractive bool
	WorkspaceID    string
	Limit          int
	Offset         int
}

// NewListCommand creates a new `collection list` command to give you details about
// your collections.
func NewListCommand(c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all collections in a workspace",
		Long: `List collections in a workspace

This command allows you to see all collections within a workspace on the
Prolific Platform.`,
		Example: `
You can list all collections in an interactive manner. This is a searchable
interface. When you have found the collection you want to look into in more
detail, press enter.
$ prolific collection list -w <workspace-id>

You can provide a non-interactive experience, if you want to get details in the
terminal, or into another application
$ prolific collection list -w <workspace-id> -n

You can render the results as JSON for machine-readable output
$ prolific collection list -w <workspace-id> --json

You can render the results as a CSV format
$ prolific collection list -w <workspace-id> -c

You can specify the fields you want to render in either the non-interactive or CSV
view
$ prolific collection list -w <workspace-id> -f ID,Name,ItemCount -c

The fields you can use are:
- ID
- Name
- CreatedAt
- CreatedBy
- ItemCount`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.WorkspaceID == "" {
				return fmt.Errorf("workspace ID is required")
			}

			renderer := collection.ListRenderer{}

			if opts.Json {
				renderer.SetStrategy(&collection.JSONRenderer{})
			} else if opts.Csv {
				renderer.SetStrategy(&collection.CsvRenderer{})
			} else if opts.NonInteractive {
				renderer.SetStrategy(&collection.NonInteractiveRenderer{})
			} else {
				renderer.SetStrategy(&collection.InteractiveRenderer{})
			}

			err := renderer.Render(c, collection.ListUsedOptions{
				WorkspaceID: opts.WorkspaceID,
				Fields:      opts.Fields,
				Limit:       opts.Limit,
				Offset:      opts.Offset,
			}, w)

			if err != nil {
				if shared.IsFeatureNotEnabledError(err) {
					ui.RenderFeatureAccessMessage(FeatureNameAITBCollection, FeatureContactEmailAITBCollection)
					return nil
				}
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", "", "The workspace ID to list collections for (required).")
	flags.BoolVarP(&opts.NonInteractive, "non-interactive", "n", false, "Render the list details straight to the terminal.")
	flags.BoolVar(&opts.Json, "json", false, "Render the list details in JSON format for machine-readable output.")
	flags.BoolVarP(&opts.Csv, "csv", "c", false, "Render the list details in a CSV format.")
	flags.StringVarP(&opts.Fields, "fields", "f", "", "Comma separated list of fields you want to display in non-interactive/csv mode.")
	flags.IntVar(&opts.Limit, "limit", client.DefaultRecordLimit, "Limit the number of results returned.")
	flags.IntVar(&opts.Offset, "offset", client.DefaultRecordOffset, "Offset for pagination.")

	return cmd
}
