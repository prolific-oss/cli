package study

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateOptions is the options for the creating a study command.
type CreateOptions struct {
	Args         []string
	TemplatePath string
}

// NewCreateCommand creates a new `study create` command to allow you to create
// a study
func NewCreateCommand(client client.API) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creation of studies",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Args = args

			if opts.TemplatePath == "" {
				fmt.Println("Error: Can only create via a template YAML file at the moment.")
				os.Exit(1)
			}

			err := createStudy(client, opts, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %s", strings.ReplaceAll(err.Error(), "\n", ""))
				os.Exit(1)
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a YAML file containing your studies you want to create")

	return cmd
}

func createStudy(client client.API, opts CreateOptions, w io.Writer) error {
	v := viper.New()
	v.SetConfigFile(opts.TemplatePath)
	v.ReadInConfig()

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

	fmt.Fprintf(w, "Study created: %s\n", study.Name)

	return nil
}
