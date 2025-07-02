package filtersets

import (
	"errors"
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

// ViewOptions is the options for the detail view of a filter set.
type ViewOptions struct {
	Args []string
	Web  bool
}

// NewViewCommand creates a new command to show a filter set.
func NewViewCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ViewOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Args:  cobra.MinimumNArgs(1),
		Short: "Provide details about your filter set",
		Long: `View your filter set

A detailed view of a specific filter set.
`,
		Example: `
View the details of a specific filter set

$ prolific filter-sets view 64efb93c7788944088864cec
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderProject(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&opts.Web, "web", "W", false, "Open the filter set in the web application")

	return cmd
}

// renderProject will show your project
func renderProject(client client.API, opts ViewOptions, w io.Writer) error {
	if len(opts.Args) < 1 || opts.Args[0] == "" {
		return errors.New("please provide a filter set ID")
	}

	filterSet, err := client.GetFilterSet(opts.Args[0])
	if err != nil {
		return err
	}

	if opts.Web {
		return browser.OpenURL(GetFilterSetURL(filterSet.WorkspaceID, opts.Args[0]))
	}

	content := fmt.Sprintln(ui.RenderHeading(filterSet.Name))

	content += "\n"
	content += fmt.Sprintf("Organisation:               %v\n", filterSet.OrganisationID)
	content += fmt.Sprintf("Workspace:                  %v\n", filterSet.WorkspaceID)
	content += fmt.Sprintf("Version:                    %v\n", filterSet.Version)
	content += fmt.Sprintf("Eligible participant count: %v\n", filterSet.EligibleParticipantCount)
	content += fmt.Sprintf("Locked:                     %v\n", filterSet.IsLocked)
	content += fmt.Sprintf("Deleted:                    %v\n", filterSet.IsDeleted)

	content += ui.RenderSectionMarker()

	filterLength := len(filterSet.Filters)
	for _, filter := range filterSet.Filters {
		filterLength--
		content += fmt.Sprintf("Filter ID: %v", filter.FilterID)

		if len(filter.SelectedValues) > 0 {
			content += "\nSelected values:"
			for _, value := range filter.SelectedValues {
				content += fmt.Sprintf("\n- %v", value)
			}
		}

		if filter.SelectedRange.Lower != nil || filter.SelectedRange.Upper != nil {
			content += "\nSelected range:"

			if filter.SelectedRange.Upper != nil {
				content += fmt.Sprintf("\n- Upper: %v", filter.SelectedRange.Upper)
			}

			if filter.SelectedRange.Lower != nil {
				content += fmt.Sprintf("\n- Lower: %v", filter.SelectedRange.Lower)
			}
		}

		if filterLength > 0 {
			content += "\n\n"
		}
	}

	if len(filterSet.Filters) == 0 {
		content += "No filters"
	}

	fmt.Fprintln(w, content)
	fmt.Fprintln(w, ui.RenderApplicationLink("filter set", GetFilterSetPath(filterSet.WorkspaceID, filterSet.ID)))

	return nil
}
