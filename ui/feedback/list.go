package feedback

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
)

// StudyRating pairs the API's rating id with a human-readable label, in
// display order. The API keys the ratings summary by rating id (e.g.
// "clarity_rating"), not by the shorter names used in per-record ratings
// (e.g. "clarity").
type studyRatingID struct {
	ID    string
	Label string
}

// StudyRatingIDs are the known study-level rating categories, in display order.
var StudyRatingIDs = []studyRatingID{
	{ID: "clarity_rating", Label: "clarity"},
	{ID: "difficulty_rating", Label: "ease"},
	{ID: "fairness_rating", Label: "fairness"},
}

// ListUsedOptions are the options selected by the user.
type ListUsedOptions struct {
	StudyID            string
	HasWrittenFeedback bool
	Limit              int
	Offset             int
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

// Render will use the render strategy to render the study feedback.
func (r *ListRenderer) Render(client client.API, opts ListUsedOptions, w io.Writer) error {
	return r.strategy.Render(client, opts, w)
}

// fetchFeedback retrieves feedback records for the study. If the user hasn't
// specified a limit or offset, both are omitted from the request entirely,
// so the API returns every record for the study in one response. It returns
// the records fetched and the total number of records available for the
// study.
func fetchFeedback(c client.API, opts ListUsedOptions) ([]model.StudyFeedback, int, error) {
	response, err := c.GetStudyFeedback(opts.StudyID, opts.HasWrittenFeedback, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, err
	}

	if len(response.Results) == 0 {
		return nil, response.Meta.Count, errors.New("no feedback found for this study")
	}

	return response.Results, response.Meta.Count, nil
}

func formatRating(rating *float64) string {
	if rating == nil {
		return "-"
	}
	return fmt.Sprintf("%v", *rating)
}

func formatOptionalString(value *string) string {
	if value == nil || *value == "" {
		return "-"
	}
	return *value
}

// NonInteractiveRenderer outputs the study's aggregated ratings, followed by
// its written feedback records, straight to the terminal.
type NonInteractiveRenderer struct{}

// Render displays the study ratings summary and feedback records as tables.
func (r *NonInteractiveRenderer) Render(c client.API, opts ListUsedOptions, w io.Writer) error {
	ratings, err := c.GetStudyRatings(opts.StudyID)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "Study ratings:")
	rtw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprint(rtw, "Rating\tAverage\tResponses\n")
	for _, rating := range StudyRatingIDs {
		summary := (*ratings)[rating.ID]
		fmt.Fprintf(rtw, "%s\t%s\t%d\n", rating.Label, formatRating(summary.AverageRating), summary.TotalCount)
	}
	rtw.Flush()

	records, total, err := fetchFeedback(c, opts)
	if err != nil {
		return err
	}

	fmt.Fprintln(w)

	fmt.Fprintln(w, "Feedback:")
	ftw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprint(ftw, "ParticipantID\tCategory\tClarity\tEase\tFairness\tText\n")
	for _, record := range records {
		fmt.Fprintf(
			ftw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			record.ParticipantID,
			formatOptionalString(record.Category),
			formatRating(record.Ratings.Clarity),
			formatRating(record.Ratings.Ease),
			formatRating(record.Ratings.Fairness),
			formatOptionalString(record.Text),
		)
	}
	ftw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(records), total))

	return nil
}

// CsvRenderer renders the feedback records in CSV format.
type CsvRenderer struct{}

// Render renders the feedback records in CSV format.
func (r *CsvRenderer) Render(c client.API, opts ListUsedOptions, w io.Writer) error {
	records, _, err := fetchFeedback(c, opts)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	if err := csvWriter.Write([]string{"ParticipantID", "Category", "Clarity", "Ease", "Fairness", "Text"}); err != nil {
		return err
	}

	for _, record := range records {
		row := []string{
			record.ParticipantID,
			formatOptionalString(record.Category),
			formatRating(record.Ratings.Clarity),
			formatRating(record.Ratings.Ease),
			formatRating(record.Ratings.Fairness),
			formatOptionalString(record.Text),
		}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// JSONRenderer renders the feedback records in JSON format, matching the
// record-level contract: participant_id, category, text, ratings.
type JSONRenderer struct{}

// Render renders the feedback records in JSON format.
func (r *JSONRenderer) Render(c client.API, opts ListUsedOptions, w io.Writer) error {
	records, _, err := fetchFeedback(c, opts)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(records)
}
