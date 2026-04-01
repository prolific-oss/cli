package study_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewDemographicExportCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := study.NewDemographicExportCommand(c, os.Stdout)

	use := "demographic-export <study-id>"
	short := "Trigger a demographic data export for a study"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestDemographicExportCallsClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	response := client.DemographicExportResponse{
		ID:     "export-5566",
		Status: "generating",
	}

	c.
		EXPECT().
		ExportDemographics(gomock.Eq(studyID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewDemographicExportCommand(c, writer)
	_ = cmd.RunE(cmd, []string{studyID})
	writer.Flush()

	expected := fmt.Sprintf("%s\n", response.ID)
	actual := b.String()

	if actual != expected {
		t.Fatalf("expected \n'%s'\ngot\n'%s'", expected, actual)
	}
}

func TestDemographicExportHandlesApiErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	c.
		EXPECT().
		ExportDemographics(gomock.Eq(studyID)).
		Return(nil, errors.New("export failed")).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewDemographicExportCommand(c, writer)
	err := cmd.RunE(cmd, []string{studyID})
	writer.Flush()

	expected := "error: export failed"
	if err.Error() != expected {
		t.Fatalf("expected %s, got %s", expected, err.Error())
	}
}
