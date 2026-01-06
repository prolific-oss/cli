package collection

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
)

// DefaultListFields is the default fields we should show if the user has not specified.
const DefaultListFields = "ID,Name,ItemCount"

// ListUsedOptions are the options selected by the user.
type ListUsedOptions struct {
	WorkspaceID string
	Fields      string
	Limit       int
	Offset      int
}

// ListStrategy defines an interface to allow different strategies to render the list view.
type ListStrategy interface {
	Render(client client.API, opts ListUsedOptions, w io.Writer) error
}

// ListRenderer defines an interface to allow different strategies to render the list view.
type ListRenderer struct {
	strategy ListStrategy
}

// SetStrategy allows you to set the renderer strategy for the list view.
func (r *ListRenderer) SetStrategy(s ListStrategy) {
	r.strategy = s
}

// Render will use the render strategy to render the collections.
func (r *ListRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	return r.strategy.Render(client, opts, w)
}

// InteractiveRenderer runs the bubbles UI framework to provide a rich
// UI experience for the user.
type InteractiveRenderer struct{}

// Render will render the list in an interactive manner.
func (r *InteractiveRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	collections, err := client.GetCollections(opts.WorkspaceID, opts.Limit, opts.Offset)

	if err != nil {
		return err
	}

	var items []list.Item
	var collectionMap = make(map[string]model.Collection)

	for _, collection := range collections.Results {
		items = append(items, collection)
		collectionMap[collection.ID] = collection
	}

	lv := ListView{
		List:        list.New(items, list.NewDefaultDelegate(), 0, 0),
		Collections: collectionMap,
		Client:      client,
	}
	lv.List.Title = "Collections"

	p := tea.NewProgram(lv)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cannot render collections: %s", err)
	}

	return nil
}

// NonInteractiveRenderer will just output collection data to the UI.
type NonInteractiveRenderer struct{}

// Render will just display the results in a table.
func (r *NonInteractiveRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	collections, err := client.GetCollections(opts.WorkspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	if len(opts.Fields) == 0 {
		opts.Fields = DefaultListFields
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)

	fieldList := strings.Split(opts.Fields, ",")

	for _, field := range fieldList {
		fmt.Fprintf(tw, "%s\t", strings.Trim(field, " "))
	}
	fmt.Fprint(tw, "\n")

	for _, collection := range collections.Results {
		for _, field := range fieldList {
			fmt.Fprintf(tw, "%v\t", reflect.ValueOf(collection).FieldByName(strings.Trim(field, " ")))
		}
		fmt.Fprint(tw, "\n")
	}

	return tw.Flush()
}

// CsvRenderer will render the output in a CSV format.
type CsvRenderer struct{}

// Render will render the collections in the CSV format.
func (r *CsvRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	collections, err := client.GetCollections(opts.WorkspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	if len(opts.Fields) == 0 {
		opts.Fields = DefaultListFields
	}

	fieldList := strings.Split(opts.Fields, ",")

	for _, field := range fieldList {
		fmt.Fprintf(w, "%s,", strings.Trim(field, " "))
	}
	fmt.Fprint(w, "\n")

	for _, collection := range collections.Results {
		for _, field := range fieldList {
			value := reflect.ValueOf(collection).FieldByName(strings.Trim(field, " "))
			valueString := fmt.Sprintf("%v", value)
			if strings.Contains(valueString, ",") {
				valueString = fmt.Sprintf("\"%v\"", valueString)
			}
			fmt.Fprintf(w, "%v,", valueString)
		}
		fmt.Fprint(w, "\n")
	}

	return nil
}

// JSONRenderer will render the output in JSON format.
type JSONRenderer struct{}

// Render will render the collections in JSON format.
func (r *JSONRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	collections, err := client.GetCollections(opts.WorkspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(collections.Results)
}
