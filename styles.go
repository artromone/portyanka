package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
const (
	colorWhite    = lipgloss.Color("15")
	colorBlue     = lipgloss.Color("12")
	colorYellow   = lipgloss.Color("11")
	colorDarkGrey = lipgloss.Color("8")
	colorBlack    = lipgloss.Color("0")
	colorRed      = lipgloss.Color("9")
	colorDarkRed  = lipgloss.Color("1")
	colorCyan     = lipgloss.Color("14")
	colorDarkCyan = lipgloss.Color("6")
	colorGreen    = lipgloss.Color("10")
)

// Column widths (characters)
const (
	pidWidth     = 8
	processWidth = 14
	protoWidth   = 6
	addrWidth    = 18
	portWidth    = 6
	// Command column width is calculated dynamically based on terminal width
	fixedWidth = pidWidth + processWidth + protoWidth + addrWidth + portWidth + 10 // +10 for separators
)

// Styles for different UI components
var (
	// Header style - blue background, white text
	headerStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(colorBlue).
			Padding(0, 1).
			Bold(true)

	// Column header style - yellow, bold
	columnHeaderStyle = lipgloss.NewStyle().
				Foreground(colorYellow).
				Bold(true)

	// Normal row style
	rowStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	// Selected row style - inverted colors
	selectedRowStyle = lipgloss.NewStyle().
				Foreground(colorBlack).
				Background(colorWhite).
				Bold(true)

	// Footer style - dark grey background
	footerStyle = lipgloss.NewStyle().
			Foreground(colorDarkGrey).
			Background(colorBlack).
			Padding(0, 1)

	// Confirm dialog style - red border and background
	confirmDialogStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorRed).
				Background(colorDarkRed).
				Foreground(colorWhite).
				Padding(0, 2).
				Bold(true)

	// Action menu style - cyan border
	actionMenuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Background(colorDarkCyan).
			Foreground(colorWhite).
			Padding(0, 1)

	// Expansion detail style - dimmed
	expansionStyle = lipgloss.NewStyle().
			Foreground(colorDarkGrey).
			Padding(0, 2)

	// Status message style - green
	statusStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)
)

// truncate truncates a string to the specified maximum length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// padRight pads a string to the specified width with spaces
func padRight(s string, width int) string {
	if len(s) >= width {
		return truncate(s, width)
	}
	return s + lipgloss.NewStyle().Width(width-len(s)).Render("")
}

// centerText centers text within the specified width
func centerText(s string, width int) string {
	if len(s) >= width {
		return truncate(s, width)
	}
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(s)
}
