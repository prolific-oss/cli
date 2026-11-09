package taskbuilder

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewTaskBuilderCommand creates a new `taskbuilder` command
func NewTaskBuilderCommand(client client.API, w io.Writer) *cobra.Command {
	return newCommand("taskbuilder", client, w)
}

// NewDeprecatedAITaskBuilderCommand keeps `aitaskbuilder` working after it was
// renamed to `taskbuilder`: the same subcommands, hidden from help, with a
// notice on stderr so that scripts parsing stdout are unaffected.
func NewDeprecatedAITaskBuilderCommand(client client.API, w io.Writer) *cobra.Command {
	cobra.AddTemplateFunc("movedPath", func(path string) string {
		return strings.ReplaceAll(path, "aitaskbuilder", "taskbuilder")
	})

	cmd := newCommand("aitaskbuilder", client, w)
	cmd.Hidden = true

	notice := func(cmd *cobra.Command) {
		fmt.Fprint(cmd.ErrOrStderr(), "`prolific aitaskbuilder` has moved, did you mean `prolific taskbuilder`?\n\n")
	}

	// Cobra skips the pre-run hook when only help is shown, so the notice is
	// also printed from the help function. The help itself shows the new
	// command path, so the usage lines can be copied as they are.
	help := cmd.HelpFunc()
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		notice(c)

		out := c.OutOrStdout()
		var buf bytes.Buffer
		c.SetOut(&buf)
		help(c, args)
		c.SetOut(out)

		oldPath := cmd.CommandPath()
		newPath := strings.TrimSuffix(oldPath, "aitaskbuilder") + "taskbuilder"
		text := strings.ReplaceAll(buf.String(), oldPath, newPath)
		text = strings.ReplaceAll(text, "help for aitaskbuilder", "help for taskbuilder")
		fmt.Fprint(out, text)
	})
	cmd.PersistentPreRun = func(c *cobra.Command, _ []string) { notice(c) }

	// Flag parsing fails before the pre-run hook, and the usage printed with
	// the error should also show the new command.
	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		notice(c)
		return err
	})
	cmd.SetUsageTemplate(strings.NewReplacer(
		"{{.UseLine}}", "{{movedPath .UseLine}}",
		"{{.CommandPath}}", "{{movedPath .CommandPath}}",
	).Replace(cmd.UsageTemplate()))

	return cmd
}

func newCommand(use string, client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: "Task Builder tools and utilities",
		Long:  "Manage Task Builder workflows, datasets, and batch operations for the Prolific platform",
	}

	cmd.AddCommand(
		NewDatasetsCommand(client, w),
		NewBatchesCommand(client, w),
	)

	return cmd
}
