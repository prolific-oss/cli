package aitaskbuilder

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/prolific-oss/cli/client"
)

// validDatasetSchemaFieldTypes lists the field types accepted by the API.
var validDatasetSchemaFieldTypes = map[string]bool{
	"text":          true,
	"image_url":     true,
	"metadata":      true,
	"task_group_id": true,
	"audio_url":     true,
}

// rawDatasetSchema mirrors DatasetSchema but distinguishes an absent "strict"
// key from an explicit false, so we can layer the --strict flag on top.
type rawDatasetSchema struct {
	Strict *bool                                `json:"strict"`
	Fields map[string]client.DatasetSchemaField `json:"fields"`
}

// resolveDatasetSchema builds the schema payload from the --schema value and --strict flag.
// Returns (nil, nil) when schemaInput is empty
func resolveDatasetSchema(schemaInput string, strict, strictSet bool) (*client.DatasetSchema, error) {
	if schemaInput == "" {
		if strictSet {
			return nil, errors.New(ErrStrictRequiresSchema)
		}
		return nil, nil
	}

	raw, err := readDatasetSchemaInput(schemaInput)
	if err != nil {
		return nil, err
	}

	parsed, err := parseDatasetSchema(raw)
	if err != nil {
		return nil, err
	}

	if len(parsed.Fields) == 0 {
		return nil, errors.New(ErrSchemaFieldsRequired)
	}

	taskGroupIDCount := 0
	for key, field := range parsed.Fields {
		if !validDatasetSchemaFieldTypes[field.Type] {
			return nil, fmt.Errorf("field %q has invalid type %q; must be one of text, image_url, metadata, task_group_id, audio_url", key, field.Type)
		}
		if field.Type == "task_group_id" {
			taskGroupIDCount++
		}
	}
	if taskGroupIDCount > 1 {
		return nil, errors.New(ErrSchemaMultipleTaskGroupID)
	}

	// Resolve strict. When the researcher does not specify it in either place,
	// default to false so the payload always satisfies the API contract.
	var strictValue *bool
	switch {
	case parsed.Strict != nil && strictSet:
		return nil, errors.New(ErrSchemaStrictSetInBoth)
	case parsed.Strict != nil:
		strictValue = parsed.Strict
	case strictSet:
		strictValue = &strict
	default:
		strictValue = &strict
	}

	return &client.DatasetSchema{Strict: strictValue, Fields: parsed.Fields}, nil
}

func parseDatasetSchema(raw []byte) (*rawDatasetSchema, error) {
	var parsed rawDatasetSchema

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&parsed); err != nil {
		if strings.HasPrefix(err.Error(), "json: unknown field ") {
			return nil, fmt.Errorf("schema contains %s", strings.TrimPrefix(err.Error(), "json: "))
		}
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, errors.New(ErrSchemaInvalidJSON)
		}
		return nil, errors.New(ErrSchemaMustBeObject)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, errors.New(ErrSchemaMustBeObject)
	}

	return &parsed, nil
}

// readDatasetSchemaInput returns the raw schema bytes. A trimmed value starting
// with "{" is treated as inline JSON; otherwise it is read from a file path.
func readDatasetSchemaInput(schemaInput string) ([]byte, error) {
	if strings.HasPrefix(strings.TrimSpace(schemaInput), "{") {
		return []byte(schemaInput), nil
	}

	raw, err := os.ReadFile(schemaInput)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("schema file does not exist: %s", schemaInput)
		}
		return nil, fmt.Errorf("failed to read schema file %s: %w", schemaInput, err)
	}
	return raw, nil
}
