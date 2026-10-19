package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen  = lipgloss.Color("2")
	colorYellow = lipgloss.Color("3")
	colorRed    = lipgloss.Color("1")
	colorGray   = lipgloss.Color("8")

	styleGreen  = lipgloss.NewStyle().Foreground(colorGreen)
	styleYellow = lipgloss.NewStyle().Foreground(colorYellow)
	styleRed    = lipgloss.NewStyle().Foreground(colorRed)
	styleGray   = lipgloss.NewStyle().Foreground(colorGray)

	styleHeader   = lipgloss.NewStyle().Bold(true)
	styleHeading  = lipgloss.NewStyle().Bold(true)
	styleSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))

	styleHelp  = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder())
	styleError = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
)
