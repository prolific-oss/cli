package study

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// ListUsedOptions are the options selected by the user.
type ListUsedOptions struct {
	Status         string
	NonInteractive bool
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
	studies, err := client.GetStudies(opts.Status)
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

	if err := tea.NewProgram(lv).Start(); err != nil {
		return fmt.Errorf("cannot render studies: %s", err)
	}

	return nil
}

// NonInteractiveRenderer will just output study data to the UI.
type NonInteractiveRenderer struct{}

// Render will just display the results in a table.
func (r *NonInteractiveRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	studies, err := client.GetStudies(opts.Status)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "ID", "Title", "Status")

	for _, study := range studies.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", study.ID, study.Title(), study.Status)
	}

	tw.Flush()
	return nil
}
