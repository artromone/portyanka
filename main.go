package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var Version = "dev"

func main() {
	// Create and initialize the model
	m := NewModel()

	// Create the Bubble Tea program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
