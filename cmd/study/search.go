package study

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SearchOptions struct {
	WorkspaceID string
	Page        int
	JSON        bool
}

// NewSearchCommand searches studies without opening the interactive list.
func NewSearchCommand(c client.API, w io.Writer) *cobra.Command {
	var opts SearchOptions
	cmd := &cobra.Command{
		Use:     "search <query>",
		Short:   "Search accessible studies by name, internal name, or ID",
		Long:    "Search studies on the server. Results are scoped to your access and optionally a workspace. Use --page for another page; the studies endpoint ignores limit/offset. JSON retains results and pagination metadata.",
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
			if opts.JSON {
				return json.NewEncoder(w).Encode(response)
			}
			if err := (ui.TableRenderer[model.Study]{}).Render(response.Results, "ID,Name,InternalName,Status", w); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if response.JSONAPIMeta != nil {
				if _, err := fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(response.Results), response.Meta.Count)); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Scope search to a workspace.")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Result page (starting at 1).")
	cmd.Flags().BoolVarP(&opts.JSON, "json", "j", false, "Output results and pagination metadata as JSON.")
	return cmd
}
