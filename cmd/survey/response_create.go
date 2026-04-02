package survey

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ResponseCreateOptions are the options for creating a survey response.
type ResponseCreateOptions struct {
	Args         []string
	TemplatePath string
}

// NewResponseCreateCommand creates a new command for creating a survey response.
func NewResponseCreateCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ResponseCreateOptions

	cmd := &cobra.Command{
		Use:   commandName + " <survey_id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Create a survey response",
		Long: `Create a survey response

Define your response as a JSON or YAML template file, specifying the participant_id,
submission_id, and answers to questions (either in sections or as a flat list).`,
		Example: `
To create a survey response from a template
$ prolific survey response create 6261321e223a605c7a4f7678 -t /path/to/response.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.TemplatePath == "" {
				return fmt.Errorf("error: a template file is required, use -t to specify the path")
			}

			err := createSurveyResponse(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a JSON/YAML file defining the survey response")

	return cmd
}

func createSurveyResponse(c client.API, opts ResponseCreateOptions, w io.Writer) error {
	v := viper.New()
	v.SetConfigFile(opts.TemplatePath)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	var r model.CreateSurveyResponseRequest
	err = v.Unmarshal(&r)
	if err != nil {
		return fmt.Errorf("unable to map %s to survey response model: %s", opts.TemplatePath, err)
	}

	surveyID := opts.Args[0]

	record, err := c.CreateSurveyResponse(surveyID, r)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Created survey response: %s\n", record.ID)

	return nil
}
