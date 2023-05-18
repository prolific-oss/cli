package study

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prolific-oss/prolificli/client"
	"github.com/prolific-oss/prolificli/model"
)

// DefaultListFields is the default fields we should show if the user has not specified.
const DefaultListFields = "ID,Name,Status"

// ListUsedOptions are the options selected by the user.
type ListUsedOptions struct {
	Status         string
	NonInteractive bool
	Fields         string
	ProjectID      string
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

// Render will use the render strategy to render the studies.
func (r *ListRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	return r.strategy.Render(client, opts, w)
}

// InteractiveRenderer runs the bubbles UI framework to provide a rich
// UI experience for the user.
type InteractiveRenderer struct{}

// Render will render the list in an interactive manner.
func (r *InteractiveRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	studies, err := client.GetStudies(opts.Status, opts.ProjectID)

	if err != nil {
		return err
	}

	var items []list.Item
	var studyMap = make(map[string]model.Study)

	for _, study := range studies.Results {
		items = append(items, study)
		studyMap[study.ID] = study
	}

	lv := ListView{
		List:    list.New(items, list.NewDefaultDelegate(), 0, 0),
		Studies: studyMap,
		Client:  client,
	}
	lv.List.Title = "My studies"

	p := tea.NewProgram(lv)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cannot render studies: %s", err)
	}

	return nil
}

// NonInteractiveRenderer will just output study data to the UI.
type NonInteractiveRenderer struct{}

// Render will just display the results in a table.
func (r *NonInteractiveRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	studies, err := client.GetStudies(opts.Status, opts.ProjectID)
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

	for _, study := range studies.Results {
		for _, field := range fieldList {
			fmt.Fprintf(tw, "%v\t", reflect.ValueOf(study).FieldByName(strings.Trim(field, " ")))
		}
		fmt.Fprint(tw, "\n")
	}

	tw.Flush()
	return nil
}

// CsvRenderer will render the output in a CSV format.
type CsvRenderer struct{}

// Render will render the studies in the CSV format.
func (r *CsvRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	studies, err := client.GetStudies(opts.Status, opts.ProjectID)
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

	for _, study := range studies.Results {
		for _, field := range fieldList {
			value := reflect.ValueOf(study).FieldByName(strings.Trim(field, " "))
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
