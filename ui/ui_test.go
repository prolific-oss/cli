package ui_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/prolific-oss/cli/ui"
)

func TestRenderMoney(t *testing.T) {
	tt := []struct {
		name     string
		amount   float64
		currency string
		expected string
	}{{
		name:     "Pound sterling",
		amount:   10.00,
		currency: "GBP",
		expected: "£10.00",
	}, {
		name:     "Default sterling if nothing passed",
		amount:   1.99,
		currency: "",
		expected: "£1.99",
	}, {
		name:     "Dollar",
		amount:   80001.01,
		currency: "USD",
		expected: "$80001.01",
	}}

	for _, tc := range tt {
		actual := ui.RenderMoney(tc.amount, tc.currency)

		if tc.expected != actual {
			t.Fatalf("expected %v got %v", tc.expected, actual)
		}
	}
}

func TestRenderRecordCounter(t *testing.T) {
	tt := []struct {
		name     string
		count    int
		total    int
		expected string
	}{
		{
			name:     "Single record count",
			count:    1,
			total:    1,
			expected: "Showing 1 record of 1",
		},
		{
			name:     "Showing more than one record",
			count:    2,
			total:    10,
			expected: "Showing 2 records of 10",
		},
	}

	for _, tc := range tt {
		actual := ui.RenderRecordCounter(tc.count, tc.total)

		if tc.expected != actual {
			t.Fatalf("expected '%v' got '%v'", tc.expected, actual)
		}
	}
}

func TestRenderFeatureAccessMessage(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	featureName := "AI Task Builder Collections"
	contactEmail := "support@prolific.com"

	ui.RenderFeatureAccessMessage(featureName, contactEmail)

	// Restore stderr and read captured output
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected key phrases
	expectedStrings := []string{
		"EARLY ACCESS",
		featureName,
		"early-access feature",
		"upon request",
		"contribute towards the feature's roadmap",
		contactEmail,
		"activation request will be reviewed",
		"under active development",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("expected output to contain '%s', got:\n%s", expected, output)
		}
	}
}
