package feedback_test

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/feedback"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func sampleRatingsResponse() client.StudyRatingsResponse {
	clarity := 4.5
	difficulty := 3.0
	fairness := 5.0
	return client.StudyRatingsResponse{
		"clarity_rating":    model.StudyRating{AverageRating: &clarity, TotalCount: 2},
		"difficulty_rating": model.StudyRating{AverageRating: &difficulty, TotalCount: 3},
		"fairness_rating":   model.StudyRating{AverageRating: &fairness, TotalCount: 4},
	}
}

func TestNewRatingsCommandRequiresStudyID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := feedback.NewRatingsCommand(c, &bytes.Buffer{})
	err := cmd.RunE(cmd, nil)

	if err == nil || err.Error() != "please provide a study ID" {
		t.Fatalf("expected missing study ID error, got %v", err)
	}
}

func TestNewRatingsCommandOutputFormats(t *testing.T) {
	tests := []struct {
		name     string
		flag     string
		expected []string
	}{
		{
			name:     "table",
			flag:     "table",
			expected: []string{"Rating", "Average", "Responses", "clarity", "4.5", "difficulty", "fairness"},
		},
		{
			name:     "JSON",
			flag:     "json",
			expected: []string{`"rating": "clarity"`, `"rating": "difficulty"`, `"rating": "fairness"`, `"average": 4.5`, `"responses": 2`},
		},
		{
			name:     "CSV",
			flag:     "csv",
			expected: []string{"Rating,Average,Responses", "clarity,4.5,2", "difficulty,3,3", "fairness,5,4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			ratings := sampleRatingsResponse()
			c.EXPECT().
				GetStudyRatings(gomock.Eq(feedbackTestStudyID)).
				Return(&ratings, nil)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)
			cmd := feedback.NewRatingsCommand(c, writer)
			_ = cmd.Flags().Set("study", feedbackTestStudyID)
			_ = cmd.Flags().Set(tt.flag, "true")

			if err := cmd.RunE(cmd, nil); err != nil {
				t.Fatalf("did not expect error, got %v", err)
			}
			writer.Flush()

			actual := b.String()
			for _, expected := range tt.expected {
				if !strings.Contains(actual, expected) {
					t.Fatalf("expected output to contain %q, got: %s", expected, actual)
				}
			}
			if tt.flag == "json" && (strings.Contains(actual, "clarity_rating") || strings.Contains(actual, "difficulty_rating") || strings.Contains(actual, "fairness_rating")) {
				t.Fatalf("expected JSON output to use display labels, not raw API rating IDs, got: %s", actual)
			}
		})
	}
}

func TestNewRatingsCommandReturnsAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		GetStudyRatings(gomock.Eq(feedbackTestStudyID)).
		Return(nil, errors.New("API error"))

	cmd := feedback.NewRatingsCommand(c, &bytes.Buffer{})
	_ = cmd.Flags().Set("study", feedbackTestStudyID)

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "error: API error" {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestNewRatingsCommandReturnsLimitedAccessMessageForForbiddenResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		GetStudyRatings(gomock.Eq(feedbackTestStudyID)).
		Return(nil, &client.UnrecognizedAPIError{
			StatusCode: http.StatusForbidden,
			Body:       []byte("You do not currently have permission to access this feature"),
		})

	cmd := feedback.NewRatingsCommand(c, &bytes.Buffer{})
	_ = cmd.Flags().Set("study", feedbackTestStudyID)

	err := cmd.RunE(cmd, nil)
	expected := "We’re currently testing participant feedback with a limited number of researchers. It’ll be available more widely soon."
	if err == nil || err.Error() != expected {
		t.Fatalf("expected limited access message %q, got %v", expected, err)
	}
}

func TestFeedbackCommandRegistersRatings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := feedback.NewFeedbackCommand(c, &bytes.Buffer{})
	ratings, _, err := cmd.Find([]string{"ratings"})
	if err != nil {
		t.Fatalf("expected ratings command, got error %v", err)
	}
	if ratings.Name() != "ratings" {
		t.Fatalf("expected ratings command, got %q", ratings.Name())
	}
}
