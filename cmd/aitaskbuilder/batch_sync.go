package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

const (
	batchSyncPollInterval   = 3 * time.Second
	batchSyncDefaultTimeout = 10 * time.Minute
	// batchSyncMaxPollErrors is how many consecutive polling errors are tolerated
	// before giving up — polling should survive transient network/API blips.
	batchSyncMaxPollErrors = 3

	batchSyncStatusQueued     = "queued"
	batchSyncStatusProcessing = "processing"
	batchSyncStatusComplete   = "complete"
	batchSyncStatusFailed     = "failed"
)

// batchSyncPollSleep is the sleep function used between sync status polls.
// Replaced in tests via SetBatchSyncPollSleepForTesting to avoid real delays.
var batchSyncPollSleep func(time.Duration) = time.Sleep

// BatchSyncOptions holds the options for the batch sync command.
type BatchSyncOptions struct {
	Args    []string
	Timeout time.Duration
}

// NewBatchSyncCommand creates a new `aitaskbuilder batch sync` command to extend
// a batch with tasks for datapoints appended to its dataset since the last sync.
func NewBatchSyncCommand(c client.API, w io.Writer) *cobra.Command {
	var opts BatchSyncOptions

	cmd := &cobra.Command{
		Use:   "sync <batch-id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Sync a batch with datapoints appended to its dataset",
		Long: `Sync a batch with datapoints appended to its dataset

Starts an asynchronous sync job that extends an already set-up batch with
tasks created from datapoints added to its attached dataset since setup or
the last sync. This command polls until the job reaches a terminal state and
then reports how much work was materialised.`,
		Example: `
Sync a batch and wait for completion:

$ prolific aitaskbuilder batch sync 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a

Sync with a shorter timeout:

$ prolific aitaskbuilder batch sync 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a --timeout 2m
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 1 || opts.Args[0] == "" {
				return errors.New(ErrBatchIDRequired)
			}

			return syncBatch(c, opts, w)
		},
	}

	cmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", batchSyncDefaultTimeout,
		"Maximum time to wait for the sync to complete")

	return cmd
}

func syncBatch(c client.API, opts BatchSyncOptions, w io.Writer) error {
	batchID := opts.Args[0]

	fmt.Fprintf(w, "Starting sync for batch %s...\n", batchID)

	// Step 1: POST to start the sync job.
	initResult, err := c.SyncAITaskBuilderBatch(batchID)
	if err != nil {
		return fmt.Errorf("error: %s", err)
	}

	// A sync could in principle come back already terminal; handle both here.
	switch initResult.Status {
	case batchSyncStatusComplete:
		return reportSyncComplete(initResult, batchID, w)
	case batchSyncStatusFailed:
		return fmt.Errorf("error: sync failed for batch %s: %s", batchID, syncFailureReason(initResult))
	case batchSyncStatusQueued, batchSyncStatusProcessing:
		// fall through to polling
	default:
		return fmt.Errorf("error: unexpected sync status %q for batch %s", initResult.Status, batchID)
	}

	syncID := initResult.SyncID
	if syncID == "" {
		return fmt.Errorf("error: sync started but no sync ID was returned for batch %s", batchID)
	}

	// Step 2: Poll GET until complete or failed, tolerating transient errors.
	deadline := time.Now().Add(opts.Timeout)
	consecutiveErrors := 0

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("error: sync timed out after %s for batch %s", opts.Timeout, batchID)
		}

		fmt.Fprint(w, ".")
		// Cap the sleep at the time remaining so --timeout bounds the total wait rather than being
		// overshot by a whole poll interval.
		sleep := batchSyncPollInterval
		if remaining := time.Until(deadline); remaining < sleep {
			sleep = remaining
		}
		batchSyncPollSleep(sleep)

		pollResult, err := c.GetAITaskBuilderBatchSyncStatus(batchID, syncID)
		if err != nil {
			consecutiveErrors++
			if consecutiveErrors >= batchSyncMaxPollErrors {
				return fmt.Errorf("error: %s", err)
			}
			// Transient blip — keep polling until the deadline or the error cap.
			continue
		}
		consecutiveErrors = 0

		switch pollResult.Status {
		case batchSyncStatusComplete:
			return reportSyncComplete(pollResult, batchID, w)
		case batchSyncStatusFailed:
			return fmt.Errorf("error: sync failed for batch %s: %s", batchID, syncFailureReason(pollResult))
		case batchSyncStatusQueued, batchSyncStatusProcessing:
			// continue polling
		default:
			return fmt.Errorf("error: unexpected sync status %q for batch %s", pollResult.Status, batchID)
		}
	}
}

func reportSyncComplete(r *client.AITaskBuilderBatchSyncResponse, batchID string, w io.Writer) error {
	fmt.Fprintf(w, "\nSync complete for batch %s.\n", batchID)
	fmt.Fprintf(w, "  Datapoints processed: %d\n", r.DatapointsProcessed)
	fmt.Fprintf(w, "  Tasks created:        %d\n", r.TasksCreated)
	fmt.Fprintf(w, "  Groups created:       %d\n", r.GroupsCreated)
	fmt.Fprintf(w, "  Groups expanded:      %d\n", r.GroupsExpanded)
	return nil
}

func syncFailureReason(r *client.AITaskBuilderBatchSyncResponse) string {
	if r.Reason != "" {
		return r.Reason
	}
	return "no reason provided"
}
