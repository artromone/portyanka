package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// Clear status message on any key press
		m.statusMsg = ""
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcLayout()
		return m, nil

	case tickMsg:
		// Periodic refresh
		return m, tea.Batch(tickCmd(), loadPortsCmd())

	case portListMsg:
		m.entries = []PortEntry(msg)
		m.applySorting()
		m.applyFilter()
		return m, nil

	case killResultMsg:
		if msg.success {
			m.statusMsg = fmt.Sprintf("Killed %s (PID %d)", msg.name, msg.pid)
		} else {
			m.statusMsg = fmt.Sprintf("Failed to kill %s (PID %d): %v", msg.name, msg.pid, msg.err)
		}
		m.viewMode = ModeNormal
		m.confirmKill = nil
		return m, loadPortsCmd()
	}

	return m, nil
}

// handleKeyPress routes key presses based on current view mode
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Priority order: confirm dialog > action menu > filter mode > normal mode

	if m.viewMode == ModeConfirmKill {
		return m.handleConfirmKeys(msg)
	}

	if m.viewMode == ModeActionMenu {
		return m.handleActionMenuKeys(msg)
	}

	if m.viewMode == ModeFilter {
		return m.handleFilterKeys(msg)
	}

	// Normal mode
	return m.handleNormalKeys(msg)
}

// handleNormalKeys handles key presses in normal mode
func (m Model) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "j", "down":
		m.nextRow()
		return m, nil

	case "k", "up":
		m.prevRow()
		return m, nil

	case "e":
		// Toggle expansion
		m.toggleExpansion()
		return m, nil

	case "enter":
		// Open action menu
		m.openActionMenu()
		return m, nil

	case "/":
		// Enter filter mode
		m.filterMode = true
		m.viewMode = ModeFilter
		m.filterInput.Focus()
		m.filterInput.SetValue(m.filter)
		return m, nil

	case "p":
		// Sort by PID
		m.toggleSort(SortByPID)
		return m, nil

	case "n":
		// Sort by Name
		m.toggleSort(SortByName)
		return m, nil

	case "o":
		// Sort by pOrt
		m.toggleSort(SortByPort)
		return m, nil

	case "K":
		// Kill with SIGTERM
		m.requestKill(false)
		return m, nil

	case "F":
		// Force kill with SIGKILL
		m.requestKill(true)
		return m, nil

	case "r":
		// Refresh
		m.statusMsg = "Refreshing..."
		return m, loadPortsCmd()
	}

	return m, nil
}

// handleFilterKeys handles key presses in filter mode
func (m Model) handleFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		// Apply filter
		m.filter = m.filterInput.Value()
		m.applyFilter()
		m.filterMode = false
		m.viewMode = ModeNormal
		m.filterInput.Blur()
		return m, nil

	case "esc":
		// Cancel filter
		m.filterInput.SetValue(m.filter)
		m.filterMode = false
		m.viewMode = ModeNormal
		m.filterInput.Blur()
		return m, nil

	default:
		// Pass key to textinput
		m.filterInput, cmd = m.filterInput.Update(msg)
		return m, cmd
	}
}

// handleActionMenuKeys handles key presses in action menu mode
func (m Model) handleActionMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.actionMenu == nil {
		m.viewMode = ModeNormal
		return m, nil
	}

	switch msg.String() {
	case "j", "down":
		// Next action
		m.actionMenu.selected = (m.actionMenu.selected + 1) % 2
		return m, nil

	case "k", "up":
		// Previous action
		m.actionMenu.selected = (m.actionMenu.selected - 1 + 2) % 2
		return m, nil

	case "enter":
		// Confirm action
		force := m.actionMenu.selected == 1
		m.confirmKill = &ConfirmDialog{
			pid:   m.actionMenu.pid,
			name:  m.actionMenu.name,
			force: force,
		}
		m.viewMode = ModeConfirmKill
		m.actionMenu = nil
		return m, nil

	case "esc", "q":
		// Close action menu
		m.viewMode = ModeNormal
		m.actionMenu = nil
		return m, nil
	}

	return m, nil
}

// handleConfirmKeys handles key presses in confirm kill mode
func (m Model) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmKill == nil {
		m.viewMode = ModeNormal
		return m, nil
	}

	switch msg.String() {
	case "y", "Y":
		// Confirm kill
		cmd := killProcessCmd(m.confirmKill.pid, m.confirmKill.name, m.confirmKill.force)
		m.statusMsg = "Killing process..."
		return m, cmd

	case "n", "N", "esc", "q":
		// Cancel kill
		m.viewMode = ModeNormal
		m.confirmKill = nil
		return m, nil
	}

	return m, nil
}

// recalcLayout recalculates the visible rows based on terminal size
func (m *Model) recalcLayout() {
	// Header (1) + Column headers (1) + Footer (1) = 3 lines overhead
	// Reserve 1 line for status messages
	// Available rows for table
	availableRows := m.height - 4

	if availableRows < 1 {
		availableRows = 1
	}

	// Limit to max 20 rows for inline display
	if availableRows > 20 {
		availableRows = 20
	}

	// Account for expanded row (adds 2 extra lines)
	if m.expandedRow >= 0 && m.expandedRow < len(m.filteredIndices) {
		availableRows -= 2
		if availableRows < 1 {
			availableRows = 1
		}
	}

	m.visibleRows = availableRows
	m.ensureVisible()
}
