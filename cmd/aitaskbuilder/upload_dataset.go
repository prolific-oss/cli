package aitaskbuilder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// DatasetUploadOptions are the options for uploading to an AI Task Builder dataset.
type DatasetUploadOptions struct {
	Args      []string
	DatasetID string
	FilePath  string
}

// NewDatasetUploadCommand creates a new command for uploading to an AI Task Builder dataset.
func NewDatasetUploadCommand(client client.API, w io.Writer) *cobra.Command {
	var opts DatasetUploadOptions

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload data to a dataset",
		Long: `Upload csv file contents to a dataset.

Upload csv file contents to an existing dataset to be processed for use as part of a batch.`,
		Example: `
Upload a file to a dataset:
$ prolific aitaskbuilder dataset upload -d <dataset_id> -f docs/examples/aitb-model-evaluation.csv
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := uploadDatasetFile(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.DatasetID, "dataset-id", "d", "", "The ID of the dataset to upload to (required)")
	flags.StringVarP(&opts.FilePath, "file", "f", "", "The path to the CSV file to upload (required)")

	_ = cmd.MarkFlagRequired("dataset-id")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

// uploadDatasetFile uploads a file to an AI Task Builder dataset
func uploadDatasetFile(client client.API, opts DatasetUploadOptions, w io.Writer) error {
	if opts.DatasetID == "" {
		return errors.New("dataset ID is required")
	}

	if opts.FilePath == "" {
		return errors.New("file path is required")
	}

	// Check if file exists
	if _, err := os.Stat(opts.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", opts.FilePath)
	}

	// Extract filename without extension for the API call
	fileName := filepath.Base(opts.FilePath)
	if !strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		return errors.New("file must be a CSV file")
	}
	// Remove .csv extension for the API call
	fileNameWithoutExt := strings.TrimSuffix(fileName, ".csv")

	fmt.Fprintf(w, "Getting upload URL for dataset %s and file %s...\n", opts.DatasetID, fileName)

	// Get upload URL from API
	uploadResponse, err := client.GetAITaskBuilderDatasetUploadURL(opts.DatasetID, fileNameWithoutExt)
	if err != nil {
		return fmt.Errorf("failed to get upload URL: %w", err)
	}

	fmt.Fprintf(w, "Upload URL obtained, expires at: %s\n", uploadResponse.ExpiresAt)

	// Upload file to the presigned URL
	err = uploadFileToPresignedURL(opts.FilePath, uploadResponse.UploadURL)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	fmt.Fprintf(w, "Successfully uploaded %s to dataset %s\n", fileName, opts.DatasetID)
	return nil
}

// uploadFileToPresignedURL uploads a file to a presigned URL using PUT request
func uploadFileToPresignedURL(filePath, uploadURL string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(context.Background(), http.MethodPut, uploadURL, buffer)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "multipart/form-data")
	client := &http.Client{}
	_, err = client.Do(request)
	return err
}
