package campaign

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/spf13/cobra"
)

// CampaignListOptions is the options for the listing campaigns command.
type CampaignListOptions struct {
	Args        []string
	WorkspaceID string
	Limit       int
	Offset      int
}

// NewListCommand creates a new command to deal with campaigns
func NewListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts CampaignListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your campaigns",
		Long: `List your campaigns

Bring your own participants.

Researchers can bring participants to their workspace by creating a campaign.
A campaign is a unique URL that can be shared with potential participants to sign
up for a Prolific participant account.`,
		Example: `
List your campaigns
$ prolific campaign list -w <workspace_id>

Utilise the paging options to limit your campaigns, for example one campaign
$ prolific campaign list -w <workspace_id> -l 1

Offset records in the result set, for example by 2
$ prolific campaign list -w <workspace_id> -l 1 -o 2
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderCampaigns(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", "", "Filter campaigns by workspace.")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of campaigns returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of campaigns to offset")

	return cmd
}

// renderCampaigns will show your campaigns
func renderCampaigns(c client.API, opts CampaignListOptions, w io.Writer) error {
	if opts.WorkspaceID == "" {
		return errors.New("please provide a workspace ID")
	}
	campaigns, err := c.GetCampaigns(opts.WorkspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "ID", "Name", "Link")
	for _, campaign := range campaigns.Results {
		fmt.Fprintf(tw, "%s\t%s\t%v\n", campaign.ID, campaign.Name, campaign.SignupLink)
	}

	_ = tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(campaigns.Results), campaigns.Meta.Count))

	return nil
}
