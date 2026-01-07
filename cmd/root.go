package cmd

import (
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/cmd/campaign"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/cmd/credentials"
	"github.com/prolific-oss/cli/cmd/filters"
	"github.com/prolific-oss/cli/cmd/filtersets"
	"github.com/prolific-oss/cli/cmd/hook"
	"github.com/prolific-oss/cli/cmd/message"
	"github.com/prolific-oss/cli/cmd/participantgroup"
	"github.com/prolific-oss/cli/cmd/project"
	requirement "github.com/prolific-oss/cli/cmd/requirements"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/cmd/submission"
	"github.com/prolific-oss/cli/cmd/user"
	"github.com/prolific-oss/cli/cmd/workspace"
	"github.com/prolific-oss/cli/ui"
	"github.com/prolific-oss/cli/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// ApplicationName is the name of the cli binary
const ApplicationName = "prolific"

// BannerFilePath is the path to the ASCII art banner file
const BannerFilePath = "banner.txt"

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once
func Execute() {
	// Register custom template functions before building commands
	registerTemplateFuncs()

	// We need the configuration loaded before we create a NewCli
	// as that needs the viper configuration up and running
	initConfig()

	// Build the root command
	cmd := NewRootCommand()

	// Execute the application
	if err := cmd.Execute(); err != nil {
		ui.WriteError(err.Error())
		os.Exit(1)
	}
}

// NewRootCommand builds the main cli application and
// adds all the child commands
func NewRootCommand() *cobra.Command {
	// Load banner from file
	banner, err := os.ReadFile(BannerFilePath)
	bannerText := ""
	if err == nil {
		bannerText = string(banner)
	}

	var cmd = &cobra.Command{
		Use:     ApplicationName,
		Short:   "CLI application for retrieving data from the Prolific Platform",
		Long:    ui.RenderBanner(bannerText) + "\nCLI application for retrieving data from the Prolific Platform",
		Version: version.GITCOMMIT,
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.config/prolific-oss/%s.yaml)", ApplicationName))

	client := client.New()

	w := os.Stdout

	cmd.AddCommand(
		aitaskbuilder.NewAITaskBuilderCommand(&client, w),
		campaign.NewListCommand("campaign", &client, w),
		collection.NewCollectionCommand(&client, w),
		credentials.NewCredentialsCommand(&client, w),
		filters.NewListCommand(&client, w),
		filtersets.NewFilterSetCommand(&client, w),
		hook.NewHookCommand(&client, w),
		message.NewMessageCommand(&client, w),
		participantgroup.NewParticipantCommand(&client, w),
		project.NewProjectCommand(&client, w),
		requirement.NewListCommand(&client, w),
		study.NewListCommand("studies", &client, w),
		study.NewStudyCommand(&client, w),
		submission.NewSubmissionCommand(&client, w),
		user.NewMeCommand(&client, w),
		workspace.NewWorkspaceCommand(&client, w),
	)

	// Apply custom templates to all commands recursively (including root)
	applyTemplateToAllCommands(cmd)

	return cmd
}

// applyTemplateToAllCommands recursively applies custom help templates to all commands
func applyTemplateToAllCommands(cmd *cobra.Command) {
	cmd.SetHelpTemplate(getHelpTemplate())
	cmd.SetUsageTemplate(getUsageTemplate())

	for _, subCmd := range cmd.Commands() {
		applyTemplateToAllCommands(subCmd)
	}
}

// getHelpTemplate returns a custom help template with colors
func getHelpTemplate() string {
	return `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
}

// getUsageTemplate returns a custom usage template with colors
func getUsageTemplate() string {
	return `{{bold "Usage:"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{bold "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{bold "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{bold "Available Commands:"}}{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{highlight (rpad .Name .NamePadding)}}  {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{highlight (rpad .Name .NamePadding)}}  {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{highlight (rpad .Name .NamePadding)}}  {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{bold "Flags:"}}
{{.LocalFlags.FlagUsagesWrapped 100 | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{bold "Global Flags:"}}
{{.InheritedFlags.FlagUsagesWrapped 100 | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{highlight (rpad .CommandPath .CommandPathPadding)}}  {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

{{dim (print "Use \"" .CommandPath " [command] --help\" for more information about a command.")}}{{end}}
`
}

// registerTemplateFuncs registers custom template functions for cobra commands
func registerTemplateFuncs() {
	cobra.AddTemplateFunc("trimTrailingWhitespaces", func(s string) string {
		return strings.TrimRight(s, " \t")
	})
	cobra.AddTemplateFunc("bold", ui.Bold)
	cobra.AddTemplateFunc("highlight", ui.Highlight)
	cobra.AddTemplateFunc("dim", ui.Dim)
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
			ui.WriteError(err.Error())
			os.Exit(1)
		}

		viper.AddConfigPath(strings.Join([]string{home, ".config/prolific-oss"}, "/"))
		viper.SetConfigName(ApplicationName)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
}
