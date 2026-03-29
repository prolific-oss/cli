package filters_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	filters "github.com/prolific-oss/cli/cmd/filters"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/stretchr/testify/assert"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filters.NewListCommand(c, nil)

	assert.Equal(t, "filters", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	flag := cmd.Flags().Lookup("non-interactive")
	assert.NotNil(t, flag)
	assert.Equal(t, "n", flag.Shorthand)
	assert.Equal(t, "false", flag.DefValue)
}

func TestListFiltersNonInteractive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListFiltersResponse{
		Results: []model.Filter{
			{
				FilterID:          "age",
				FilterTitle:       "Age",
				FilterDescription: "Filter by age",
				Question:          "How old are you?",
				Type:              "range",
				DataType:          "integer",
			},
			{
				FilterID:          "handedness",
				FilterTitle:       "Handedness",
				FilterDescription: "Filter by handedness",
				Question:          "Are you left or right handed?",
				Type:              "select",
				DataType:          "string",
				Choices: map[string]string{
					"1": "Right-handed",
					"2": "Left-handed",
					"3": "Ambidextrous",
				},
			},
		},
	}

	c.EXPECT().
		GetFilters().
		Return(&response, nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewListCommand(c, w)
	cmd.SetArgs([]string{"-n"})
	err := cmd.Execute()
	w.Flush()

	assert.NoError(t, err)

	output := b.String()
	assert.Contains(t, output, "Age")
	assert.Contains(t, output, "age")
	assert.Contains(t, output, "Handedness")
	assert.Contains(t, output, "handedness")
	assert.Contains(t, output, "Right-handed")
	assert.Contains(t, output, "Left-handed")
}

func TestListFiltersNonInteractiveEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListFiltersResponse{Results: []model.Filter{}}
	c.EXPECT().GetFilters().Return(&response, nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := filters.NewListCommand(c, w)
	cmd.SetArgs([]string{"-n"})
	err := cmd.Execute()
	w.Flush()

	assert.NoError(t, err)
	assert.Empty(t, b.String())
}

func TestListFiltersNonInteractiveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		GetFilters().
		Return(nil, assert.AnError)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewListCommand(c, w)
	cmd.SetArgs([]string{"-n"})
	err := cmd.Execute()

	assert.Error(t, err)
}
