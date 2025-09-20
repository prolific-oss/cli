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

func TestNewBatchSetupCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewBatchSetupCommand(c, os.Stdout)

	use := "setup"
	short := "Setup an AI Task Builder batch"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewBatchSetupCommandCallsAPI(t *testing.T) {
	testCases := []struct {
		name          string
		tasksPerGroup int
	}{
		{
			name:          "successful setup with 3 tasks per group",
			tasksPerGroup: 3,
		},
		{
			name:          "successful setup with 5 tasks per group",
			tasksPerGroup: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			batchID := "12354894-65b3-779e-aaf6-348698e23619"
			datasetID := "8c4c51f1-f6f3-43bc-b65d-7415e8ef22c0"

			response := &client.SetupAITaskBuilderBatchResponse{
				// Empty response body
			}

			c.EXPECT().SetupAITaskBuilderBatch(batchID, datasetID, tc.tasksPerGroup).Return(response, nil)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := aitaskbuilder.NewBatchSetupCommand(c, writer)
			cmd.SetArgs([]string{
				"--batch-id", batchID,
				"--dataset-id", datasetID,
				"--tasks-per-group", fmt.Sprintf("%d", tc.tasksPerGroup),
			})

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("expected error to be nil; got %v", err)
			}

			writer.Flush()

			expected := fmt.Sprintf("AI Task Builder Batch Setup Complete:\nBatch ID: %s\nDataset ID: %s\nTasks per Group: %d\n",
				batchID, datasetID, tc.tasksPerGroup)

			if b.String() != expected {
				t.Fatalf("expected output:\n%s\ngot output:\n%s", expected, b.String())
			}
		})
	}
}

func TestNewBatchSetupCommandAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-321e-aaf6-348698e23321"
	datasetID := "8c4c51f1-f6f3-43bc-b65d-7415e8ef22c0"
	tasksPerGroup := 3

	c.EXPECT().SetupAITaskBuilderBatch(batchID, datasetID, tasksPerGroup).Return(nil, errors.New("API error"))

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchSetupCommand(c, writer)
	cmd.SetArgs([]string{
		"--batch-id", batchID,
		"--dataset-id", datasetID,
		"--tasks-per-group", fmt.Sprintf("%d", tasksPerGroup),
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error; got nil")
	}

	expectedError := "error: API error"
	if err.Error() != expectedError {
		t.Fatalf("expected error: %s; got %s", expectedError, err.Error())
	}
}

func TestNewBatchSetupCommandMissingRequiredFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	testCases := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "missing batch-id flag",
			args:        []string{"--dataset-id", "8c4c51f1-f6f3-43bc-b65d-7415e8ef22c0", "--tasks-per-group", "3"},
			expectedErr: `required flag(s) "batch-id" not set`,
		},
		{
			name:        "missing dataset-id flag",
			args:        []string{"--batch-id", "01954894-65b3-779e-aaf6-348698e23634", "--tasks-per-group", "3"},
			expectedErr: `required flag(s) "dataset-id" not set`,
		},
		{
			name:        "missing tasks-per-group flag",
			args:        []string{"--batch-id", "01954894-65b3-779e-aaf6-348698e23634", "--dataset-id", "8c4c51f1-f6f3-43bc-b65d-7415e8ef22c0"},
			expectedErr: `required flag(s) "tasks-per-group" not set`,
		},
		{
			name:        "missing all flags",
			args:        []string{},
			expectedErr: `required flag(s) "batch-id", "dataset-id", "tasks-per-group" not set`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := aitaskbuilder.NewBatchSetupCommand(c, writer)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error; got nil")
			}

			if err.Error() != tc.expectedErr {
				t.Fatalf("expected error: %s; got %s", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestNewBatchSetupCommandInvalidTasksPerGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	testCases := []struct {
		name          string
		tasksPerGroup string
		expectedErr   string
	}{
		{
			name:          "zero tasks per group",
			tasksPerGroup: "0",
			expectedErr:   "error: tasks per group must be at least 1",
		},
		{
			name:          "negative tasks per group",
			tasksPerGroup: "-1",
			expectedErr:   "error: tasks per group must be at least 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := aitaskbuilder.NewBatchSetupCommand(c, writer)
			cmd.SetArgs([]string{
				"--batch-id", "12354894-65b3-888e-aaf1-348698e99321",
				"--dataset-id", "8c4c51f1-f6f3-43bc-b65d-7415e8ef22c0",
				"--tasks-per-group", tc.tasksPerGroup,
			})

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error; got nil")
			}

			if err.Error() != tc.expectedErr {
				t.Fatalf("expected error: %s; got %s", tc.expectedErr, err.Error())
			}
		})
	}
}
