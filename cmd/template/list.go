package template

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/prolific-oss/cli/docs/examples"
	"github.com/spf13/cobra"
)

// Template holds metadata about an embedded template file.
type Template struct {
	ID       string
	Filename string
	Category string
	Format   string
}

// listTemplates returns all available study and collection templates from the
// embedded filesystem.
func listTemplates() []Template {
	var templates []Template

	_ = fs.WalkDir(examples.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			return nil
		}

		name := strings.TrimSuffix(path, ext)

		category := categorise(name)
		if category == "" {
			return nil
		}

		format := strings.TrimPrefix(ext, ".")
		if format == "yml" {
			format = "yaml"
		}

		templates = append(templates, Template{
			ID:       name,
			Filename: path,
			Category: category,
			Format:   format,
		})

		return nil
	})

	sort.Slice(templates, func(i, j int) bool {
		if templates[i].Category != templates[j].Category {
			return templates[i].Category < templates[j].Category
		}
		return templates[i].ID < templates[j].ID
	})

	return templates
}

// categorise returns the category for a template based on its filename, or
// empty string if the file should be excluded.
func categorise(name string) string {
	if strings.HasPrefix(name, "collection") {
		return "collection"
	}
	if strings.HasPrefix(name, "study-") ||
		strings.HasPrefix(name, "standard-sample") ||
		strings.HasPrefix(name, "minimal-study") ||
		strings.HasPrefix(name, "multi-submission") ||
		strings.HasPrefix(name, "multiple-participant") ||
		strings.HasPrefix(name, "star-") {
		return "study"
	}
	return ""
}

// NewListCommand creates the `template list` subcommand.
func NewListCommand(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available templates",
		Long:  `List all bundled study and collection templates with their ID, category, and format.`,
		Example: `
List all templates
$ prolific template list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			templates := listTemplates()

			tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tCategory\tFormat")

			for _, t := range templates {
				fmt.Fprintf(tw, "%s\t%s\t%s\n", t.ID, t.Category, t.Format)
			}

			return tw.Flush()
		},
	}

	return cmd
}
