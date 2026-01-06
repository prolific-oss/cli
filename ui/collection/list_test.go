package collection_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui/collection"
)

type csvRendererTestSetup struct {
	client *mock_client.MockAPI
	opts   collection.ListUsedOptions
	buffer *bytes.Buffer
	writer *bufio.Writer
}

func setupCsvRendererTest(t *testing.T, collectionName string) csvRendererTestSetup {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	c := mock_client.NewMockAPI(ctrl)

	opts := collection.ListUsedOptions{
		WorkspaceID: "workspace-123",
		Limit:       200,
		Offset:      0,
	}

	actualCollection := model.Collection{
		ID:        "coll-1234",
		Name:      collectionName,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		CreatedBy: "user@example.com",
		ItemCount: 5,
	}
	collectionResponse := client.ListCollectionsResponse{
		Results: []model.Collection{actualCollection},
	}

	c.
		EXPECT().
		GetCollections(gomock.Eq(opts.WorkspaceID), gomock.Eq(opts.Limit), gomock.Eq(opts.Offset)).
		Return(&collectionResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	return csvRendererTestSetup{
		client: c,
		opts:   opts,
		buffer: &b,
		writer: writer,
	}
}

func TestCsvRendererRendersInCsvFormat(t *testing.T) {
	setup := setupCsvRendererTest(t, "My first collection")

	renderer := collection.CsvRenderer{}
	err := renderer.Render(setup.client, setup.opts, setup.writer)

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	setup.writer.Flush()

	expected := `ID,Name,ItemCount,
coll-1234,My first collection,5,
`

	if setup.buffer.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, setup.buffer.String())
	}
}

func TestCsvRendererReturnsErrorIfCannotGetCollections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := collection.ListUsedOptions{
		WorkspaceID: "workspace-123",
		Limit:       200,
		Offset:      0,
	}

	expected := errors.New("something went wrong")

	c.
		EXPECT().
		GetCollections(gomock.Eq(opts.WorkspaceID), gomock.Eq(opts.Limit), gomock.Eq(opts.Offset)).
		Return(nil, expected).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := collection.CsvRenderer{}
	err := renderer.Render(c, opts, writer)

	if err != expected {
		t.Fatalf("Expected \n%v\n, got \n%v\n", expected, err)
	}

	writer.Flush()
}

func TestCsvRendererRendersInCsvFormatRespectingFieldOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := collection.ListUsedOptions{
		WorkspaceID: "workspace-123",
		Fields:      "ID,Name",
		Limit:       200,
		Offset:      0,
	}

	actualCollection := model.Collection{
		ID:        "coll-1234",
		Name:      "My first collection",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		CreatedBy: "user@example.com",
		ItemCount: 5,
	}
	collectionResponse := client.ListCollectionsResponse{
		Results: []model.Collection{actualCollection},
	}

	c.
		EXPECT().
		GetCollections(gomock.Eq(opts.WorkspaceID), gomock.Eq(opts.Limit), gomock.Eq(opts.Offset)).
		Return(&collectionResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := collection.CsvRenderer{}
	err := renderer.Render(c, opts, writer)

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	expected := `ID,Name,
coll-1234,My first collection,
`

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}

func TestCsvRendererHandlesCommasInName(t *testing.T) {
	setup := setupCsvRendererTest(t, "My first, standard, collection")

	renderer := collection.CsvRenderer{}
	err := renderer.Render(setup.client, setup.opts, setup.writer)

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	setup.writer.Flush()

	expected := `ID,Name,ItemCount,
coll-1234,"My first, standard, collection",5,
`

	if setup.buffer.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, setup.buffer.String())
	}
}

func TestNonInteractiveRendererRendersCollections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := collection.ListUsedOptions{
		WorkspaceID: "workspace-123",
		Fields:      "ID,Name,ItemCount",
		Limit:       200,
		Offset:      0,
	}

	actualCollection := model.Collection{
		ID:        "coll-1234",
		Name:      "My collection",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		CreatedBy: "user@example.com",
		ItemCount: 10,
	}
	collectionResponse := client.ListCollectionsResponse{
		Results: []model.Collection{actualCollection},
	}

	c.
		EXPECT().
		GetCollections(gomock.Eq(opts.WorkspaceID), gomock.Eq(opts.Limit), gomock.Eq(opts.Offset)).
		Return(&collectionResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := collection.NonInteractiveRenderer{}
	err := renderer.Render(c, opts, writer)

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	output := b.String()

	if !strings.Contains(output, "ID") || !strings.Contains(output, "Name") || !strings.Contains(output, "ItemCount") {
		t.Fatalf("expected headers ID, Name, ItemCount in output, got '%v'", output)
	}

	if !strings.Contains(output, "coll-1234") || !strings.Contains(output, "My collection") || !strings.Contains(output, "10") {
		t.Fatalf("expected collection data in output, got '%v'", output)
	}
}

func TestJSONRendererRendersInJSONFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := collection.ListUsedOptions{
		WorkspaceID: "workspace-123",
		Limit:       200,
		Offset:      0,
	}

	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	actualCollection := model.Collection{
		ID:        "coll-1234",
		Name:      "My first collection",
		CreatedAt: createdAt,
		CreatedBy: "user@example.com",
		ItemCount: 5,
	}
	collectionResponse := client.ListCollectionsResponse{
		Results: []model.Collection{actualCollection},
	}

	c.
		EXPECT().
		GetCollections(gomock.Eq(opts.WorkspaceID), gomock.Eq(opts.Limit), gomock.Eq(opts.Offset)).
		Return(&collectionResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := collection.JSONRenderer{}
	err := renderer.Render(c, opts, writer)

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	var result []model.Collection
	if err := json.Unmarshal(b.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 collection, got %d", len(result))
	}

	if result[0].ID != "coll-1234" {
		t.Fatalf("expected ID 'coll-1234', got '%s'", result[0].ID)
	}

	if result[0].Name != "My first collection" {
		t.Fatalf("expected Name 'My first collection', got '%s'", result[0].Name)
	}

	if result[0].ItemCount != 5 {
		t.Fatalf("expected ItemCount 5, got %d", result[0].ItemCount)
	}
}

func TestJSONRendererReturnsErrorIfCannotGetCollections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := collection.ListUsedOptions{
		WorkspaceID: "workspace-123",
		Limit:       200,
		Offset:      0,
	}

	expected := errors.New("something went wrong")

	c.
		EXPECT().
		GetCollections(gomock.Eq(opts.WorkspaceID), gomock.Eq(opts.Limit), gomock.Eq(opts.Offset)).
		Return(nil, expected).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := collection.JSONRenderer{}
	err := renderer.Render(c, opts, writer)

	if err != expected {
		t.Fatalf("Expected \n%v\n, got \n%v\n", expected, err)
	}

	writer.Flush()
}
