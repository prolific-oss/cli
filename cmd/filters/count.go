package filters

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CountOptions are the options for previewing eligible participant counts.
type CountOptions struct {
	TemplatePath   string
	WorkspaceID    string
	OrganisationID string
}

// NewCountCommand creates a new command for counting eligible participants without saving a filter set.
func NewCountCommand(apiClient client.API, w io.Writer) *cobra.Command {
	var opts CountOptions

	cmd := &cobra.Command{
		Use:   "count",
		Short: "Preview eligible participant count for a filter payload",
		Long: `Preview the number of eligible participants for a set of filters.

This does not persist a filter set. Provide a JSON or YAML template that
contains a filters array and optional workspace or organisation IDs.`,
		Example: `
Preview eligibility count from a template file
$ prolific filters count -t /path/to/filters.json

Override workspace for count accuracy
$ prolific filters count -t /path/to/filters.json -w 644aaabfaf6bbc363b9d47c6
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.TemplatePath == "" {
				return fmt.Errorf("error: a template file is required, use -t to specify the path")
			}

			err := countEligibleParticipants(apiClient, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a JSON/YAML file defining filters")
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Override the workspace ID for count accuracy")
	flags.StringVarP(&opts.OrganisationID, "organisation", "o", "", "Override the organisation ID")

	return cmd
}

func countEligibleParticipants(apiClient client.API, opts CountOptions, w io.Writer) error {
	v := viper.New()
	v.SetConfigFile(opts.TemplatePath)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	var payload client.EligibilityCountPayload
	err = v.Unmarshal(&payload)
	if err != nil {
		return fmt.Errorf("unable to map %s to eligibility count payload: %s", opts.TemplatePath, err)
	}

	if opts.WorkspaceID != "" {
		payload.WorkspaceID = opts.WorkspaceID
	}

	if opts.OrganisationID != "" {
		payload.OrganisationID = opts.OrganisationID
	}

	count, err := apiClient.GetEligibleCount(payload)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Eligible participants: %d\n", count.Count)

	return nil
}
