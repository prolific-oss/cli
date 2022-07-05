package study

import (
	"fmt"
	"io"
	"strings"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	studyui "github.com/benmatselby/prolificli/ui/study"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.TemplatePath == "" {
				return fmt.Errorf("Error: Can only create via a template YAML file at the moment.")
			}

			err := createStudy(client, opts, w)
			if err != nil {
				return fmt.Errorf("Error: %s\n", strings.ReplaceAll(err.Error(), "\n", ""))
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

	s := model.CreateStudy{
		Name:                    v.GetString("name"),
		InternalName:            v.GetString("internal_name"),
		Description:             v.GetString("description"),
		ExternalStudyURL:        v.GetString("external_study_url"),
		ProlificIDOption:        v.GetString("prolific_id_option"),
		CompletionCode:          v.GetString("completion_code"),
		CompletionOption:        v.GetString("completion_option"),
		TotalAvailablePlaces:    v.GetInt("total_available_places"),
		EstimatedCompletionTime: v.GetInt("estimated_completion_time"),
		MaximumAllowedTime:      v.GetInt("maximum_allowed_time"),
		Reward:                  v.GetFloat64("reward"),
		DeviceCompatibility:     v.GetStringSlice("device_compatibility"),
		PeripheralRequirements:  v.GetStringSlice("peripheral_requirements"),
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
		fmt.Fprintln(w, studyui.RenderStudy(client, *study))
	}

	return nil
}
