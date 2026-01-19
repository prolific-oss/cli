// Package ui provides functions for rendering styled terminal output,
// including headings, section markers, formatted currency, record counters,
// and application links.
package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/cli/config"
	"github.com/prolific-oss/cli/model"
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

func RenderHighlightedText(text string) string {
	return lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#FFA500")).Foreground(lipgloss.Color("#000000")).Render(text)
}

// RenderFeatureAccessMessage renders a styled early-access feature message to stderr.
// This is used when a feature is gated behind a feature flag and returns 404 feature not enabled errors.
// Output goes to stderr so it doesn't interfere with JSON/CSV output piping.
func RenderFeatureAccessMessage(featureName, contactURL string) {
	warningStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFA500"))

	urlStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00BFFF"))

	betaStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color(DarkGrey))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, warningStyle.Render("EARLY ACCESS"))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s is an early-access feature that may be enabled on your workspace upon request.\n", featureName)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "To request access or contribute towards the feature's roadmap, visit our help center at %s and drop us a message in the chat.\n", urlStyle.Render(contactURL))
	fmt.Fprintln(os.Stderr, "Your activation request will be reviewed by our team.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, betaStyle.Render("Note: This feature is under active development and you may encounter bugs."))
	fmt.Fprintln(os.Stderr)
}
