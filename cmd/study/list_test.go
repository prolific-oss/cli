package study_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewViewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewListCommand("list", client, os.Stdout)

	use := "list"
	short := "List all of your studies"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestListCommandPassesBothProjectIDAndStatusToGetStudies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	projectID := "6655b8281cc82a88996f0bbb"
	status := model.StatusUnpublished

	studyResponse := client.ListStudiesResponse{
		Results: []model.Study{},
	}

	mockClient.
		EXPECT().
		GetStudies(gomock.Eq(status), gomock.Eq(projectID)).
		Return(&studyResponse, nil).
		Times(1)

	cmd := study.NewListCommand("list", mockClient, os.Stdout)

	_ = cmd.Flags().Set("project", projectID)
	_ = cmd.Flags().Set("status", status)
	_ = cmd.Flags().Set("non-interactive", "true")

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestListCommandFiltersByUnderpaying(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	underpayingStudy := model.Study{
		ID:                      "1234",
		Name:                    "An underpaying study",
		InternalName:            "Underpaying study",
		Desc:                    "An underpaying study.",
		Status:                  model.StatusActive,
		ExternalStudyURL:        "https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}",
		TotalAvailablePlaces:    10,
		EstimatedCompletionTime: 10,
		MaximumAllowedTime:      10,
		Reward:                  400,
		DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
		IsUnderpaying:           true,
	}
	payingFineStudy := model.Study{
		ID:                      "5678",
		Name:                    "Normal paying study",
		InternalName:            "Normal pay",
		Desc:                    "A normal paying study.",
		Status:                  model.StatusActive,
		ExternalStudyURL:        "https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}",
		TotalAvailablePlaces:    10,
		EstimatedCompletionTime: 10,
		MaximumAllowedTime:      10,
		Reward:                  400,
		DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
		IsUnderpaying:           false,
	}
	studyResponse := client.ListStudiesResponse{
		Results: []model.Study{underpayingStudy, payingFineStudy},
	}

	c.
		EXPECT().
		GetStudies(gomock.Eq(model.StatusAll), gomock.Eq("")).
		Return(&studyResponse, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewListCommand("list", c, writer)

	_ = cmd.Flags().Set("underpaying", "true")
	_ = cmd.Flags().Set("csv", "true")

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	writer.Flush()

	expected := "ID,Name,Status\n1234,An underpaying study,active\n"

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}

func TestListRawPreservesCompleteAPIResponse(t *testing.T) {
	response := json.RawMessage(`{
		"results": [{
			"id": "study-1",
			"unknown_field": {"nested": [null, false, {"value": "<new>&value"}]},
			"unknown_integer": 9007199254740993,
			"unknown_decimal": 1.2300e+10
		}],
		"meta": {"count": 1, "unknown_meta": "preserved"},
		"_links": {"next": {"href": null}},
		"unknown_top_level": {"enabled": true}
	}`)

	tests := []struct {
		name        string
		commandName string
		args        []string
	}{
		{name: "raw", commandName: "list", args: []string{"--raw"}},
		{name: "raw with json", commandName: "list", args: []string{"--json", "--raw"}},
		{name: "studies alias", commandName: "studies", args: []string{"--raw"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			c.EXPECT().GetStudiesRaw(model.StatusAll, "").Return(response, nil)
			var output bytes.Buffer
			writer := bufio.NewWriter(&output)
			cmd := study.NewListCommand(tt.commandName, c, writer)
			cmd.SetArgs(tt.args)

			require.NoError(t, cmd.Execute())
			require.NoError(t, writer.Flush())

			assert.JSONEq(t, string(response), output.String())
			assert.Contains(t, output.String(), "9007199254740993")
			assert.Contains(t, output.String(), "1.2300e+10")
			assert.Contains(t, output.String(), "<new>&value")
			assert.NotContains(t, output.String(), "date_created")
			assert.NotContains(t, output.String(), "internal_name")
		})
	}
}

func TestListRawPassesProjectAndStatus(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().GetStudiesRaw(model.StatusUnpublished, "project-1").
		Return(json.RawMessage(`{"results":[]}`), nil)
	var output bytes.Buffer
	cmd := study.NewListCommand("list", c, &output)
	cmd.SetArgs([]string{"--raw", "--project", "project-1", "--status", model.StatusUnpublished})

	require.NoError(t, cmd.Execute())
	assert.JSONEq(t, `{"results":[]}`, output.String())
}

func TestListRawPreservesEmptyAndNullValues(t *testing.T) {
	for _, response := range []string{
		`{"results":[],"meta":{"count":0}}`,
		`{"results":null,"meta":null}`,
		`{}`,
		`null`,
	} {
		t.Run(response, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			c.EXPECT().GetStudiesRaw(model.StatusAll, "").Return(json.RawMessage(response), nil)
			var output bytes.Buffer
			cmd := study.NewListCommand("list", c, &output)
			cmd.SetArgs([]string{"--raw"})

			require.NoError(t, cmd.Execute())
			assert.JSONEq(t, response, output.String())
		})
	}
}

