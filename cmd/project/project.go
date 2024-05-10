package project

import (
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/config"
	"github.com/spf13/cobra"
)

// NewProjectCommand creates a new `project` command
func NewProjectCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage and view your projects in a workspace",
		Long: `Manage your projects
Projects are a way to organise studies in a workspace.
`,
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
		NewCreateCommand("create", client, w),
		NewViewCommand("view", client, w),
	)
	return cmd
}

// GetProjectPath returns the URL path to a project, agnostic of domain
func GetProjectPath(ID string) string {
	return fmt.Sprintf("researcher/workspaces/projects/%s/", ID)
}

// GetProjectURL returns the full URL to a project using configuration
func GetProjectURL(ID string) string {
	return fmt.Sprintf("%s/%s", config.GetApplicationURL(), GetProjectPath(ID))
}
