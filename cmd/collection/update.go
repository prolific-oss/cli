package collection

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
$ prolific collection update collec12345 -t collection.yaml

Update a collection using a JSON config file:
$ prolific collection update collec12345 -t collection.json

Example YAML config file:
---
name: My Updated Collection
items:
  - order: 0
    items:
      - type: free_text
        description: "What is your feedback?"
        order: 0

Example JSON config file:
{
  "name": "My Updated Collection",
  "items": []
}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			collectionID := args[0]

			if opts.Config == "" {
				return fmt.Errorf("config file is required, use -c to specify a YAML or JSON file")
			}

			file, err := os.Open(opts.Config)
			if err != nil {
				return fmt.Errorf("unable to open config file: %s", err)
			}
			defer file.Close()

			var updatePayload model.UpdateCollection
			ext := strings.ToLower(filepath.Ext(opts.Config))

			switch ext {
			case ".json":
				decoder := json.NewDecoder(file)
				decoder.DisallowUnknownFields()
				if err := decoder.Decode(&updatePayload); err != nil {
					return fmt.Errorf("unable to parse JSON config file: %s", err)
				}
			case ".yaml", ".yml":
				decoder := yaml.NewDecoder(file)
				decoder.KnownFields(true)
				if err := decoder.Decode(&updatePayload); err != nil {
					return fmt.Errorf("unable to parse YAML config file: %s", err)
				}
			default:
				return fmt.Errorf("unsupported config file format '%s': use .json, .yaml, or .yml", ext)
			}

			if updatePayload.Name == "" {
				return errors.New(ErrNameRequired)
			}
			if updatePayload.Items == nil {
				return errors.New(ErrCollectionItemsRequired)
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

	cmd.Flags().StringVarP(&opts.Config, "template-path", "t", "", "Path to a YAML or JSON file containing your collection updates")

	return cmd
}
