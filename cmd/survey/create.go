package survey

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateOptions are the options for creating a survey.
type CreateOptions struct {
	Args         []string
	TemplatePath string
	Title        string
}

// NewCreateCommand creates a new command for creating a survey.
func NewCreateCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Create a survey",
		Long: `Create a survey

Define your survey as a JSON or YAML template file, specifying the questions
and optional sections. You can override the title using a flag.

Surveys accept either sections containing questions, or a flat list of questions.`,
		Example: `
To create a survey from a template
$ prolific survey create -t /path/to/survey.json

To create a survey and override the title
$ prolific survey create -t /path/to/survey.json --title "My Survey"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.TemplatePath == "" {
				return fmt.Errorf("error: a template file is required, use -t to specify the path")
			}

			err := createSurvey(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a JSON/YAML file defining the survey")
	flags.StringVar(&opts.Title, "title", "", "Override the title of the survey")

	return cmd
}

func createSurvey(c client.API, opts CreateOptions, w io.Writer) error {
	v := viper.New()
	v.SetConfigFile(opts.TemplatePath)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	var s model.CreateSurvey
	err = v.Unmarshal(&s)
	if err != nil {
		return fmt.Errorf("unable to map %s to survey model: %s", opts.TemplatePath, err)
	}

	if opts.Title != "" {
		s.Title = opts.Title
	}

	if s.ResearcherID == "" {
		me, err := c.GetMe()
		if err != nil {
			return err
		}
		s.ResearcherID = me.ID
	}

	record, err := c.CreateSurvey(s)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Created survey: %s\n", record.ID)

	return nil
}
