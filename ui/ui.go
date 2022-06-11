package ui

import (
	"github.com/benmatselby/prolificli/client"
	"github.com/charmbracelet/lipgloss"
)

const (
	// DarkBlue is a colour used in the UI.
	DarkBlue = "#083759"
	// Green is a colour used in the UI.
	Green = "#008033"
)

// RenderStatus will render a nice coloured UI for the status
func RenderStatus(status string) lipgloss.Style {

	var color = ""
	if status == client.StatusActive {
		color = Green
	} else if status == client.StatusAwaitingReview {
		color = DarkBlue
	} else if status == client.StatusCompleted {
		color = Green
	} else if status == client.StatusScheduled {
		color = DarkBlue
	} else if status == client.StatusUnpublished {
		color = DarkBlue
	}

	var style = lipgloss.NewStyle().
		Background(lipgloss.Color(color)).
		PaddingLeft(4).
		Width(22).
		SetString(status)

	return style
}
