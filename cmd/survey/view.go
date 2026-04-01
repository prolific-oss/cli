package survey

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// ViewOptions is the options for the detail view of a survey.
type ViewOptions struct {
	Args []string
}

// NewViewCommand creates a new command to show a survey.
func NewViewCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ViewOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Args:  cobra.MinimumNArgs(1),
		Short: "Provide details about your survey",
		Long: `View your survey

A detailed view of how your survey is configured.
`,
		Example: `
View the details of a specific survey

$ prolific survey view 6261321e223a605c7a4f7678
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderSurvey(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

// renderSurvey will show your survey
func renderSurvey(client client.API, opts ViewOptions, w io.Writer) error {
	if len(opts.Args) < 1 || opts.Args[0] == "" {
		return errors.New("please provide a survey ID")
	}

	survey, err := client.GetSurvey(opts.Args[0])
	if err != nil {
		return err
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintln(ui.RenderHeading(survey.Title)))

	content.WriteString(fmt.Sprintf("ID:            %v\n", survey.ID))
	content.WriteString(fmt.Sprintf("Researcher:    %v\n", survey.ResearcherID))
	content.WriteString(fmt.Sprintf("Date Created:  %v\n", survey.DateCreated.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Date Modified: %v\n", survey.DateModified.Format("2006-01-02 15:04:05")))

	if len(survey.Sections) > 0 {
		content.WriteString(ui.RenderSectionMarker())
		for i, section := range survey.Sections {
			content.WriteString(fmt.Sprintf("Section: %s\n", section.Title))
			renderQuestions(&content, section.Questions)
			if i < len(survey.Sections)-1 {
				content.WriteString("\n")
			}
		}
	}

	if len(survey.Questions) > 0 {
		content.WriteString(ui.RenderSectionMarker())
		renderQuestions(&content, survey.Questions)
	}

	if len(survey.Sections) == 0 && len(survey.Questions) == 0 {
		content.WriteString("\nNo questions defined\n")
	}

	fmt.Fprint(w, content.String())

	return nil
}

func renderQuestions(content *strings.Builder, questions []model.SurveyQuestion) {
	for _, q := range questions {
		content.WriteString(fmt.Sprintf("  Question: %s (%s)\n", q.Title, q.Type))
		for _, a := range q.Answers {
			content.WriteString(fmt.Sprintf("    - %s\n", a.Value))
		}
	}
}
