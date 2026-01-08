package collection

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/ui"
	"github.com/prolific-oss/cli/ui/collection"
	"github.com/spf13/cobra"
)

// GetOptions is the options for the get collection command.
type GetOptions struct {
	Args []string
}

// NewGetCommand creates a new `collection get` command to retrieve details about
// a specific collection.
func NewGetCommand(c client.API, w io.Writer) *cobra.Command {
	var opts GetOptions

	cmd := &cobra.Command{
		Use:   "get <collection-id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Get details of a specific collection",
		Long: `Get collection details

This command allows you to view the details of a specific collection
on the Prolific Platform by providing its ID.`,
		Example: `
View the details of a specific collection

$ prolific collection get 123456789
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 1 || opts.Args[0] == "" {
				return errors.New("please provide a collection ID")
			}

			coll, err := c.GetCollection(opts.Args[0])
			if err != nil {
				if isFeatureNotEnabledError(err) {
					ui.RenderFeatureAccessMessage(FeatureNameDCP2152AITBCollection, FeatureContactEmailDCP2152AITBCollection)
					return nil
				}
				return fmt.Errorf("error: %s", err.Error())
			}

			return collection.RenderCollection(coll, w)
		},
	}

	return cmd
}
