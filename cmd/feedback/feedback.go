package feedback

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
)

const limitedAccessMessage = "We’re currently testing participant feedback with a limited number of researchers. It’ll be available more widely soon."

var errLimitedAccess = errors.New(limitedAccessMessage) //nolint:staticcheck

func handleAPIError(err error) error {
	if client.IsHTTPStatusError(err, http.StatusForbidden) && shared.IsFeatureNotEnabledError(err) {
		return errLimitedAccess
	}
	return fmt.Errorf("error: %s", err)
}

// NewFeedbackCommand creates a new `feedback` command
func NewFeedbackCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "View participant feedback for your studies",
		Long: `View participant feedback for your studies

Participants can leave feedback, including written comments and clarity,
difficulty and fairness ratings, after taking part in a study. These commands allow
you to retrieve that feedback programmatically.`,
	}

	cmd.AddCommand(
		NewListCommand(client, w),
		NewRatingsCommand(client, w),
	)
	return cmd
}
