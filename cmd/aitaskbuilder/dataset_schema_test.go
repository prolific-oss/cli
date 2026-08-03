package aitaskbuilder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// minimalSchemaJSON is a valid single-field schema with no "strict" key, reused
// across the strict-resolution tests.
const minimalSchemaJSON = `{"fields":{"q":{"type":"text"}}}`

func TestResolveDatasetSchemaInlineValid(t *testing.T) {
	input := `{
		"strict": true,
		"fields": {
			"question": { "type": "text", "label": "Question" },
			"image":    { "type": "image_url" },
			"audio":    { "type": "audio_url" },
			"video":    { "type": "video_url" },
			"source":   { "type": "metadata" },
			"group":    { "type": "task_group_id" }
		}
	}`

	schema, err := resolveDatasetSchema(input, false, false)
	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
	if schema == nil {
		t.Fatal("expected schema; got nil")
	}
	if schema.Strict == nil || !*schema.Strict {
		t.Fatal("expected strict to be true from JSON")
	}
	if len(schema.Fields) != 6 {
		t.Fatalf("expected 6 fields; got %d", len(schema.Fields))
	}
	if schema.Fields["audio"].Type != "audio_url" {
		t.Fatalf("unexpected audio field: %+v", schema.Fields["audio"])
	}
	if schema.Fields["video"].Type != "video_url" {
		t.Fatalf("unexpected video field: %+v", schema.Fields["video"])
	}
	if schema.Fields["question"].Type != "text" || schema.Fields["question"].Label != "Question" {
		t.Fatalf("unexpected question field: %+v", schema.Fields["question"])
	}
}

func TestResolveDatasetSchemaFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	content := `{"fields":{"question":{"type":"text"}}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	schema, err := resolveDatasetSchema(path, false, false)
	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
	if schema == nil || len(schema.Fields) != 1 {
		t.Fatalf("expected one field from file; got %+v", schema)
	}
}

func TestResolveDatasetSchemaLeadingWhitespaceIsInline(t *testing.T) {
	input := "   \n\t{\"fields\":{\"question\":{\"type\":\"text\"}}}"
	schema, err := resolveDatasetSchema(input, false, false)
	if err != nil {
		t.Fatalf("expected inline detection with leading whitespace; got %v", err)
	}
	if schema == nil {
		t.Fatal("expected schema; got nil")
	}
}

func TestResolveDatasetSchemaMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	_, err := resolveDatasetSchema(path, false, false)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("expected error to include path %q; got %v", path, err)
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected file-not-found error; got %v", err)
	}
}

func TestResolveDatasetSchemaInvalidJSON(t *testing.T) {
	_, err := resolveDatasetSchema(`{not valid json`, false, false)
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
	if err.Error() != ErrSchemaInvalidJSON {
		t.Fatalf("expected %q; got %v", ErrSchemaInvalidJSON, err)
	}
}

func TestResolveDatasetSchemaNonObject(t *testing.T) {
	// Leading "{" so it is treated as inline, but it is not a valid object.
	_, err := resolveDatasetSchema(`{}[]`, false, false)
	if err == nil || err.Error() != ErrSchemaMustBeObject {
		t.Fatalf("expected %q; got %v", ErrSchemaMustBeObject, err)
	}
}

func TestResolveDatasetSchemaTypeMismatch(t *testing.T) {
	_, err := resolveDatasetSchema(`{"fields":[]}`, false, false)
	if err == nil {
		t.Fatal("expected invalid JSON type error")
	}
	if err.Error() != ErrSchemaInvalidJSON {
		t.Fatalf("expected %q; got %v", ErrSchemaInvalidJSON, err)
	}
}

func TestResolveDatasetSchemaEmptyFields(t *testing.T) {
	_, err := resolveDatasetSchema(`{"fields":{}}`, false, false)
	if err == nil || err.Error() != ErrSchemaFieldsRequired {
		t.Fatalf("expected %q; got %v", ErrSchemaFieldsRequired, err)
	}
}

func TestResolveDatasetSchemaMissingFields(t *testing.T) {
	_, err := resolveDatasetSchema(`{"strict":true}`, false, false)
	if err == nil || err.Error() != ErrSchemaFieldsRequired {
		t.Fatalf("expected %q; got %v", ErrSchemaFieldsRequired, err)
	}
}

func TestResolveDatasetSchemaUnknownTopLevelField(t *testing.T) {
	_, err := resolveDatasetSchema(`{"bla":false,"fields":{"q":{"type":"text"}}}`, false, false)
	if err == nil {
		t.Fatal("expected unknown top-level field error")
	}
	if !strings.Contains(err.Error(), `unknown field "bla"`) {
		t.Fatalf("expected unknown field error for bla; got %v", err)
	}
}

func TestResolveDatasetSchemaUnknownFieldProperty(t *testing.T) {
	_, err := resolveDatasetSchema(`{"fields":{"q":{"type":"text","extra":true}}}`, false, false)
	if err == nil {
		t.Fatal("expected unknown field property error")
	}
	if !strings.Contains(err.Error(), `unknown field "extra"`) {
		t.Fatalf("expected unknown field error for extra; got %v", err)
	}
}

func TestResolveDatasetSchemaInvalidFieldType(t *testing.T) {
	_, err := resolveDatasetSchema(`{"fields":{"q":{"type":"number"}}}`, false, false)
	if err == nil {
		t.Fatal("expected invalid field type error")
	}
	msg := err.Error()
	if !strings.Contains(msg, `"q"`) || !strings.Contains(msg, `"number"`) {
		t.Fatalf("expected error to name field and type; got %v", err)
	}
	if !strings.Contains(msg, "text, image_url, metadata, task_group_id, audio_url, video_url") {
		t.Fatalf("expected error to list allowed types; got %v", err)
	}
}

func TestResolveDatasetSchemaMultipleTaskGroupID(t *testing.T) {
	input := `{"fields":{"a":{"type":"task_group_id"},"b":{"type":"task_group_id"}}}`
	_, err := resolveDatasetSchema(input, false, false)
	if err == nil || err.Error() != ErrSchemaMultipleTaskGroupID {
		t.Fatalf("expected %q; got %v", ErrSchemaMultipleTaskGroupID, err)
	}
}

func TestResolveDatasetSchemaStrictBothSet(t *testing.T) {
	input := `{"strict":true,"fields":{"q":{"type":"text"}}}`
	_, err := resolveDatasetSchema(input, true, true)
	if err == nil || err.Error() != ErrSchemaStrictSetInBoth {
		t.Fatalf("expected %q; got %v", ErrSchemaStrictSetInBoth, err)
	}
}

func TestResolveDatasetSchemaStrictFromJSONOnly(t *testing.T) {
	input := `{"strict":false,"fields":{"q":{"type":"text"}}}`
	// --strict not passed; JSON value (false) wins and is sent explicitly.
	schema, err := resolveDatasetSchema(input, false, false)
	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
	if schema.Strict == nil || *schema.Strict {
		t.Fatal("expected strict false from JSON")
	}

	input = `{"strict":true,"fields":{"q":{"type":"text"}}}`
	schema, err = resolveDatasetSchema(input, false, false)
	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
	if schema.Strict == nil || !*schema.Strict {
		t.Fatal("expected strict true from JSON")
	}
}

func TestResolveDatasetSchemaStrictFromFlagOnly(t *testing.T) {
	input := minimalSchemaJSON
	schema, err := resolveDatasetSchema(input, true, true)
	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
	if schema.Strict == nil || !*schema.Strict {
		t.Fatal("expected strict true from --strict flag")
	}
}

func TestResolveDatasetSchemaStrictDefaultsFalseWhenUnspecified(t *testing.T) {
	input := minimalSchemaJSON
	schema, err := resolveDatasetSchema(input, false, false)
	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
	if schema.Strict == nil || *schema.Strict {
		t.Fatal("expected strict false by default")
	}
}

func TestResolveDatasetSchemaStrictExplicitFalseFlag(t *testing.T) {
	// --strict=false explicitly passed: honour it and send false.
	input := minimalSchemaJSON
	schema, err := resolveDatasetSchema(input, false, true)
	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
	if schema.Strict == nil || *schema.Strict {
		t.Fatal("expected strict false from explicit --strict=false")
	}
}

func TestResolveDatasetSchemaMarshalsStrictWhenUnspecified(t *testing.T) {
	schema, err := resolveDatasetSchema(minimalSchemaJSON, false, false)
	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}
	if !strings.Contains(string(raw), `"strict":false`) {
		t.Fatalf("expected marshalled schema to include strict false; got %s", raw)
	}
}

func TestResolveDatasetSchemaEmptyInputWithStrict(t *testing.T) {
	_, err := resolveDatasetSchema("", false, true)
	if err == nil || err.Error() != ErrStrictRequiresSchema {
		t.Fatalf("expected %q; got %v", ErrStrictRequiresSchema, err)
	}
}

func TestResolveDatasetSchemaEmptyInputNoFlag(t *testing.T) {
	schema, err := resolveDatasetSchema("", false, false)
	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
	if schema != nil {
		t.Fatalf("expected nil schema when --schema is omitted; got %+v", schema)
	}
}
