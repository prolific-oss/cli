package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

func setupMockClient(t *testing.T) *mock_client.MockAPI {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })
	return mock_client.NewMockAPI(ctrl)
}

func TestNewDatasetStatusCommand(t *testing.T) {
	c := setupMockClient(t)

	cmd := aitaskbuilder.NewDatasetStatusCommand(c, os.Stdout)

	use := "status"
	short := "Get an AI Task Builder dataset status"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewDatasetStatusCommandCallsAPI(t *testing.T) {
	testCases := []struct {
		name      string
		datasetID string
		status    string
	}{
		{
			name:      "returns READY status",
			datasetID: "01954894-65b3-779e-aaf6-348698e23612",
			status:    "READY",
		},
		{
			name:      "returns PROCESSING status",
			datasetID: "01954894-65b3-779e-aaf6-348698e23613",
			status:    "PROCESSING",
		},
		{
			name:      "returns UNINITIALISED status",
			datasetID: "01954894-65b3-779e-aaf6-348698e23614",
			status:    "UNINITIALISED",
		},
		{
			name:      "returns ERROR status",
			datasetID: "01954894-65b3-779e-aaf6-348698e23615",
			status:    "ERROR",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := setupMockClient(t)

			response := client.GetAITaskBuilderDatasetStatusResponse{
				Status: tc.status,
			}

			c.
				EXPECT().
				GetAITaskBuilderDatasetStatus(gomock.Eq(tc.datasetID)).
				Return(&response, nil).
				AnyTimes()

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)
			cmd := aitaskbuilder.NewDatasetStatusCommand(c, writer)

			_ = cmd.Flags().Set("dataset-id", tc.datasetID)
			_ = cmd.RunE(cmd, nil)
			writer.Flush()

			expected := fmt.Sprintf(`AI Task Builder Dataset Status:
Dataset ID: %s
Status: %s
`, tc.datasetID, tc.status)
			actual := b.String()
			if actual != expected {
				t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
			}
		})
	}
}

func TestNewDatasetStatusCommandHandlesErrors(t *testing.T) {
	c := setupMockClient(t)

	datasetID := "the-invalid-dataset-id"
	errorMessage := "dataset not found"

	c.
		EXPECT().
		GetAITaskBuilderDatasetStatus(gomock.Eq(datasetID)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := aitaskbuilder.NewDatasetStatusCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", datasetID)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewDatasetStatusCommandRequiresDatasetID(t *testing.T) {
	c := setupMockClient(t)

	cmd := aitaskbuilder.NewDatasetStatusCommand(c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when dataset-id is missing")
	}

	if !cmd.Flags().Changed("dataset-id") {
		expected := aitaskbuilder.ErrDatasetIDRequired
		if err.Error() != "error: "+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}

func TestNewDatasetStatusCommandHelpText(t *testing.T) {
	c := setupMockClient(t)

	cmd := aitaskbuilder.NewDatasetStatusCommand(c, os.Stdout)

	// Check that the long description contains status information
	if !strings.Contains(cmd.Long, "UNINITIALISED") {
		t.Fatal("expected long description to contain UNINITIALISED status")
	}
	if !strings.Contains(cmd.Long, "PROCESSING") {
		t.Fatal("expected long description to contain PROCESSING status")
	}
	if !strings.Contains(cmd.Long, "READY") {
		t.Fatal("expected long description to contain READY status")
	}
	if !strings.Contains(cmd.Long, "ERROR") {
		t.Fatal("expected long description to contain ERROR status")
	}

	// Check example contains correct flag
	if !strings.Contains(cmd.Example, "-d <dataset_id>") {
		t.Fatal("expected example to contain '-d <dataset_id>' flag usage")
	}
}
