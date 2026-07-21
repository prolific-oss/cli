package aitaskbuilder

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/config"
	"github.com/spf13/cobra"
)

// NewBatchesCommand creates a new `batch` command
func NewBatchesCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Manage your batches",
		Long:  "Create and manage AI Task Builder batches in your workspace",
	}

	cmd.AddCommand(
		NewBatchCreateCommand(client, w),
		NewBatchUpdateCommand(client, w),
		NewBatchExportCommand(client, w),
		NewBatchSyncCommand(client, w),
		NewBatchInstructionsCommand(client, w),
		NewBatchSetupCommand(client, w),
		NewGetBatchCommand(client, w),
		NewGetBatchStatusCommand(client, w),
		NewGetBatchesListCommand(client, w),
		NewGetResponsesCommand(client, w),
		NewBatchTasksCommand(client, w),
		NewBatchPreviewCommand(client, w),
	)

	return cmd
}

// GetBatchPreviewPath returns the URL path to a batch preview, agnostic of domain.
func GetBatchPreviewPath(batchID, taskGroupID string) string {
	return fmt.Sprintf("data-collection-tool/batches/%s/task-groups/%s?preview=true", batchID, taskGroupID)
}

// GetBatchPreviewURL returns the full URL to a batch preview using configuration.
func GetBatchPreviewURL(batchID, taskGroupID string) string {
	return fmt.Sprintf("%s/%s", config.GetApplicationURL(), GetBatchPreviewPath(batchID, taskGroupID))
}
