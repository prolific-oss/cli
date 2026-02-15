package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewBatchInstructionsCommandWithFileUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e12350"

	maxFileSizeMB := 10.0
	minFileCount := 1
	maxFileCount := 5

	response := client.CreateAITaskBuilderInstructionsResponse{
		model.Instruction{
			ID:                "inst-upload-1",
			Type:              "file_upload",
			BatchID:           batchID,
			CreatedBy:         "Sean",
			CreatedAt:         "2024-09-18T07:50:15.055Z",
			Description:       "Please upload photos of the product",
			AcceptedFileTypes: []string{".jpg", ".png", ".heic"},
			MaxFileSizeMB:     &maxFileSizeMB,
			MinFileCount:      &minFileCount,
			MaxFileCount:      &maxFileCount,
		},
	}

	instructions := client.CreateAITaskBuilderInstructionsPayload{
		Instructions: []client.Instruction{
			{
				Type:              "file_upload",
				CreatedBy:         "Sean",
				Description:       "Please upload photos of the product",
				AcceptedFileTypes: []string{".jpg", ".png", ".heic"},
				MaxFileSizeMB:     &maxFileSizeMB,
				MinFileCount:      &minFileCount,
				MaxFileCount:      &maxFileCount,
			},
		},
	}

	c.EXPECT().CreateAITaskBuilderInstructions(batchID, instructions).Return(&response, nil)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	instructionsJSON := `[{
		"type": "file_upload",
		"created_by": "Sean",
		"description": "Please upload photos of the product",
		"accepted_file_types": [".jpg", ".png", ".heic"],
		"max_file_size_mb": 10.0,
		"min_file_count": 1,
		"max_file_count": 5
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error; got %s", err.Error())
	}

	writer.Flush()

	expectedOutput := "Successfully added 1 instruction(s) to batch " + batchID
	if !strings.Contains(buf.String(), expectedOutput) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedOutput, buf.String())
	}

	expectedID := "ID: inst-upload-1"
	if !strings.Contains(buf.String(), expectedID) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedID, buf.String())
	}
}

func TestNewBatchInstructionsCommandFileUploadMinimal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e12351"

	response := client.CreateAITaskBuilderInstructionsResponse{
		model.Instruction{
			ID:          "inst-upload-2",
			Type:        "file_upload",
			BatchID:     batchID,
			CreatedBy:   "Sean",
			CreatedAt:   "2024-09-18T07:50:15.055Z",
			Description: "Upload your receipt",
		},
	}

	instructions := client.CreateAITaskBuilderInstructionsPayload{
		Instructions: []client.Instruction{
			{
				Type:        "file_upload",
				CreatedBy:   "Sean",
				Description: "Upload your receipt",
			},
		},
	}

	c.EXPECT().CreateAITaskBuilderInstructions(batchID, instructions).Return(&response, nil)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// Minimal file_upload - all fields are optional except type, created_by, description
	instructionsJSON := `[{
		"type": "file_upload",
		"created_by": "Sean",
		"description": "Upload your receipt"
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error; got %s", err.Error())
	}

	writer.Flush()

	expectedOutput := "Successfully added 1 instruction(s) to batch " + batchID
	if !strings.Contains(buf.String(), expectedOutput) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedOutput, buf.String())
	}
}

func TestNewBatchInstructionsCommandFileUploadInvalidFileExtension(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23710"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// File extension without leading dot
	instructionsJSON := `[{
		"type": "file_upload",
		"created_by": "Sean",
		"description": "Upload your file",
		"accepted_file_types": ["jpg", ".png"]
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "file extension 'jpg' must start with a dot") {
		t.Fatalf("expected error about invalid file extension; got %s", err.Error())
	}
}

func TestNewBatchInstructionsCommandFileUploadInvalidMaxFileSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23711"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// Negative max_file_size_mb
	instructionsJSON := `[{
		"type": "file_upload",
		"created_by": "Sean",
		"description": "Upload your file",
		"max_file_size_mb": -5.0
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "max_file_size_mb must be a positive number") {
		t.Fatalf("expected error about invalid max_file_size_mb; got %s", err.Error())
	}
}

func TestNewBatchInstructionsCommandFileUploadInvalidMinFileCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23712"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// min_file_count less than 1
	instructionsJSON := `[{
		"type": "file_upload",
		"created_by": "Sean",
		"description": "Upload your file",
		"min_file_count": 0
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "min_file_count must be at least 1") {
		t.Fatalf("expected error about invalid min_file_count; got %s", err.Error())
	}
}

func TestNewBatchInstructionsCommandFileUploadMaxLessThanMin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23713"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// max_file_count < min_file_count
	instructionsJSON := `[{
		"type": "file_upload",
		"created_by": "Sean",
		"description": "Upload your files",
		"min_file_count": 5,
		"max_file_count": 2
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "max_file_count (2) must be greater than or equal to min_file_count (5)") {
		t.Fatalf("expected error about max < min file count; got %s", err.Error())
	}
}

func TestNewBatchInstructionsCommandFileUploadZeroMaxFileSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23714"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// Zero max_file_size_mb
	instructionsJSON := `[{
		"type": "file_upload",
		"created_by": "Sean",
		"description": "Upload your file",
		"max_file_size_mb": 0
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "max_file_size_mb must be a positive number") {
		t.Fatalf("expected error about zero max_file_size_mb; got %s", err.Error())
	}
}
