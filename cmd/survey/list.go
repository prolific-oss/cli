package survey

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing surveys command.
type ListOptions struct {
	Args   []string
	Limit  int
	Offset int
}

// NewListCommand creates a new command to list surveys
func NewListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide a list of your surveys",
		Long: `List your surveys

Surveys are associated with your researcher account.
`,
		Example: `
List your surveys

$ prolific survey list
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderList(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of surveys returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of surveys to offset")

	return cmd
}

// renderList will list your surveys
func renderList(client client.API, opts ListOptions, w io.Writer) error {
	me, err := client.GetMe()
	if err != nil {
		return err
	}

	records, err := client.GetSurveys(me.ID, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "ID", "Title", "Date Created")
	for _, record := range records.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", record.ID, record.Title, record.DateCreated.Format("2006-01-02"))
	}

	_ = tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(records.Results), len(records.Results)))

	return nil
}
