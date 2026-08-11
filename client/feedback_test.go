package client_test

import (
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
					`{"results":[{"participant_id":"participant-1","category":"other","text":%q,"ratings":{"clarity_rating":4,"difficulty_rating":4,"fairness_rating":4}}],"meta":{"count":1}}`,
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
	})

	if output != "" {
		t.Fatalf("expected no debug output for feedback endpoints, got %q", output)
	}
}
