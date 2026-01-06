package collection

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type UpdateOptions struct {
	Config string
}

// NewUpdateCommand creates a new `collection update` command to update a collection
func NewUpdateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts UpdateOptions

	cmd := &cobra.Command{
		Use:   "update <collection-id>",
		Short: "Update a collection",
		Long:  "Update a collection with new values from a YAML or JSON config file",
		Example: `
Update a collection using a YAML config file:
$ prolific collection update 64395e9c2332b8a59a65d51e -t collection.yaml

Update a collection using a JSON config file:
$ prolific collection update 64395e9c2332b8a59a65d51e -t collection.json

Example YAML config file:
---
name: My Updated Collection

Example JSON config file:
{
  "name": "My Updated Collection"
}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			collectionID := args[0]

			if opts.Config == "" {
				return fmt.Errorf("config file is required, use -t to specify a YAML or JSON file")
			}

			v := viper.New()
			v.SetConfigFile(opts.Config)
			if err := v.ReadInConfig(); err != nil {
				return fmt.Errorf("unable to read config file: %s", err)
			}

			var updatePayload model.UpdateCollection
			if err := v.Unmarshal(&updatePayload); err != nil {
				return fmt.Errorf("unable to parse config file: %s", err)
			}

			collection, err := client.UpdateCollection(collectionID, updatePayload)
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "Collection updated successfully\n")
			fmt.Fprintf(w, "ID: %s\n", collection.ID)
			fmt.Fprintf(w, "Name: %s\n", collection.Name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Config, "config", "t", "", "Path to a YAML or JSON file containing your collection updates")

	return cmd
}
