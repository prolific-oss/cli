package cmd

import (
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/prolific-oss/prolificli/client"
	"github.com/prolific-oss/prolificli/cmd/hook"
	"github.com/prolific-oss/prolificli/cmd/participantgroup"
	"github.com/prolific-oss/prolificli/cmd/project"
	requirement "github.com/prolific-oss/prolificli/cmd/requirements"
	"github.com/prolific-oss/prolificli/cmd/study"
	"github.com/prolific-oss/prolificli/cmd/submission"
	"github.com/prolific-oss/prolificli/cmd/user"
	"github.com/prolific-oss/prolificli/cmd/workspace"
	"github.com/prolific-oss/prolificli/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// ApplicationName is the name of the cli binary
const ApplicationName = "prolificli"

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once
func Execute() {
	// We need the configuration loaded before we create a NewCli
	// as that needs the viper configration up and running
	initConfig()

	// Build the root command
	cmd := NewRootCommand()

	// Execute the application
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// NewRootCommand builds the main cli application and
// adds all the child commands
func NewRootCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     ApplicationName,
		Short:   "CLI application for retrieving data from Prolific",
		Version: version.GITCOMMIT,
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.config/prolific/%s.yaml)", ApplicationName))

	client := client.New()

	w := os.Stdout

	cmd.AddCommand(
		user.NewMeCommand(&client, w),
		study.NewStudyCommand(&client, w),
		study.NewListCommand("studies", &client, w),
		submission.NewSubmissionCommand(&client, w),
		requirement.NewListCommand(&client, w),
		hook.NewHookCommand(&client, w),
		workspace.NewWorkspaceCommand(&client, w),
		project.NewProjectCommand(&client, w),
		participantgroup.NewParticipantCommand(&client, w),
	)

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(strings.Join([]string{home, ".config/prolific"}, "/"))
		viper.SetConfigName(ApplicationName)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
}
