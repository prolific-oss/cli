package aitaskbuilder

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
)

const (
	datasetUploadDefaultTimeout           = 10 * time.Minute
	datasetUploadPollInterval             = 3 * time.Second
	datasetUploadMaxConsecutivePollErrors = 3
	datasetUploadRecordErrorDisplayLimit  = 20
)

// DatasetUploadPollSleep is the sleep function used between dataset import status polls.
var DatasetUploadPollSleep func(time.Duration) = time.Sleep

var validAudioURLFileExtensions = map[string]bool{
	".aac": true,
	".m4a": true,
	".mp3": true,
	".wav": true,
}

const supportedAudioURLFileExtensions = ".aac, .m4a, .mp3, .wav"

var validVideoURLFileExtensions = map[string]bool{
	".mp4":  true,
	".mov":  true,
	".webm": true,
	".avi":  true,
}

const supportedVideoURLFileExtensions = ".mp4, .mov, .webm, .avi"

// DatasetUploadOptions are the options for uploading to an AI Task Builder dataset.
type DatasetUploadOptions struct {
	Args      []string
	DatasetID string
	FilePath  string
	Format    string
	Timeout   time.Duration
}

// NewDatasetUploadCommand creates a new command for uploading to an AI Task Builder dataset.
func NewDatasetUploadCommand(client client.API, w io.Writer) *cobra.Command {
	var opts DatasetUploadOptions

	cmd := &cobra.Command{
		Use:           "upload",
		Short:         "Upload data to a dataset",
		SilenceErrors: true,
		Long: `Upload CSV or JSONL file contents to a dataset.

Upload file contents to an existing dataset to be processed for use as part of a batch.

The command auto-detects CSV and JSONL formats from the file extension unless
--format is provided. After the upload completes, the command polls the import
job until it completes, partially completes, fails, requires a schema, or
times out.`,
		Example: `
Upload a CSV file with format auto-detected from the extension:
$ prolific aitaskbuilder dataset upload -d <dataset_id> -f docs/examples/aitb-model-evaluation.csv

Upload a JSONL file with an explicit format override:
$ prolific aitaskbuilder dataset upload -d <dataset_id> -f /tmp/records --format jsonl
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			opts.Args = args

			err := uploadDatasetFile(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.DatasetID, "dataset-id", "d", "", "The ID of the dataset to upload to (required)")
	flags.StringVarP(&opts.FilePath, "file", "f", "", "The path to the CSV or JSONL file to upload (required)")
	flags.StringVar(&opts.Format, "format", "", "Override the detected file format (csv or jsonl)")
	flags.DurationVar(&opts.Timeout, "timeout", datasetUploadDefaultTimeout, "Maximum time to wait for import processing. Defaults to 10 minutes.")

	_ = cmd.MarkFlagRequired("dataset-id")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

type datasetUploadRequest struct {
	DisplayName    string
	LocalPath      string
	UploadFilename string
	Format         model.DatasetImportFormat
}

// uploadDatasetFile uploads a file to an AI Task Builder dataset
func uploadDatasetFile(client client.API, opts DatasetUploadOptions, w io.Writer) error {
	if opts.DatasetID == "" {
		return errors.New(ErrDatasetIDRequired)
	}

	if opts.FilePath == "" {
		return errors.New(ErrFilePathRequired)
	}

	if opts.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	uploadRequest, err := prepareDatasetUploadRequest(opts.FilePath, opts.Format)
	if err != nil {
		return err
	}

	dataset, err := client.GetAITaskBuilderDataset(opts.DatasetID)
	if err != nil {
		return fmt.Errorf("failed to get dataset: %w", err)
	}

	if err := validateAudioURLFields(opts.FilePath, uploadRequest.Format, dataset.Schema); err != nil {
		return err
	}

	if err := validateVideoURLFields(opts.FilePath, uploadRequest.Format, dataset.Schema); err != nil {
		return err
	}

	fmt.Fprintf(w, "Getting upload URL for dataset %s and file %s...\n", opts.DatasetID, uploadRequest.UploadFilename)

	// Get upload URL from API
	uploadResponse, err := client.GetAITaskBuilderDatasetUploadURL(opts.DatasetID, uploadRequest.UploadFilename)
	if err != nil {
		return fmt.Errorf("failed to get upload URL: %w", err)
	}
	if strings.TrimSpace(uploadResponse.ImportID) == "" {
		return errors.New("import_id is missing in response")
	}

	fmt.Fprintf(w, "Upload URL obtained for import %s, expires at: %s\n", uploadResponse.ImportID, uploadResponse.ExpiresAt)
	fmt.Fprintf(w, "Uploading %s...\n", uploadRequest.DisplayName)

	// Upload file to the presigned URL
	err = uploadFileToPresignedURL(uploadRequest.LocalPath, uploadResponse.UploadURL, uploadResponse.HTTPMethod, uploadResponse.ContentType)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	fmt.Fprintf(w, "Upload received for dataset %s\nImport ID: %s\n", opts.DatasetID, uploadResponse.ImportID)

	return waitForDatasetImport(client, opts.DatasetID, uploadResponse.ImportID, uploadRequest.Format, opts.Timeout, w)
}

func prepareDatasetUploadRequest(filePath, formatOverride string) (*datasetUploadRequest, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %s", filePath)
		}

		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect file %s: %w", filePath, err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("file path is a directory: %s", filePath)
	}

	if info.Size() == 0 {
		return nil, errors.New(ErrDatasetUploadFileEmpty)
	}

	displayName := filepath.Base(filePath)
	format, err := resolveDatasetUploadFormat(displayName, formatOverride)
	if err != nil {
		return nil, err
	}

	uploadFilename := displayName
	if formatOverride != "" {
		uploadFilename = fmt.Sprintf("%s.%s", displayName, format)
	}

	return &datasetUploadRequest{
		DisplayName:    displayName,
		LocalPath:      filePath,
		UploadFilename: uploadFilename,
		Format:         format,
	}, nil
}

func resolveDatasetUploadFormat(fileName, formatOverride string) (model.DatasetImportFormat, error) {
	if formatOverride != "" {
		return parseDatasetUploadFormat(formatOverride)
	}

	return detectDatasetUploadFormat(fileName)
}

func parseDatasetUploadFormat(format string) (model.DatasetImportFormat, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case string(model.DatasetImportFormatCSV):
		return model.DatasetImportFormatCSV, nil
	case string(model.DatasetImportFormatJSONL):
		return model.DatasetImportFormatJSONL, nil
	default:
		return "", fmt.Errorf("unsupported format %q; use --format csv or --format jsonl", format)
	}
}

func detectDatasetUploadFormat(fileName string) (model.DatasetImportFormat, error) {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".csv":
		return model.DatasetImportFormatCSV, nil
	case ".jsonl":
		return model.DatasetImportFormatJSONL, nil
	case "":
		return "", errors.New(ErrDatasetUploadFormatRequired)
	default:
		return "", fmt.Errorf("unsupported file extension %q; use --format csv or --format jsonl", filepath.Ext(fileName))
	}
}

// uploadFileToPresignedURL uploads a file to a presigned URL using the API-provided method and content type.
func uploadFileToPresignedURL(filePath, uploadURL, method, contentType string) error {
	if method == "" {
		return errors.New("upload response missing http method")
	}

	if contentType == "" {
		return errors.New("upload response missing content type")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to inspect file %s: %w", filePath, err)
	}

	request, err := http.NewRequestWithContext(context.Background(), method, uploadURL, file)
	if err != nil {
		return err
	}

	request.ContentLength = info.Size()
	request.Header.Set("Content-Type", contentType)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(response.Body)
		if len(body) == 0 {
			return fmt.Errorf("upload request returned status %d", response.StatusCode)
		}

		return fmt.Errorf("upload request returned status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

func validateAudioURLFields(filePath string, format model.DatasetImportFormat, schema *client.DatasetSchema) error {
	return validateMediaURLFields(filePath, format, schema, "audio_url", "audio", validAudioURLFileExtensions, supportedAudioURLFileExtensions)
}

func validateVideoURLFields(filePath string, format model.DatasetImportFormat, schema *client.DatasetSchema) error {
	return validateMediaURLFields(filePath, format, schema, "video_url", "video", validVideoURLFileExtensions, supportedVideoURLFileExtensions)
}

func validateMediaURLFields(
	filePath string,
	format model.DatasetImportFormat,
	schema *client.DatasetSchema,
	fieldType, mediaLabel string,
	extensions map[string]bool,
	supportedExtensions string,
) error {
	if schema == nil {
		return nil
	}

	mediaFields := make(map[string]struct{})
	for fieldName, field := range schema.Fields {
		if field.Type == fieldType {
			mediaFields[fieldName] = struct{}{}
		}
	}

	if len(mediaFields) == 0 {
		return nil
	}

	switch format {
	case model.DatasetImportFormatCSV:
		return validateMediaURLFieldsInCSV(filePath, mediaFields, mediaLabel, extensions, supportedExtensions)
	case model.DatasetImportFormatJSONL:
		return validateMediaURLFieldsInJSONL(filePath, mediaFields, mediaLabel, extensions, supportedExtensions)
	default:
		return nil
	}
}

func validateMediaURLFieldsInCSV(
	filePath string,
	mediaFields map[string]struct{},
	mediaLabel string,
	extensions map[string]bool,
	supportedExtensions string,
) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header from %s: %w", filePath, err)
	}

	mediaColumnIndexes := make(map[int]string)
	for idx, header := range headers {
		fieldName := strings.TrimSpace(header)
		if _, ok := mediaFields[fieldName]; ok {
			mediaColumnIndexes[idx] = fieldName
		}
	}

	if len(mediaColumnIndexes) == 0 {
		return nil
	}

	recordIndex := 1
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV record %d from %s: %w", recordIndex, filePath, err)
		}

		for idx, fieldName := range mediaColumnIndexes {
			if idx >= len(record) {
				continue
			}

			if err := validateMediaURLValue(recordIndex, fieldName, record[idx], mediaLabel, extensions, supportedExtensions); err != nil {
				return err
			}
		}

		recordIndex++
	}
}

func ValidateAudioURLFieldsInJSONL(filePath string, audioFields map[string]struct{}) error {
	return validateMediaURLFieldsInJSONL(filePath, audioFields, "audio", validAudioURLFileExtensions, supportedAudioURLFileExtensions)
}

func ValidateVideoURLFieldsInJSONL(filePath string, videoFields map[string]struct{}) error {
	return validateMediaURLFieldsInJSONL(filePath, videoFields, "video", validVideoURLFileExtensions, supportedVideoURLFileExtensions)
}

func validateMediaURLFieldsInJSONL(
	filePath string,
	mediaFields map[string]struct{},
	mediaLabel string,
	extensions map[string]bool,
	supportedExtensions string,
) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	recordIndex := 1
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			recordIndex++
			continue
		}

		record := make(map[string]any)
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return fmt.Errorf("failed to parse JSONL record %d from %s: %w", recordIndex, filePath, err)
		}

		for fieldName := range mediaFields {
			value, ok := record[fieldName]
			if !ok || value == nil {
				continue
			}

			valueString, ok := value.(string)
			if !ok {
				return fmt.Errorf(
					"record %d field %s: %s URL must be a string ending with one of %s",
					recordIndex,
					fieldName,
					mediaLabel,
					supportedExtensions,
				)
			}

			if err := validateMediaURLValue(recordIndex, fieldName, valueString, mediaLabel, extensions, supportedExtensions); err != nil {
				return err
			}
		}

		recordIndex++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read JSONL file %s: %w", filePath, err)
	}

	return nil
}

func validateMediaURLValue(recordIndex int, fieldName, value, mediaLabel string, extensions map[string]bool, supportedExtensions string) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil
	}

	if !hasSupportedURLExtension(trimmedValue, extensions) {
		return fmt.Errorf(
			"record %d field %s: %s URL %q must end with one of %s",
			recordIndex,
			fieldName,
			mediaLabel,
			trimmedValue,
			supportedExtensions,
		)
	}

	return nil
}

func hasSupportedURLExtension(value string, extensions map[string]bool) bool {
	parsedURL, err := url.ParseRequestURI(value)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	extension := strings.ToLower(filepath.Ext(parsedURL.Path))
	return extensions[extension]
}

func waitForDatasetImport(
	client client.API,
	datasetID, importID string,
	format model.DatasetImportFormat,
	timeout time.Duration,
	w io.Writer,
) error {
	fmt.Fprint(w, "Processing import")

	deadline := time.Now().Add(timeout)
	consecutivePollErrors := 0

	for {
		if time.Now().After(deadline) {
			fmt.Fprintln(w)
			return datasetImportTimeoutError(datasetID, importID, timeout)
		}

		status, err := client.GetAITaskBuilderDatasetImportStatus(datasetID, importID)
		if err != nil {
			consecutivePollErrors++
			if consecutivePollErrors >= datasetUploadMaxConsecutivePollErrors {
				fmt.Fprintln(w)
				return fmt.Errorf(
					"failed to retrieve dataset %s import %s status after %d consecutive attempts: %s",
					datasetID,
					importID,
					datasetUploadMaxConsecutivePollErrors,
					err,
				)
			}

			fmt.Fprint(w, ".")
			if !sleepUntilNextDatasetImportPoll(deadline) {
				fmt.Fprintln(w)
				return datasetImportTimeoutError(datasetID, importID, timeout)
			}
			continue
		}

		consecutivePollErrors = 0
		job := normalizeDatasetImportJob(status.DatasetImportJob, datasetID, importID)

		switch job.Status {
		case model.DatasetImportJobStatusUninitialised,
			model.DatasetImportJobStatusQueued,
			model.DatasetImportJobStatusProcessing:
			fmt.Fprint(w, ".")
			if !sleepUntilNextDatasetImportPoll(deadline) {
				fmt.Fprintln(w)
				return datasetImportTimeoutError(datasetID, importID, timeout)
			}
		case model.DatasetImportJobStatusComplete:
			fmt.Fprintln(w)
			renderDatasetImportSuccess(w, job, format)
			return nil
		case model.DatasetImportJobStatusPartial:
			fmt.Fprintln(w)
			renderDatasetImportPartial(w, job, format)
			return nil
		case model.DatasetImportJobStatusFailed:
			fmt.Fprintln(w)
			return errors.New(formatDatasetImportFailure(job))
		case model.DatasetImportJobStatusPendingSchema:
			fmt.Fprintln(w)
			renderDatasetImportPendingSchema(w, job)
			return nil
		default:
			fmt.Fprintln(w)
			return fmt.Errorf("unexpected import status %q for dataset %s import %s", job.Status, job.DatasetID, job.ImportID)
		}
	}
}

func sleepUntilNextDatasetImportPoll(deadline time.Time) bool {
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return false
	}

	sleepDuration := datasetUploadPollInterval
	if remaining < sleepDuration {
		sleepDuration = remaining
	}

	DatasetUploadPollSleep(sleepDuration)
	return true
}

func datasetImportTimeoutError(datasetID, importID string, timeout time.Duration) error {
	return fmt.Errorf(
		"timed out waiting for dataset %s import %s after %s; processing may still continue server-side",
		datasetID,
		importID,
		timeout,
	)
}

func normalizeDatasetImportJob(job model.DatasetImportJob, datasetID, importID string) model.DatasetImportJob {
	if job.DatasetID == "" {
		job.DatasetID = datasetID
	}

	if job.ImportID == "" {
		job.ImportID = importID
	}

	return job
}

func renderDatasetImportSuccess(w io.Writer, job model.DatasetImportJob, format model.DatasetImportFormat) {
	renderDatasetImportSummary(w, "Import complete.", job, format)
}

func renderDatasetImportPartial(w io.Writer, job model.DatasetImportJob, format model.DatasetImportFormat) {
	renderDatasetImportSummary(w, "Import completed with warnings.", job, format)

	if job.Reason != "" {
		fmt.Fprintf(w, "Warning: %s\n", job.Reason)
	}

	writeDatasetImportRecordErrors(w, job)
}

func renderDatasetImportPendingSchema(w io.Writer, job model.DatasetImportJob) {
	fmt.Fprintf(
		w,
		"Warning: upload was received for dataset %s import %s, but processing is waiting for a dataset schema; define a schema for the dataset rather than re-uploading the file\n",
		job.DatasetID,
		job.ImportID,
	)
}

func renderDatasetImportSummary(w io.Writer, heading string, job model.DatasetImportJob, format model.DatasetImportFormat) {
	acceptedCount := datasetImportCount(job.AcceptedCount)
	rejectedCount := datasetImportCount(job.RejectedCount)
	duplicateCount := datasetImportCount(job.DuplicateCount)
	newlyImportedCount := max(acceptedCount-duplicateCount, 0)

	fmt.Fprintf(w, "%s\n", heading)
	fmt.Fprintf(w, "Dataset ID: %s\n", job.DatasetID)
	fmt.Fprintf(w, "Import ID: %s\n", job.ImportID)
	fmt.Fprintf(w, "Format: %s\n", format)
	fmt.Fprintf(w, "Accepted: %d\n", acceptedCount)
	if duplicateCount > 0 {
		fmt.Fprintf(w, "Newly Imported: %d\n", newlyImportedCount)
		fmt.Fprintf(w, "Duplicates: %d\n", duplicateCount)
	} else {
		fmt.Fprintf(w, "Imported: %d\n", acceptedCount)
	}
	fmt.Fprintf(w, "Rejected: %d\n", rejectedCount)

	if duplicateCount > 0 && acceptedCount == duplicateCount {
		fmt.Fprintln(w, "Note: all accepted records were duplicates; no new records were imported.")
	}
}

func formatDatasetImportFailure(job model.DatasetImportJob) string {
	var summary strings.Builder

	reason := job.Reason
	if reason == "" {
		reason = "processing failed"
	}

	fmt.Fprintf(
		&summary,
		"dataset import failed for dataset %s import %s: %s",
		job.DatasetID,
		job.ImportID,
		reason,
	)

	appendDatasetImportRecordErrors(&summary, job)

	return summary.String()
}

func writeDatasetImportRecordErrors(w io.Writer, job model.DatasetImportJob) {
	var summary strings.Builder
	appendDatasetImportRecordErrors(&summary, job)
	if summary.Len() == 0 {
		return
	}

	fmt.Fprint(w, summary.String())
}

func appendDatasetImportRecordErrors(w io.StringWriter, job model.DatasetImportJob) {
	if len(job.Errors) == 0 {
		return
	}

	errorsToDisplay := slices.Min([]int{len(job.Errors), datasetUploadRecordErrorDisplayLimit})
	_, _ = w.WriteString("Record errors:\n")
	for _, recordError := range job.Errors[:errorsToDisplay] {
		if recordError.Field != "" {
			_, _ = w.WriteString(
				fmt.Sprintf("- record %d field %s: %s\n", recordError.RecordIndex+1, recordError.Field, recordError.Reason),
			)
			continue
		}

		_, _ = w.WriteString(fmt.Sprintf("- record %d: %s\n", recordError.RecordIndex+1, recordError.Reason))
	}

	rejectedCount := datasetImportCount(job.RejectedCount)
	if rejectedCount > len(job.Errors) || len(job.Errors) > datasetUploadRecordErrorDisplayLimit {
		_, _ = w.WriteString(
			fmt.Sprintf(
				"Detailed errors are a partial view: %d records were rejected but only %d detailed errors were returned.\n",
				rejectedCount,
				errorsToDisplay,
			),
		)
	}
}

func datasetImportCount(value *int) int {
	if value == nil {
		return 0
	}

	return *value
}
