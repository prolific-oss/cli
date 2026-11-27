package participantgroup_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/participantgroup"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
)

const memberGroupID = "group-1"

// memberCase describes how a membership command (add/remove) differs from its sibling.
type memberCase struct {
	name        string
	newCommand  func(string, client.API, *bytes.Buffer) *cobra.Command
	expectCall  func(c *mock_client.MockAPI, ids gomock.Matcher) *gomock.Call
	expectShort string
	summary     func(changed int, total int) string
}

func memberCases() []memberCase {
	return []memberCase{
		{
			name: "add",
			newCommand: func(n string, c client.API, b *bytes.Buffer) *cobra.Command {
				return participantgroup.NewAddCommand(n, c, b)
			},
			expectCall: func(c *mock_client.MockAPI, ids gomock.Matcher) *gomock.Call {
				return c.EXPECT().AddParticipantGroupMembers(gomock.Eq(memberGroupID), ids)
			},
			expectShort: "Add participants to a participant group",
			summary: func(changed, total int) string {
				return fmt.Sprintf("Added %d participant(s) to group group-1 (%d total)\n", changed, total)
			},
		},
		{
			name: "remove",
			newCommand: func(n string, c client.API, b *bytes.Buffer) *cobra.Command {
				return participantgroup.NewRemoveCommand(n, c, b)
			},
			expectCall: func(c *mock_client.MockAPI, ids gomock.Matcher) *gomock.Call {
				return c.EXPECT().RemoveParticipantGroupMembers(gomock.Eq(memberGroupID), ids)
			},
			expectShort: "Remove participants from a participant group",
			summary: func(changed, total int) string {
				return fmt.Sprintf("Removed %d participant(s) from group group-1 (%d remaining)\n", changed, total)
			},
		},
	}
}

func memberResponse(remaining ...string) *client.ViewParticipantGroupResponse {
	memberships := make([]model.ParticipantGroupMembership, len(remaining))
	for i, id := range remaining {
		memberships[i] = model.ParticipantGroupMembership{ParticipantID: id}
	}
	return &client.ViewParticipantGroupResponse{Results: memberships}
}

func writeIDFile(t *testing.T, content string) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), "participants.csv")
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}
	return filePath
}

// runMemberCommand executes the command with the given flags and returns its output and error.
func runMemberCommand(tc memberCase, c client.API, flags map[string][]string) (string, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := tc.newCommand(tc.name, c, &b)
	cmd.SetOut(writer)
	for name, values := range flags {
		for _, v := range values {
			_ = cmd.Flags().Set(name, v)
		}
	}
	cmd.SetArgs([]string{memberGroupID})
	err := cmd.Execute()
	writer.Flush()

	return b.String(), err
}

func TestNewMemberCommands(t *testing.T) {
	for _, tc := range memberCases() {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cmd := tc.newCommand(tc.name, mock_client.NewMockAPI(ctrl), &bytes.Buffer{})

			if cmd.Use != tc.name+" <group-id>" {
				t.Fatalf("expected use: '%s <group-id>'; got %q", tc.name, cmd.Use)
			}
			if cmd.Short != tc.expectShort {
				t.Fatalf("unexpected short description: %q", cmd.Short)
			}
		})
	}
}

func TestMemberCommandValidation(t *testing.T) {
	for _, tc := range memberCases() {
		t.Run(tc.name, func(t *testing.T) {
			filePath := writeIDFile(t, "part-1\n")

			cases := map[string]struct {
				flags    map[string][]string
				expected string
			}{
				"no participants": {
					flags:    nil,
					expected: "error: you must provide at least one participant ID via --participant-id or --file",
				},
				"file and inline ids": {
					flags:    map[string][]string{"file": {filePath}, "participant-id": {"part-1"}},
					expected: "error: cannot use --file together with --participant-id",
				},
			}

			for name, v := range cases {
				t.Run(name, func(t *testing.T) {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					c := mock_client.NewMockAPI(ctrl)
					tc.expectCall(c, gomock.Any()).MaxTimes(0)

					_, err := runMemberCommand(tc, c, v.flags)

					if err == nil {
						t.Fatal("expected error, got nil")
					}
					if err.Error() != v.expected {
						t.Fatalf("expected\n'%s'\ngot\n'%s'\n", v.expected, err.Error())
					}
				})
			}
		})
	}
}

func TestMemberCommandSuccess(t *testing.T) {
	for _, tc := range memberCases() {
		t.Run(tc.name, func(t *testing.T) {
			cases := map[string]struct {
				flags     map[string][]string
				fileBody  string
				ids       []string
				remaining []string
			}{
				"inline ids": {
					flags:     map[string][]string{"participant-id": {"part-1", "part-2"}},
					ids:       []string{"part-1", "part-2"},
					remaining: []string{"remaining-1"},
				},
				"file": {
					fileBody: "part-1\npart-2\npart-3\n",
					ids:      []string{"part-1", "part-2", "part-3"},
				},
				"file skips blank lines": {
					fileBody:  "part-1\n\n  \npart-2\n",
					ids:       []string{"part-1", "part-2"},
					remaining: []string{"remaining-1", "remaining-2"},
				},
			}

			for name, v := range cases {
				t.Run(name, func(t *testing.T) {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					c := mock_client.NewMockAPI(ctrl)
					tc.expectCall(c, gomock.Eq(v.ids)).
						Return(memberResponse(v.remaining...), nil).
						MaxTimes(1)

					flags := v.flags
					if v.fileBody != "" {
						flags = map[string][]string{"file": {writeIDFile(t, v.fileBody)}}
					}

					out, err := runMemberCommand(tc, c, flags)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}

					expected := tc.summary(len(v.ids), len(v.remaining))
					if out != expected {
						t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, out)
					}
				})
			}
		})
	}
}

func TestMemberCommandReturnsErrorIfAPIFails(t *testing.T) {
	for _, tc := range memberCases() {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)
			tc.expectCall(c, gomock.Any()).Return(nil, errors.New("api error")).MaxTimes(1)

			_, err := runMemberCommand(tc, c, map[string][]string{"participant-id": {"part-1"}})

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if expected := "error: api error"; err.Error() != expected {
				t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
			}
		})
	}
}
