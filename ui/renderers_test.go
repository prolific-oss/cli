package ui_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
)

// testCredPoolID is a test fixture representing a credential pool ID in the format {workspace_id}_{uuid}
const testCredPoolID = "679271425fe00981084a5f58_a856d700-c495-11f0-adce-338d4126f6e8"

func TestCsvRendererRendersInCsvFormat(t *testing.T) {
	studies := []model.Study{
		{
			ID:                      "1234",
			Name:                    "My first, standard, sample",
			InternalName:            "Standard sample",
			Desc:                    "This is my first standard sample study on the Prolific system.",
			Status:                  model.StatusActive,
			ExternalStudyURL:        "https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}",
			TotalAvailablePlaces:    10,
			EstimatedCompletionTime: 10,
			MaximumAllowedTime:      10,
			Reward:                  400,
			DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
		},
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	r := ui.CsvRenderer[model.Study]{}
	err := r.Render(studies, "ID,Name,Status", writer)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	expected := "ID,Name,Status\n1234,\"My first, standard, sample\",active\n"

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}

func TestCsvRendererRendersInCsvFormatRespectingFieldOrder(t *testing.T) {
	studies := []model.Study{
		{
			ID:                      "1234",
			Name:                    "My first, standard, sample",
			InternalName:            "Standard sample",
			Desc:                    "This is my first standard sample study on the Prolific system.",
			Status:                  model.StatusActive,
			ExternalStudyURL:        "https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}",
			TotalAvailablePlaces:    10,
			EstimatedCompletionTime: 10,
			MaximumAllowedTime:      10,
			Reward:                  400,
			DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
		},
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	r := ui.CsvRenderer[model.Study]{}
	err := r.Render(studies, "ID,Status", writer)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	expected := "ID,Status\n1234,active\n"

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}

func TestCsvRendererRendersCredentialPoolID(t *testing.T) {
	studies := []model.Study{
		{
			ID:                      "1234",
			Name:                    "Study with credential pool",
			InternalName:            "Study with credential pool",
			Desc:                    "This study demonstrates how to attach a credential pool for participant authentication",
			Status:                  model.StatusActive,
			ExternalStudyURL:        "https://example.com/my-study-id",
			TotalAvailablePlaces:    50,
			EstimatedCompletionTime: 15,
			MaximumAllowedTime:      30,
			Reward:                  600,
			DeviceCompatibility:     []string{"desktop"},
			CredentialPoolID:        testCredPoolID,
		},
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	r := ui.CsvRenderer[model.Study]{}
	err := r.Render(studies, "ID,Name,CredentialPoolID", writer)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	expected := "ID,Name,CredentialPoolID\n1234,Study with credential pool,679271425fe00981084a5f58_a856d700-c495-11f0-adce-338d4126f6e8\n"

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}

func TestTableRendererRendersCredentialPoolID(t *testing.T) {
	studies := []model.Study{
		{
			ID:                      "1234",
			Name:                    "Study with credential pool",
			InternalName:            "Study with credential pool",
			Desc:                    "This study demonstrates how to attach a credential pool for participant authentication",
			Status:                  model.StatusActive,
			ExternalStudyURL:        "https://example.com/my-study-id",
			TotalAvailablePlaces:    50,
			EstimatedCompletionTime: 15,
			MaximumAllowedTime:      30,
			Reward:                  600,
			DeviceCompatibility:     []string{"desktop"},
			CredentialPoolID:        testCredPoolID,
		},
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	r := ui.TableRenderer[model.Study]{}
	err := r.Render(studies, "ID,Name,CredentialPoolID", writer)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	output := b.String()

	if !strings.Contains(output, "ID") || !strings.Contains(output, "Name") || !strings.Contains(output, "CredentialPoolID") {
		t.Fatalf("expected headers ID, Name, CredentialPoolID in output, got '%v'", output)
	}

	if !strings.Contains(output, "1234") || !strings.Contains(output, "Study with credential pool") || !strings.Contains(output, testCredPoolID) {
		t.Fatalf("expected study data in output, got '%v'", output)
	}
}
