package aitaskbuilder

import "github.com/prolific-oss/cli/client"

// BatchExportListFields are the fields shown in batch export list table and
// CSV output.
const BatchExportListFields = "ExportID,Filter,Status,CreatedAt"

// BatchExportListItem is the flat presentation model used by table and CSV
// renderers. The API model remains nested (with a pointer Filter field) so
// JSON output preserves the API response shape.
type BatchExportListItem struct {
	ExportID  string
	Filter    string
	Status    string
	CreatedAt string
}

// NewBatchExportListItems converts export job API models into display-ready rows.
func NewBatchExportListItems(jobs []client.ExportJobListItem) []BatchExportListItem {
	items := make([]BatchExportListItem, 0, len(jobs))
	for _, job := range jobs {
		items = append(items, BatchExportListItem{
			ExportID:  job.ExportID,
			Filter:    formatExportJobFilter(job.Filter),
			Status:    job.Status,
			CreatedAt: job.CreatedAt,
		})
	}
	return items
}
