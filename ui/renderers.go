package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

// TableRenderer renders a slice of items as a tab-aligned table using reflection for field access.
type TableRenderer[T any] struct{}

// Render writes items as a tab-aligned table to w.
func (r TableRenderer[T]) Render(items []T, fields string, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)

	fieldList := splitFields(fields)

	for _, field := range fieldList {
		fmt.Fprintf(tw, "%s\t", strings.Trim(field, " "))
	}
	fmt.Fprint(tw, "\n")

	for _, item := range items {
		for _, field := range fieldList {
			fmt.Fprintf(tw, "%v\t", reflect.ValueOf(item).FieldByName(strings.Trim(field, " ")))
		}
		fmt.Fprint(tw, "\n")
	}

	return tw.Flush()
}

// CsvRenderer renders a slice of items as CSV using reflection for field access.
type CsvRenderer[T any] struct{}

// Render writes items as CSV to w. Values containing commas are wrapped in double quotes.
func (r CsvRenderer[T]) Render(items []T, fields string, w io.Writer) error {
	fieldList := splitFields(fields)

	for _, field := range fieldList {
		fmt.Fprintf(w, "%s,", strings.Trim(field, " "))
	}
	fmt.Fprint(w, "\n")

	for _, item := range items {
		for _, field := range fieldList {
			value := fmt.Sprintf("%v", reflect.ValueOf(item).FieldByName(strings.Trim(field, " ")))
			if strings.Contains(value, ",") {
				value = fmt.Sprintf("%q", value)
			}
			fmt.Fprintf(w, "%s,", value)
		}
		fmt.Fprint(w, "\n")
	}

	return nil
}

// JSONRenderer renders a slice of items as indented JSON.
type JSONRenderer[T any] struct{}

// Render writes items as indented JSON to w.
func (r JSONRenderer[T]) Render(items []T, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}

func splitFields(fields string) []string {
	parts := strings.Split(fields, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
