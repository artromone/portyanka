package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// tickMsg is sent periodically for automatic refresh
type tickMsg time.Time

// tickCmd returns a command that sends tickMsg after 250ms
func tickCmd() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// portListMsg contains the list of discovered ports
type portListMsg []PortEntry

// loadPortsCmd returns a command that loads the port list
func loadPortsCmd() tea.Cmd {
	return func() tea.Msg {
		return portListMsg(ListListeningPorts())
	}
}

// killResultMsg contains the result of a kill operation
type killResultMsg struct {
	success bool
	pid     uint32
	name    string
	err     error
}

// killProcessCmd returns a command that kills a process
func killProcessCmd(pid uint32, name string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := KillProcess(pid, force)
		return killResultMsg{
			success: err == nil,
			pid:     pid,
			name:    name,
			err:     err,
		}
	}
}
