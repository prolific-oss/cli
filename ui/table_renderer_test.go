package ui_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
)

func TestTableRendererDoesNotPadTheFinalColumn(t *testing.T) {
	studies := []model.Study{
		{ID: "1234", Name: "A short name"},
		{ID: "5678", Name: "A considerably longer study name"},
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	if err := (ui.TableRenderer[model.Study]{}).Render(studies, "ID,Name", writer); err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	_ = writer.Flush()

	expected := "ID   Name\n1234 A short name\n5678 A considerably longer study name\n"
	if actual := b.String(); actual != expected {
		t.Fatalf("expected\n%q\ngot\n%q", expected, actual)
	}
}

func TestTableRendererLeavesUnknownFieldsBlank(t *testing.T) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	if err := (ui.TableRenderer[model.Study]{}).Render([]model.Study{{ID: "1234"}}, "ID,NoSuchField,Name", writer); err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	_ = writer.Flush()

	// Trailing empty cells are still written, so the columns stay aligned with
	// the heading row; only the padding after a populated final column is gone.
	expected := "ID   NoSuchField Name\n1234             \n"
	if actual := b.String(); actual != expected {
		t.Fatalf("expected\n%q\ngot\n%q", expected, actual)
	}
}
