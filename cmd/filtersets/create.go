package filtersets

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateOptions are the options for creating a filter set.
type CreateOptions struct {
	Args         []string
	TemplatePath string
	Name         string
	Workspace    string
}

// NewCreateCommand creates a new command for creating a filter set.
func NewCreateCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Create a filter set",
		Long: `Create a filter set

Define your filter set as a JSON or YAML template file, specifying the filters
you want to apply. You can override the name and workspace ID using flags.

Filter sets allow you to save and reuse preset filter configurations across
multiple studies.`,
		Example: `
To create a filter set from a template
$ prolific filter-sets create -t /path/to/filter-set.json

To create a filter set and override the name
$ prolific filter-sets create -t /path/to/filter-set.json -N "My filter set"

To create a filter set in a specific workspace
$ prolific filter-sets create -t /path/to/filter-set.json -w 644aaabfaf6bbc363b9d47c6
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.TemplatePath == "" {
				return fmt.Errorf("error: a template file is required, use -t to specify the path")
			}

			err := createFilterSet(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a JSON/YAML file defining the filter set")
	flags.StringVarP(&opts.Name, "name", "N", "", "Override the name of the filter set")
	flags.StringVarP(&opts.Workspace, "workspace", "w", "", "Override the workspace ID for the filter set")

	return cmd
}

func createFilterSet(client client.API, opts CreateOptions, w io.Writer) error {
	v := viper.New()
	v.SetConfigFile(opts.TemplatePath)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	var fs model.CreateFilterSet
	err = v.Unmarshal(&fs)
	if err != nil {
		return fmt.Errorf("unable to map %s to filter set model: %s", opts.TemplatePath, err)
	}

	for i := range fs.Filters {
		if fs.Filters[i].SelectedRange != nil && fs.Filters[i].SelectedRange.Lower == nil && fs.Filters[i].SelectedRange.Upper == nil {
			fs.Filters[i].SelectedRange = nil
		}
	}

	if opts.Name != "" {
		fs.Name = opts.Name
	}

	if opts.Workspace != "" {
		fs.WorkspaceID = opts.Workspace
	}

	record, err := client.CreateFilterSet(fs)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Created filter set: %s (eligible participants: %d)\n", record.ID, record.EligibleParticipantCount)

	return nil
}
