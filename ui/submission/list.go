package submission

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
)

// DefaultListFields is the default fields we should show if the user has not specified.
const DefaultListFields = "ParticipantID,StartedAt,TimeTaken,StudyCode,Status"

// ListUsedOptions are the options selected by the user.
type ListUsedOptions struct {
	StudyID        string
	Status         string
	Csv            bool
	NonInteractive bool
	Fields         string
	Limit          int
	Offset         int
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

// NonInteractiveRenderer will just output submission data to the UI.
type NonInteractiveRenderer struct{}

// Render will just display the results in a table.
func (r *NonInteractiveRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	submissions, err := client.GetSubmissions(opts.StudyID, opts.Limit, opts.Offset)
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

	for _, submission := range submissions.Results {
		for _, field := range fieldList {
			fmt.Fprintf(tw, "%v\t", reflect.ValueOf(submission).FieldByName(strings.Trim(field, " ")))
		}
		fmt.Fprint(tw, "\n")
	}

	_ = tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(submissions.Results), submissions.Meta.Count))

	return nil
}

// CsvRenderer will render the output in a CSV format.
type CsvRenderer struct{}

// Render will render the studies in the CSV format.
func (r *CsvRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	submissions, err := client.GetSubmissions(opts.StudyID, opts.Limit, opts.Offset)
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

	for _, submission := range submissions.Results {
		for _, field := range fieldList {
			value := reflect.ValueOf(submission).FieldByName(strings.Trim(field, " "))
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
