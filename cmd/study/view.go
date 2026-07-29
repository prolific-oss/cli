package study

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/browser"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/config"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"

	"github.com/spf13/cobra"
)

// ViewOptions is the options for the detail view of a project.
type ViewOptions struct {
	Args []string
	Web  bool
}

// NewViewCommand creates a new `study view` command to give you details about
// your studies.
func NewViewCommand(client client.API, w io.Writer) *cobra.Command {
	var opts ViewOptions

	cmd := &cobra.Command{
		Use:   "view",
		Short: "Provide details about your study, requires a Study ID",
		Long:  `View study details`,
		Example: `
To get details about a study
$ prolific study view 64395e9c2332b8a59a65d51e`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.Web {
				return browser.OpenURL(GetStudyURL(opts.Args[0]))
			}

			study, err := client.GetStudy(args[0])
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			fmt.Fprintln(w, RenderStudy(*study))

			return nil
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&opts.Web, "web", "W", false, "Open the study in the web application")

	return cmd
}

// RenderStudy will produce a detailed view of the selected study.
func RenderStudy(study model.Study) string {
	underpaying := ""
	if study.IsUnderpaying == true {
		underpaying = " " + ui.RenderHighlightedText("(Underpaying)")
	}

	content := fmt.Sprintln(ui.RenderHeading(study.Name))
	content += fmt.Sprintf("%s\n\n", study.Desc)
	content += fmt.Sprintf("ID:                        %s\n", study.ID)
	content += fmt.Sprintf("Status:                    %s\n", study.Status)
	content += fmt.Sprintf("Type:                      %s\n", study.StudyType)
	content += fmt.Sprintf("Total cost:                %s\n", ui.RenderMoney((study.TotalCost/100), study.GetCurrencyCode()))
	content += fmt.Sprintf("Reward:                    %s%s\n", ui.RenderMoney((study.Reward/100), study.GetCurrencyCode()), underpaying)
	content += fmt.Sprintf("Hourly rate:               %s\n", ui.RenderMoney((study.AverageRewardPerHour/100), study.GetCurrencyCode()))
	content += fmt.Sprintf("Estimated completion time: %d\n", study.EstimatedCompletionTime)
	content += fmt.Sprintf("Maximum allowed time:      %d\n", study.MaximumAllowedTime)
	content += fmt.Sprintf("Study URL:                 %s\n", study.ExternalStudyURL)
	content += fmt.Sprintf("Places taken:              %d\n", study.PlacesTaken)
	content += fmt.Sprintf("Available places:          %d\n", study.TotalAvailablePlaces)
	if study.CredentialPoolID != "" {
		content += fmt.Sprintf("Credential Pool ID:        %s\n", study.CredentialPoolID)
	}

	content += ui.RenderSectionMarker()

	content += fmt.Sprintln(ui.RenderHeading("Submissions configuration"))

	content += fmt.Sprintf("Max submissions per participant: %v\n", study.SubmissionsConfig.MaxSubmissionsPerParticipant)
	content += fmt.Sprintf("Max concurrent submissions:      %v\n", study.SubmissionsConfig.MaxConcurrentSubmissions)

	content += ui.RenderSectionMarker()

	content += fmt.Sprintln(ui.RenderHeading("Filters"))

	filterCount := 0
	var filterContent strings.Builder
	for _, filter := range study.Filters {
		filterContent.WriteString(fmt.Sprintf("\n%s\n", filter.FilterID))

		for _, value := range filter.SelectedValues {
			filterContent.WriteString(fmt.Sprintf("- %s\n", value))
		}
		filterCount++
	}

	if filterCount == 0 {
		content += fmt.Sprintln("No filters are defined for this study.")
	} else {
		content += filterContent.String()
	}

	content += ui.RenderApplicationLink("study", GetStudyPath(study.ID))

	return content
}

// GetStudyPath returns the URL path to a study, agnostic of domain
func GetStudyPath(ID string) string {
	return fmt.Sprintf("researcher/studies/%s", ID)
}

// GetStudyURL returns the full URL to a study using configuration
func GetStudyURL(ID string) string {
	return fmt.Sprintf("%s/%s", config.GetApplicationURL(), GetStudyPath(ID))
}
