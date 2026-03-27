package filters_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/filters"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filters.NewListCommand(c, os.Stdout)

	use := "filters"
	short := "List all filters available for your study"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewListCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListFiltersResponse{
		Results: []model.Filter{
			{
				FilterID:    "age",
				FilterTitle: "Age",
				Type:        "range",
				DataType:    "integer",
				Min:         18,
				Max:         100,
			},
			{
				FilterID:    "handedness",
				FilterTitle: "Handedness",
				Type:        "select",
				DataType:    "string",
			},
		},
	}

	c.
		EXPECT().
		GetFilters().
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := filters.NewListCommand(c, writer)
	_ = cmd.Flags().Set("non-interactive", "true")
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	actual := b.String()

	for _, expected := range []string{
		"Title",
		"FilterID",
		"Type",
		"DataType",
		"Age",
		"age",
		"range",
		"integer (18\u2013100)",
		"Handedness",
		"handedness",
		"select",
		"string",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected output to contain %q, got:\n%s", expected, actual)
		}
	}
}

func TestNewListCommandCallsAPIWithMinOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListFiltersResponse{
		Results: []model.Filter{
			{
				FilterID:    "score",
				FilterTitle: "Score",
				Type:        "range",
				DataType:    "integer",
				Min:         5,
			},
		},
	}

	c.
		EXPECT().
		GetFilters().
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := filters.NewListCommand(c, writer)
	_ = cmd.Flags().Set("non-interactive", "true")
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	actual := b.String()
	if !strings.Contains(actual, "integer (min: 5)") {
		t.Fatalf("expected output to contain 'integer (min: 5)', got:\n%s", actual)
	}
}

func TestNewListCommandCallsAPIWithMaxOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListFiltersResponse{
		Results: []model.Filter{
			{
				FilterID:    "score",
				FilterTitle: "Score",
				Type:        "range",
				DataType:    "integer",
				Max:         100,
			},
		},
	}

	c.
		EXPECT().
		GetFilters().
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := filters.NewListCommand(c, writer)
	_ = cmd.Flags().Set("non-interactive", "true")
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	actual := b.String()
	if !strings.Contains(actual, "integer (max: 100)") {
		t.Fatalf("expected output to contain 'integer (max: 100)', got:\n%s", actual)
	}
}

func TestNewListCommandHandlesAnAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "something went wrong"

	c.
		EXPECT().
		GetFilters().
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := filters.NewListCommand(c, writer)
	_ = cmd.Flags().Set("non-interactive", "true")
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
