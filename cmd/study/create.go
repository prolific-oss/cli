package study

import (
	"fmt"
	"io"
	"log"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	studyui "github.com/prolific-oss/cli/ui/study"
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
		Long:  `Create studies on the Prolific Platform`,
		Example: `
To create studies via the CLI, you define your study as a JSON/YAML file
$ prolific study create -t /path/to/study.json
$ prolific study create -t /path/to/study.yml

You can also create and publish a study at the same time
$ prolific study create -t /path/to/study.json -p

If you are using the CLI in other tooling, you may want to silence the returned
output of the study creation, so you can use the "-s" flag.
$ prolific study create -t /path/to/study.json -p -s

An example of a JSON study file, with completion codes and filters

{
  "name": "Study with completion codes",
  "internal_name": "Study with completion codes",
  "description": "This study uses the new completion_codes array format",
  "external_study_url": "https://google.com",
  "prolific_id_option": "url_parameters",
  "completion_codes": [
    {
      "code": "C1234567",
      "code_type": "COMPLETED",
      "actions": [{"action": "AUTOMATICALLY_APPROVE"}]
    },
    {
      "code": "C7654321",
      "code_type": "REJECTED",
      "actions": [{"action": "AUTOMATICALLY_REJECT"}]
    }
  ],
  "total_available_places": 10,
  "estimated_completion_time": 10,
  "maximum_allowed_time": 10,
  "reward": 400,
  "device_compatibility": ["desktop", "tablet", "mobile"],
  "peripheral_requirements": ["audio", "camera", "download", "microphone"],
  "credential_pool_id": "64a1b2c3d4e5f6a7b8c9d0e1_12345678-1234-11e0-8000-0a1b2c3d4e5f"
}

Note: The old completion_code and completion_option fields are DEPRECATED.
Use completion_codes array instead for new studies.

An example of a YAML study file

---
name: My first standard sample
internal_name: Standard sample
description: This is my first standard sample study on the Prolific system.
external_study_url: https://eggs-experriment.com
# Enum: "question", "url_parameters" (Recommended), "not_required"
prolific_id_option: url_parameters
# New completion_codes array format (recommended)
completion_codes:
  - code: C1234567
    code_type: COMPLETED
    actions:
      - action: AUTOMATICALLY_APPROVE
  - code: C7654321
    code_type: REJECTED
    actions:
      - action: AUTOMATICALLY_REJECT
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
# For taskflow studies with multiple URLs
# access_details:
#   - external_url: https://example.com/task1
#     total_allocation: 50
#   - external_url: https://example.com/task2
#     total_allocation: 50
# Use predefined filter sets
# filter_set_id: filter-set-123
# filter_set_version: 1
# Content warnings
# content_warnings:
#   - VIOLENCE
#   - EXPLICIT_LANGUAGE
# content_warning_details: May contain violent imagery
# Custom metadata
# metadata:
#   project_id: proj-123
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
		log.Fatalf("unable to map %s to study model: %s", opts.TemplatePath, err)
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
		fmt.Fprintln(w, studyui.RenderStudy(*study))
	}

	return nil
}
