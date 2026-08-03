package rewardrecommendations

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// allowedCurrencies are the ISO 4217 currency codes Prolific's reward
// recommendations endpoint supports.
var allowedCurrencies = []string{"USD", "GBP"}

// Options are the options for the reward recommendations command.
type Options struct {
	WorkspaceID string
	Currency    string
	ScreenerIDs []string
}

// NewCommand creates a new `reward-recommendations` command to calculate
// recommended participant reward rates.
func NewCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts Options

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Calculate recommended participant reward rates",
		Long: `Calculate reward recommendations

Returns Prolific's recommended minimum and "good" hourly reward rates for a
workspace and currency, optionally scoped to a set of screener filter IDs
(e.g. custom group filter IDs) you plan to apply to your study.

We recommend calling this before creating a draft study, then using the
returned rates to determine your study's reward.`,
		Example: `
Calculate reward recommendations for a workspace in GBP
$ prolific reward-recommendations -w 6261321e223a605c7a4f7623 -c GBP

Scope the recommendation to specific screeners
$ prolific reward-recommendations -w 6261321e223a605c7a4f7623 -c GBP -s mandarin -s spanish

Use the workspace set in your config file
$ prolific reward-recommendations -c USD
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := render(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The ID of the workspace you'll be creating the study in.")
	flags.StringVarP(&opts.Currency, "currency", "c", "", fmt.Sprintf("The currency for the recommendation. One of: %s", strings.Join(allowedCurrencies, ", ")))
	flags.StringArrayVarP(&opts.ScreenerIDs, "screener-id", "s", nil, "A screener/filter ID to scope the recommendation to. Can be specified multiple times.")

	return cmd
}

// render will fetch and display the reward recommendation.
func render(c client.API, opts Options, w io.Writer) error {
	if opts.WorkspaceID == "" {
		return errors.New("please provide a workspace ID")
	}

	if opts.Currency == "" {
		return errors.New("please provide a currency")
	}

	if !slices.Contains(allowedCurrencies, opts.Currency) {
		return fmt.Errorf("currency must be one of: %s", strings.Join(allowedCurrencies, ", "))
	}

	response, err := c.GetRewardRecommendations(opts.WorkspaceID, opts.Currency, opts.ScreenerIDs)
	if err != nil {
		return err
	}

	if len(*response) == 0 {
		return errors.New("no reward recommendations returned")
	}

	// The API guarantees the first item is the most recent set of rates.
	recommendation := (*response)[0]

	toCurrency := func(amount int) float64 {
		return float64(amount) / 100
	}

	tw := tabwriter.NewWriter(w, 0, 1, 2, ' ', 0)
	fmt.Fprintf(tw, "Currency:\t%s\n", recommendation.Currency)
	fmt.Fprintf(tw, "Minimum reward per hour:\t%.2f\n", toCurrency(recommendation.MinRewardPerHour))
	fmt.Fprintf(tw, "Recommended reward per hour:\t%.2f\n", toCurrency(recommendation.RecommendedRewardPerHour))

	return tw.Flush()
}
