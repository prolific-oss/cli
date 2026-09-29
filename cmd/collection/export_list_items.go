package collection

import "github.com/prolific-oss/cli/client"

// ExportListFields are the fields shown in collection export list table and
// CSV output.
const ExportListFields = "ExportID,Filter,Status,CreatedAt"

// ExportListItem is the flat presentation model used by table and CSV
// renderers. The API model remains nested (with a pointer Filter field) so
// JSON output preserves the API response shape.
type ExportListItem struct {
	ExportID  string
	Filter    string
	Status    string
	CreatedAt string
}

// NewExportListItems converts export job API models into display-ready rows.
func NewExportListItems(jobs []client.ExportJobListItem) []ExportListItem {
	items := make([]ExportListItem, 0, len(jobs))
	for _, job := range jobs {
		items = append(items, ExportListItem{
			ExportID:  job.ExportID,
			Filter:    formatExportJobFilter(job.Filter),
			Status:    job.Status,
			CreatedAt: job.CreatedAt,
		})
	}
	return items
}
