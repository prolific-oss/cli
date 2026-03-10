package study

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/prolific-oss/cli/client"
	studyui "github.com/prolific-oss/cli/ui/study"
	"github.com/spf13/cobra"
)

// UpdateOptions is the options for the update study command.
type UpdateOptions struct {
	Args         []string
	TemplatePath string
	Json         bool
	Silent       bool
}

// NewUpdateCommand creates a new `study update` command to allow you to update
// a study using a JSON template file.
func NewUpdateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts UpdateOptions

	cmd := &cobra.Command{
		Use:   "update <study_id>",
		Short: "Update a study using a JSON template",
		Long: `Update study attributes using a JSON template file.

The template should contain only the fields you want to update. The JSON payload
is sent directly to the PATCH /api/v1/studies/:id/ endpoint.

For draft (UNPUBLISHED) studies, all fields can be updated.
For published studies, only certain fields can be updated (e.g., internal_name,
total_available_places, submissions_config).`,
		Example: `
# Update a study from a JSON file
$ prolific study update 60d9aadeb86739de712faee0 -t ./updates.json

# Pipe JSON from stdin
$ echo '{"internal_name": "Updated Name"}' | prolific study update 60d9aadeb86739de712faee0 -t -

# Inline JSON via heredoc
$ prolific study update 60d9aadeb86739de712faee0 -t - <<EOF
{
  "internal_name": "Updated Study v2",
  "total_available_places": 150
}
EOF

# JSON output for scripting
$ prolific study update 60d9aadeb86739de712faee0 -t ./updates.json --json

# Silent mode (exit code only)
$ prolific study update 60d9aadeb86739de712faee0 -t ./updates.json -s`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			return updateStudy(client, opts, cmd.InOrStdin(), w)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template", "t", "", "Path to a JSON file containing the update payload, or - for stdin")
	flags.BoolVar(&opts.Json, "json", false, "Output the full API response as JSON")
	flags.BoolVarP(&opts.Silent, "silent", "s", false, "Suppress output (exit code only)")
	_ = cmd.MarkFlagRequired("template")

	return cmd
}

func updateStudy(client client.API, opts UpdateOptions, stdin io.Reader, w io.Writer) error {
	var data []byte
	var err error

	if opts.TemplatePath == "-" {
		data, err = io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("error reading from stdin: %s", err.Error())
		}
	} else {
		data, err = os.ReadFile(opts.TemplatePath)
		if err != nil {
			return fmt.Errorf("error reading template file: %s", err.Error())
		}
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("error parsing JSON template: %s", err.Error())
	}

	if len(payload) == 0 {
		return fmt.Errorf("error: template contains no fields to update")
	}

	studyID := opts.Args[0]
	study, err := client.UpdateStudy(studyID, payload)
	if err != nil {
		return err
	}

	if opts.Silent {
		return nil
	}

	if opts.Json {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(study)
	}

	fmt.Fprintln(w, studyui.RenderStudy(*study))

	return nil
}
