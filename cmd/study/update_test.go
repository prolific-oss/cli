package study_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "update-test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	return tmpFile.Name()
}

func TestNewUpdateCommandRendersBasicUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewUpdateCommand(client, os.Stdout)

	use := "update <study_id>"
	short := "Update a study using a JSON template"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewUpdateCommandRequiresTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewUpdateCommand(c, writer)
	cmd.SetArgs([]string{"study-123"})
	err := cmd.Execute()

	if err == nil || !strings.Contains(err.Error(), `required flag(s) "template" not set`) {
		t.Fatalf("expected required flag error, got '%v'", err)
	}
}

func TestNewUpdateCommandHandlesFileNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewUpdateCommand(c, writer)
	_ = cmd.Flags().Set("template", "/nonexistent/path/updates.json")
	err := cmd.RunE(cmd, []string{"study-123"})

	if err == nil || !strings.Contains(err.Error(), "error reading template file") {
		t.Fatalf("expected file read error, got '%v'", err)
	}
}

func TestNewUpdateCommandHandlesInvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	path := writeTempJSON(t, "{bad json")

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewUpdateCommand(c, writer)
	_ = cmd.Flags().Set("template", path)
	err := cmd.RunE(cmd, []string{"study-123"})

	if err == nil || !strings.Contains(err.Error(), "error parsing JSON template") {
		t.Fatalf("expected JSON parse error, got '%v'", err)
	}
}

func TestNewUpdateCommandHandlesEmptyPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	path := writeTempJSON(t, `{}`)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewUpdateCommand(c, writer)
	_ = cmd.Flags().Set("template", path)
	err := cmd.RunE(cmd, []string{"study-123"})

	expected := "error: template contains no fields to update"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error '%s', got '%v'", expected, err)
	}
}

func TestNewUpdateCommandUpdatesStudyFromFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"
	expectedPayload := map[string]interface{}{
		"internal_name":          "Updated Study",
		"total_available_places": float64(150),
	}

	updatedStudy := model.Study{
		ID:                   studyID,
		Name:                 "My Study",
		InternalName:         "Updated Study",
		TotalAvailablePlaces: 150,
	}

	c.
		EXPECT().
		UpdateStudy(gomock.Eq(studyID), gomock.Eq(expectedPayload)).
		Return(&updatedStudy, nil).
		Times(1)

	path := writeTempJSON(t, `{"internal_name": "Updated Study", "total_available_places": 150}`)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewUpdateCommand(c, writer)
	_ = cmd.Flags().Set("template", path)
	err := cmd.RunE(cmd, []string{studyID})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writer.Flush()

	if b.Len() == 0 {
		t.Fatal("expected output, got empty")
	}
}

func TestNewUpdateCommandUpdatesStudyFromStdin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"
	expectedPayload := map[string]interface{}{
		"internal_name": "Stdin Update",
	}

	updatedStudy := model.Study{
		ID:           studyID,
		Name:         "My Study",
		InternalName: "Stdin Update",
	}

	c.
		EXPECT().
		UpdateStudy(gomock.Eq(studyID), gomock.Eq(expectedPayload)).
		Return(&updatedStudy, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewUpdateCommand(c, writer)
	_ = cmd.Flags().Set("template", "-")
	cmd.SetIn(strings.NewReader(`{"internal_name": "Stdin Update"}`))

	err := cmd.RunE(cmd, []string{studyID})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writer.Flush()

	if b.Len() == 0 {
		t.Fatal("expected output, got empty")
	}
}

func TestNewUpdateCommandOutputsJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	updatedStudy := model.Study{
		ID:           studyID,
		Name:         "My Study",
		InternalName: "JSON Test",
	}

	c.
		EXPECT().
		UpdateStudy(gomock.Eq(studyID), gomock.Any()).
		Return(&updatedStudy, nil).
		Times(1)

	path := writeTempJSON(t, `{"internal_name": "JSON Test"}`)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewUpdateCommand(c, writer)
	_ = cmd.Flags().Set("template", path)
	_ = cmd.Flags().Set("json", "true")
	err := cmd.RunE(cmd, []string{studyID})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writer.Flush()

	var result map[string]interface{}
	if err := json.Unmarshal(b.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v\noutput: %s", err, b.String())
	}

	if result["id"] != studyID {
		t.Fatalf("expected id %s, got %v", studyID, result["id"])
	}
}

func TestNewUpdateCommandSilentMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	updatedStudy := model.Study{
		ID:           studyID,
		InternalName: "Silent Test",
	}

	c.
		EXPECT().
		UpdateStudy(gomock.Eq(studyID), gomock.Any()).
		Return(&updatedStudy, nil).
		Times(1)

	path := writeTempJSON(t, `{"internal_name": "Silent Test"}`)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewUpdateCommand(c, writer)
	_ = cmd.Flags().Set("template", path)
	_ = cmd.Flags().Set("silent", "true")
	err := cmd.RunE(cmd, []string{studyID})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writer.Flush()

	if b.String() != "" {
		t.Fatalf("expected no output in silent mode, got: %s", b.String())
	}
}

func TestNewUpdateCommandHandlesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	c.
		EXPECT().
		UpdateStudy(gomock.Eq(studyID), gomock.Any()).
		Return(nil, errors.New("unable to update study: request failed: field not allowed")).
		Times(1)

	path := writeTempJSON(t, `{"reward": 100}`)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewUpdateCommand(c, writer)
	_ = cmd.Flags().Set("template", path)
	err := cmd.RunE(cmd, []string{studyID})

	if err == nil {
		t.Fatal("expected error from API, got nil")
	}

	if !strings.Contains(err.Error(), "unable to update study") {
		t.Fatalf("expected API error message, got: %s", err.Error())
	}
}
