package aitaskbuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

type BatchUpdateOptions struct {
	Args                    []string
	BatchID                 string
	Name                    string
	DatasetID               string
	TaskName                string
	TaskIntroduction        string
	TaskSteps               string
	TaskNameChanged         bool
	TaskIntroductionChanged bool
	TaskStepsChanged        bool
	BatchItemsFile          string
	BatchItemsJSON          string
	ClearBatchItems         bool
	EnableAutoSync          bool
	DisableAutoSync         bool
}

func NewBatchUpdateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchUpdateOptions

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a batch",
		Long: `Update an existing AI Task Builder batch

This command updates a batch's name, dataset, and/or task details. At least one
field must be provided. Task detail flags can be provided individually — any
omitted task detail fields will be preserved from the existing batch.

batch_items can be set from a file (-f) or inline JSON (-j), or cleared with
--clear-batch-items. Clearing batch_items also deletes all associated instructions
and content blocks for the batch.

--auto-sync enables automatic sync; --no-auto-sync disables it. They are mutually exclusive.`,
		Example: `
Update a batch name:
$ prolific aitaskbuilder batch update -b 497f6eca-6276-4993-bfeb-53cbbbba6f08 -n "Updated Batch Name"

Update a single task detail field:
$ prolific aitaskbuilder batch update -b 497f6eca-6276-4993-bfeb-53cbbbba6f08 --task-name "New Task Name"

Update all task details:
$ prolific aitaskbuilder batch update -b 497f6eca-6276-4993-bfeb-53cbbbba6f08 --task-name "New Task" --task-introduction "New introduction" --task-steps "1. Step one\n2. Step two"

Update name and dataset:
$ prolific aitaskbuilder batch update -b 497f6eca-6276-4993-bfeb-53cbbbba6f08 -n "Updated Name" -d 1234acb09999db4b99bcded1

Set batch_items from a file:
$ prolific aitaskbuilder batch update -b 497f6eca-6276-4993-bfeb-53cbbbba6f08 -f batch-items.json

Clear batch_items:
$ prolific aitaskbuilder batch update -b 497f6eca-6276-4993-bfeb-53cbbbba6f08 --clear-batch-items

Enable auto-sync:
$ prolific aitaskbuilder batch update -b 497f6eca-6276-4993-bfeb-53cbbbba6f08 --auto-sync

Disable auto-sync:
$ prolific aitaskbuilder batch update -b 497f6eca-6276-4993-bfeb-53cbbbba6f08 --no-auto-sync
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.TaskNameChanged = cmd.Flags().Changed("task-name")
			opts.TaskIntroductionChanged = cmd.Flags().Changed("task-introduction")
			opts.TaskStepsChanged = cmd.Flags().Changed("task-steps")

			err := updateAITaskBuilderBatch(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.BatchID, "batch-id", "b", "", "Batch ID (required) - The ID of the batch to update.")
	flags.StringVarP(&opts.Name, "name", "n", "", "Batch name - The new name for the batch.")
	flags.StringVarP(&opts.DatasetID, "dataset-id", "d", "", "Dataset ID - The ID of the new dataset to link to the batch.")
	flags.StringVar(&opts.TaskName, "task-name", "", "Task name - The new name of the task.")
	flags.StringVar(&opts.TaskIntroduction, "task-introduction", "", "Task introduction - The new introduction text for the task.")
	flags.StringVar(&opts.TaskSteps, "task-steps", "", "Task steps - The new steps for completing the task.")
	flags.StringVarP(&opts.BatchItemsFile, "batch-items-file", "f", "", "Path to JSON file containing batch_items layout.")
	flags.StringVarP(&opts.BatchItemsJSON, "batch-items-json", "j", "", "Inline JSON string containing batch_items layout.")
	flags.BoolVar(&opts.ClearBatchItems, "clear-batch-items", false, "Set batch_items to null, removing the configured task layout and deleting all associated instructions and content blocks.")
	flags.BoolVar(&opts.EnableAutoSync, "auto-sync", false, "Enable automatic synchronization of new dataset datapoints into the batch.")
	flags.BoolVar(&opts.DisableAutoSync, "no-auto-sync", false, "Disable automatic synchronization of new dataset datapoints into the batch.")

	_ = cmd.MarkFlagRequired("batch-id")

	return cmd
}

// updateAITaskBuilderBatch will update an existing AI Task Builder batch
func updateAITaskBuilderBatch(c client.API, opts BatchUpdateOptions, w io.Writer) error {
	if opts.BatchID == "" {
		return errors.New(ErrBatchIDRequired)
	}

	anyTaskDetailChanged := opts.TaskNameChanged || opts.TaskIntroductionChanged || opts.TaskStepsChanged
	allTaskDetailsChanged := opts.TaskNameChanged && opts.TaskIntroductionChanged && opts.TaskStepsChanged
	anyBatchItemsChanged := opts.BatchItemsFile != "" || opts.BatchItemsJSON != "" || opts.ClearBatchItems
	anyAutoSyncChanged := opts.EnableAutoSync || opts.DisableAutoSync

	if opts.EnableAutoSync && opts.DisableAutoSync {
		return errors.New(ErrAutoSyncFlagsMutuallyExclusive)
	}

	if opts.Name == "" && opts.DatasetID == "" && !anyTaskDetailChanged && !anyBatchItemsChanged && !anyAutoSyncChanged {
		return errors.New(ErrAtLeastOneUpdateFieldRequired)
	}

	var batchItems json.RawMessage
	if opts.ClearBatchItems {
		if opts.BatchItemsFile != "" || opts.BatchItemsJSON != "" {
			return errors.New(ErrBatchItemsMutuallyExclusive)
		}
		fmt.Fprintln(os.Stderr, "Warning: clearing batch_items will delete all associated instructions and content blocks.")
		batchItems = json.RawMessage("null")
	} else if anyBatchItemsChanged {
		var err error
		batchItems, err = parseBatchItemsInput(opts.BatchItemsFile, opts.BatchItemsJSON)
		if err != nil {
			return err
		}
	}

	params := client.UpdateBatchParams{
		BatchID:    opts.BatchID,
		Name:       opts.Name,
		DatasetID:  opts.DatasetID,
		BatchItems: batchItems,
	}

	if anyAutoSyncChanged {
		autoSyncVal := opts.EnableAutoSync
		params.AutoSync = &autoSyncVal
	}

	if anyTaskDetailChanged {
		taskDetails, err := resolveTaskDetails(c, opts, allTaskDetailsChanged)
		if err != nil {
			return err
		}
		params.TaskDetails = taskDetails
	}

	response, err := c.UpdateAITaskBuilderBatch(params)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "AI Task Builder Batch Updated Successfully:\n")
	fmt.Fprintf(w, "ID: %s\n", response.ID)
	fmt.Fprintf(w, "Name: %s\n", response.Name)
	fmt.Fprintf(w, "Status: %s\n", response.Status)
	fmt.Fprintf(w, "Auto Sync Enabled: %t\n", response.AutoSyncEnabled)
	fmt.Fprintf(w, "Total Task Count: %d\n", response.TotalTaskCount)
	fmt.Fprintf(w, "Total Instruction Count: %d\n", response.TotalInstructionCount)
	fmt.Fprintf(w, "Workspace ID: %s\n", response.WorkspaceID)
	fmt.Fprintf(w, "Created By: %s\n", response.CreatedBy)
	fmt.Fprintf(w, "Created At: %s\n", response.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Schema Version: %d\n", response.SchemaVersion)

	if len(response.Datasets) > 0 {
		fmt.Fprintf(w, "Datasets: %d\n", len(response.Datasets))
		for i, dataset := range response.Datasets {
			fmt.Fprintf(w, "  Dataset %d: %s (%d datapoints)\n", i+1, dataset.ID, dataset.TotalDatapointCount)
		}
	}

	if response.TaskDetails.TaskName != "" {
		fmt.Fprintf(w, "\nTask Details:\n")
		fmt.Fprintf(w, "  Name: %s\n", response.TaskDetails.TaskName)
		fmt.Fprintf(w, "  Introduction: %s\n", response.TaskDetails.TaskIntroduction)
		fmt.Fprintf(w, "  Steps: %s\n", response.TaskDetails.TaskSteps)
	}

	return nil
}

// resolveTaskDetails returns the task details to send in the update request.
// When all three fields are provided they are used directly; otherwise the
// existing batch is fetched and only the changed fields are overwritten.
func resolveTaskDetails(c client.API, opts BatchUpdateOptions, allChanged bool) (*client.TaskDetails, error) {
	if allChanged {
		return &client.TaskDetails{
			TaskName:         opts.TaskName,
			TaskIntroduction: opts.TaskIntroduction,
			TaskSteps:        opts.TaskSteps,
		}, nil
	}

	existing, err := c.GetAITaskBuilderBatch(opts.BatchID)
	if err != nil {
		return nil, err
	}

	taskName := existing.TaskDetails.TaskName
	taskIntroduction := existing.TaskDetails.TaskIntroduction
	taskSteps := existing.TaskDetails.TaskSteps

	if opts.TaskNameChanged {
		taskName = opts.TaskName
	}
	if opts.TaskIntroductionChanged {
		taskIntroduction = opts.TaskIntroduction
	}
	if opts.TaskStepsChanged {
		taskSteps = opts.TaskSteps
	}

	return &client.TaskDetails{
		TaskName:         taskName,
		TaskIntroduction: taskIntroduction,
		TaskSteps:        taskSteps,
	}, nil
}
