package taskbuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/prolific-oss/cli/client"
)

const (
	maxReviewSampleRateBps = 10000
	maxReviewRubrics       = 20
)

// Error messages for creating a review batch.
const (
	ErrReviewSourceBatchIDRequired     = "--review-source-batch-id is required to create a review batch"
	ErrReviewSampleRateInvalid         = "--review-sample-rate-bps must be a whole number from 1 to 10000 (10000 is every submission)"
	ErrReviewRubricsRequired           = "a review batch needs rubrics: give --review-rubrics-file or --review-rubrics-json"
	ErrBothReviewRubricsInputsProvided = "cannot specify both review-rubrics file and review-rubrics JSON"
	ErrReviewRubricsMustBeArray        = `review rubrics must be a JSON array of {"description", "scoring_scheme"}`
	ErrReviewRubricsCount              = "a review batch takes from 1 to 20 rubrics"
	ErrReviewDatasetNotAllowed         = "a review batch creates its own dataset, so --dataset-id cannot be used with the review flags"
	ErrReviewBatchItemsNotAllowed      = "a review batch's form is built from its rubrics, so batch-items flags cannot be used with the review flags"
)

// ReviewCreateOptions are the flags that make a batch a review of another batch.
type ReviewCreateOptions struct {
	SourceBatchID string
	SampleRateBps int
	RubricsFile   string
	RubricsJSON   string
	Notes         bool
}

// isSet reports whether any review flag was given, which is what makes the batch a review batch.
func (o ReviewCreateOptions) isSet() bool {
	return o.SourceBatchID != "" || o.SampleRateBps != 0 || o.RubricsFile != "" || o.RubricsJSON != "" || o.Notes
}

// parseReviewInput builds the review request, or returns nil when no review flag was given.
func parseReviewInput(o ReviewCreateOptions) (*client.ReviewParams, error) {
	if !o.isSet() {
		return nil, nil
	}
	if o.SourceBatchID == "" {
		return nil, errors.New(ErrReviewSourceBatchIDRequired)
	}
	if o.SampleRateBps < 1 || o.SampleRateBps > maxReviewSampleRateBps {
		return nil, errors.New(ErrReviewSampleRateInvalid)
	}
	rubrics, err := parseReviewRubrics(o.RubricsFile, o.RubricsJSON)
	if err != nil {
		return nil, err
	}
	return &client.ReviewParams{
		SourceBatchID: o.SourceBatchID,
		SampleRateBps: o.SampleRateBps,
		Rubrics:       rubrics,
		Justification: o.Notes,
	}, nil
}

func parseReviewRubrics(filePath, jsonStr string) ([]client.ReviewRubric, error) {
	if filePath != "" && jsonStr != "" {
		return nil, errors.New(ErrBothReviewRubricsInputsProvided)
	}

	var raw []byte
	switch {
	case filePath != "":
		var err error
		raw, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read review-rubrics file: %w", err)
		}
	case jsonStr != "":
		raw = []byte(jsonStr)
	default:
		return nil, errors.New(ErrReviewRubricsRequired)
	}

	var rubrics []client.ReviewRubric
	if err := json.Unmarshal(raw, &rubrics); err != nil {
		return nil, errors.New(ErrReviewRubricsMustBeArray)
	}
	if len(rubrics) == 0 || len(rubrics) > maxReviewRubrics {
		return nil, errors.New(ErrReviewRubricsCount)
	}
	for i, rubric := range rubrics {
		if rubric.Description == "" || rubric.ScoringScheme == "" {
			return nil, fmt.Errorf("review rubric %d needs a description and a scoring_scheme", i+1)
		}
	}
	return rubrics, nil
}
