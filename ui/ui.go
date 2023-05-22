package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/prolificli/model"
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

func RenderSectionMarker() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(DarkGrey)).Render("---")
}

// RenderHeading will render a heading in the output.
func RenderHeading(heading string) string {
	return lipgloss.NewStyle().Bold(true).Render(heading)
}

// RenderMoney will return a symbolised string of money.
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
