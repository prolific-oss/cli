package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewGetDatasetStatusCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetDatasetStatusCommand(c, os.Stdout)

	use := "getdatasetstatus"
	short := "Get an AI Task Builder dataset status"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewGetDatasetStatusCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	datasetID := "01954894-65b3-779e-aaf6-348698e23612"

	response := client.GetAITaskBuilderDatasetStatusResponse{
		Status: "READY",
	}

	c.
		EXPECT().
		GetAITaskBuilderDatasetStatus(gomock.Eq(datasetID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetDatasetStatusCommand(c, writer)
	_ = cmd.Flags().Set("dataset-id", datasetID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Dataset Status:
Dataset ID: 01954894-65b3-779e-aaf6-348698e23612
Status: READY
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewGetDatasetStatusCommandCallsAPIWithProcessingStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	datasetID := "01954894-65b3-779e-aaf6-348698e23613"

	response := client.GetAITaskBuilderDatasetStatusResponse{
		Status: "PROCESSING",
	}

	c.
		EXPECT().
		GetAITaskBuilderDatasetStatus(gomock.Eq(datasetID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetDatasetStatusCommand(c, writer)
	_ = cmd.Flags().Set("dataset-id", datasetID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Dataset Status:
Dataset ID: 01954894-65b3-779e-aaf6-348698e23613
Status: PROCESSING
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewGetDatasetStatusCommandCallsAPIWithUninitialisedStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	datasetID := "01954894-65b3-779e-aaf6-348698e23614"

	response := client.GetAITaskBuilderDatasetStatusResponse{
		Status: "UNINITIALISED",
	}

	c.
		EXPECT().
		GetAITaskBuilderDatasetStatus(gomock.Eq(datasetID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetDatasetStatusCommand(c, writer)
	_ = cmd.Flags().Set("dataset-id", datasetID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Dataset Status:
Dataset ID: 01954894-65b3-779e-aaf6-348698e23614
Status: UNINITIALISED
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewGetDatasetStatusCommandCallsAPIWithErrorStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	datasetID := "01954894-65b3-779e-aaf6-348698e23615"

	response := client.GetAITaskBuilderDatasetStatusResponse{
		Status: "ERROR",
	}

	c.
		EXPECT().
		GetAITaskBuilderDatasetStatus(gomock.Eq(datasetID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetDatasetStatusCommand(c, writer)
	_ = cmd.Flags().Set("dataset-id", datasetID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Dataset Status:
Dataset ID: 01954894-65b3-779e-aaf6-348698e23615
Status: ERROR
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewGetDatasetStatusCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	datasetID := "the-invalid-dataset-id"
	errorMessage := "dataset not found"

	c.
		EXPECT().
		GetAITaskBuilderDatasetStatus(gomock.Eq(datasetID)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := aitaskbuilder.NewGetDatasetStatusCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", datasetID)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewGetDatasetStatusCommandRequiresDatasetID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetDatasetStatusCommand(c, os.Stdout)
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

func TestNewGetDatasetStatusCommandHelpText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetDatasetStatusCommand(c, os.Stdout)

	// Check that the long description contains status information
	if !contains(cmd.Long, "UNINITIALISED") {
		t.Fatal("expected long description to contain UNINITIALISED status")
	}
	if !contains(cmd.Long, "PROCESSING") {
		t.Fatal("expected long description to contain PROCESSING status")
	}
	if !contains(cmd.Long, "READY") {
		t.Fatal("expected long description to contain READY status")
	}
	if !contains(cmd.Long, "ERROR") {
		t.Fatal("expected long description to contain ERROR status")
	}

	// Check example contains correct flag
	if !contains(cmd.Example, "-d <dataset_id>") {
		t.Fatal("expected example to contain '-d <dataset_id>' flag usage")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
