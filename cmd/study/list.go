package study

import (
	"fmt"
	"io"
	"strings"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	"github.com/benmatselby/prolificli/ui/study"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing studies command.
type ListOptions struct {
	Args           []string
	Csv            bool
	Fields         string
	NonInteractive bool
	ProjectID      string
	Status         string
}

// NewListCommand creates a new `study list` command to give you details about
// your studies.
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "List all of your studies",
		Long: `List your studies

This command allows you to understand what is happening with your studies on the
Prolific Platform.`,
		Example: `
You can list all your studies in an interactive manner. This is a searchable
interface. When you have found the study you want to look into in more detail,
press enter.
$ prolific study list

You can provide a non-iterative experience, if you want to get details in the
terminal, or into another application
$ prolific study list -n

You can filter the studies by the project they are assigned to
$ prolific study list -p 6261321e223a605c7a4f7561

You can filter the studies by their status, for example your active studies
$ prolific study list -s active

You can render the results as a CSV format
$ prolific study list -c

You can specify the fields you want to render in either the non-iterative or CSV
view
$ prolific study list -f ID,InternalName,TotalCost -c

The fields you can use are
- ID
- Name
- InternalName
- DateCreated
- TotalAvailablePlaces
- Reward
- CanAutoReview
- Desc
- EstimatedCompletionTime
- MaximumAllowedTime
- CompletionURL
- ExternalStudyURL
- PublishedAt
- StartedPublishingAt
- AwardPoints
- PresentmentCurrencyCode
- Status
- AverageRewardPerHour
- DeviceCompatibility
- PeripheralRequirements
- PlacesTaken
- EstimatedRewardPerHour
- Ref
- StudyType
- TotalCost
- PublishAt
- IsPilot
- IsUnderpaying`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			renderer := study.ListRenderer{}

			if opts.Csv {
				renderer.SetStrategy(&study.CsvRenderer{})
			} else if opts.NonInteractive {
				renderer.SetStrategy(&study.NonInteractiveRenderer{})
			} else {
				renderer.SetStrategy(&study.InteractiveRenderer{})
			}

			err := renderer.Render(client, study.ListUsedOptions{
				Status: opts.Status, NonInteractive: opts.NonInteractive, Fields: opts.Fields, ProjectID: opts.ProjectID,
			}, w)

			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Status, "status", "s", model.StatusAll, fmt.Sprintf("The status you want to filter on: %s.", strings.Join(model.StudyListStatus, ", ")))
	flags.BoolVarP(&opts.NonInteractive, "non-interactive", "n", false, "Render the list details straight to the terminal.")
	flags.BoolVarP(&opts.Csv, "csv", "c", false, "Render the list details in a CSV format.")
	flags.StringVarP(&opts.Fields, "fields", "f", "", "Comma separated list of fields you want to display in non-interactive/csv mode.")
	flags.StringVarP(&opts.ProjectID, "project", "p", "", "Get studies for a given project ID.")

	return cmd
}