func TestListRawRejectsConflictingFlagsBeforeRequest(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "csv", args: []string{"--csv"}, wantErr: "--csv"},
		{name: "table", args: []string{"--table"}, wantErr: "--table"},
		{name: "non-interactive", args: []string{"--non-interactive"}, wantErr: "--non-interactive"},
		{name: "non-interactive shorthand", args: []string{"-n"}, wantErr: "--non-interactive"},
		{name: "disabled table before alias", args: []string{"--table=false", "--non-interactive"}, wantErr: "--non-interactive"},
		{name: "disabled table after alias", args: []string{"--non-interactive", "--table=false"}, wantErr: "--non-interactive"},
		{name: "disabled alias before table", args: []string{"--non-interactive=false", "--table"}, wantErr: "--table"},
		{name: "disabled alias after table", args: []string{"--table", "--non-interactive=false"}, wantErr: "--table"},
		{name: "disabled short table before alias", args: []string{"-t=false", "-n"}, wantErr: "--non-interactive"},
		{name: "disabled short table after alias", args: []string{"-n", "-t=false"}, wantErr: "--non-interactive"},
		{name: "disabled short alias before table", args: []string{"-n=false", "-t"}, wantErr: "--table"},
		{name: "disabled short alias after table", args: []string{"-t", "-n=false"}, wantErr: "--table"},
		{name: "repeated table", args: []string{"--table", "-n", "--table=false"}, wantErr: "--non-interactive"},
		{name: "repeated alias", args: []string{"--non-interactive", "-t", "--non-interactive=false"}, wantErr: "--table"},
		{name: "both aliases enabled", args: []string{"-nt"}, wantErr: "--table, --non-interactive"},
		{name: "fields", args: []string{"--fields", "ID"}, wantErr: "--fields"},
		{name: "empty fields", args: []string{"--fields="}, wantErr: "--fields"},
		{name: "underpaying", args: []string{"--underpaying"}, wantErr: "--underpaying"},
		{name: "several", args: []string{"-c", "-u", "-f", "ID"}, wantErr: "--csv, --underpaying, --fields"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			var output bytes.Buffer
			cmd := study.NewListCommand("list", c, &output)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			cmd.SetArgs(append([]string{"--raw"}, tt.args...))

			err := cmd.Execute()

			require.EqualError(t, err, "error: --raw cannot be combined with "+tt.wantErr)
			assert.Empty(t, output.String())
		})
	}
}

func TestListRawAllowsDisabledBoolFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "other flags", args: []string{"--csv=false", "--underpaying=false"}},
		{name: "both aliases", args: []string{"--table=false", "--non-interactive=false"}},
		{name: "short aliases", args: []string{"-n=false", "-t=false"}},
		{name: "repeated aliases", args: []string{"-nt", "--table=false", "--non-interactive=false"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			c.EXPECT().GetStudiesRaw(model.StatusAll, "").Return(json.RawMessage(`{"results":[]}`), nil)
			var output bytes.Buffer
			cmd := study.NewListCommand("list", c, &output)
			cmd.SetArgs(append([]string{"--raw"}, tt.args...))

			require.NoError(t, cmd.Execute())
			assert.JSONEq(t, `{"results":[]}`, output.String())
		})
	}
}

