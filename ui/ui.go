package ui

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
	"github.com/prolific-oss/cli/config"
	"github.com/prolific-oss/cli/model"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	// DarkGrey is the colour for Grey
	DarkGrey = "#989898"

	// Color constants for minimal formatting
	Green        = "#00D787"
	Red          = "#FF6B6B"
	Yellow       = "#FFD93D"
	Cyan         = "#6FC3DF"
	ProlificBlue = "#0F2BC9" // Official Prolific brand blue
)

// Symbol constants
const (
	// Unicode symbols for TTY output
	SymbolSuccess = "✓"
	SymbolError   = "✗"
	SymbolWarning = "!"
	SymbolInfo    = "•"

	// ASCII fallback symbols for non-TTY
	SymbolSuccessAscii = "[ok]"
	SymbolErrorAscii   = "[error]"
	SymbolWarningAscii = "[warn]"
	SymbolInfoAscii    = "[info]"
)

// AppDateTimeFormat The format for date/times in the application.
const AppDateTimeFormat string = "02-01-2006 15:04"

var (
	isTTYStdout      = isatty.IsTerminal(os.Stdout.Fd())
	isTTYStderr      = isatty.IsTerminal(os.Stderr.Fd())
	noColor          = os.Getenv("NO_COLOR") != ""
	colorProfileOnce sync.Once
)

// initColorProfile sets lipgloss color profile based on terminal capabilities
// This is called lazily using sync.Once to ensure it runs exactly once
func initColorProfile() {
	colorProfileOnce.Do(func() {
		if noColor {
			lipgloss.SetColorProfile(termenv.Ascii)
		} else if isTTYStdout {
			// Auto-detect color profile for the terminal
			lipgloss.SetColorProfile(termenv.EnvColorProfile())
		} else {
			lipgloss.SetColorProfile(termenv.Ascii)
		}
	})
}

// Pre-configured lipgloss styles for consistent, vibrant output
var (
	successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(Green)).Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(Red)).Bold(true)
	warningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(Yellow)).Bold(true)
	infoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(Cyan))
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color(DarkGrey))
	boldStyle      = lipgloss.NewStyle().Bold(true)
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(Cyan)) // Cyan color, no underline
)

// RenderSectionMarker will render a section marker in the output.
func RenderSectionMarker() string {
	if !shouldUseColor(isTTYStdout) {
		return "\n---\n\n"
	}
	return fmt.Sprintf("\n%s\n\n", dimStyle.Render("---"))
}

// RenderHeading will render a heading in the output.
func RenderHeading(heading string) string {
	if !shouldUseColor(isTTYStdout) {
		return heading
	}
	return boldStyle.Render(heading)
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

// shouldUseColor returns true if color output should be used based on TTY detection and NO_COLOR env.
func shouldUseColor(isTTY bool) bool {
	initColorProfile() // Lazy initialization
	return isTTY && !noColor
}

// Success renders a success message with a green checkmark.
// Output format: "✓ message" (or "[ok] message" for non-TTY).
func Success(msg string) string {
	if !shouldUseColor(isTTYStdout) {
		return fmt.Sprintf("%s %s", SymbolSuccessAscii, msg)
	}
	symbol := successStyle.Render(SymbolSuccess)
	return fmt.Sprintf("%s %s", symbol, msg)
}

// Error renders an error message with a red X.
// Output format: "✗ Error: message" (or "[error] Error: message" for non-TTY).
// This should be written to stderr.
func Error(msg string) string {
	if !shouldUseColor(isTTYStderr) {
		return fmt.Sprintf("%s Error: %s", SymbolErrorAscii, msg)
	}
	symbol := errorStyle.Render(SymbolError)
	errorText := errorStyle.Render("Error:")
	return fmt.Sprintf("%s %s %s", symbol, errorText, msg)
}

// ErrorWithHint renders an error message with a hint for next steps.
// The hint is displayed on a new line with 2-space indentation.
func ErrorWithHint(msg, hint string) string {
	errorLine := Error(msg)
	dimmedHint := Dim(hint)
	return fmt.Sprintf("%s\n\n  %s", errorLine, dimmedHint)
}

// Warn renders a warning message with a yellow exclamation mark.
// Output format: "! Warning: message" (or "[warn] Warning: message" for non-TTY).
func Warn(msg string) string {
	if !shouldUseColor(isTTYStdout) {
		return fmt.Sprintf("%s Warning: %s", SymbolWarningAscii, msg)
	}
	symbol := warningStyle.Render(SymbolWarning)
	warningText := warningStyle.Render("Warning:")
	return fmt.Sprintf("%s %s %s", symbol, warningText, msg)
}

// Info renders an informational message with a cyan bullet.
// Output format: "• message" (or "[info] message" for non-TTY).
func Info(msg string) string {
	if !shouldUseColor(isTTYStdout) {
		return fmt.Sprintf("%s %s", SymbolInfoAscii, msg)
	}
	symbol := infoStyle.Render(SymbolInfo)
	return fmt.Sprintf("%s %s", symbol, msg)
}

// Dim renders text in a dimmed/grey color for secondary information.
func Dim(text string) string {
	if !shouldUseColor(isTTYStdout) {
		return text
	}
	return dimStyle.Render(text)
}

// Bold renders text in bold (no color).
func Bold(text string) string {
	if !shouldUseColor(isTTYStdout) {
		return text
	}
	return boldStyle.Render(text)
}

// Highlight renders text in cyan for emphasis (URLs, commands, etc.).
func Highlight(text string) string {
	if !shouldUseColor(isTTYStdout) {
		return text
	}
	return highlightStyle.Render(text)
}

// WriteSuccess writes a success message to the provided writer.
func WriteSuccess(w io.Writer, msg string) {
	fmt.Fprintln(w, Success(msg))
}

// WriteError writes an error message to stderr.
func WriteError(msg string) {
	fmt.Fprintln(os.Stderr, Error(msg))
}

// WriteErrorWithHint writes an error message with a hint to stderr.
func WriteErrorWithHint(msg, hint string) {
	fmt.Fprintln(os.Stderr, ErrorWithHint(msg, hint))
}

// WriteWarn writes a warning message to the provided writer.
func WriteWarn(w io.Writer, msg string) {
	fmt.Fprintln(w, Warn(msg))
}

// WriteInfo writes an informational message to the provided writer.
func WriteInfo(w io.Writer, msg string) {
	fmt.Fprintln(w, Info(msg))
}

// RenderBanner renders the ASCII banner in Prolific blue.
func RenderBanner(banner string) string {
	if !shouldUseColor(isTTYStdout) {
		return banner
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(ProlificBlue)).Render(banner)
}
