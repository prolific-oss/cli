package study_test

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/benmatselby/prolificli/ui/study"
	"github.com/golang/mock/gomock"
)

func TestCsvRendererRendersInCsvFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := study.ListUsedOptions{
		Status: model.StatusActive,
	}

	actualStudy := model.Study{
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
	}
	studyResponse := client.ListStudiesResponse{
		Results: []model.Study{actualStudy},
	}

	c.
		EXPECT().
		GetStudies(gomock.Eq(opts.Status)).
		Return(&studyResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := study.CsvRenderer{}
	err := renderer.Render(c, opts, writer)

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	expected := `ID,Name,Status,
1234,"My first, standard, sample",active,
`

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}

func TestCsvRendererRendersReturnsErrorIfCannotGetStudies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := study.ListUsedOptions{
		Status: model.StatusActive,
	}

	expected := errors.New("What in the blazes!!!")

	c.
		EXPECT().
		GetStudies(gomock.Eq(opts.Status)).
		Return(nil, expected).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := study.CsvRenderer{}
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

	opts := study.ListUsedOptions{
		Status: model.StatusActive,
		Fields: "ID,Status",
	}

	actualStudy := model.Study{
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
	}
	studyResponse := client.ListStudiesResponse{
		Results: []model.Study{actualStudy},
	}

	c.
		EXPECT().
		GetStudies(gomock.Eq(opts.Status)).
		Return(&studyResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := study.CsvRenderer{}
	err := renderer.Render(c, opts, writer)

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	expected := `ID,Status,
1234,active,
`

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}
