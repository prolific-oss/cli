package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewBatchInstructionsCommandWithStarRating(t *testing.T) {
	fiveStars := 5

	testCases := []struct {
		name             string
		instructionsJSON string
		payloadMaxStars  *int
	}{
		{
			name: "with explicit max_stars",
			instructionsJSON: `[{
				"type": "star_rating",
				"created_by": "Sean",
				"description": "How would you rate the overall quality of this response?",
				"max_stars": 5
			}]`,
			payloadMaxStars: &fiveStars,
		},
		{
			// max_stars is optional - the API defaults it to 5 when omitted.
			name: "without max_stars",
			instructionsJSON: `[{
				"type": "star_rating",
				"created_by": "Sean",
				"description": "How would you rate the overall quality of this response?"
			}]`,
			payloadMaxStars: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			batchID := "01954894-65b3-779e-aaf6-348698e12360"

			expectedPayload := client.CreateAITaskBuilderInstructionsPayload{
				Instructions: []client.Instruction{
					{
						Type:        "star_rating",
						CreatedBy:   "Sean",
						Description: "How would you rate the overall quality of this response?",
						MaxStars:    tc.payloadMaxStars,
					},
				},
			}

			response := client.CreateAITaskBuilderInstructionsResponse{
				model.Instruction{
					ID:          "inst-star-1",
					Type:        "star_rating",
					BatchID:     batchID,
					CreatedBy:   "Sean",
					Description: "How would you rate the overall quality of this response?",
					MaxStars:    tc.payloadMaxStars,
				},
			}

			c.EXPECT().CreateAITaskBuilderInstructions(batchID, expectedPayload).Return(&response, nil)

			var buf bytes.Buffer
			writer := bufio.NewWriter(&buf)
			cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)
			cmd.SetArgs([]string{"-b", batchID, "-j", tc.instructionsJSON})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("expected no error; got %s", err.Error())
			}

			writer.Flush()

			expectedOutput := "Successfully added 1 instruction(s) to batch " + batchID
			if !strings.Contains(buf.String(), expectedOutput) {
				t.Fatalf("expected output to contain '%s'; got %s", expectedOutput, buf.String())
			}
		})
	}
}

func TestNewBatchInstructionsCommandStarRatingInvalidMaxStars(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e12362"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	instructionsJSON := `[{
		"type": "star_rating",
		"created_by": "Sean",
		"description": "Rate this response",
		"max_stars": 11
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error for max_stars out of range; got none")
	}

	if !strings.Contains(err.Error(), "max_stars must be between 1 and 10") {
		t.Fatalf("expected error about max_stars range; got %s", err.Error())
	}
}
