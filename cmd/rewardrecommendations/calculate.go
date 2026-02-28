package rewardrecommendations

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
)

// CalculateOptions is the options for calculating reward recommendations.
type CalculateOptions struct {
	Currency      string
	EstimatedTime int
	FilterIDs     []string
}

// NewCalculateCommand creates a new command to calculate reward recommendations
func NewCalculateCommand(c client.API, w io.Writer) *cobra.Command {
	var opts CalculateOptions

	cmd := &cobra.Command{
		Use:   "reward-recommendations",
		Short: "Calculate recommended reward rates for participants",
		Long: `Calculate recommended participant reward rates for a given currency and study parameters.

At Prolific, we encourage data collectors to pay their participants as fairly as possible 
for the particular skills they provide. This command provides recommended participant reward 
rates for a given currency across a set of filters.

If you're using custom groups, we heavily recommend that you calculate reward recommendations 
with the custom group filter IDs before creating a draft study. You should then use the 
returned values to determine the reward for your study.`,
		Example: `
Calculate reward recommendations for a 10 minute study in GBP
$ prolific reward-recommendations -c GBP -t 10

Calculate reward recommendations with filter IDs
$ prolific reward-recommendations -c USD -t 15 -f filter1,filter2

Calculate reward recommendations for a 20 minute study in EUR
$ prolific reward-recommendations -c EUR -t 20
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := renderRewardRecommendations(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Currency, "currency", "c", model.DefaultCurrency, "Currency code (e.g., GBP, USD, EUR)")
	flags.IntVarP(&opts.EstimatedTime, "time", "t", 10, "Estimated completion time in minutes")
	flags.StringSliceVarP(&opts.FilterIDs, "filters", "f", []string{}, "Comma-separated list of filter IDs")

	return cmd
}

// renderRewardRecommendations will show reward recommendations
func renderRewardRecommendations(c client.API, opts CalculateOptions, w io.Writer) error {
	recommendations, err := c.GetRewardRecommendations(opts.Currency, opts.EstimatedTime, opts.FilterIDs)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\n", "Metric", "Value")
	fmt.Fprintf(tw, "%s\t%s\n", "Currency", recommendations.CurrencyCode)
	fmt.Fprintf(tw, "%s\t%.2f\n", "Min Reward Per Hour", recommendations.MinRewardPerHour)
	fmt.Fprintf(tw, "%s\t%.2f\n", "Estimated Reward Per Hour", recommendations.EstimatedRewardPerHour)
	fmt.Fprintf(tw, "%s\t%.2f\n", "Max Reward Per Hour", recommendations.MaxRewardPerHour)
	fmt.Fprintf(tw, "%s\t%.2f\n", "Min Reward For Study", recommendations.MinRewardForEstimatedTime)
	fmt.Fprintf(tw, "%s\t%.2f\n", "Estimated Reward For Study", recommendations.EstimatedReward)
	fmt.Fprintf(tw, "%s\t%.2f\n", "Max Reward For Study", recommendations.MaxRewardForEstimatedTime)

	_ = tw.Flush()

	if len(opts.FilterIDs) > 0 {
		fmt.Fprintf(w, "\nFilters: %s\n", strings.Join(opts.FilterIDs, ", "))
	}

	return nil
}
