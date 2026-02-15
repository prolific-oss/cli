package bonus

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

type CreateOptions struct {
	Bonuses        []string
	File           string
	NonInteractive bool
	Csv            bool
}

func NewCreateCommand(commandName string, apiClient client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Create bonus payments for study participants",
		Long: `Create bonus payment records for participants in a study.

Provide participant-amount or submission-amount pairs either inline via 
repeatable --bonus flags or via a CSV file. The system creates bonus 
records and returns a summary showing the bonus ID, amounts, fees, VAT, 
and total cost.

Bonus records must be paid separately using the 'bonus pay' command.`,
		Example: `  # Create with inline flags
  prolific bonus create <study_id> --bonus "pid1,4.25" --bonus "pid2,3.50"
  prolific bonus create <study_id> --bonus "subid1,4.25" --bonus "subid2,3.50"

  # Create from CSV file
  prolific bonus create <study_id> --file bonuses.csv

  # Non-interactive output (for scripting)
  prolific bonus create <study_id> --file bonuses.csv -n

  # CSV output format
  prolific bonus create <study_id> --file bonuses.csv -c`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			studyID := args[0]

			err := createBonusPayments(apiClient, studyID, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVarP(&opts.Bonuses, "bonus", "b", nil, "Participant bonus entry in format 'id,amount' (repeatable)")
	flags.StringVarP(&opts.File, "file", "f", "", "Path to CSV file containing bonus entries")
	flags.BoolVarP(&opts.NonInteractive, "non-interactive", "n", false, "Non-interactive output for scripting")
	flags.BoolVarP(&opts.Csv, "csv", "c", false, "Output in CSV format")

	return cmd
}

func createBonusPayments(apiClient client.API, studyID string, opts CreateOptions, w io.Writer) error {
	// Validate mutual exclusivity
	if len(opts.Bonuses) > 0 && opts.File != "" {
		return fmt.Errorf("cannot use both --bonus and --file flags")
	}

	if opts.Csv && opts.NonInteractive {
		return fmt.Errorf("cannot use both --csv and --non-interactive flags")
	}

	if len(opts.Bonuses) == 0 && opts.File == "" {
		return fmt.Errorf("either --bonus or --file flag is required")
	}

	// Parse input to csv_bonuses string
	var csvBonuses string
	var err error

	if opts.File != "" {
		csvBonuses, err = parseBonusFile(opts.File)
	} else {
		csvBonuses, err = parseBonusEntries(opts.Bonuses)
	}

	if err != nil {
		return err
	}

	payload := client.CreateBonusPaymentsPayload{
		StudyID:    studyID,
		CSVBonuses: csvBonuses,
	}

	response, err := apiClient.CreateBonusPayments(payload)
	if err != nil {
		return err
	}

	if opts.Csv {
		return renderCSVOutput(response, w)
	}

	if opts.NonInteractive {
		return renderNonInteractiveOutput(response, w)
	}

	return renderInteractiveOutput(response, csvBonuses, w)
}

func renderInteractiveOutput(resp *client.CreateBonusPaymentsResponse, csvBonuses string, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)

	fmt.Fprintf(tw, "%s\t%s\n", ui.RenderHeading("Bonus ID"), resp.ID)
	fmt.Fprintf(tw, "%s\t%s\n", ui.RenderHeading("Study"), resp.Study)
	fmt.Fprintln(tw)
	fmt.Fprintf(tw, "%s\t%s\n", "Participant", "Amount")
	fmt.Fprintf(tw, "%s\t%s\n", "───────────", "──────")

	// Echo per-participant breakdown from input data
	for _, line := range strings.Split(csvBonuses, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ",", 2)
		if len(parts) == 2 {
			fmt.Fprintf(tw, "%s\t%s\n", parts[0], parts[1])
		}
	}

	fmt.Fprintln(tw)
	fmt.Fprintf(tw, "%s\t%s\n", "Amount", ui.RenderMoney(resp.Amount/100, model.DefaultCurrency))
	fmt.Fprintf(tw, "%s\t%s\n", "Fees", ui.RenderMoney(resp.Fees/100, model.DefaultCurrency))
	fmt.Fprintf(tw, "%s\t%s\n", "VAT", ui.RenderMoney(resp.VAT/100, model.DefaultCurrency))
	fmt.Fprintf(tw, "%s\t%s\n", "Total", ui.RenderMoney(resp.TotalAmount/100, model.DefaultCurrency))

	return tw.Flush()
}

func renderNonInteractiveOutput(resp *client.CreateBonusPaymentsResponse, w io.Writer) error {
	// First line: bonus ID for pipe extraction
	fmt.Fprintln(w, resp.ID)
	fmt.Fprintf(w, "study=%s\n", resp.Study)
	fmt.Fprintf(w, "amount=%s\n", ui.RenderMoney(resp.Amount/100, model.DefaultCurrency))
	fmt.Fprintf(w, "fees=%s\n", ui.RenderMoney(resp.Fees/100, model.DefaultCurrency))
	fmt.Fprintf(w, "vat=%s\n", ui.RenderMoney(resp.VAT/100, model.DefaultCurrency))
	fmt.Fprintf(w, "total=%s\n", ui.RenderMoney(resp.TotalAmount/100, model.DefaultCurrency))

	return nil
}

func renderCSVOutput(resp *client.CreateBonusPaymentsResponse, w io.Writer) error {
	csvWriter := csv.NewWriter(w)

	if err := csvWriter.Write([]string{"id", "study", "amount", "fees", "vat", "total_amount"}); err != nil {
		return err
	}

	if err := csvWriter.Write([]string{
		resp.ID,
		resp.Study,
		ui.RenderMoney(resp.Amount/100, model.DefaultCurrency),
		ui.RenderMoney(resp.Fees/100, model.DefaultCurrency),
		ui.RenderMoney(resp.VAT/100, model.DefaultCurrency),
		ui.RenderMoney(resp.TotalAmount/100, model.DefaultCurrency),
	}); err != nil {
		return err
	}

	csvWriter.Flush()
	return csvWriter.Error()
}
