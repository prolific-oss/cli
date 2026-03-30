package collection

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	collectionui "github.com/prolific-oss/cli/ui/collection"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// defaultListFields is the default fields shown when the user has not specified --fields.
const defaultListFields = "ID,Name,ItemCount"

// ListOptions is the options for the listing collections command.
type ListOptions struct {
	Args        []string
	Fields      string
	Output      shared.OutputOptions
	WorkspaceID string
	Limit       int
	Offset      int
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

You can output as a table, useful for terminal, or into another application
$ prolific collection list -w <workspace-id> --table
$ prolific collection list -w <workspace-id> -t

You can output as CSV
$ prolific collection list -w <workspace-id> --csv
$ prolific collection list -w <workspace-id> -c

You can output as JSON
$ prolific collection list -w <workspace-id> --json
$ prolific collection list -w <workspace-id> -j

You can specify the fields you want to render in table or CSV output
$ prolific collection list -w <workspace-id> -f ID,Name,ItemCount -t
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

			collections, err := c.GetCollections(opts.WorkspaceID, opts.Limit, opts.Offset)
			if err != nil {
				if shared.IsFeatureNotEnabledError(err) {
					ui.RenderFeatureAccessMessage(FeatureNameAITBCollection, FeatureContactURLAITBCollection)
					return nil
				}
				return fmt.Errorf("error: %s", err.Error())
			}

			format := shared.ResolveFormat(opts.Output)
			fields := opts.Fields
			if fields == "" {
				fields = defaultListFields
			}
			switch format {
			case "json":
				r := ui.JSONRenderer[model.Collection]{}
				if err := r.Render(collections.Results, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "csv":
				r := ui.CsvRenderer[model.Collection]{}
				if err := r.Render(collections.Results, fields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "table":
				r := ui.TableRenderer[model.Collection]{}
				if err := r.Render(collections.Results, fields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			default:
				r := &collectionui.InteractiveRenderer{}
				if err := r.Render(c, *collections, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The workspace ID to list collections for (required).")
	flags.StringVarP(&opts.Fields, "fields", "f", "", "Comma separated list of fields you want to display in table/csv mode.")
	flags.IntVar(&opts.Limit, "limit", client.DefaultRecordLimit, "Limit the number of results returned.")
	flags.IntVar(&opts.Offset, "offset", client.DefaultRecordOffset, "Offset for pagination.")
	shared.AddOutputFlags(cmd, &opts.Output)

	return cmd
}
