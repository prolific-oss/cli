package taskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

type BatchCreateOptions struct {
	Args             []string
	Name             string
	WorkspaceID      string
	DatasetID        string
	TaskName         string
	TaskIntroduction string
	TaskSteps        string
	BatchItemsFile   string
	BatchItemsJSON   string
	AutoSync         bool
	Review           ReviewCreateOptions
}

func NewBatchCreateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchCreateOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a batch",
		Long: `Create a new Task Builder batch

This command creates a batch in a workspace and links it to a dataset. The dataset must
be in READY status before you can create a batch with it. You must provide the batch name,
workspace ID, dataset ID, and task details (name, introduction, and steps).

Optionally, supply batch_items to define the participant task layout as a structured
JSON schema (pages → rows → columns → items). Provide batch_items from a file with
-f or as an inline JSON string with -j. See docs/examples/batch-items.json for an example.

Each page may also set display_position to "intro" or "outro" to render once — before
or after the repeating task loop — instead of once per datapoint. Omit the field (or
set it to "body") for the default repeating behaviour. intro/outro pages may only
contain rich_text/image content blocks, not instructions or dataset_field refs. See
docs/examples/batch-items-with-intro-outro.json for an example.

Set --auto-sync to automatically add new dataset datapoints to the batch as they
arrive; it defaults to false.

To create a review batch, which has researchers' reviewers score a sample of the work
done on another batch, give the --review-* flags instead of --dataset-id and batch items.
A review batch creates its own dataset and builds its form from the rubrics. Each rubric
is {"description": "...", "scoring_scheme": "scale_1_5" | "pass_fail_unsure"}. The sample
rate is in basis points: 1000 reviews one submission in ten, 10000 reviews every one.
Add --review-notes to give reviewers an optional notes box.`,
		Example: `
Create a batch:
$ prolific taskbuilder batch create -n "My Data Collection Batch" -w 6278acb09062db3b35bcbeb0 -d 1234acb09999db4b99bcded1 --task-name "Sample Task" --task-introduction "This is a sample task for testing" --task-steps "1. Review the data\n2. Provide your response"

Create a batch with batch_items from a file:
$ prolific taskbuilder batch create -n "My Batch" -w <workspace-id> -d <dataset-id> --task-name "Task" --task-introduction "Intro" --task-steps "Steps" -f batch-items.json

Create a batch with inline batch_items JSON:
$ prolific taskbuilder batch create -n "My Batch" -w <workspace-id> -d <dataset-id> --task-name "Task" --task-introduction "Intro" --task-steps "Steps" -j '[{"rows":[{"columns":[{"items":[{"type":"free_text","description":"Describe your observations."}]}]}]}]'

Create a batch with intro/outro pages that render once instead of once per datapoint:
$ prolific taskbuilder batch create -n "My Batch" -w <workspace-id> -d <dataset-id> --task-name "Task" --task-introduction "Intro" --task-steps "Steps" -j '[{"display_position":"intro","rows":[{"columns":[{"items":[{"type":"rich_text","content":"<p>Welcome!</p>"}]}]}]},{"rows":[{"columns":[{"items":[{"type":"free_text","description":"Describe your observations."}]}]}]},{"display_position":"outro","rows":[{"columns":[{"items":[{"type":"rich_text","content":"<p>Thank you!</p>"}]}]}]}]'

Create a review batch that samples one submission in ten of another batch:
$ prolific taskbuilder batch create -n "My Review" -w <workspace-id> --task-name "Review" --task-introduction "Intro" --task-steps "Steps" --review-source-batch-id <source-batch-id> --review-sample-rate-bps 1000 --review-notes --review-rubrics-json '[{"description":"Is the narration accurate?","scoring_scheme":"pass_fail_unsure"}]'

Create a batch with auto-sync enabled:
$ prolific taskbuilder batch create -n "My Batch" -w <workspace-id> -d <dataset-id> --task-name "Task" --task-introduction "Intro" --task-steps "Steps" --auto-sync
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createAITaskBuilderBatch(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Name, "name", "n", "", "Batch name (required) - The name of the batch to create.")
	flags.StringVarP(&opts.WorkspaceID, "workspace-id", "w", "", "Workspace ID (required) - The ID of the workspace where the batch will be created.")
	flags.StringVarP(&opts.DatasetID, "dataset-id", "d", "", "Dataset ID (required) - The ID of the dataset to use for the batch.")
	flags.StringVar(&opts.TaskName, "task-name", "", "Task name (required) - The name of the task.")
	flags.StringVar(&opts.TaskIntroduction, "task-introduction", "", "Task introduction (required) - The introduction text for the task.")
	flags.StringVar(&opts.TaskSteps, "task-steps", "", "Task steps (required) - The steps for completing the task.")
	flags.StringVarP(&opts.BatchItemsFile, "batch-items-file", "f", "", "Path to JSON file containing batch_items layout.")
	flags.StringVarP(&opts.BatchItemsJSON, "batch-items-json", "j", "", "Inline JSON string containing batch_items layout.")
	flags.BoolVar(&opts.AutoSync, "auto-sync", false, "Automatically synchronize new dataset datapoints into the batch as they are added.")

	flags.StringVar(&opts.Review.SourceBatchID, "review-source-batch-id", "", "Review: the ID of the batch whose work is reviewed. Makes this a review batch.")
	flags.IntVar(&opts.Review.SampleRateBps, "review-sample-rate-bps", 0, "Review: how much of the work to review in basis points, from 1 to 10000 (1000 is one in ten).")
	flags.StringVar(&opts.Review.RubricsFile, "review-rubrics-file", "", "Review: path to a JSON file of rubrics, [{\"description\", \"scoring_scheme\"}].")
	flags.StringVar(&opts.Review.RubricsJSON, "review-rubrics-json", "", "Review: inline JSON of rubrics, [{\"description\", \"scoring_scheme\"}].")
	flags.BoolVar(&opts.Review.Notes, "review-notes", false, "Review: give reviewers an optional notes box.")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("workspace-id")
	_ = cmd.MarkFlagRequired("task-name")
	_ = cmd.MarkFlagRequired("task-introduction")
	_ = cmd.MarkFlagRequired("task-steps")

	return cmd
}

// createAITaskBuilderBatch will create a new Task Builder batch
func createAITaskBuilderBatch(c client.API, opts BatchCreateOptions, w io.Writer) error {
	if opts.Name == "" {
		return errors.New(ErrNameRequired)
	}
	if opts.WorkspaceID == "" {
		return errors.New(ErrWorkspaceIDRequired)
	}
	if opts.TaskName == "" {
		return errors.New(ErrTaskNameRequired)
	}
	if opts.TaskIntroduction == "" {
		return errors.New(ErrTaskIntroductionRequired)
	}
	if opts.TaskSteps == "" {
		return errors.New(ErrTaskStepsRequired)
	}

	review, err := parseReviewInput(opts.Review)
	if err != nil {
		return err
	}
	if review == nil && opts.DatasetID == "" {
		return errors.New(ErrDatasetIDRequired)
	}
	if review != nil && opts.DatasetID != "" {
		return errors.New(ErrReviewDatasetNotAllowed)
	}

	batchItems, err := parseBatchItemsInput(opts.BatchItemsFile, opts.BatchItemsJSON)
	if err != nil {
		return err
	}
	if review != nil && batchItems != nil {
		return errors.New(ErrReviewBatchItemsNotAllowed)
	}

	params := client.CreateBatchParams{
		Name:             opts.Name,
		WorkspaceID:      opts.WorkspaceID,
		DatasetID:        opts.DatasetID,
		TaskName:         opts.TaskName,
		TaskIntroduction: opts.TaskIntroduction,
		TaskSteps:        opts.TaskSteps,
		BatchItems:       batchItems,
		AutoSync:         opts.AutoSync,
		Review:           review,
	}

	response, err := c.CreateAITaskBuilderBatch(params)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Task Builder Batch Created Successfully:\n")
	fmt.Fprintf(w, "ID: %s\n", response.ID)
	fmt.Fprintf(w, "Name: %s\n", response.Name)
	fmt.Fprintf(w, "Status: %s\n", response.Status)
	fmt.Fprintf(w, "Auto Sync Enabled: %t\n", response.AutoSyncEnabled)
	fmt.Fprintf(w, "Total Task Count: %d\n", response.TotalTaskCount)
	fmt.Fprintf(w, "Total Instruction Count: %d\n", response.TotalInstructionCount)
	if response.TotalTaskGroups != nil {
		fmt.Fprintf(w, "Total Task Groups: %d\n", *response.TotalTaskGroups)
	}
	fmt.Fprintf(w, "Workspace ID: %s\n", response.WorkspaceID)
	fmt.Fprintf(w, "Created By: %s\n", response.CreatedBy)
	fmt.Fprintf(w, "Created At: %s\n", response.CreatedAt)

	if len(response.Datasets) > 0 {
		fmt.Fprintf(w, "Datasets: %d\n", len(response.Datasets))
		for i, dataset := range response.Datasets {
			fmt.Fprintf(w, "  Dataset %d: %s (%d datapoints)\n", i+1, dataset.ID, dataset.TotalDatapointCount)
		}
	}

	if response.Review != nil {
		fmt.Fprintf(w, "\nReview:\n")
		fmt.Fprintf(w, "  Round: %d\n", response.Review.Round)
		fmt.Fprintf(w, "  Sample Rate: %d bps\n", response.Review.SampleRateBps)
		fmt.Fprintf(w, "  Review Dataset ID: %s\n", response.Review.ReviewDatasetID)
		fmt.Fprintf(w, "  Backfill Requested: %t\n", response.Review.BackfillRequested)
	}

	fmt.Fprintf(w, "\nTask Details:\n")
	fmt.Fprintf(w, "  Name: %s\n", response.TaskDetails.TaskName)
	fmt.Fprintf(w, "  Introduction: %s\n", response.TaskDetails.TaskIntroduction)
	fmt.Fprintf(w, "  Steps: %s\n", response.TaskDetails.TaskSteps)

	return nil
}
