package taskbuilder_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/taskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

const helpFlag = "--help"

func TestNewTaskBuilderCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := taskbuilder.NewTaskBuilderCommand(c, os.Stdout)

	use := "taskbuilder"
	short := "Task Builder tools and utilities"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}

	if cmd.Hidden {
		t.Fatal("expected taskbuilder command not to be hidden")
	}
}

func TestNewTaskBuilderCommandDoesNotPrintMovedNotice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var stdout, stderr bytes.Buffer
	cmd := taskbuilder.NewTaskBuilderCommand(c, &stdout)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"dataset", helpFlag})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(stderr.String(), "has moved") {
		t.Fatalf("did not expect moved notice; got %q", stderr.String())
	}
}

func TestNewDeprecatedAITaskBuilderCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := taskbuilder.NewDeprecatedAITaskBuilderCommand(c, os.Stdout)

	if cmd.Use != "aitaskbuilder" {
		t.Fatalf("expected use: aitaskbuilder; got %s", cmd.Use)
	}

	if !cmd.Hidden {
		t.Fatal("expected aitaskbuilder command to be hidden")
	}
}

func TestNewDeprecatedAITaskBuilderCommandPrintsMovedNoticeToStderr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var stdout, stderr bytes.Buffer
	cmd := taskbuilder.NewDeprecatedAITaskBuilderCommand(c, &stdout)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	// Running a leaf command (rather than --help) proves the notice fires for
	// subcommands, which cobra's own Deprecated field would not do.
	cmd.SetArgs([]string{"dataset", "create"})

	_ = cmd.Execute()

	want := "`prolific aitaskbuilder` has moved, did you mean `prolific taskbuilder`?"
	if !strings.Contains(stderr.String(), want) {
		t.Fatalf("expected stderr to contain %q; got %q", want, stderr.String())
	}

	if strings.Contains(stdout.String(), "has moved") {
		t.Fatalf("expected notice only on stderr; stdout was %q", stdout.String())
	}
}

func TestNewDeprecatedAITaskBuilderCommandPrintsMovedNoticeOnHelp(t *testing.T) {
	for _, args := range [][]string{{}, {helpFlag}, {"batch", helpFlag}} {
		ctrl := gomock.NewController(t)
		c := mock_client.NewMockAPI(ctrl)

		var stdout, stderr bytes.Buffer
		cmd := taskbuilder.NewDeprecatedAITaskBuilderCommand(c, &stdout)
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs(args)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("args %v: unexpected error: %v", args, err)
		}

		if got := strings.Count(stderr.String(), "has moved"); got != 1 {
			t.Fatalf("args %v: expected the notice once on stderr; got %d in %q", args, got, stderr.String())
		}

		if strings.Contains(stdout.String(), "has moved") {
			t.Fatalf("args %v: expected notice only on stderr; stdout was %q", args, stdout.String())
		}

		if strings.Contains(stdout.String(), "aitaskbuilder") {
			t.Fatalf("args %v: expected help to show the new command name; got %q", args, stdout.String())
		}

		if !strings.Contains(stdout.String(), "Task Builder") {
			t.Fatalf("args %v: expected help on stdout; got %q", args, stdout.String())
		}
		ctrl.Finish()
	}
}

func TestNewDeprecatedAITaskBuilderCommandPrintsMovedNoticeOnFlagError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var out, errOut bytes.Buffer
	cmd := taskbuilder.NewDeprecatedAITaskBuilderCommand(c, &out)
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"batch", "list", "-w"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected a flag error")
	}

	if got := strings.Count(errOut.String(), "has moved"); got != 1 {
		t.Fatalf("expected the notice once on stderr; got %d in %q", got, errOut.String())
	}

	all := out.String() + errOut.String()
	if strings.Contains(strings.ReplaceAll(all, "`prolific aitaskbuilder` has moved", ""), "aitaskbuilder") {
		t.Fatalf("expected usage to show the new command name; got %q", all)
	}
}
