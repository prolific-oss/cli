package study

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateOptions is the options for the creating a study command.
type CreateOptions struct {
	Args         []string
	TemplatePath string
	Publish      bool
	Silent       bool
}

// NewCreateCommand creates a new `study create` command to allow you to create
// a study
func NewCreateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creation of studies",
		Long: `Create studies on the Prolific Platform.

The template's filters array contains filter_id objects with selected_values
(for choice filters) or selected_range (lower/upper bounds for range filters).
Top-level filters are combined with AND. Use filter_id "and" or "or" with
selected_filters containing child filters to express nested groups. The CLI
preserves these groups; the API validates allowed nesting and filter combinations
for your workspace. Nested groups do not support weightings.

See docs/examples/study-with-nested-filters.json for an OR of two age ranges.
The same filters array format is used by audience count, audience breakdown's
base filters, and filter-set templates.`,
		Example: `
To create studies via the CLI, you define your study as a JSON/YAML file
$ prolific study create -t /path/to/study.json
$ prolific study create -t /path/to/study.yml

You can also create and publish a study at the same time
$ prolific study create -t /path/to/study.json -p

If you are using the CLI in other tooling, you may want to silence the returned
output of the study creation, so you can use the "-s" flag.
$ prolific study create -t /path/to/study.json -p -s

An example of a JSON study file, with an ethnicity screener

{
  "name": "Study with a ethnicity screener",
  "internal_name": "Study with a ethnicity screener",
  "description": "This study will be published to the participants with the selected ethnicity",
  "external_study_url": "https://google.com",
  "prolific_id_option": "question",
  "completion_code": "COMPLE01",
  "completion_option": "code",
  "total_available_places": 10,
  "estimated_completion_time": 10,
  "maximum_allowed_time": 10,
  "reward": 400,
  "device_compatibility": ["desktop", "tablet", "mobile"],
  "peripheral_requirements": ["audio", "camera", "download", "microphone"],
  "eligibility_requirements": [
    {
      "attributes": [{ "index": 3, "value": true }],
      "query": { "id": "5950c8413e9d730001924f2a" },
      "_cls": "web.eligibility.models.SelectAnswerEligibilityRequirement"
    }
  ],
  "credential_pool_id": "64a1b2c3d4e5f6a7b8c9d0e1_12345678-1234-11e0-8000-0a1b2c3d4e5f"
}

An example of a YAML study file

---
name: My first standard sample
internal_name: Standard sample
description: This is my first standard sample study on the Prolific system.
external_study_url: https://eggs-experriment.com
# Enum: "question", "url_parameters" (Recommended), "not_required"
prolific_id_option: question
completion_code: COMPLE01
# Enum: "url", "code"
completion_option: code
total_available_places: 10
# In minutes
estimated_completion_time: 10

###
# Optional fields
###
# In minutes
maximum_allowed_time: 10
# In cents
reward: 400
# Enum: "desktop", "tablet", "mobile"
device_compatibility:
  - desktop
  - tablet
  - mobile
# Enum: "audio", "camera", "download", "microphone"
peripheral_requirements:
  - audio
  - camera
  - download
  - microphone
---`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.TemplatePath == "" {
				return fmt.Errorf("error: Can only create via a template YAML file at the moment")
			}

			err := createStudy(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a YAML file containing your studies you want to create")
	flags.BoolVarP(&opts.Publish, "publish", "p", false, "Publish the study once created.")
	flags.BoolVarP(&opts.Silent, "silent", "s", false, "Silently create the study. It will not render the study once created.")

	return cmd
}

func createStudy(client client.API, opts CreateOptions, w io.Writer) error {
	v := viper.New()
	v.SetConfigFile(opts.TemplatePath)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	var s model.CreateStudy
	err = v.Unmarshal(&s)
	if err != nil {
		return fmt.Errorf("unable to map %s to study model: %s", opts.TemplatePath, err)
	}

	study, err := client.CreateStudy(s)
	if err != nil {
		return err
	}

	if opts.Publish {
		_, err = client.TransitionStudy(study.ID, model.TransitionStudyPublish)
		if err != nil {
			return err
		}

		study, err = client.GetStudy(study.ID)
		if err != nil {
			return err
		}
	}

	if !opts.Silent {
		fmt.Fprintln(w, RenderStudy(*study))
	}

	return nil
}
