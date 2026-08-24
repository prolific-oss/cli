package client_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/prolific-oss/cli/client"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer

	defer func() {
		os.Stdout = originalStdout
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}

	return string(output)
}

func TestFeedbackEndpointsDoNotWriteDebugLogs(t *testing.T) {
	const feedbackText = "sensitive participant feedback"

	httpClient := &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(fmt.Sprintf(
					`{"results":[{"participant_id":null,"category":"other","text":%q,"ratings":{"clarity_rating":4,"difficulty_rating":4,"fairness_rating":4}}],"meta":{"count":1}}`,
					feedbackText,
				))),
				Header: make(http.Header),
			}, nil
		}),
	}

	c := client.Client{
		Client:  httpClient,
		BaseURL: "https://example.test",
		Token:   "token",
		Debug:   true,
	}

	output := captureStdout(t, func() {
		response, err := c.GetStudyFeedback("63c123af913a974f87e8e7fc", false, 0, 0)
		if err != nil {
			t.Fatalf("get study feedback: %v", err)
		}
		if response.Results[0].Text == nil || *response.Results[0].Text != feedbackText {
			t.Fatalf("expected feedback response to be decoded")
		}
		if response.Results[0].ParticipantID != nil {
			t.Fatalf("expected optional participant ID to decode as nil")
		}
	})

	if output != "" {
		t.Fatalf("expected no debug output for feedback endpoints, got %q", output)
	}
}

func TestGetStudyFeedbackRequest(t *testing.T) {
	const studyID = "63c123af913a974f87e8e7fc"

	tests := []struct {
		name               string
		hasWrittenFeedback bool
		limit              int
		offset             int
		query              string
	}{
		{
			name:               "includes filters and pagination",
			hasWrittenFeedback: true,
			limit:              25,
			offset:             10,
			query:              "has_written_feedback=true&limit=25&offset=10",
		},
		{
			name:  "omits zero pagination values",
			query: "has_written_feedback=false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := &http.Client{
				Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodGet {
						t.Errorf("expected GET request, got %s", request.Method)
					}
					expectedPath := "/api/v1/studies/" + studyID + "/feedback/responses/"
					if request.URL.Path != expectedPath {
						t.Errorf("expected path %q, got %q", expectedPath, request.URL.Path)
					}
					if request.URL.RawQuery != tt.query {
						t.Errorf("expected query %q, got %q", tt.query, request.URL.RawQuery)
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"results":[]}`)),
						Header:     make(http.Header),
					}, nil
				}),
			}

			c := client.Client{
				Client:  httpClient,
				BaseURL: "https://example.test",
				Token:   "token",
			}

			if _, err := c.GetStudyFeedback(studyID, tt.hasWrittenFeedback, tt.limit, tt.offset); err != nil {
				t.Fatalf("get study feedback: %v", err)
			}
		})
	}
}

func TestFeedbackEndpointsDoNotDoubleWrapRequestErrors(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("request failed")
		}),
	}

	c := client.Client{
		Client:  httpClient,
		BaseURL: "https://example.test",
		Token:   "token",
	}

	_, err := c.GetStudyRatings("63c123af913a974f87e8e7fc")
	if err == nil {
		t.Fatal("expected request error")
	}
	if count := strings.Count(err.Error(), "unable to fulfil request"); count != 1 {
		t.Fatalf("expected request error to be wrapped once, got %q", err)
	}
}
