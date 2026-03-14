package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the entire UI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Column headers
	sections = append(sections, m.renderColumnHeaders())

	// Table rows
	sections = append(sections, m.renderTable()...)

	// Footer
	sections = append(sections, m.renderFooter())

	base := strings.Join(sections, "\n")

	// Overlay popups
	if m.viewMode == ModeConfirmKill && m.confirmKill != nil {
		return m.overlayConfirmDialog(base)
	} else if m.viewMode == ModeActionMenu && m.actionMenu != nil {
		return m.overlayActionMenu(base)
	}

	return base
}

// renderHeader renders the top header bar
func (m Model) renderHeader() string {
	var title string

	if m.statusMsg != "" {
		title = fmt.Sprintf(" portyanka — %s", statusStyle.Render(m.statusMsg))
	} else if m.viewMode == ModeFilter && m.filterMode {
		title = fmt.Sprintf(" portyanka — filter: %s", m.filterInput.View())
	} else if m.filter != "" {
		title = fmt.Sprintf(" portyanka — filter: [%s]", m.filter)
	} else {
		title = fmt.Sprintf(" portyanka — %d ports", len(m.filteredIndices))
	}

	return headerStyle.Width(m.width).Render(title)
}

// renderColumnHeaders renders the column headers with sort indicators
func (m Model) renderColumnHeaders() string {
	cmdWidth := m.width - fixedWidth
	if cmdWidth < 8 {
		cmdWidth = 8
	}

	// Add sort indicator to the active sort column
	pidHeader := "PID"
	processHeader := "Process"
	portHeader := "Port"

	switch m.sortBy {
	case SortByPID:
		pidHeader = fmt.Sprintf("PID %s", m.sortDirectionIcon())
	case SortByName:
		processHeader = fmt.Sprintf("Process %s", m.sortDirectionIcon())
	case SortByPort:
		portHeader = fmt.Sprintf("Port %s", m.sortDirectionIcon())
	}

	header := fmt.Sprintf(" %s | %s | %s | %s | %s | %s",
		padRight(pidHeader, pidWidth),
		padRight(processHeader, processWidth),
		padRight("Proto", protoWidth),
		padRight("Address", addrWidth),
		padRight(portHeader, portWidth),
		padRight("Command", cmdWidth),
	)

	return columnHeaderStyle.Render(header)
}

// renderTable renders all visible table rows
func (m Model) renderTable() []string {
	var rows []string

	if len(m.filteredIndices) == 0 {
		emptyMsg := "No listening TCP ports found"
		if m.filter != "" {
			emptyMsg = fmt.Sprintf("No ports match filter: %s", m.filter)
		}
		rows = append(rows, lipgloss.NewStyle().Foreground(colorDarkGrey).Render(emptyMsg))
		return rows
	}

	// Calculate visible range
	start := m.scrollOffset
	end := start + m.visibleRows
	if end > len(m.filteredIndices) {
		end = len(m.filteredIndices)
	}

	// Render visible rows
	for i := start; i < end; i++ {
		idx := m.filteredIndices[i]
		if idx < 0 || idx >= len(m.entries) {
			continue
		}

		entry := m.entries[idx]
		isSelected := i == m.selected

		// Render main row
		rows = append(rows, m.renderRow(entry, isSelected))

		// Render expansion details if this row is expanded
		if m.expandedRow == i {
			rows = append(rows, m.renderExpansion(entry)...)
		}
	}

	return rows
}

// renderRow renders a single table row
func (m Model) renderRow(entry PortEntry, isSelected bool) string {
	cmdWidth := m.width - fixedWidth
	if cmdWidth < 8 {
		cmdWidth = 8
	}

	row := fmt.Sprintf(" %s | %s | %s | %s | %s | %s",
		padRight(fmt.Sprintf("%d", entry.PID), pidWidth),
		padRight(entry.ProcessName, processWidth),
		padRight("TCP", protoWidth),
		padRight(entry.Address, addrWidth),
		padRight(fmt.Sprintf("%d", entry.Port), portWidth),
		padRight(entry.Command, cmdWidth),
	)

	if isSelected {
		return selectedRowStyle.Width(m.width).Render(row)
	}
	return rowStyle.Render(row)
}

// renderExpansion renders the expansion details for a port entry
func (m Model) renderExpansion(entry PortEntry) []string {
	var lines []string

	// Full command line (wrapped if necessary)
	cmdLine := fmt.Sprintf("   └─ Full command: %s", entry.Command)
	lines = append(lines, expansionStyle.Render(cmdLine))

	// Connection count
	connMsg := fmt.Sprintf("      Established connections: %d", entry.ConnectionCount)
	lines = append(lines, expansionStyle.Render(connMsg))

	return lines
}

// renderFooter renders the bottom help bar
func (m Model) renderFooter() string {
	var helpText string

	if m.viewMode == ModeFilter && m.filterMode {
		helpText = " Type to filter · Enter to apply · Esc to cancel"
	} else {
		helpText = " q quit · j/k nav · e expand · Enter menu · / filter · p/n/o sort · K kill · F force · r refresh"
	}

	return footerStyle.Width(m.width).Render(helpText)
}

// overlayConfirmDialog overlays the kill confirmation dialog on the base view
func (m Model) overlayConfirmDialog(base string) string {
	sig := "SIGTERM"
	if m.confirmKill.force {
		sig = "SIGKILL"
	}

	msg := fmt.Sprintf("Kill %s (PID %d) with %s? [y/n]",
		m.confirmKill.name, m.confirmKill.pid, sig)

	dialog := confirmDialogStyle.Render(msg)

	// Center the dialog
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		dialog, lipgloss.WithWhitespaceForeground(colorBlack))
}

// overlayActionMenu overlays the action menu on the base view
func (m Model) overlayActionMenu(base string) string {
	actions := []string{
		"Kill (SIGTERM)",
		"Force Kill (SIGKILL)",
	}

	var menuItems []string
	menuItems = append(menuItems, fmt.Sprintf(" Actions for %s (PID %d):",
		m.actionMenu.name, m.actionMenu.pid))
	menuItems = append(menuItems, "")

	for i, action := range actions {
		if i == m.actionMenu.selected {
			menuItems = append(menuItems, fmt.Sprintf(" ▸ %s", action))
		} else {
			menuItems = append(menuItems, fmt.Sprintf("   %s", action))
		}
	}

	menuItems = append(menuItems, "")
	menuItems = append(menuItems, " j/k navigate · Enter confirm · Esc cancel")

	menu := actionMenuStyle.Render(strings.Join(menuItems, "\n"))

	// Center the menu
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		menu, lipgloss.WithWhitespaceForeground(colorBlack))
}
