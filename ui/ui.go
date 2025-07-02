package ui

import (
	"fmt"

	"github.com/benmatselby/prolificli/config"
	"github.com/benmatselby/prolificli/model"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	// DarkGrey is the colour for Grey
	DarkGrey = "#989898"
)

// AppDateTimeFormat The format for date/times in the application.
const AppDateTimeFormat string = "02-01-2006 15:04"

// RenderSectionMarker will render a section marker in the output.
func RenderSectionMarker() string {
	return fmt.Sprintf("\n%s\n\n", lipgloss.NewStyle().Foreground(lipgloss.Color(DarkGrey)).Render("---"))
}

// RenderHeading will render a heading in the output.
func RenderHeading(heading string) string {
	return lipgloss.NewStyle().Bold(true).Render(heading)
}

// RenderMoney will return a symbolised string of money, e.g. £10.00
func RenderMoney(amount float64, currencyCode string) string {
	if currencyCode == "" {
		currencyCode = model.DefaultCurrency
	}

	p := message.NewPrinter(language.English)
	cur := currency.MustParseISO(currencyCode)
	return fmt.Sprintf("%s%.2f", p.Sprint(currency.Symbol(cur)), amount)
}

// RenderRecordCounter will render a common string to explain how many records
// are being shown out of the total collection. This will take care of pluralisation
// for you.
func RenderRecordCounter(count, total int) string {
	word := "record"

	if count > 1 {
		word = "records"
	}

	return fmt.Sprintf("Showing %v %s of %v", count, word, total)
}

// RenderApplicationLink will standardise the way we render application links.
func RenderApplicationLink(entity, slug string) string {
	content := RenderSectionMarker()
	content += fmt.Sprintf("View %s in the application: %s/%s", entity, config.GetApplicationURL(), slug)

	return content
}
