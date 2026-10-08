package study

import (
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SearchOptions struct {
	WorkspaceID string
	Page        int
	Fields      string
	Output      shared.OutputOptions
}

// NewSearchCommand searches studies without opening the interactive list.
func NewSearchCommand(c client.API, w io.Writer) *cobra.Command {
	var opts SearchOptions
	cmd := &cobra.Command{
		Use:     "search <query>",
		Short:   "Search accessible studies by name, internal name, or ID",
		Long:    "Search studies on the server. Results are scoped to your access and optionally a workspace. Studies paginate by page number rather than limit/offset, so use --page for another page. JSON emits the CLI envelope: results, plus the count, limit and offset describing the window they came from.",
		Example: "prolific study search 'memory task' --workspace WORKSPACE_ID --json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return fmt.Errorf("error: search query must not be empty")
			}
			if opts.Page < 1 {
				return fmt.Errorf("error: page must be positive")
			}
			response, err := c.SearchStudies(query, opts.WorkspaceID, opts.Page)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if response.Results == nil {
				response.Results = []model.Study{}
			}
			// One total for every format, so the JSON count and the table
			// footer cannot disagree about the same response.
			total := client.ReportedTotal(response.JSONAPIMeta, len(response.Results))

			switch shared.ResolveFormat(opts.Output) {
			case shared.FormatJSON:
				// Studies paginate by page number, so the window a page
				// covers comes from the page size the client sends.
				offset := (opts.Page - 1) * client.StudyPageSize
				envelope := ui.NewEnvelope(response.Results, total, client.StudyPageSize, offset)
				if err := (ui.JSONEnvelopeRenderer[model.Study]{}).Render(envelope, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
				return nil
			case shared.FormatCSV:
				if err := (ui.CsvRenderer[model.Study]{}).Render(response.Results, opts.Fields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
				return nil
			}
			if err := (ui.TableRenderer[model.Study]{}).Render(response.Results, opts.Fields, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if _, err := fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(response.Results), total)); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Scope search to a workspace.")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Result page (starting at 1).")
	cmd.Flags().StringVarP(&opts.Fields, "fields", "f", "ID,Name,InternalName,Status", "Comma separated fields to display in table or CSV output.")
	shared.AddOutputFlags(cmd, &opts.Output)
	return cmd
}
