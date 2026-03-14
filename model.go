package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewMode represents the current input mode
type ViewMode int

const (
	ModeNormal ViewMode = iota
	ModeFilter
	ModeActionMenu
	ModeConfirmKill
)

// ActionMenu represents the action selection menu
type ActionMenu struct {
	pid      uint32
	name     string
	selected int // 0 = SIGTERM, 1 = SIGKILL
}

// ConfirmDialog represents the kill confirmation dialog
type ConfirmDialog struct {
	pid   uint32
	name  string
	force bool
}

// Model is the Bubble Tea model containing all application state
type Model struct {
	// Data
	entries         []PortEntry
	filteredIndices []int

	// Selection & Navigation
	selected     int
	scrollOffset int
	visibleRows  int

	// Sorting
	sortBy        SortBy
	sortAscending bool

	// Expansion
	expandedRow int // -1 when none expanded

	// Filter
	filter      string
	filterMode  bool
	filterInput textinput.Model

	// UI State
	viewMode    ViewMode
	actionMenu  *ActionMenu
	confirmKill *ConfirmDialog
	statusMsg   string

	// Layout
	width  int
	height int

	// Internal flags
	quitting bool
}

// NewModel creates and initializes a new Model
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Filter by name or port..."
	ti.CharLimit = 50

	return Model{
		entries:         []PortEntry{},
		filteredIndices: []int{},
		selected:        0,
		scrollOffset:    0,
		visibleRows:     0,
		sortBy:          SortByPort, // Default sort by port
		sortAscending:   true,
		expandedRow:     -1,
		filter:          "",
		filterMode:      false,
		filterInput:     ti,
		viewMode:        ModeNormal,
		actionMenu:      nil,
		confirmKill:     nil,
		statusMsg:       "",
		width:           80,
		height:          24,
		quitting:        false,
	}
}

// Init initializes the Model and returns initial commands
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadPortsCmd(),
		tickCmd(),
	)
}

// getFilteredEntries returns the filtered port entries
func (m *Model) getFilteredEntries() []PortEntry {
	entries := make([]PortEntry, len(m.filteredIndices))
	for i, idx := range m.filteredIndices {
		entries[i] = m.entries[idx]
	}
	return entries
}

// selectedEntry returns the currently selected port entry
func (m *Model) selectedEntry() *PortEntry {
	if len(m.filteredIndices) == 0 || m.selected < 0 || m.selected >= len(m.filteredIndices) {
		return nil
	}
	idx := m.filteredIndices[m.selected]
	if idx < 0 || idx >= len(m.entries) {
		return nil
	}
	return &m.entries[idx]
}

// applyFilter updates the filtered indices based on the current filter
func (m *Model) applyFilter() {
	query := strings.ToLower(m.filter)

	if query == "" {
		// Show all entries
		m.filteredIndices = make([]int, len(m.entries))
		for i := range m.entries {
			m.filteredIndices[i] = i
		}
	} else {
		// Filter by process name, port, or command
		m.filteredIndices = []int{}
		for i, entry := range m.entries {
			if strings.Contains(strings.ToLower(entry.ProcessName), query) ||
				strings.Contains(strconv.Itoa(int(entry.Port)), query) ||
				strings.Contains(strings.ToLower(entry.Command), query) {
				m.filteredIndices = append(m.filteredIndices, i)
			}
		}
	}

	// Clamp selection
	if len(m.filteredIndices) == 0 {
		m.selected = 0
	} else if m.selected >= len(m.filteredIndices) {
		m.selected = len(m.filteredIndices) - 1
	}

	m.scrollOffset = 0
}

// applySorting sorts the entries based on current sort settings
func (m *Model) applySorting() {
	SortEntries(m.entries, m.sortBy, m.sortAscending)
}

// nextRow moves selection down by one, wrapping at the bottom
func (m *Model) nextRow() {
	if len(m.filteredIndices) == 0 {
		return
	}
	m.selected = (m.selected + 1) % len(m.filteredIndices)
	m.ensureVisible()
}

// prevRow moves selection up by one, wrapping at the top
func (m *Model) prevRow() {
	if len(m.filteredIndices) == 0 {
		return
	}
	m.selected = (m.selected - 1 + len(m.filteredIndices)) % len(m.filteredIndices)
	m.ensureVisible()
}

// ensureVisible scrolls to keep the selected row visible
func (m *Model) ensureVisible() {
	if m.selected < m.scrollOffset {
		m.scrollOffset = m.selected
	} else if m.selected >= m.scrollOffset+m.visibleRows {
		m.scrollOffset = m.selected - m.visibleRows + 1
	}
}

// openActionMenu opens the action menu for the selected process
func (m *Model) openActionMenu() {
	entry := m.selectedEntry()
	if entry == nil {
		return
	}

	m.actionMenu = &ActionMenu{
		pid:      entry.PID,
		name:     entry.ProcessName,
		selected: 0,
	}
	m.viewMode = ModeActionMenu
}

// requestKill opens the kill confirmation dialog
func (m *Model) requestKill(force bool) {
	entry := m.selectedEntry()
	if entry == nil {
		return
	}

	m.confirmKill = &ConfirmDialog{
		pid:   entry.PID,
		name:  entry.ProcessName,
		force: force,
	}
	m.viewMode = ModeConfirmKill
}

// toggleSort changes the sort field or toggles direction
func (m *Model) toggleSort(newSortBy SortBy) {
	if m.sortBy == newSortBy {
		// Toggle direction
		m.sortAscending = !m.sortAscending
	} else {
		// Change sort field, reset to ascending
		m.sortBy = newSortBy
		m.sortAscending = true
	}

	m.applySorting()
	m.applyFilter() // Reapply filter to update indices
	m.statusMsg = fmt.Sprintf("Sorted by %s %s",
		m.sortFieldName(), m.sortDirectionIcon())
}

// sortFieldName returns the human-readable name of the current sort field
func (m *Model) sortFieldName() string {
	switch m.sortBy {
	case SortByPID:
		return "PID"
	case SortByPort:
		return "Port"
	case SortByName:
		return "Name"
	default:
		return "Unknown"
	}
}

// sortDirectionIcon returns the icon for the current sort direction
func (m *Model) sortDirectionIcon() string {
	if m.sortAscending {
		return "↑"
	}
	return "↓"
}

// toggleExpansion toggles the expansion state of the selected row
func (m *Model) toggleExpansion() {
	if m.expandedRow == m.selected {
		// Collapse if already expanded
		m.expandedRow = -1
	} else {
		// Expand selected row
		m.expandedRow = m.selected
	}
}
