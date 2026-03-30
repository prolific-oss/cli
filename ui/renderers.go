package ui

import (
	"encoding/csv"
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
			v := reflect.ValueOf(item).FieldByName(strings.Trim(field, " "))
			if !v.IsValid() {
				fmt.Fprintf(tw, "\t")
				continue
			}
			fmt.Fprintf(tw, "%v\t", v)
		}
		fmt.Fprint(tw, "\n")
	}

	return tw.Flush()
}

// CsvRenderer renders a slice of items as CSV using reflection for field access.
type CsvRenderer[T any] struct{}

// Render writes items as CSV to w.
func (r CsvRenderer[T]) Render(items []T, fields string, w io.Writer) error {
	fieldList := splitFields(fields)

	cw := csv.NewWriter(w)

	if err := cw.Write(fieldList); err != nil {
		return err
	}

	for _, item := range items {
		row := make([]string, len(fieldList))
		for i, field := range fieldList {
			v := reflect.ValueOf(item).FieldByName(field)
			if v.IsValid() {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
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
