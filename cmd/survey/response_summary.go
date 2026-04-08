package survey

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// ResponseSummaryOptions is the options for the survey response summary command.
type ResponseSummaryOptions struct {
	Args []string
	Json bool
}

// NewResponseSummaryCommand creates a new command to view the response summary for a survey.
func NewResponseSummaryCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ResponseSummaryOptions

	cmd := &cobra.Command{
		Use:   commandName + " <survey_id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "View a summary of survey responses",
		Long: `View a summary of survey responses

Shows aggregated response data for each question in the survey,
including the total number of answers and the count for each answer option.
`,
		Example: `
View the response summary for a survey
$ prolific survey response summary 6261321e223a605c7a4f7678

View the response summary as JSON
$ prolific survey response summary 6261321e223a605c7a4f7678 --json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderSurveyResponseSummary(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.Json, "json", "j", false, "Output as JSON")

	return cmd
}

func renderSurveyResponseSummary(client client.API, opts ResponseSummaryOptions, w io.Writer) error {
	surveyID := opts.Args[0]

	summary, err := client.GetSurveyResponseSummary(surveyID)
	if err != nil {
		return err
	}

	if opts.Json {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(summary)
	}

	fmt.Fprint(w, renderSummaryString(*summary))

	return nil
}

func renderSummaryString(s model.SurveySummary) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintln(ui.RenderHeading("Survey Response Summary")))
	content.WriteString(fmt.Sprintf("Survey ID: %s\n", s.SurveyID))

	if len(s.Questions) == 0 {
		content.WriteString("\nNo responses recorded\n")
		return content.String()
	}

	content.WriteString(ui.RenderSectionMarker())

	for i, q := range s.Questions {
		content.WriteString(fmt.Sprintf("  Question: %s\n", q.Question))
		content.WriteString(fmt.Sprintf("  Total Answers: %d\n", q.TotalAnswers))
		for _, a := range q.Answers {
			content.WriteString(fmt.Sprintf("    - %s: %d\n", a.Answer, a.Count))
		}
		if i < len(s.Questions)-1 {
			content.WriteString("\n")
		}
	}

	return content.String()
}
