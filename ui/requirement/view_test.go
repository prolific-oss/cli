package requirement_test

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/benmatselby/prolificli/model"
	"github.com/benmatselby/prolificli/ui/requirement"
)

func TestRenderRequirement(t *testing.T) {
	req := model.Requirement{
		ID:          "id",
		Cls:         "this is the cls",
		Category:    "category",
		Subcategory: "sub-category",
		Query:       model.RequirementQuestion{ID: "query-id", Title: "requirement title", Question: "query-quest", Description: "query-description"},
		Attributes:  []model.RequirementAttribute{{Label: "attribute-label"}},
	}

	actual := requirement.RenderRequirement(req)

	expected := `query-quest
ID:id
CLS(_cls):thisisthecls
Category:category
Subcategory:sub-category

---

Query
ID:query-id
Question:query-quest
Title:requirementtitle
Description:query-description

---

Attributes
Name:
Label:attribute-label
Index:0
Value:<nil>


`

	actual = stripansi.Strip(actual)
	actual = strings.ReplaceAll(actual, " ", "")
	expected = strings.ReplaceAll(expected, " ", "")

	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}
