package feedback

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// RatingsOptions is the options for viewing aggregate study ratings.
type RatingsOptions struct {
	Study  string
	Output shared.OutputOptions
}

// NewRatingsCommand creates a command for viewing aggregate ratings for a study.
func NewRatingsCommand(c client.API, w io.Writer) *cobra.Command {
	var opts RatingsOptions

	cmd := &cobra.Command{
		Use:   "ratings",
		Short: "View aggregate participant ratings for a study",
		Long: `View aggregate participant ratings for a study

Shows the average clarity, difficulty and fairness ratings and the number of
responses contributing to each rating.`,
		Example: `
View aggregate ratings for a study
$ prolific feedback ratings -s 63c123af913a974f87e8e7fc

Output aggregate ratings as JSON
$ prolific feedback ratings -s 63c123af913a974f87e8e7fc --json

Output aggregate ratings as CSV
$ prolific feedback ratings -s 63c123af913a974f87e8e7fc --csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Study == "" {
				return errors.New("please provide a study ID")
			}

			ratings, err := c.GetStudyRatings(opts.Study)
			if err != nil {
				return handleAPIError(err)
			}

			switch shared.ResolveFormat(opts.Output) {
			case "json":
				renderer := ui.JSONRenderer[RatingRow]{}
				if err := renderer.Render(NewRatingRows(*ratings), w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "csv":
				renderer := ui.CsvRenderer[RatingItem]{}
				if err := renderer.Render(NewRatingItems(*ratings), RatingFields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			default:
				renderer := ui.TableRenderer[RatingItem]{}
				if err := renderer.Render(NewRatingItems(*ratings), RatingFields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Study, "study", "s", "", "The study you want aggregate ratings for (required).")
	shared.AddOutputFlags(cmd, &opts.Output)

	return cmd
}
