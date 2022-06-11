package ui

import (
	"github.com/benmatselby/prolificli/model"
	"github.com/charmbracelet/lipgloss"
)

const (
	// DarkBlue is a colour used in the UI.
	DarkBlue = "#083759"
	// Green is a colour used in the UI.
	Green = "#008033"
)

// RenderTitle will render a nice coloured UI for a title based on status
func RenderTitle(title, status string) lipgloss.Style {

	var color = ""
	if status == model.StatusActive {
		color = Green
	} else if status == model.StatusAwaitingReview {
		color = DarkBlue
	} else if status == model.StatusCompleted {
		color = Green
	} else if status == model.StatusScheduled {
		color = DarkBlue
	} else if status == model.StatusUnpublished {
		color = DarkBlue
	}

	var style = lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Background(lipgloss.Color(color)).
		MarginBottom(1).
		Padding(1).
		Align(lipgloss.Center).
		SetString(title)

	return style
}