func TestListTableAliasesKeepLastValue(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "table then disabled alias", args: []string{"--table", "--non-interactive=false"}, want: false},
		{name: "alias then disabled table", args: []string{"--non-interactive", "--table=false"}, want: false},
		{name: "disabled table then alias", args: []string{"--table=false", "--non-interactive"}, want: true},
		{name: "disabled alias then table", args: []string{"--non-interactive=false", "--table"}, want: true},
		{name: "short table then disabled alias", args: []string{"-t", "-n=false"}, want: false},
		{name: "short alias then disabled table", args: []string{"-n", "-t=false"}, want: false},
		{name: "repeated table enables output", args: []string{"-t", "-n=false", "-t"}, want: true},
		{name: "repeated alias disables output", args: []string{"-n", "-t", "-n=false"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			cmd := study.NewListCommand("list", c, io.Discard)
			// Inspect the parsed output setting without starting the interactive UI.
			require.NoError(t, cmd.ParseFlags(tt.args))
			for _, name := range []string{"table", "non-interactive"} {
				enabled, err := cmd.Flags().GetBool(name)
				require.NoError(t, err)
				assert.Equal(t, tt.want, enabled, name)
			}
		})
	}
}

func TestListRawReturnsAPIErrors(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().GetStudiesRaw(model.StatusAll, "").Return(nil, errors.New("request failed"))
	var output bytes.Buffer
	cmd := study.NewListCommand("list", c, &output)
	require.NoError(t, cmd.Flags().Set("raw", "true"))

	err := cmd.RunE(cmd, nil)

	assert.EqualError(t, err, "error: request failed")
	assert.Empty(t, output.String())
}

func TestListRawReturnsWriterErrors(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().GetStudiesRaw(model.StatusAll, "").Return(json.RawMessage(`{"results":[]}`), nil)
	cmd := study.NewListCommand("list", c, rawListErrorWriter{})
	require.NoError(t, cmd.Flags().Set("raw", "true"))

	assert.EqualError(t, cmd.RunE(cmd, nil), "error: output unavailable")
}

func TestListRawReturnsJSONEncodingErrors(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().GetStudiesRaw(model.StatusAll, "").Return(json.RawMessage(`{"results":`), nil)
	var output bytes.Buffer
	cmd := study.NewListCommand("list", c, &output)
	require.NoError(t, cmd.Flags().Set("raw", "true"))

	err := cmd.RunE(cmd, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error:")
	assert.Empty(t, output.String())
}

func TestListJSONRemainsTypedArrayWithoutRaw(t *testing.T) {
	for _, flag := range []string{"--json", "-j"} {
		t.Run(flag, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			response := &client.ListStudiesResponse{
				Results: []model.Study{{ID: "study-1", Name: "A study"}},
			}
			c.EXPECT().GetStudies(model.StatusAll, "").Return(response, nil)
			var output bytes.Buffer
			cmd := study.NewListCommand("list", c, &output)
			cmd.SetArgs([]string{flag})

			require.NoError(t, cmd.Execute())

			var records []map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(output.Bytes(), &records))
			require.Len(t, records, 1)
			assert.JSONEq(t, `"study-1"`, string(records[0]["id"]))
			assert.JSONEq(t, `"A study"`, string(records[0]["name"]))
			assert.JSONEq(t, `"0001-01-01T00:00:00Z"`, string(records[0]["date_created"]))
			assert.JSONEq(t, `0`, string(records[0]["reward"]))
			assert.NotContains(t, output.String(), `"results"`)
			assert.NotContains(t, output.String(), `"meta"`)
		})
	}
}

func TestListJSONStillSupportsUnderpayingFilter(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().GetStudies(model.StatusAll, "").Return(&client.ListStudiesResponse{
		Results: []model.Study{
			{ID: "underpaying", IsUnderpaying: true},
			{ID: "paying-fine", IsUnderpaying: false},
		},
	}, nil)
	var output bytes.Buffer
	cmd := study.NewListCommand("list", c, &output)
	cmd.SetArgs([]string{"--json", "--underpaying"})

	require.NoError(t, cmd.Execute())

	var records []model.Study
	require.NoError(t, json.Unmarshal(output.Bytes(), &records))
	require.Len(t, records, 1)
	assert.Equal(t, "underpaying", records[0].ID)
}

type rawListErrorWriter struct{}

func (rawListErrorWriter) Write([]byte) (int, error) {
	return 0, errors.New("output unavailable")
}

var _ io.Writer = rawListErrorWriter{}
